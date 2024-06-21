package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func du(dirNameOrFileName string) (size_of_unlinked_files, size_of_linked_files uint64) {
	if ok, err := isDirectory(dirNameOrFileName); ok {
		if err != nil {
			fmt.Printf("Error identifying %v: %v", dirNameOrFileName, err)
			return
		}
		size_of_unlinked_files, size_of_linked_files, err = duInternalDirectory(dirNameOrFileName)
		if err != nil {
			fmt.Printf("Error reading directory %v: %v", dirNameOrFileName, err)
			return
		}
	} else {
		size_of_unlinked_files, size_of_linked_files, err = duInternalFile(dirNameOrFileName)
		if err != nil {
			fmt.Printf("Error reading file %v: %v", dirNameOrFileName, err)
			return
		}
	}
	fmt.Printf("total size of unlinked files: %v bytes; total size of linked files: %v bytes", size_of_unlinked_files, size_of_linked_files)
	return size_of_unlinked_files, size_of_linked_files
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func duInternalFile(fileName string) (size_of_unlinked_files, size_of_linked_files uint64, err error) {
	size_of_unlinked_files = 0
	size_of_linked_files = 0
	f_size, f_links, err := getSizeAndLinkCount(fileName)
	if err != nil {
		return 0, 0, err
	}
	if f_links == 1 {
		size_of_unlinked_files = f_size
	} else {
		size_of_linked_files = f_size
	}
	return size_of_unlinked_files, size_of_linked_files, nil
}

func duInternalDirectory(directoryName string) (size_of_unlinked_files, size_of_linked_files uint64, err error) {
	size_of_unlinked_files = 0
	size_of_linked_files = 0
	files, err := os.ReadDir(directoryName)
	if err != nil {
		return size_of_unlinked_files, size_of_linked_files, err
	}
	subdirs := []string{}
	for _, file := range files {
		fullPath := filepath.Join(directoryName, file.Name())
		if file.Type().IsRegular() {
			f_size_of_unlinked_files, f_size_of_linked_files, err := duInternalFile(fullPath)
			if err != nil {
				fmt.Printf("Error getting info for regular file %v: %v", fullPath, err)
			}
			size_of_unlinked_files += f_size_of_unlinked_files
			size_of_linked_files += f_size_of_linked_files
		} else if file.Type().IsDir() {
			// store directories for later descending to be a little more cache efficient
			subdirs = append(subdirs, fullPath)
		} else {
			fmt.Printf("Skipping file of type %v: %v", file.Type(), fullPath)
		}
	}
	// now descend into the directories
	for _, subdir := range subdirs {
		sd_size_of_unlinked_files, sd_size_of_linked_files, err := duInternalDirectory(subdir)
		if err != nil {
			fmt.Printf("Error getting info for directory %v: %v", subdir, err)
		} else {
			size_of_unlinked_files += sd_size_of_unlinked_files
			size_of_linked_files += sd_size_of_linked_files
		}
	}
	return size_of_unlinked_files, size_of_linked_files, nil
}
