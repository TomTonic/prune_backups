//go:build darwin

package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func getSizeAndLinkCount(filename string) (size, link_count uint64, err error) {
	// https://pkg.go.dev/golang.org/x/sys@v0.21.0/unix#Fsopen
	// https://pkg.go.dev/golang.org/x/sys@v0.21.0/unix#Fstat
	// https://pkg.go.dev/golang.org/x/sys@v0.21.0/unix#Stat_t Nlink

	/* does not seem to work with the current implementation of golang.org/x/sys/unix
	flags := unix.O_RDONLY
	fd, err := unix.Fsopen(filename, flags)
	if err != nil {
		f, err2 := os.Open(filename)
		if err2 != nil {
			return 0, 0, fmt.Errorf("Error calling unix.Fsopen (%s) and os.Open (%s).", err, err2)
		}
		fdtest := f.Fd()
		var stat unix.Stat_t
		err3 := unix.Fstat(int(fdtest), &stat)
		if err3 != nil {
			return 0, 0, fmt.Errorf("Error calling unix.Fsopen (%s) and os.Open works. However, unix.Fstat gives an error with the converted filedescriptor ().", err, err3)
		}
		return 0, 0, fmt.Errorf("Error calling unix.Fsopen (%s). os.Open seems to work though and unix.Fstat would accept the filedescriptor...", err)
	}
	defer unix.Close(fd)
	*/

	// alternative code
	f, err := openFileWithRetry(filename, 1000000, 2)
	if err != nil {
		return 0, 0, fmt.Errorf("error calling os.Open (%s)", err)
	}
	defer f.Close()
	fd_ptrtype := f.Fd()
	fd := int(fd_ptrtype) // this cast seems to work currently
	// end alternative code

	var stat unix.Stat_t
	err = unix.Fstat(fd, &stat)
	if err != nil {
		return 0, 0, fmt.Errorf("error calling unix.Fstat (%s)", err)
	}

	size = uint64(stat.Size)
	link_count = uint64(stat.Nlink)

	return size, link_count, err
}
