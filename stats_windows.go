//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func getSizeAndLinkCount(filename string) (size, link_count uint64, err error) {
	// see https://pkg.go.dev/golang.org/x/sys@v0.21.0/windows#pkg-functions
	// GetFileInformationByHandle -> https://pkg.go.dev/golang.org/x/sys@v0.21.0/windows#ByHandleFileInformation
	// FileSizeHigh und FileSizeLow

	file, err := os.Open(filename)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	handle := windows.Handle(file.Fd())

	var fileInfo windows.ByHandleFileInformation
	err = windows.GetFileInformationByHandle(handle, &fileInfo)
	if err != nil {
		return 0, 0, err
	}

	size = uint64(fileInfo.FileSizeHigh)<<32 | uint64(fileInfo.FileSizeLow)
	link_count = uint64(fileInfo.NumberOfLinks)

	return
}
