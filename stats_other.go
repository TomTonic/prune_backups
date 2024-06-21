//go:build !(linux || windows)

package main

import (
	"fmt"
	"runtime"
)

func getSizeAndLinkCount(filename string) (size, link_count uint64, err error) {
	err = fmt.Errorf("Sorry, the necessary low level file system operations are not implemented for this this operating system (%s).", runtime.GOOS)
	return 0, 0, err
}
