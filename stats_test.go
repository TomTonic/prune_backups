package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
			number_of_subdirs:        4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := du(tt.dir)
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
	defer os.RemoveAll(dir) // clean up

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
	number_of_subdirs := 1

	got := du(dir)
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
	defer os.RemoveAll(dir) // clean up

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
	number_of_subdirs := (1)

	got := du(dir)
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
	defer os.RemoveAll(dir) // clean up

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
	number_of_subdirs := (1)

	got := du(dir)
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

func createTestfile(name string, size int) (err error) {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

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
