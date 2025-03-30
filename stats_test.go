package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsDirectory(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"./testdata", true},     // Assuming testdata is a directory
		{"./go.mod", false},      // Assuming go.mod is a file
		{"./nonexistent", false}, // Non-existent path
	}

	for _, test := range tests {
		result, err := isDirectory(test.path)
		if err != nil && test.expected {
			t.Errorf("isDirectory(%s) returned error: %v", test.path, err)
		}
		if result != test.expected {
			t.Errorf("isDirectory(%s) = %v; want %v", test.path, result, test.expected)
		}
	}
}

func TestOpenFileWithRetry(t *testing.T) {
	tests := []struct {
		filename        string
		retries         int
		maxwait_seconds int
		expectError     bool
	}{
		{"go.mod", 3, 1, false},             // Assuming go.mod exists
		{"nonexistentfile.txt", 3, 1, true}, // Non-existent file
		{"irrelevant", 0, 1, true},          // try 0 times -> shall return cannot open file error
	}

	for _, test := range tests {
		file, err := openFileWithRetry(test.filename, test.retries, test.maxwait_seconds)
		if test.expectError {
			if err == nil {
				t.Errorf("expected error for file %s, but got none", test.filename)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for file %s: %v", test.filename, err)
			}
			if file != nil {
				err2 := file.Close()
				if err2 != nil {
					t.Errorf("unexpected error for file %s: %v", test.filename, err2)
				}
			}
		}
	}
}

func TestReadDirWithRetry(t *testing.T) {
	tests := []struct {
		directoryname   string
		retries         int
		maxwait_seconds int
		expectError     bool
	}{
		{"testdata", 3, 1, false},      // Assuming testdata exists
		{"nonexistentdir", 3, 1, true}, // Non-existent directory
		{"irrelevant", 0, 1, true},     // try 0 times -> shall return cannot open file error
	}

	for _, test := range tests {
		direntries, err := readDirWithRetry(test.directoryname, test.retries, test.maxwait_seconds)
		if test.expectError {
			if err == nil {
				t.Errorf("expected error for directory %s, but got none", test.directoryname)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for directory %s: %v", test.directoryname, err)
			} else {
				t.Logf("Directory: %s, Entries: %d", test.directoryname, len(direntries))
			}
		}
	}
}

func TestGetSizeAndLinkCount(t *testing.T) {
	tests := []struct {
		filename    string
		expectError bool
	}{
		{"go.mod", false},             // Assuming go.mod exists
		{"nonexistentfile.txt", true}, // Non-existent file
	}

	for _, test := range tests {
		size, linkCount, err := getSizeAndLinkCount(test.filename)
		if test.expectError {
			if err == nil {
				t.Errorf("expected error for file %s, but got none", test.filename)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for file %s: %v", test.filename, err)
			} else {
				t.Logf("File: %s, Size: %d, Link Count: %d", test.filename, size, linkCount)
			}
		}
	}
}

func Test_du0(t *testing.T) {
	expectedOutput := "open nonexistingfileordirectoryname5648623485762456: no such file or directory"
	if runtime.GOOS == "windows" {
		expectedOutput = "open nonexistingfileordirectoryname5648623485762456: The system cannot find the file specified."
	}

	_, err := du("nonexistingfileordirectoryname5648623485762456")

	if err != nil {
		if !strings.HasPrefix(err.Error(), expectedOutput) {
			t.Fatalf("expected %q, got %q", expectedOutput, err.Error())
		}
	} else {
		t.Fatalf("expected error")
	}
}

