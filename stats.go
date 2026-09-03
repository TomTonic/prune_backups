package main

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type Infoblock struct {
	size_of_unlinked_files            uint64
	size_of_linked_files              uint64
	number_of_unlinked_files          int
	number_of_linked_files            int
	number_of_subdirs                 int
	number_of_permission_errors_files int
	number_of_permission_errors_dirs  int
	number_of_other_errors_files      int
	number_of_other_errors_dirs       int
	nr_apnd                           int
	nr_excl                           int
	nr_tmp                            int
	nr_sym                            int
	nr_dev                            int
	nr_pipe                           int
	nr_sock                           int
}

type infoblock_internal struct {
	ib    Infoblock
	mutex sync.Mutex
}

var (
	Stats_SupportedOS = runtime.GOOS == "linux" || runtime.GOOS == "windows" || runtime.GOOS == "darwin"
)

func DiskUsage(path string) (Infoblock, error) {
	result := infoblock_internal{}
	const limit = 4000
	// limit bounds how many directories may be spawned-but-not-yet-locally-scanned
	// at once (see duInternalDirectory / scanDirectory below), not how many
	// goroutines exist in total.
	semaphore := NewSemaphore(limit)
	debug.SetMaxThreads(2 * limit) // Ensure the thread limit is high enough

	nevermind, err := os.Open(path)
	defer func() {
		if nevermind != nil {
			_ = nevermind.Close()
		}
	}()
	if err != nil {
		return result.ib, err
	}

	if ok, err := isDirectory(path); ok {
		// duInternalDirectory expects the caller to have already acquired one
		// token on its behalf (the recursion loop below does this for every
		// call it spawns); DiskUsage does it here for the initial, un-spawned
		// root call.
		semaphore.Acquire()
		duInternalDirectory(path, &result, semaphore)
	} else {
		if err != nil {
			errorMessage := fmt.Sprintf("Error identifying %v: %v\n", path, err)
			return result.ib, errors.New(errorMessage)
		}
		duInternalFile(path, &result)
	}
	return result.ib, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func duInternalFile(fileName string, info *infoblock_internal) {
	fSize, fLinks, err := getSizeAndLinkCount(fileName)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			(*info).ib.number_of_permission_errors_files += 1
		} else {
			(*info).ib.number_of_other_errors_files += 1
		}
		return
	}
	if fLinks == 1 {
		(*info).ib.number_of_unlinked_files += 1
		(*info).ib.size_of_unlinked_files += fSize
	} else {
		(*info).ib.number_of_linked_files += 1
		(*info).ib.size_of_linked_files += fSize
	}
}

// duInternalDirectory processes directoryName. The caller must have already
// acquired one semaphore token on its behalf - either DiskUsage, for the
// un-spawned root call, or the recursion loop below, for every call it
// spawns. duInternalDirectory releases that token itself, right after its own
// local scanDirectory work completes, before it recurses into subdirectories.
// That keeps the original bound on how many directories can be spawned-and-
// scanning (or spawned-and-waiting-to-scan) at once, while fixing a deadlock a
// prior version of this function had: it released the token only after its
// entire subtree had finished, so once enough directories were in flight at
// once, every token ended up held by a goroutine blocked in wg.Wait() for a
// child that could never acquire a token of its own.
func duInternalDirectory(directoryName string, globalInfo *infoblock_internal, semaphore Semaphore) {
	localInfo := infoblock_internal{}
	defer addAll(globalInfo, &localInfo) // this is synchronized

	subdirs, err := scanDirectory(directoryName, &localInfo)
	semaphore.Release()
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			localInfo.ib.number_of_permission_errors_dirs += 1
		} else {
			localInfo.ib.number_of_other_errors_dirs += 1
		}
		return
	}

	localInfo.ib.number_of_subdirs += len(subdirs)

	// now descend into the directories
	var wg sync.WaitGroup

	for _, subdir := range subdirs {
		semaphore.Acquire() // released by the spawned call above, not here
		wg.Go(func() {
			duInternalDirectory(subdir, globalInfo, semaphore)
		})
	}

	wg.Wait() // Wait for all child directories to complete

}

// scanDirectory lists directoryName and processes its regular-file and other
// (non-directory) entries, returning the subdirectory paths to recurse into.
func scanDirectory(directoryName string, localInfo *infoblock_internal) ([]string, error) {
	files, err := readDirWithRetry(directoryName, 1000000, 2)
	if err != nil {
		return nil, err
	}

	subdirs := []string{}
	for _, file := range files {
		fullPath := filepath.Join(directoryName, file.Name())
		if file.Type().IsRegular() {
			duInternalFile(fullPath, localInfo)
		} else if file.Type().IsDir() {
			subdirs = append(subdirs, fullPath)
		} else {
			countAccordingType(file.Type(), localInfo)
		}
	}
	return subdirs, nil
}

