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
	size_of_unlinked_files   uint64
	size_of_linked_files     uint64
	number_of_unlinked_files int
	number_of_linked_files   int
	number_of_subdirs        int
	nr_apnd                  int
	nr_excl                  int
	nr_tmp                   int
	nr_sym                   int
	nr_dev                   int
	nr_pipe                  int
	nr_sock                  int
}

type infoblock_internal struct {
	ib    Infoblock
	mutex sync.Mutex
}

var (
	Stats_SupportedOS = runtime.GOOS == "linux" || runtime.GOOS == "windows"
)

func du(dir_name_or_file_name string) (Infoblock, error) {
	result := infoblock_internal{Infoblock{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, sync.Mutex{}}
	if ok, err := isDirectory(dir_name_or_file_name); ok {
		if err != nil {
			errorMessage := fmt.Sprintf("Error identifying %v: %v\n", dir_name_or_file_name, err)
			return result.ib, errors.New(errorMessage)
		}
		err = duInternalDirectory(dir_name_or_file_name, &result)
		if err != nil {
			errorMessage := fmt.Sprintf("Error reading directory %v: %v\n", dir_name_or_file_name, err)
			return result.ib, errors.New(errorMessage)
		}
	} else {
		err = duInternalFile(dir_name_or_file_name, &result)
		if err != nil {
			errorMessage := fmt.Sprintf("Error reading file %v: %v\n", dir_name_or_file_name, err)
			return result.ib, errors.New(errorMessage)
		}
	}
	//fmt.Printf("total size of unlinked files: %v bytes; total size of linked files: %v bytes", size_of_unlinked_files, size_of_linked_files)
	return result.ib, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func duInternalFile(fileName string, info *infoblock_internal) (err error) {
	f_size, f_links, err := getSizeAndLinkCount(fileName)
	if err != nil {
		return err
	}
	if f_links == 1 {
		(*info).ib.number_of_unlinked_files += 1
		(*info).ib.size_of_unlinked_files += f_size
	} else {
		(*info).ib.number_of_linked_files += 1
		(*info).ib.size_of_linked_files += f_size
	}
	return nil
}

func duInternalDirectory(directoryName string, globalinfo *infoblock_internal) (err error) {
	localinfo := infoblock_internal{Infoblock{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, sync.Mutex{}}
	localinfo.ib.number_of_subdirs += 1 // this directory
	files, err := readDirWithRetry(directoryName, 1000000, 2)
	if err != nil {
		return err
	}
	subdirs := []string{}
	for _, file := range files {
		fullPath := filepath.Join(directoryName, file.Name())
		if file.Type().IsRegular() {
			err := duInternalFile(fullPath, &localinfo)
			if err != nil {
				fmt.Printf("Error getting info for regular file %v: %v\n", fullPath, err)
			}
		} else if file.Type().IsDir() {
			subdirs = append(subdirs, fullPath)
		} else {
			countAccordingType(file.Type(), &localinfo)
			// fmt.Printf("Skipping file of type %v: %v\n", typeToString(file.Type()), fullPath)
		}
	}
	// now descend into the directories
	var wg sync.WaitGroup
	for _, subdir := range subdirs {
		// descend children in parallel
		wg.Add(1) // Increment the WaitGroup counter.
		go func() {
			defer wg.Done() // Decrement the counter when the goroutine completes.
			err := duInternalDirectory(subdir, globalinfo)
			if err != nil {
				fmt.Printf("Error getting info for directory %v: %v\n", subdir, err)
			}
		}()
	}
	wg.Wait()                      // Wait for all child directories to complete
	addAll(globalinfo, &localinfo) // this is synchronized
	return nil
}

func addAll(globalinfo, localinfo *infoblock_internal) {
	(*globalinfo).mutex.Lock()
	(*globalinfo).ib.size_of_unlinked_files += (*localinfo).ib.size_of_unlinked_files
	(*globalinfo).ib.size_of_linked_files += (*localinfo).ib.size_of_linked_files
	(*globalinfo).ib.number_of_unlinked_files += (*localinfo).ib.number_of_unlinked_files
	(*globalinfo).ib.number_of_linked_files += (*localinfo).ib.number_of_linked_files
	(*globalinfo).ib.number_of_subdirs += (*localinfo).ib.number_of_subdirs
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

/*
func typeToString(mode fs.FileMode) string {
	switch mode {
	case fs.ModeDir:
		return "directory"
	case fs.ModeAppend:
		return "append-only file"
	case fs.ModeExclusive:
		return "exclusive file"
	case fs.ModeTemporary:
		return "temporary file"
	case fs.ModeSymlink:
		return "symlink"
	case fs.ModeDevice:
		return "device"
	case fs.ModeNamedPipe:
		return "named pipe"
	case fs.ModeSocket:
		return "socket"
	default:
		result := checkBit(mode, fs.ModeDir, "DIR|", "dir|")
		result += checkBit(mode, fs.ModeAppend, "APND|", "apnd|")
		result += checkBit(mode, fs.ModeExclusive, "EXCL|", "excl|")
		result += checkBit(mode, fs.ModeTemporary, "TMP|", "tmp|")
		result += checkBit(mode, fs.ModeSymlink, "SYM|", "sym|")
		result += checkBit(mode, fs.ModeDevice, "DEV|", "dev|")
		result += checkBit(mode, fs.ModeNamedPipe, "PIPE|", "pipe|")
		result += checkBit(mode, fs.ModeSocket, "SOCK", "sock")
		return result
	}
}

func checkBit(mode fs.FileMode, cmp fs.FileMode, if_set string, if_not_set string) string {
	if mode&cmp > 1 {
		return if_set
	}
	return if_not_set
}
*/

func readDirWithRetry(directoryname string, retries, maxwait_seconds int) ([]fs.DirEntry, error) {
	for i := 0; i < retries; i++ {
		direntries, err := os.ReadDir(directoryname)
		if err != nil {
			pathErr, ok := err.(*os.PathError)
			if ok && pathErr.Err.Error() == "too many open files" {
				// Wait a random time before retrying
				rnd := rand.IntN(maxwait_seconds * 1000)
				time.Sleep(time.Duration(200+rnd) * time.Millisecond) // wait at leas 200ms
				continue
			} else {
				return nil, fmt.Errorf("error reading directory (error type %s): %v", err, reflect.TypeOf(err))
				// return nil, err
			}
		}
		return direntries, nil
	}
	return nil, fmt.Errorf("giving up - failed to read directory after %v retries: %s", retries, directoryname)
}

func openFileWithRetry(filename string, retries, maxwait_seconds int) (*os.File, error) {
	for i := 0; i < retries; i++ {
		file, err := os.Open(filename)
		if err != nil {
			pathErr, ok := err.(*os.PathError)
			if ok && pathErr.Err.Error() == "too many open files" {
				// Wait a random time before retrying
				rnd := rand.IntN(maxwait_seconds * 1000)
				time.Sleep(time.Duration(200+rnd) * time.Millisecond) // wait at leas 200ms
				continue
			} else {
				return nil, fmt.Errorf("error reading file (error type %s): %v", err, reflect.TypeOf(err))
				// return nil, err
			}
		}
		return file, nil
	}
	return nil, fmt.Errorf("giving up - failed to open file after %v retries: %s", retries, filename)
}
