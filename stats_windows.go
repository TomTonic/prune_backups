//go:build windows

package main

import (
	"errors"
	"fmt"
	"io/fs"

	"golang.org/x/sys/windows"
)

func getSizeAndLinkCount(filename string) (size, link_count uint64, err error) {
	// see https://pkg.go.dev/golang.org/x/sys@v0.21.0/windows#pkg-functions
	// GetFileInformationByHandle -> https://pkg.go.dev/golang.org/x/sys@v0.21.0/windows#ByHandleFileInformation
	// FileSizeHigh und FileSizeLow

	//file, err := os.Open(filename)
	file, err := openFileWithRetry(filename, 1000000, 2)
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return 0, 0, err
		}
		return 0, 0, fmt.Errorf("error opening file: %v", err)
	}

	handle := windows.Handle(file.Fd())

	var fileInfo windows.ByHandleFileInformation
	err = windows.GetFileInformationByHandle(handle, &fileInfo)
	if err != nil {
		return 0, 0, fmt.Errorf("error calling windows.GetFileInformationByHandle (%s)", err)
	}

	size = uint64(fileInfo.FileSizeHigh)<<32 | uint64(fileInfo.FileSizeLow)
	link_count = uint64(fileInfo.NumberOfLinks)

	return
}
