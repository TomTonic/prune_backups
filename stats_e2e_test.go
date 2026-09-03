//go:build e2e && !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/TomTonic/rtcompare"
)

// Content/stats E2E test: generates one large, deterministic directory tree
// (same DPRNG seed every run) mixing unlinked files, hardlinked files,
// subdirectories, symlinks, named pipes, and permission-denied entries, then
// asserts DiskUsage's Infoblock matches the counts accumulated while
// generating the tree. Unlike the retention E2E test, no golden file is
// needed here: the expected result is exactly what we just built.
const (
	e2eStatsSeed = 0xC0FFEE

	e2eNumTopDirs         = 40
	e2eNumSubDirsPerTop   = 50
	e2eFilesPerLeafMin    = 90
	e2eFilesPerLeafMax    = 110
	e2eMaxFileSize        = 4096
	e2eNumHardlinkGroups  = 500
	e2eHardlinkMinLinks   = 2
	e2eHardlinkMaxLinks   = 4
	e2eNumSymlinks        = 200
	e2eNumNamedPipes      = 100
	e2eNumPermDeniedFiles = 50
	e2eNumPermDeniedDirs  = 10
)

func Test_E2E_DiskUsage_ContentMix(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission-denied fixtures would not actually deny access")
	}

	root := t.TempDir()
	dprng := rtcompare.NewDPRNG(e2eStatsSeed)
	expected := Infoblock{}

	leafDirs := make([]string, 0, e2eNumTopDirs*e2eNumSubDirsPerTop)
	for i := range e2eNumTopDirs {
		topDir := filepath.Join(root, fmt.Sprintf("top%d", i))
		mustMkdirAllE2E(t, topDir)
		expected.number_of_subdirs++
		for j := range e2eNumSubDirsPerTop {
			leafDir := filepath.Join(topDir, fmt.Sprintf("leaf%d", j))
			mustMkdirAllE2E(t, leafDir)
			expected.number_of_subdirs++
			leafDirs = append(leafDirs, leafDir)
		}
	}

	// Unlinked files, scattered across every leaf directory.
	for _, leafDir := range leafDirs {
		numFiles := e2eFilesPerLeafMin + int(dprng.UInt32N(e2eFilesPerLeafMax-e2eFilesPerLeafMin+1))
		for f := range numFiles {
			size := int(dprng.UInt32N(e2eMaxFileSize))
			path := filepath.Join(leafDir, fmt.Sprintf("file%d", f))
			if err := createE2EFile(path, size); err != nil {
				t.Fatalf("failed to create file %s: %v", path, err)
			}
			expected.number_of_unlinked_files++
			expected.size_of_unlinked_files += uint64(size)
		}
	}

	// Hardlink groups: every path in a group (the original plus its links)
	// is counted as "linked" once its own link count is > 1 - see
	// stats_test.go's Test_du4 for the same accounting.
	for g := range e2eNumHardlinkGroups {
		leafDir := leafDirs[dprng.UInt32N(uint32(len(leafDirs)))]
		size := int(dprng.UInt32N(e2eMaxFileSize))
		numLinks := e2eHardlinkMinLinks + int(dprng.UInt32N(e2eHardlinkMaxLinks-e2eHardlinkMinLinks+1))

		base := filepath.Join(leafDir, fmt.Sprintf("hardlink_base%d", g))
		if err := createE2EFile(base, size); err != nil {
			t.Fatalf("failed to create hardlink base %s: %v", base, err)
		}
		for l := 1; l < numLinks; l++ {
			link := filepath.Join(leafDir, fmt.Sprintf("hardlink_base%d_link%d", g, l))
			if err := os.Link(base, link); err != nil {
				t.Fatalf("failed to link %s -> %s: %v", link, base, err)
			}
		}
		expected.number_of_linked_files += numLinks
		expected.size_of_linked_files += uint64(size) * uint64(numLinks)
	}

	// Symlinks. The target need not exist - only the symlink entry itself is
	// classified (via its own file mode), not what it points to.
	for s := range e2eNumSymlinks {
		leafDir := leafDirs[dprng.UInt32N(uint32(len(leafDirs)))]
		link := filepath.Join(leafDir, fmt.Sprintf("symlink%d", s))
		if err := os.Symlink("/nonexistent-e2e-target", link); err != nil {
			t.Fatalf("failed to create symlink %s: %v", link, err)
		}
		expected.nr_sym++
	}

	// Named pipes (FIFOs).
	for p := range e2eNumNamedPipes {
		leafDir := leafDirs[dprng.UInt32N(uint32(len(leafDirs)))]
		pipePath := filepath.Join(leafDir, fmt.Sprintf("pipe%d", p))
		if err := syscall.Mkfifo(pipePath, 0644); err != nil {
			t.Fatalf("failed to create named pipe %s: %v", pipePath, err)
		}
		expected.nr_pipe++
	}

	// Permission-denied files: readable parent directory, unreadable file.
	for p := range e2eNumPermDeniedFiles {
		leafDir := leafDirs[dprng.UInt32N(uint32(len(leafDirs)))]
		path := filepath.Join(leafDir, fmt.Sprintf("noperm_file%d", p))
		if err := createE2EFile(path, int(dprng.UInt32N(e2eMaxFileSize))); err != nil {
			t.Fatalf("failed to create file %s: %v", path, err)
		}
		if err := os.Chmod(path, 0000); err != nil {
			t.Fatalf("failed to chmod %s: %v", path, err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0644) }) // let t.TempDir() clean up afterwards
		expected.number_of_permission_errors_files++
	}

	// Permission-denied directories: empty, then made unreadable, so nothing
	// inside them needs to be (or could be) accounted for.
	for d := range e2eNumPermDeniedDirs {
		leafDir := leafDirs[dprng.UInt32N(uint32(len(leafDirs)))]
		path := filepath.Join(leafDir, fmt.Sprintf("noperm_dir%d", d))
		mustMkdirAllE2E(t, path)
		if err := os.Chmod(path, 0000); err != nil {
			t.Fatalf("failed to chmod %s: %v", path, err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0755) })
		expected.number_of_subdirs++
		expected.number_of_permission_errors_dirs++
	}

	got, err := DiskUsage(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.number_of_unlinked_files != expected.number_of_unlinked_files {
		t.Errorf("number_of_unlinked_files = %d, want %d", got.number_of_unlinked_files, expected.number_of_unlinked_files)
	}
	if got.size_of_unlinked_files != expected.size_of_unlinked_files {
		t.Errorf("size_of_unlinked_files = %d, want %d", got.size_of_unlinked_files, expected.size_of_unlinked_files)
	}
	if got.number_of_linked_files != expected.number_of_linked_files {
		t.Errorf("number_of_linked_files = %d, want %d", got.number_of_linked_files, expected.number_of_linked_files)
	}
	if got.size_of_linked_files != expected.size_of_linked_files {
		t.Errorf("size_of_linked_files = %d, want %d", got.size_of_linked_files, expected.size_of_linked_files)
	}
	if got.number_of_subdirs != expected.number_of_subdirs {
		t.Errorf("number_of_subdirs = %d, want %d", got.number_of_subdirs, expected.number_of_subdirs)
	}
	if got.nr_sym != expected.nr_sym {
		t.Errorf("nr_sym = %d, want %d", got.nr_sym, expected.nr_sym)
	}
	if got.nr_pipe != expected.nr_pipe {
		t.Errorf("nr_pipe = %d, want %d", got.nr_pipe, expected.nr_pipe)
	}
	if got.number_of_permission_errors_files != expected.number_of_permission_errors_files {
		t.Errorf("number_of_permission_errors_files = %d, want %d", got.number_of_permission_errors_files, expected.number_of_permission_errors_files)
	}
	if got.number_of_permission_errors_dirs != expected.number_of_permission_errors_dirs {
		t.Errorf("number_of_permission_errors_dirs = %d, want %d", got.number_of_permission_errors_dirs, expected.number_of_permission_errors_dirs)
	}
	if got.number_of_other_errors_files != 0 {
		t.Errorf("number_of_other_errors_files = %d, want 0", got.number_of_other_errors_files)
	}
	if got.number_of_other_errors_dirs != 0 {
		t.Errorf("number_of_other_errors_dirs = %d, want 0", got.number_of_other_errors_dirs)
	}
}
