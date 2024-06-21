package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func du(dirNameOrFileName string) (number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, number_of_subdirs uint64) {
	if ok, err := isDirectory(dirNameOrFileName); ok {
		if err != nil {
			fmt.Printf("Error identifying %v: %v\n", dirNameOrFileName, err)
			return
		}
		number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, number_of_subdirs, err = duInternalDirectory(dirNameOrFileName)
		if err != nil {
			fmt.Printf("Error reading directory %v: %v\n", dirNameOrFileName, err)
			return
		}
	} else {
		number_of_subdirs = 0
		number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, err = duInternalFile(dirNameOrFileName)
		if err != nil {
			fmt.Printf("Error reading file %v: %v\n", dirNameOrFileName, err)
			return
		}
	}
	//fmt.Printf("total size of unlinked files: %v bytes; total size of linked files: %v bytes", size_of_unlinked_files, size_of_linked_files)
	return number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, number_of_subdirs
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func duInternalFile(fileName string) (number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files uint64, err error) {
	number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files = 0, 0, 0, 0
	f_size, f_links, err := getSizeAndLinkCount(fileName)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if f_links == 1 {
		number_of_unlinked_files = 1
		size_of_unlinked_files = f_size
	} else {
		number_of_linked_files = 1
		size_of_linked_files = f_size
	}
	return number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, nil
}

func duInternalDirectory(directoryName string) (number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, number_of_subdirs uint64, err error) {
	number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files = 0, 0, 0, 0
	number_of_subdirs = 1 // this directory
	files, err := os.ReadDir(directoryName)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	subdirs := []string{}
	for _, file := range files {
		fullPath := filepath.Join(directoryName, file.Name())
		if file.Type().IsRegular() {
			f_number_of_unlinked_files, f_size_of_unlinked_files, f_number_of_linked_files, f_size_of_linked_files, err := duInternalFile(fullPath)
			if err != nil {
				fmt.Printf("Error getting info for regular file %v: %v\n", fullPath, err)
				err = nil
			}
			number_of_unlinked_files += f_number_of_unlinked_files
			size_of_unlinked_files += f_size_of_unlinked_files
			number_of_linked_files += f_number_of_linked_files
			size_of_linked_files += f_size_of_linked_files
		} else if file.Type().IsDir() {
			// store directories for later descending to be a little more cache efficient
			subdirs = append(subdirs, fullPath)
		} else {
			// fmt.Printf("Skipping file of type %v: %v\n", typeToString(file.Type()), fullPath)
		}
	}
	// now descend into the directories
	for _, subdir := range subdirs {
		sd_number_of_unlinked_files, sd_size_of_unlinked_files, sd_number_of_linked_files, sd_size_of_linked_files, sd_number_of_subdirs, err := duInternalDirectory(subdir)
		if err != nil {
			fmt.Printf("Error getting info for directory %v: %v\n", subdir, err)
			err = nil
		} else {
			number_of_unlinked_files += sd_number_of_unlinked_files
			size_of_unlinked_files += sd_size_of_unlinked_files
			number_of_linked_files += sd_number_of_linked_files
			size_of_linked_files += sd_size_of_linked_files
			number_of_subdirs += sd_number_of_subdirs
		}
	}
	return number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, number_of_subdirs, nil
}

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
