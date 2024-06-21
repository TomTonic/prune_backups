//go:build linux || darwin

package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func getSizeAndLinkCount(filename string) (size, link_count uint64, err error) {
	// https://pkg.go.dev/golang.org/x/sys@v0.21.0/unix#Fsopen
	// https://pkg.go.dev/golang.org/x/sys@v0.21.0/unix#Fstat
	// https://pkg.go.dev/golang.org/x/sys@v0.21.0/unix#Stat_t Nlink

	flags := unix.O_RDONLY
	fd, err := unix.Fsopen(filename, flags)
	if err != nil {
		return 0, 0, err
	}
	defer unix.Close(fd)

	fmt.Println("File descriptor:", fd)

	var stat unix.Stat_t
	err = unix.Fstat(fd, &stat)
	if err != nil {
		return 0, 0, err
	}

	size = uint64(stat.Size)
	link_count = stat.Nlink

	return size, link_count, err
}
