package main

import (
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
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
	mutex                    sync.Mutex
}

func du(dirNameOrFileName string) Infoblock {
	result := Infoblock{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, sync.Mutex{}}
	if ok, err := isDirectory(dirNameOrFileName); ok {
		if err != nil {
			fmt.Printf("Error identifying %v: %v\n", dirNameOrFileName, err)
			return result
		}
		err = duInternalDirectory(dirNameOrFileName, &result)
		if err != nil {
			fmt.Printf("Error reading directory %v: %v\n", dirNameOrFileName, err)
			return result
		}
	} else {
		err = duInternalFile(dirNameOrFileName, &result)
		if err != nil {
			fmt.Printf("Error reading file %v: %v\n", dirNameOrFileName, err)
			return result
		}
	}
	//fmt.Printf("total size of unlinked files: %v bytes; total size of linked files: %v bytes", size_of_unlinked_files, size_of_linked_files)
	return result
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func duInternalFile(fileName string, info *Infoblock) (err error) {
	f_size, f_links, err := getSizeAndLinkCount(fileName)
	if err != nil {
		return err
	}
	if f_links == 1 {
		(*info).number_of_unlinked_files += 1
		(*info).size_of_unlinked_files += f_size
	} else {
		(*info).number_of_linked_files += 1
		(*info).size_of_linked_files += f_size
	}
	return nil
}

func duInternalDirectory(directoryName string, globalinfo *Infoblock) (err error) {
	localinfo := Infoblock{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, sync.Mutex{}}
	localinfo.number_of_subdirs += 1 // this directory
	//files, err := os.ReadDir(directoryName)
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
			// store directories for later descending to be a little more cache efficient
			subdirs = append(subdirs, fullPath)
		} else {
			addAccordingType(file.Type(), &localinfo)
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

func addAll(globalinfo, localinfo *Infoblock) {
	(*globalinfo).mutex.Lock()
	(*globalinfo).size_of_unlinked_files += (*localinfo).size_of_unlinked_files
	(*globalinfo).size_of_linked_files += (*localinfo).size_of_linked_files
	(*globalinfo).number_of_unlinked_files += (*localinfo).number_of_unlinked_files
	(*globalinfo).number_of_linked_files += (*localinfo).number_of_linked_files
	(*globalinfo).number_of_subdirs += (*localinfo).number_of_subdirs
	(*globalinfo).nr_apnd += (*localinfo).nr_apnd
	(*globalinfo).nr_excl += (*localinfo).nr_excl
	(*globalinfo).nr_tmp += (*localinfo).nr_tmp
	(*globalinfo).nr_sym += (*localinfo).nr_sym
	(*globalinfo).nr_dev += (*localinfo).nr_dev
	(*globalinfo).nr_pipe += (*localinfo).nr_pipe
	(*globalinfo).nr_sock += (*localinfo).nr_sock
	(*globalinfo).mutex.Unlock()
}

func addAccordingType(mode fs.FileMode, info *Infoblock) {
	switch mode {
	case fs.ModeDir:
		(*info).number_of_subdirs++
		return
	case fs.ModeAppend:
		(*info).nr_apnd++
		return
	case fs.ModeExclusive:
		(*info).nr_excl++
		return
	case fs.ModeTemporary:
		(*info).nr_tmp++
		return
	case fs.ModeSymlink:
		(*info).nr_sym++
		return
	case fs.ModeDevice:
		(*info).nr_dev++
		return
	case fs.ModeNamedPipe:
		(*info).nr_pipe++
		return
	case fs.ModeSocket:
		(*info).nr_sock++
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