func addAll(globalInfo, localInfo *infoblock_internal) {
	globalInfo.mutex.Lock()
	globalInfo.ib.size_of_unlinked_files += localInfo.ib.size_of_unlinked_files
	globalInfo.ib.size_of_linked_files += localInfo.ib.size_of_linked_files
	globalInfo.ib.number_of_unlinked_files += localInfo.ib.number_of_unlinked_files
	globalInfo.ib.number_of_linked_files += localInfo.ib.number_of_linked_files
	globalInfo.ib.number_of_subdirs += localInfo.ib.number_of_subdirs
	globalInfo.ib.number_of_permission_errors_files += localInfo.ib.number_of_permission_errors_files
	globalInfo.ib.number_of_permission_errors_dirs += localInfo.ib.number_of_permission_errors_dirs
	globalInfo.ib.number_of_other_errors_files += localInfo.ib.number_of_other_errors_files
	globalInfo.ib.number_of_other_errors_dirs += localInfo.ib.number_of_other_errors_dirs
	globalInfo.ib.nr_apnd += localInfo.ib.nr_apnd
	globalInfo.ib.nr_excl += localInfo.ib.nr_excl
	globalInfo.ib.nr_tmp += localInfo.ib.nr_tmp
	globalInfo.ib.nr_sym += localInfo.ib.nr_sym
	globalInfo.ib.nr_dev += localInfo.ib.nr_dev
	globalInfo.ib.nr_pipe += localInfo.ib.nr_pipe
	globalInfo.ib.nr_sock += localInfo.ib.nr_sock
	globalInfo.mutex.Unlock()
}

func countAccordingType(mode fs.FileMode, info *infoblock_internal) {
	switch mode {
	case fs.ModeDir:
		(*info).ib.number_of_subdirs++
		return
	case fs.ModeAppend:
		(*info).ib.nr_apnd++
		return
	case fs.ModeExclusive:
		(*info).ib.nr_excl++
		return
	case fs.ModeTemporary:
		(*info).ib.nr_tmp++
		return
	case fs.ModeSymlink:
		(*info).ib.nr_sym++
		return
	case fs.ModeDevice:
		(*info).ib.nr_dev++
		return
	case fs.ModeNamedPipe:
		(*info).ib.nr_pipe++
		return
	case fs.ModeSocket:
		(*info).ib.nr_sock++
		return
	}
}

func readDirWithRetry(directoryName string, retries, maxWaitSeconds int) ([]fs.DirEntry, error) {
	for range retries {
		direntries, err := os.ReadDir(directoryName)
		if err != nil {
			pathErr, ok := err.(*os.PathError)
			if !ok {
				return nil, fmt.Errorf("error reading directory (error type %s): %v", reflect.TypeOf(err), err)
			}
			if pathErr.Err.Error() == "too many open files" {
				// Wait a random time before retrying
				rnd := rand.IntN(maxWaitSeconds * 1000)
				time.Sleep(time.Duration(200+rnd) * time.Millisecond) // wait at least 200ms
				continue
			}
			if pathErr.Err.Error() == "permission denied" {
				return nil, pathErr.Err
			}
			return nil, err
		}
		return direntries, nil
	}
	return nil, fmt.Errorf("failed to read directory after %v retries: %s", retries, directoryName)
}

func openFileWithRetry(filename string, retries, maxWaitSeconds int) (*os.File, error) {
	for range retries {
		file, err := os.Open(filename)
		if err != nil {
			pathErr, ok := err.(*os.PathError)
			if !ok {
				return nil, fmt.Errorf("error reading file (error type %s): %v", reflect.TypeOf(err), err)
			}
			if pathErr.Err.Error() == "too many open files" {
				// Wait a random time before retrying
				rnd := rand.IntN(maxWaitSeconds * 1000)
				time.Sleep(time.Duration(200+rnd) * time.Millisecond) // wait at least 200ms
				continue
			}
			if pathErr.Err.Error() == "permission denied" {
				return nil, pathErr.Err
			}
			return nil, err
		}
		return file, nil
	}
	return nil, fmt.Errorf("failed to open file after %v retries: %s", retries, filename)
}

// Semaphore is a counting semaphore backed by a buffered channel. Callers
// park on Acquire when the limit is reached and are woken immediately when a
// token is released — no busy-waiting, no polling delay. A token acquired by
// one goroutine may safely be released by another; there is no
// ownership/affinity requirement, since it's backed by a plain channel. What
// exactly a token bounds (concurrent goroutines, concurrent I/O, or something
// narrower) is up to where the caller places Acquire/Release - see
// duInternalDirectory for a case where the two are placed asymmetrically on
// purpose.
type Semaphore chan struct{}

func NewSemaphore(limit int) Semaphore {
	return make(Semaphore, limit)
}

func (s Semaphore) Acquire() { s <- struct{}{} }
func (s Semaphore) Release() { <-s }
