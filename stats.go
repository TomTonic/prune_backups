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
	Stats_SupportedOS = runtime.GOOS == "linux" || runtime.GOOS == "windows"
)

func du(dir_name_or_file_name string) (Infoblock, error) {
	result := infoblock_internal{Infoblock{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, sync.Mutex{}}

	_, err := os.Open(dir_name_or_file_name)
	if err != nil {
		return result.ib, err
	}

	if ok, err := isDirectory(dir_name_or_file_name); ok {
		if err != nil {
			errorMessage := fmt.Sprintf("Error identifying %v: %v\n", dir_name_or_file_name, err)
			return result.ib, errors.New(errorMessage)
		}
		duInternalDirectory(dir_name_or_file_name, &result)
	} else {
		duInternalFile(dir_name_or_file_name, &result)
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
	f_size, f_links, err := getSizeAndLinkCount(fileName)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			(*info).ib.number_of_permission_errors_files += 1
		} else {
			(*info).ib.number_of_other_errors_files += 1
		}
		return
	}
	if f_links == 1 {
		(*info).ib.number_of_unlinked_files += 1
		(*info).ib.size_of_unlinked_files += f_size
	} else {
		(*info).ib.number_of_linked_files += 1
		(*info).ib.size_of_linked_files += f_size
	}
}

func duInternalDirectory(directoryName string, globalinfo *infoblock_internal) {
	localinfo := infoblock_internal{Infoblock{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, sync.Mutex{}}
	defer addAll(globalinfo, &localinfo) // this is synchronized

	files, err := readDirWithRetry(directoryName, 1000000, 2)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			localinfo.ib.number_of_permission_errors_dirs += 1
		} else {
			localinfo.ib.number_of_other_errors_dirs += 1
		}
		return
	}

	subdirs := []string{}
	for _, file := range files {
		fullPath := filepath.Join(directoryName, file.Name())
		if file.Type().IsRegular() {
			duInternalFile(fullPath, &localinfo)
		} else if file.Type().IsDir() {
			subdirs = append(subdirs, fullPath)
		} else {
			countAccordingType(file.Type(), &localinfo)
		}
	}

	localinfo.ib.number_of_subdirs += len(subdirs)

	// now descend into the directories
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.NumCPU()*500) // Limit the number of concurrent goroutines

	for _, subdir := range subdirs {
		semaphore <- struct{}{} // Acquire a token before starting a goroutine
		wg.Add(1)
		go func(subdir string) {
			defer func() { <-semaphore }() // Release the token when done
			defer wg.Done()
			duInternalDirectory(subdir, globalinfo)
		}(subdir)
	}

	wg.Wait() // Wait for all child directories to complete

}

func addAll(globalinfo, localinfo *infoblock_internal) {
	(*globalinfo).mutex.Lock()
	(*globalinfo).ib.size_of_unlinked_files += (*localinfo).ib.size_of_unlinked_files
	(*globalinfo).ib.size_of_linked_files += (*localinfo).ib.size_of_linked_files
	(*globalinfo).ib.number_of_unlinked_files += (*localinfo).ib.number_of_unlinked_files
	(*globalinfo).ib.number_of_linked_files += (*localinfo).ib.number_of_linked_files
	(*globalinfo).ib.number_of_subdirs += (*localinfo).ib.number_of_subdirs
	(*globalinfo).ib.number_of_permission_errors_files += (*localinfo).ib.number_of_permission_errors_files
	(*globalinfo).ib.number_of_permission_errors_dirs += (*localinfo).ib.number_of_permission_errors_dirs
	(*globalinfo).ib.number_of_other_errors_files += (*localinfo).ib.number_of_other_errors_files
	(*globalinfo).ib.number_of_other_errors_dirs += (*localinfo).ib.number_of_other_errors_dirs
	(*globalinfo).ib.nr_apnd += (*localinfo).ib.nr_apnd
	(*globalinfo).ib.nr_excl += (*localinfo).ib.nr_excl
	(*globalinfo).ib.nr_tmp += (*localinfo).ib.nr_tmp
	(*globalinfo).ib.nr_sym += (*localinfo).ib.nr_sym
	(*globalinfo).ib.nr_dev += (*localinfo).ib.nr_dev
	(*globalinfo).ib.nr_pipe += (*localinfo).ib.nr_pipe
	(*globalinfo).ib.nr_sock += (*localinfo).ib.nr_sock
	(*globalinfo).mutex.Unlock()
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

func readDirWithRetry(directoryname string, retries, maxwait_seconds int) ([]fs.DirEntry, error) {
	for range retries {
		direntries, err := os.ReadDir(directoryname)
		if err != nil {
			pathErr, ok := err.(*os.PathError)
			if !ok {
				return nil, fmt.Errorf("error reading directory (error type %s): %v", reflect.TypeOf(err), err)
			}
			if pathErr.Err.Error() == "too many open files" {
				// Wait a random time before retrying
				rnd := rand.IntN(maxwait_seconds * 1000)
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
	return nil, fmt.Errorf("failed to read directory after %v retries: %s", retries, directoryname)
}

func openFileWithRetry(filename string, retries, maxwait_seconds int) (*os.File, error) {
	for range retries {
		file, err := os.Open(filename)
		if err != nil {
			pathErr, ok := err.(*os.PathError)
			if !ok {
				return nil, fmt.Errorf("error reading file (error type %s): %v", reflect.TypeOf(err), err)
			}
			if pathErr.Err.Error() == "too many open files" {
				// Wait a random time before retrying
				rnd := rand.IntN(maxwait_seconds * 1000)
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