func Test_du1(t *testing.T) {
	tests := []struct {
		name                     string
		dir                      string
		number_of_unlinked_files int
		size_of_unlinked_files   uint64
		number_of_linked_files   int
		size_of_linked_files     uint64
		number_of_subdirs        int
	}{
		{
			name:                     "Test 1",
			dir:                      "./testdata/du_test1",
			number_of_unlinked_files: 4,
			size_of_unlinked_files:   20,
			number_of_linked_files:   0,
			size_of_linked_files:     0,
			number_of_subdirs:        3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := du(tt.dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			//got_number_of_unlinked_files, got_size_of_unlinked_files, got_number_of_linked_files, got_size_of_linked_files, got_number_of_subdirs := du(tt.dir)
			if got.number_of_unlinked_files != tt.number_of_unlinked_files || got.size_of_unlinked_files != tt.size_of_unlinked_files || got.number_of_linked_files != tt.number_of_linked_files || got.size_of_linked_files != tt.size_of_linked_files || got.number_of_subdirs != tt.number_of_subdirs {
				t.Errorf("du got: #uf:%v/%v, size uf:%v/%v, #lf:%v/%v, size lf:%v/%v, dirs:%v/%v",
					got.number_of_unlinked_files, tt.number_of_unlinked_files,
					got.size_of_unlinked_files, tt.size_of_unlinked_files,
					got.number_of_linked_files, tt.number_of_linked_files,
					got.size_of_linked_files, tt.size_of_linked_files,
					got.number_of_subdirs, tt.number_of_subdirs)
			}
		})
	}
}

func Test_du2(t *testing.T) {
	dir, err := os.MkdirTemp(USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES, "prune_backups_testdir")
	if err != nil {
		t.Fatal("Error creating temporary directory: ", err)
	}
	defer func() {
		_ = os.RemoveAll(dir) // clean up
	}()

	fname1 := filepath.Join(dir, "testfile1")
	lname1 := filepath.Join(dir, "link1_to_testfile1")

	err = createTestfile(fname1, 4444)
	if err != nil {
		t.Fatal("Error creating testfile1: ", err)
	}

	err = os.Link(fname1, lname1)
	if err != nil {
		t.Fatal("Error linking file1: ", err)
	}

	number_of_unlinked_files := 0
	size_of_unlinked_files := uint64(0)
	number_of_linked_files := 2
	size_of_linked_files := uint64(8888)
	number_of_subdirs := 0

	got, err := du(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	//got_number_of_unlinked_files, got_size_of_unlinked_files, got_number_of_linked_files, got_size_of_linked_files, got_number_of_subdirs := du(dir)
	if got.number_of_unlinked_files != number_of_unlinked_files || got.size_of_unlinked_files != size_of_unlinked_files || got.number_of_linked_files != number_of_linked_files || got.size_of_linked_files != size_of_linked_files || got.number_of_subdirs != number_of_subdirs {
		t.Errorf("du got: #uf:%v/%v, size uf:%v/%v, #lf:%v/%v, size lf:%v/%v, dirs:%v/%v",
			got.number_of_unlinked_files, number_of_unlinked_files,
			got.size_of_unlinked_files, size_of_unlinked_files,
			got.number_of_linked_files, number_of_linked_files,
			got.size_of_linked_files, size_of_linked_files,
			got.number_of_subdirs, number_of_subdirs)
	}
}

func Test_du3(t *testing.T) {
	dir, err := os.MkdirTemp(USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES, "prune_backups_testdir")
	if err != nil {
		t.Fatal("Error creating temporary directory: ", err)
	}
	defer func() {
		_ = os.RemoveAll(dir) // clean up
	}()

	fname1 := filepath.Join(dir, "testfile1")
	fname2 := filepath.Join(dir, "testfile2")
	lname1 := filepath.Join(dir, "link1_to_testfile1")

	err = createTestfile(fname1, 4444)
	if err != nil {
		t.Fatal("Error creating testfile1: ", err)
	}

	err = createTestfile(fname2, 1111)
	if err != nil {
		t.Fatal("Error creating testfile1: ", err)
	}

	err = os.Link(fname1, lname1)
	if err != nil {
		t.Fatal("Error linking file1: ", err)
	}

	number_of_unlinked_files := (1)
	size_of_unlinked_files := uint64(1111)
	number_of_linked_files := (2)
	size_of_linked_files := uint64(8888)
	number_of_subdirs := (0)

	got, err := du(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	//got_number_of_unlinked_files, got_size_of_unlinked_files, got_number_of_linked_files, got_size_of_linked_files, got_number_of_subdirs := du(dir)
	if got.number_of_unlinked_files != number_of_unlinked_files || got.size_of_unlinked_files != size_of_unlinked_files || got.number_of_linked_files != number_of_linked_files || got.size_of_linked_files != size_of_linked_files || got.number_of_subdirs != number_of_subdirs {
		t.Errorf("du got: #uf:%v/%v, size uf:%v/%v, #lf:%v/%v, size lf:%v/%v, dirs:%v/%v",
			got.number_of_unlinked_files, number_of_unlinked_files,
			got.size_of_unlinked_files, size_of_unlinked_files,
			got.number_of_linked_files, number_of_linked_files,
			got.size_of_linked_files, size_of_linked_files,
			got.number_of_subdirs, number_of_subdirs)
	}
}

func Test_du4(t *testing.T) {
	dir, err := os.MkdirTemp(USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES, "prune_backups_testdir")
	if err != nil {
		t.Fatal("Error creating temporary directory: ", err)
	}
	defer func() {
		_ = os.RemoveAll(dir) // clean up
	}()

	fname1 := filepath.Join(dir, "testfile1")
	fname2 := filepath.Join(dir, "testfile2")
	fname3 := filepath.Join(dir, "testfile3")
	lname1 := filepath.Join(dir, "link1_to_testfile1")
	lname2 := filepath.Join(dir, "link2_to_testfile2")
	lname3 := filepath.Join(dir, "link3_to_testfile2")

	err = createTestfile(fname1, 37)
	if err != nil {
		t.Fatal("Error creating testfile1: ", err)
	}

	err = createTestfile(fname2, 41)
	if err != nil {
		t.Fatal("Error creating testfile2: ", err)
	}

	err = createTestfile(fname3, 43)
	if err != nil {
		t.Fatal("Error creating testfile3: ", err)
	}

	err = os.Link(fname1, lname1)
	if err != nil {
		t.Fatal("Error linking file1: ", err)
	}

	err = os.Link(fname2, lname2)
	if err != nil {
		t.Fatal("Error linking file2: ", err)
	}

	err = os.Link(fname2, lname3)
	if err != nil {
		t.Fatal("Error linking file2(2): ", err)
	}

	number_of_unlinked_files := (1)
	size_of_unlinked_files := uint64(43)
	number_of_linked_files := (5)
	size_of_linked_files := uint64(37 + 37 + 41 + 41 + 41)
	number_of_subdirs := (0)

	got, err := du(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	//got_number_of_unlinked_files, got_size_of_unlinked_files, got_number_of_linked_files, got_size_of_linked_files, got_number_of_subdirs := du(dir)
	if got.number_of_unlinked_files != number_of_unlinked_files || got.size_of_unlinked_files != size_of_unlinked_files || got.number_of_linked_files != number_of_linked_files || got.size_of_linked_files != size_of_linked_files || got.number_of_subdirs != number_of_subdirs {
		t.Errorf("du got: #uf:%v/%v, size uf:%v/%v, #lf:%v/%v, size lf:%v/%v, dirs:%v/%v",
			got.number_of_unlinked_files, number_of_unlinked_files,
			got.size_of_unlinked_files, size_of_unlinked_files,
			got.number_of_linked_files, number_of_linked_files,
			got.size_of_linked_files, size_of_linked_files,
			got.number_of_subdirs, number_of_subdirs)
	}
}

func Test_du5(t *testing.T) {
	t.Run("AccessDenied", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("This test does not work on Windows")
			// When you set a directory to 0444 permissions in Windows, it means that the directory
			// is readable by everyone but not writable or executable by anyone. However, Windows
			// allows the creation of child directories even with these restrictive permissions because
			// the permissions for new directories are determined by the permissions of the parent
			// directory and the user's permissions
		}

		// Setup
		testDir := t.TempDir()
		noreadDir := filepath.Join(testDir, "noread")
		err := os.Mkdir(noreadDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create noread directory: %v", err)
		}

		noreadDirSub1 := filepath.Join(noreadDir, "sub1")
		err = os.Mkdir(noreadDirSub1, 0755)
		if err != nil {
			t.Fatalf("Failed to create noreadDirSub1 directory: %v", err)
		}

		noreadDirSub2 := filepath.Join(noreadDir, "sub2")
		err = os.Mkdir(noreadDirSub2, 0755)
		if err != nil {
			t.Fatalf("Failed to create noreadDirSub2 directory: %v", err)
		}

		noreadDirTxt := filepath.Join(noreadDir, "test.txt")
		err = createTestfile(noreadDirTxt, 37)
		if err != nil {
			t.Fatal("Error creating noreadDirTxt: ", err)
		}

		noreadDirSub1Txt := filepath.Join(noreadDirSub1, "testSub.txt")
		err = createTestfile(noreadDirSub1Txt, 41)
		if err != nil {
			t.Fatal("Error creating noreadDirSub1Txt: ", err)
		}

		err = os.Chmod(noreadDirSub1, 0000)
		if err != nil {
			t.Fatalf("Failed to set noreadDirSub1 permissions: %v", err)
		}

		err = os.Chmod(noreadDirTxt, 0000)
		if err != nil {
			t.Fatalf("Failed to set noreadDirTxt permissions: %v", err)
		}

		number_of_unlinked_files := (0)
		size_of_unlinked_files := uint64(0)
		number_of_linked_files := (0)
		size_of_linked_files := uint64(0)
		number_of_subdirs := (3)

		got, err := du(testDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.number_of_unlinked_files != number_of_unlinked_files || got.size_of_unlinked_files != size_of_unlinked_files || got.number_of_linked_files != number_of_linked_files || got.size_of_linked_files != size_of_linked_files || got.number_of_subdirs != number_of_subdirs {
			t.Errorf("du got: #uf:%v/%v, size uf:%v/%v, #lf:%v/%v, size lf:%v/%v, dirs:%v/%v",
				got.number_of_unlinked_files, number_of_unlinked_files,
				got.size_of_unlinked_files, size_of_unlinked_files,
				got.number_of_linked_files, number_of_linked_files,
				got.size_of_linked_files, size_of_linked_files,
				got.number_of_subdirs, number_of_subdirs)
		}

		number_of_permission_errors_files := 1
		number_of_permission_errors_dirs := 1
		number_of_other_errors_files := 0
		number_of_other_errors_dirs := 0

		if got.number_of_permission_errors_files != number_of_permission_errors_files || got.number_of_permission_errors_dirs != number_of_permission_errors_dirs || got.number_of_other_errors_files != number_of_other_errors_files || got.number_of_other_errors_dirs != number_of_other_errors_dirs {
			t.Errorf("du got: pef:%v/%v, ped:%v/%v, oef:%v/%v, oed:%v/%v",
				got.number_of_permission_errors_files, number_of_permission_errors_files,
				got.number_of_permission_errors_dirs, number_of_permission_errors_dirs,
				got.number_of_other_errors_files, number_of_other_errors_files,
				got.number_of_other_errors_dirs, number_of_other_errors_dirs)
		}

		err = os.Chmod(noreadDirSub1, 0755)
		if err != nil {
			t.Fatalf("Failed to set noreadDirSub1 permissions: %v", err)
		}

		err = os.Chmod(noreadDirTxt, 0755)
		if err != nil {
			t.Fatalf("Failed to set noreadDirTxt permissions: %v", err)
		}

	})
}

func createTestfile(name string, size int) (err error) {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	fileSize := size // size in bytes
	dummyData := make([]byte, fileSize)

	_, err = file.Write(dummyData)
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func TestCountAccordingType(t *testing.T) {
	tests := []struct {
		mode     fs.FileMode
		expected func(info *infoblock_internal) int
	}{
		{fs.ModeDir, func(info *infoblock_internal) int { return info.ib.number_of_subdirs }},
		{fs.ModeAppend, func(info *infoblock_internal) int { return info.ib.nr_apnd }},
		{fs.ModeExclusive, func(info *infoblock_internal) int { return info.ib.nr_excl }},
		{fs.ModeTemporary, func(info *infoblock_internal) int { return info.ib.nr_tmp }},
		{fs.ModeSymlink, func(info *infoblock_internal) int { return info.ib.nr_sym }},
		{fs.ModeDevice, func(info *infoblock_internal) int { return info.ib.nr_dev }},
		{fs.ModeNamedPipe, func(info *infoblock_internal) int { return info.ib.nr_pipe }},
		{fs.ModeSocket, func(info *infoblock_internal) int { return info.ib.nr_sock }},
	}

	for _, test := range tests {
		info := &infoblock_internal{}
		countAccordingType(test.mode, info)
		if test.expected(info) != 1 {
			t.Errorf("expected 1, got %d for mode %v", test.expected(info), test.mode)
		}
	}
}
