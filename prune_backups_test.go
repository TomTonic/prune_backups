package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

func compareArrays(result []string, want []string, t *testing.T) {
	max := len(result)
	if len(want) > max {
		max = len(want)
	}
	for i := 0; i < max; i++ {
		if i < len(want) && i < len(result) {
			t.Logf("   wanted: %v, got: %v", want[i], result[i])
		} else if i < len(want) {
			t.Logf("   wanted: %v, got: <no more values>", want[i])
		} else {
			t.Logf("   wanted: <no more values>, got: %v", result[i])
		}
	}
}

func Test_pruneDirectoryHourlyForFourMonths(t *testing.T) {
	test_time_gen := time.Date(2024, 6, 17, 9, 49, 33, 0, time.UTC)
	test_time_prune := time.Date(2024, 6, 17, 9, 54, 21, 0, time.UTC)

	test_dir := generateHourlyTestDirectories(t, test_time_gen, 2800)

	want := []string{
		// when sorting lexographically descending, the to_delete directory will be the first
		"to_delete",
		// 24 for the hours
		"2024-06-17_09-49", "2024-06-17_08-49", "2024-06-17_07-49", "2024-06-17_06-49", "2024-06-17_05-49", "2024-06-17_04-49",
		"2024-06-17_03-49", "2024-06-17_02-49", "2024-06-17_01-49", "2024-06-17_00-49", "2024-06-16_23-49", "2024-06-16_22-49",
		"2024-06-16_21-49", "2024-06-16_20-49", "2024-06-16_19-49", "2024-06-16_18-49", "2024-06-16_17-49", "2024-06-16_16-49",
		"2024-06-16_15-49", "2024-06-16_14-49", "2024-06-16_13-49", "2024-06-16_12-49", "2024-06-16_11-49", "2024-06-16_10-49",
		// 30 for the days
		"2024-06-15_23-49", "2024-06-14_23-49", "2024-06-13_23-49", "2024-06-12_23-49", "2024-06-11_23-49", "2024-06-10_23-49", "2024-06-09_23-49", "2024-06-08_23-49", "2024-06-07_23-49", "2024-06-06_23-49",
		"2024-06-05_23-49", "2024-06-04_23-49", "2024-06-03_23-49", "2024-06-02_23-49", "2024-06-01_23-49", "2024-05-31_23-49", "2024-05-30_23-49", "2024-05-29_23-49", "2024-05-28_23-49", "2024-05-27_23-49",
		"2024-05-26_23-49", "2024-05-25_23-49", "2024-05-24_23-49", "2024-05-23_23-49", "2024-05-22_23-49", "2024-05-21_23-49", "2024-05-20_23-49", "2024-05-19_23-49", "2024-05-18_23-49", "2024-05-17_23-49",
		// 3 for the months
		"2024-04-30_23-49", "2024-03-31_23-49", "2024-02-29_23-49",
	}

	wanted_number_of_deleted := 2800 - 24 - 30 - 3 // please nothe that the to_delete-directory will not be moved!

	pruneAndCheck(t, test_dir, test_time_prune, want, wanted_number_of_deleted)

	defer os.RemoveAll(test_dir) // clean up
}

func Test_pruneDirectoryYesterdayMissing(t *testing.T) {
	test_time_prune := time.Date(2024, 6, 17, 9, 54, 21, 0, time.UTC)

	given := []string{
		// some hourly backups for the 17th
		"2024-06-17_09-49", "2024-06-17_06-49", "2024-06-17_00-49",
		// no hourly backups within the lase 24h on the 16th
		// but an hourly backup before the 24h time window
		"2024-06-16_03-49", "2024-06-16_02-49", "2024-06-16_01-49",
		// and some normal daily backups to prune
		"2024-06-15_23-49", "2024-06-15_13-49",
		"2024-06-14_23-49",
		"2024-06-13_23-49", "2024-06-13_22-49", "2024-06-13_21-49",
	}

	test_dir := generateTestDirectories(t, given)

	wanted := []string{
		// when sorting lexographically descending, the to_delete directory will be the first
		"to_delete",
		// 24 for the hours
		// some hourly backups for the 17th
		"2024-06-17_09-49", "2024-06-17_06-49", "2024-06-17_00-49",
		// no hourly backups within the lase 24h on the 16th
		// but keep the newest from the 16th
		"2024-06-16_03-49",
		// and some normal daily backups to prune
		"2024-06-15_23-49",
		"2024-06-14_23-49",
		"2024-06-13_23-49",
	}

	wanted_number_of_deleted := 5 // please nothe that the to_delete-directory will not be moved!

	pruneAndCheck(t, test_dir, test_time_prune, wanted, wanted_number_of_deleted)

	defer os.RemoveAll(test_dir) // clean up
}

func pruneAndCheck(t *testing.T, test_dir string, test_time_pruning time.Time, expect_remaining []string, number_expect_deleted int) {
	pruneDirectory(test_dir, test_time_pruning, "to_delete", 0, false)

	// get resulting directories and sort descending
	result := getAllDirectories(t, test_dir)
	sort.Sort(sort.Reverse(sort.StringSlice(result)))

	// compare result to expect_remaining
	if !reflect.DeepEqual(result, expect_remaining) {
		t.Errorf("Remaining directories not as expected!")
		compareArrays(result, expect_remaining, t)
	}

	// compare count of deleted directories
	deleted := getAllDirectories(t, filepath.Join(test_dir, "to_delete"))
	got_number_of_deleted := len(deleted)
	if number_expect_deleted != got_number_of_deleted {
		t.Errorf("Number of deleted directories not as expected: wanted=%v, got=%v", number_expect_deleted, got_number_of_deleted)
	}
}

func getAllDirectories(t *testing.T, dir string) []string {
	all_entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal("Error reading temporary directory: ", err)
	}

	var result []string
	for _, file := range all_entries {
		if file.IsDir() {
			result = append(result, file.Name())
		}
	}
	return result
}

const USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES = "" // see https://pkg.go.dev/os#MkdirTemp

func generateHourlyTestDirectories(t *testing.T, test_time time.Time, number int) string {
	var dir string

	dir, err := os.MkdirTemp(USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES, "prune_backups_testdir")
	if err != nil {
		t.Fatal("Error creating temporary directory: ", err)
	}

	for i := 0; i < number; i++ {
		next := test_time.Add(time.Duration(-i) * time.Hour)
		subDir := next.Format("2006-01-02_15-04")
		fullPath := filepath.Join(dir, subDir)
		err2 := os.MkdirAll(fullPath, 0755)
		if err2 != nil {
			t.Fatal("Error creating child in temporary directory: ", err2)
		}
	}

	return dir
}

func generateTestDirectories(t *testing.T, directories []string) string {
	var dir string

	dir, err := os.MkdirTemp(USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES, "prune_backups_testdir")
	if err != nil {
		t.Fatal("Error creating temporary directory: ", err)
	}

	for _, subDir := range directories {
		fullPath := filepath.Join(dir, subDir)
		err2 := os.MkdirAll(fullPath, 0755)
		if err2 != nil {
			t.Fatal("Error creating child in temporary directory: ", err2)
		}
	}

	return dir
}

func Test_getFiltersForDailys(t *testing.T) {
	for _, tt := range testsFor30Dailys {
		t.Run(tt.name, func(t *testing.T) {
			gotFilters, gotMonth := getFiltersForDailys(tt.test_time, tt.existing_dirs)
			if gotMonth != tt.next_month {
				t.Errorf("The month to continue diverges: expected=%v, got=%v", gotMonth, tt.next_month)
			}
			if !reflect.DeepEqual(gotFilters, tt.filter_dates) {
				compareArrays(gotFilters, tt.filter_dates, t)
				t.Errorf("getFiltersForDailys() result not as expected!")
			}
		})
	}
}

var testsFor30Dailys = []struct {
	name          string
	test_time     time.Time
	next_month    time.Time
	existing_dirs []string
	filter_dates  []string
}{
	{
		name:       "Test Case 1a - middle of the month, all existing",
		test_time:  time.Date(2014, 7, 15, 9, 54, 21, 0, time.UTC),
		next_month: time.Date(2014, 5, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{
			"2014-07-17_23-54", "2014-07-16_23-54", "2014-07-15_23-54", "2014-07-14_23-54", "2014-07-13_23-54", "2014-07-12_23-54",
			"2014-07-11_23-54", "2014-07-10_23-54", "2014-07-09_23-54", "2014-07-08_23-54", "2014-07-07_23-54", "2014-07-06_23-54",
			"2014-07-05_23-54", "2014-07-04_23-54", "2014-07-03_23-54", "2014-07-02_23-54", "2014-07-01_23-54", "2014-06-30_23-54",
			"2014-06-29_23-54", "2014-06-28_23-54", "2014-06-27_23-54", "2014-06-26_23-54", "2014-06-25_23-54", "2014-06-24_23-54",
			"2014-06-23_23-54", "2014-06-22_23-54", "2014-06-21_23-54", "2014-06-20_23-54", "2014-06-19_23-54", "2014-06-18_23-54",
			"2014-06-17_23-54", "2014-06-16_23-54", "2014-06-15_23-54", "2014-06-14_23-54", "2014-06-13_23-54", "2014-06-12_23-54",
		},
		filter_dates: []string{
			// today and yesterday and 15 days in a month
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ "2014-07-15", "2014-07-14", "2014-07-13", "2014-07-12", "2014-07-11",
			"2014-07-10", "2014-07-09", "2014-07-08", "2014-07-07", "2014-07-06", "2014-07-05", "2014-07-04", "2014-07-03", "2014-07-02", "2014-07-01",
			// ... and 15 days in the other month
			"2014-06-30", "2014-06-29", "2014-06-28", "2014-06-27", "2014-06-26", "2014-06-25", "2014-06-24", "2014-06-23", "2014-06-22", "2014-06-21",
			"2014-06-20", "2014-06-19", "2014-06-18", "2014-06-17", "2014-06-16", /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/
		},
	},
	{
		name:          "Test Case 1b - middle of the month, none existing",
		test_time:     time.Date(2014, 7, 15, 9, 54, 21, 0, time.UTC),
		next_month:    time.Date(2014, 5, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{},
		filter_dates: []string{
			// today and yesterday and 15 days in a month
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ "2014-07-15", "2014-07-14", "2014-07-13", "2014-07-12", "2014-07-11",
			"2014-07-10", "2014-07-09", "2014-07-08", "2014-07-07", "2014-07-06", "2014-07-05", "2014-07-04", "2014-07-03", "2014-07-02", "2014-07-01",
			// complete June
			"2014-06",
		},
	},
	{
		name:       "Test Case 2a - 1st of the month, prev 31 days, all existing",
		test_time:  time.Date(2022, 8, 1, 23, 54, 21, 0, time.UTC),
		next_month: time.Date(2022, 6, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{
			"2022-08-01_23-54", "2022-07-31_23-54", "2022-07-30_23-54", "2022-07-29_23-54", "2022-07-28_23-54", "2022-07-27_23-54",
			"2022-07-26_23-54", "2022-07-25_23-54", "2022-07-24_23-54", "2022-07-23_23-54", "2022-07-22_23-54", "2022-07-21_23-54",
			"2022-07-20_23-54", "2022-07-19_23-54", "2022-07-18_23-54", "2022-07-17_23-54", "2022-07-16_23-54", "2022-07-15_23-54",
			"2022-07-14_23-54", "2022-07-13_23-54", "2022-07-12_23-54", "2022-07-11_23-54", "2022-07-10_23-54", "2022-07-09_23-54",
			"2022-07-08_23-54", "2022-07-07_23-54", "2022-07-06_23-54", "2022-07-05_23-54", "2022-07-04_23-54", "2022-07-03_23-54",
			"2022-07-02_23-54", "2022-07-01_23-54", "2022-06-30_23-54", "2022-06-29_23-54", "2022-06-28_23-54", "2022-06-27_23-54",
		},
		filter_dates: []string{
			"2022-08-01", "2022-07-31", "2022-07-30", "2022-07-29", "2022-07-28", "2022-07-27",
			"2022-07-26", "2022-07-25", "2022-07-24", "2022-07-23", "2022-07-22", "2022-07-21",
			"2022-07-20", "2022-07-19", "2022-07-18", "2022-07-17", "2022-07-16", "2022-07-15",
			"2022-07-14", "2022-07-13", "2022-07-12", "2022-07-11", "2022-07-10", "2022-07-09",
			"2022-07-08", "2022-07-07", "2022-07-06", "2022-07-05", "2022-07-04", "2022-07-03",
		},
	},
	{
		name:          "Test Case 2b - 1st of the month, prev 31 days, none existing",
		test_time:     time.Date(2022, 8, 1, 23, 54, 21, 0, time.UTC),
		next_month:    time.Date(2022, 6, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{},
		filter_dates: []string{
			"2022-08-01",
			// complete July
			"2022-07",
		},
	},
	{
		name:       "Test Case 3a - 1st of the month, prev 29 days, all existing",
		test_time:  time.Date(2024, 3, 1, 23, 54, 21, 0, time.UTC),
		next_month: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{
			"2024-03-01_23-54", "2024-02-29_23-54", "2024-02-28_23-54", "2024-02-27_23-54", "2024-02-26_23-54", "2024-02-25_23-54",
			"2024-02-24_23-54", "2024-02-23_23-54", "2024-02-22_23-54", "2024-02-21_23-54", "2024-02-20_23-54", "2024-02-19_23-54",
			"2024-02-18_23-54", "2024-02-17_23-54", "2024-02-16_23-54", "2024-02-15_23-54", "2024-02-14_23-54", "2024-02-13_23-54",
			"2024-02-12_23-54", "2024-02-11_23-54", "2024-02-10_23-54", "2024-02-09_23-54", "2024-02-08_23-54", "2024-02-07_23-54",
			"2024-02-06_23-54", "2024-02-05_23-54", "2024-02-04_23-54", "2024-02-03_23-54", "2024-02-02_23-54", "2024-02-01_23-54",
			"2024-01-31_23-54", "2024-01-30_23-54", "2024-01-29_23-54", "2024-01-28_23-54", "2024-01-27_23-54", "2024-01-26_23-54",
		},
		filter_dates: []string{
			"2024-03-01", "2024-02-29", "2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25",
			"2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21", "2024-02-20", "2024-02-19",
			"2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13",
			"2024-02-12", "2024-02-11", "2024-02-10", "2024-02-09", "2024-02-08", "2024-02-07",
			"2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01",
		},
	},
	{
		name:          "Test Case 3a - 1st of the month, prev 29 days, none existing",
		test_time:     time.Date(2024, 3, 1, 23, 54, 21, 0, time.UTC),
		next_month:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{},
		filter_dates: []string{
			"2024-03-01", "2024-02-29", "2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25",
			"2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21", "2024-02-20", "2024-02-19",
			"2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13",
			"2024-02-12", "2024-02-11", "2024-02-10", "2024-02-09", "2024-02-08", "2024-02-07",
			"2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01",
		},
	},
	{
		name:       "Test Case 4a - 1st of the month, prev 28 days (3 months coverage), all existing",
		test_time:  time.Date(2023, 3, 1, 23, 54, 21, 0, time.UTC),
		next_month: time.Date(2022, 12, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{
			"2023-03-01_23-54", "2023-02-28_23-54", "2023-02-27_23-54", "2023-02-26_23-54", "2023-02-25_23-54", "2023-02-24_23-54",
			"2023-02-23_23-54", "2023-02-22_23-54", "2023-02-21_23-54", "2023-02-20_23-54", "2023-02-19_23-54", "2023-02-18_23-54",
			"2023-02-17_23-54", "2023-02-16_23-54", "2023-02-15_23-54", "2023-02-14_23-54", "2023-02-13_23-54", "2023-02-12_23-54",
			"2023-02-11_23-54", "2023-02-10_23-54", "2023-02-09_23-54", "2023-02-08_23-54", "2023-02-07_23-54", "2023-02-06_23-54",
			"2023-02-05_23-54", "2023-02-04_23-54", "2023-02-03_23-54", "2023-02-02_23-54", "2023-02-01_23-54", "2023-01-31_23-54",
			"2023-01-30_23-54", "2023-01-29_23-54", "2023-01-28_23-54", "2023-01-27_23-54", "2023-01-26_23-54", "2023-01-25_23-54",
		},
		filter_dates: []string{
			"2023-03-01", "2023-02-28", "2023-02-27", "2023-02-26", "2023-02-25", "2023-02-24",
			"2023-02-23", "2023-02-22", "2023-02-21", "2023-02-20", "2023-02-19", "2023-02-18",
			"2023-02-17", "2023-02-16", "2023-02-15", "2023-02-14", "2023-02-13", "2023-02-12",
			"2023-02-11", "2023-02-10", "2023-02-09", "2023-02-08", "2023-02-07", "2023-02-06",
			"2023-02-05", "2023-02-04", "2023-02-03", "2023-02-02", "2023-02-01", "2023-01-31",
		},
	},
	{
		name:          "Test Case 4a - 1st of the month, prev 28 days (3 months coverage), none existing",
		test_time:     time.Date(2023, 3, 1, 23, 54, 21, 0, time.UTC),
		next_month:    time.Date(2022, 12, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{},
		filter_dates: []string{
			"2023-03-01", "2023-02-28", "2023-02-27", "2023-02-26", "2023-02-25", "2023-02-24",
			"2023-02-23", "2023-02-22", "2023-02-21", "2023-02-20", "2023-02-19", "2023-02-18",
			"2023-02-17", "2023-02-16", "2023-02-15", "2023-02-14", "2023-02-13", "2023-02-12",
			"2023-02-11", "2023-02-10", "2023-02-09", "2023-02-08", "2023-02-07", "2023-02-06",
			"2023-02-05", "2023-02-04", "2023-02-03", "2023-02-02", "2023-02-01",
			// complete January
			"2023-01",
		},
	},
	{
		name:       "Test Case 5a - 30th of the month, has >=30 days (only 1 month coverage), all existing",
		test_time:  time.Date(2024, 4, 30, 23, 54, 21, 0, time.UTC),
		next_month: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{
			"2024-04-30_23-54", "2024-04-29_23-54", "2024-04-28_23-54", "2024-04-27_23-54", "2024-04-26_23-54", "2024-04-25_23-54",
			"2024-04-24_23-54", "2024-04-23_23-54", "2024-04-22_23-54", "2024-04-21_23-54", "2024-04-20_23-54", "2024-04-19_23-54",
			"2024-04-18_23-54", "2024-04-17_23-54", "2024-04-16_23-54", "2024-04-15_23-54", "2024-04-14_23-54", "2024-04-13_23-54",
			"2024-04-12_23-54", "2024-04-11_23-54", "2024-04-10_23-54", "2024-04-09_23-54", "2024-04-08_23-54", "2024-04-07_23-54",
			"2024-04-06_23-54", "2024-04-05_23-54", "2024-04-04_23-54", "2024-04-03_23-54", "2024-04-02_23-54", "2024-04-01_23-54",
			"2024-03-31_23-54", "2024-03-30_23-54", "2024-03-29_23-54", "2024-03-28_23-54", "2024-03-27_23-54", "2024-03-26_23-54",
		},
		filter_dates: []string{
			"2024-04-30", "2024-04-29", "2024-04-28", "2024-04-27", "2024-04-26", "2024-04-25",
			"2024-04-24", "2024-04-23", "2024-04-22", "2024-04-21", "2024-04-20", "2024-04-19",
			"2024-04-18", "2024-04-17", "2024-04-16", "2024-04-15", "2024-04-14", "2024-04-13",
			"2024-04-12", "2024-04-11", "2024-04-10", "2024-04-09", "2024-04-08", "2024-04-07",
			"2024-04-06", "2024-04-05", "2024-04-04", "2024-04-03", "2024-04-02", "2024-04-01",
		},
	},
	{
		name:          "Test Case 5b - 30th of the month, has >=30 days (only 1 month coverage), none existing",
		test_time:     time.Date(2024, 4, 30, 23, 54, 21, 0, time.UTC),
		next_month:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{},
		filter_dates: []string{
			"2024-04-30", "2024-04-29", "2024-04-28", "2024-04-27", "2024-04-26", "2024-04-25",
			"2024-04-24", "2024-04-23", "2024-04-22", "2024-04-21", "2024-04-20", "2024-04-19",
			"2024-04-18", "2024-04-17", "2024-04-16", "2024-04-15", "2024-04-14", "2024-04-13",
			"2024-04-12", "2024-04-11", "2024-04-10", "2024-04-09", "2024-04-08", "2024-04-07",
			"2024-04-06", "2024-04-05", "2024-04-04", "2024-04-03", "2024-04-02", "2024-04-01",
		},
	},
	{
		name:       "Test Case 6a - 31st of the month, all existing",
		test_time:  time.Date(2024, 5, 31, 23, 54, 21, 0, time.UTC),
		next_month: time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{
			"2024-05-31_23-54", "2024-05-30_23-54", "2024-05-29_23-54", "2024-05-28_23-54", "2024-05-27_23-54", "2024-05-26_23-54",
			"2024-05-25_23-54", "2024-05-24_23-54", "2024-05-23_23-54", "2024-05-22_23-54", "2024-05-21_23-54", "2024-05-20_23-54",
			"2024-05-19_23-54", "2024-05-18_23-54", "2024-05-17_23-54", "2024-05-16_23-54", "2024-05-15_23-54", "2024-05-14_23-54",
			"2024-05-13_23-54", "2024-05-12_23-54", "2024-05-11_23-54", "2024-05-10_23-54", "2024-05-09_23-54", "2024-05-08_23-54",
			"2024-05-07_23-54", "2024-05-06_23-54", "2024-05-05_23-54", "2024-05-04_23-54", "2024-05-03_23-54", "2024-05-02_23-54",
			"2024-05-01_23-54", "2024-04-30_23-54", "2024-04-29_23-54", "2024-04-28_23-54", "2024-04-27_23-54", "2024-04-26_23-54",
		},
		filter_dates: []string{
			"2024-05-31", "2024-05-30", "2024-05-29", "2024-05-28", "2024-05-27", "2024-05-26",
			"2024-05-25", "2024-05-24", "2024-05-23", "2024-05-22", "2024-05-21", "2024-05-20",
			"2024-05-19", "2024-05-18", "2024-05-17", "2024-05-16", "2024-05-15", "2024-05-14",
			"2024-05-13", "2024-05-12", "2024-05-11", "2024-05-10", "2024-05-09", "2024-05-08",
			"2024-05-07", "2024-05-06", "2024-05-05", "2024-05-04", "2024-05-03", "2024-05-02",
		},
	},
	{
		name:          "Test Case 6b - 31st of the month, none existing",
		test_time:     time.Date(2024, 5, 31, 23, 54, 21, 0, time.UTC),
		next_month:    time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
		existing_dirs: []string{},
		filter_dates: []string{
			"2024-05",
		},
	},
}

func Test_getFiltersForHourlys(t *testing.T) {
	for _, tt := range testsForHourlys {
		t.Run(tt.name, func(t *testing.T) {
			expected := append([]string{}, tt.filter_dates_today...)
			expected = append(expected, tt.filter_dates_yesterday...)
			got := getFiltersForHourlys(tt.test_time, tt.existing_dirs)
			if !reflect.DeepEqual(got, expected) {
				compareArrays(got, expected, t)
				t.Errorf("getFiltersForHourlys() result not as expected!")
			}
		})
	}
}

func Test_getFiltersForToday(t *testing.T) {
	for _, tt := range testsForHourlys {
		t.Run(tt.name, func(t *testing.T) {
			got := getFiltersForToday(tt.test_time)
			if !reflect.DeepEqual(got, tt.filter_dates_today) {
				compareArrays(got, tt.filter_dates_today, t)
				t.Errorf("getFiltersForToday() result not as expected!")
			}
		})
	}
}

func Test_getFiltersForYesterday(t *testing.T) {
	for _, tt := range testsForHourlys {
		t.Run(tt.name, func(t *testing.T) {
			remaining := 24 - len(tt.filter_dates_today)
			got := getFiltersForYesterday(tt.test_time, remaining, tt.existing_dirs)
			if !reflect.DeepEqual(got, tt.filter_dates_yesterday) {
				compareArrays(got, tt.filter_dates_today, t)
				t.Errorf("getFiltersForToday() result not as expected!")
			}
		})
	}
}

var testsForHourlys = []struct {
	name                   string
	test_time              time.Time
	existing_dirs          []string
	extra_daily_needed     bool
	extra_daily            string
	filter_dates_today     []string
	filter_dates_yesterday []string
}{
	{
		name:      "Test Case 1 - all and more, with Feb in leap year",
		test_time: time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			"2024-03-01_20-13", "2024-03-01_19-13", "2024-03-01_18-13", "2024-03-01_17-13", "2024-03-01_16-13", "2024-03-01_15-13" /* extra: */, "2024-03-01_15-03",
			"2024-03-01_14-13", "2024-03-01_13-13", "2024-03-01_12-13", "2024-03-01_11-13", "2024-03-01_10-13", "2024-03-01_09-13",
			"2024-03-01_08-13", "2024-03-01_07-13", "2024-03-01_06-13", "2024-03-01_05-13", "2024-03-01_04-13", "2024-03-01_03-13",
			"2024-03-01_02-13", "2024-03-01_01-13", "2024-03-01_00-13", "2024-02-29_23-13", "2024-02-29_22-13", "2024-02-29_21-13",
			// extra:
			"2024-02-29_20-13", "2024-02-29_19-13",
		},
		extra_daily_needed: false,
		extra_daily:        "2024-02-29",
		filter_dates_today: []string{
			"2024-03-01_20", "2024-03-01_19", "2024-03-01_18", "2024-03-01_17", "2024-03-01_16", "2024-03-01_15",
			"2024-03-01_14", "2024-03-01_13", "2024-03-01_12", "2024-03-01_11", "2024-03-01_10", "2024-03-01_09",
			"2024-03-01_08", "2024-03-01_07", "2024-03-01_06", "2024-03-01_05", "2024-03-01_04", "2024-03-01_03",
			"2024-03-01_02", "2024-03-01_01", "2024-03-01_00", /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/
		},
		filter_dates_yesterday: []string{
			/*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ "2024-02-29_23", "2024-02-29_22", "2024-02-29_21",
		},
	},
	{
		name:               "Test Case 2 - no existing dirs, with Feb in leap year",
		test_time:          time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC),
		existing_dirs:      []string{},
		extra_daily_needed: true,
		extra_daily:        "2024-02-29",
		filter_dates_today: []string{
			"2024-03-01_20", "2024-03-01_19", "2024-03-01_18", "2024-03-01_17", "2024-03-01_16", "2024-03-01_15",
			"2024-03-01_14", "2024-03-01_13", "2024-03-01_12", "2024-03-01_11", "2024-03-01_10", "2024-03-01_09",
			"2024-03-01_08", "2024-03-01_07", "2024-03-01_06", "2024-03-01_05", "2024-03-01_04", "2024-03-01_03",
			"2024-03-01_02", "2024-03-01_01", "2024-03-01_00", /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/
		},
		filter_dates_yesterday: []string{
			"2024-02-29",
		},
	},
	{
		name:      "Test Case 3 - sparse, one hit yesterday, with Feb in leap year",
		test_time: time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			/*XXXXXXXXXXXXXXX*/ "2024-03-01_19-13", "2024-03-01_18-13" /*XXXXXXXXXXXXXXX*/, "2024-03-01_16-13", /*XXXXXXXXXXXXXXX*/
			"2024-03-01_14-13" /*XXXXXXXXXXXXXXX*/, "2024-03-01_12-13", /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/
			/*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ "2024-03-01_05-13", "2024-03-01_04-13", /*XXXXXXXXXXXXXXX*/
			"2024-03-01_02-13", "2024-03-01_01-13" /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/, "2024-02-29_21-13",
			// extra:
			/*XXXXXXXXXXXXXXX*/ "2024-02-29_19-13",
		},
		extra_daily_needed: false,
		extra_daily:        "2024-02-29",
		filter_dates_today: []string{
			"2024-03-01_20", "2024-03-01_19", "2024-03-01_18", "2024-03-01_17", "2024-03-01_16", "2024-03-01_15",
			"2024-03-01_14", "2024-03-01_13", "2024-03-01_12", "2024-03-01_11", "2024-03-01_10", "2024-03-01_09",
			"2024-03-01_08", "2024-03-01_07", "2024-03-01_06", "2024-03-01_05", "2024-03-01_04", "2024-03-01_03",
			"2024-03-01_02", "2024-03-01_01", "2024-03-01_00", /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/
		},
		filter_dates_yesterday: []string{
			/*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ "2024-02-29_23", "2024-02-29_22", "2024-02-29_21",
		},
	},
	{
		name:      "Test Case 4 - sparse, no hit yesterday, with Feb in leap year",
		test_time: time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			/*XXXXXXXXXXXXXXX*/ "2024-03-01_19-13", "2024-03-01_18-13" /*XXXXXXXXXXXXXXX*/, "2024-03-01_16-13", /*XXXXXXXXXXXXXXX*/
			"2024-03-01_14-13" /*XXXXXXXXXXXXXXX*/, "2024-03-01_12-13", /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/
			/*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ "2024-03-01_05-13", "2024-03-01_04-13", /*XXXXXXXXXXXXXXX*/
			"2024-03-01_02-13", "2024-03-01_01-13", /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/ /*XXXXXXXXXXXXXXX*/
			// extra:
			/*XXXXXXXXXXXXXXX*/ "2024-02-29_19-13",
		},
		extra_daily_needed: true,
		extra_daily:        "2024-02-29",
		filter_dates_today: []string{
			"2024-03-01_20", "2024-03-01_19", "2024-03-01_18", "2024-03-01_17", "2024-03-01_16", "2024-03-01_15",
			"2024-03-01_14", "2024-03-01_13", "2024-03-01_12", "2024-03-01_11", "2024-03-01_10", "2024-03-01_09",
			"2024-03-01_08", "2024-03-01_07", "2024-03-01_06", "2024-03-01_05", "2024-03-01_04", "2024-03-01_03",
			"2024-03-01_02", "2024-03-01_01", "2024-03-01_00", /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/ /*XXXXXXXXXXXX*/
		},
		filter_dates_yesterday: []string{
			"2024-02-29",
		},
	},
	{
		name:      "Test Case 5 - 24 on a day, with Feb in leap year",
		test_time: time.Date(2024, 3, 1, 23, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			"2024-03-01_23-13", "2024-03-01_22-13", "2024-03-01_21-13", "2024-03-01_20-13", "2024-03-01_19-13", "2024-03-01_18-13",
			"2024-03-01_17-13", "2024-03-01_16-13", "2024-03-01_15-13", "2024-03-01_14-13", "2024-03-01_13-13", "2024-03-01_12-13",
			"2024-03-01_11-13", "2024-03-01_10-13", "2024-03-01_09-13", "2024-03-01_08-13", "2024-03-01_07-13", "2024-03-01_06-13",
			"2024-03-01_05-13", "2024-03-01_04-13", "2024-03-01_03-13", "2024-03-01_02-13", "2024-03-01_01-13", "2024-03-01_00-13",
			// extra:
			"2024-02-29_23-13", "2024-02-29_22-13",
		},
		extra_daily_needed: true,
		extra_daily:        "2024-02-29",
		filter_dates_today: []string{

			"2024-03-01_23", "2024-03-01_22", "2024-03-01_21", "2024-03-01_20", "2024-03-01_19", "2024-03-01_18",
			"2024-03-01_17", "2024-03-01_16", "2024-03-01_15", "2024-03-01_14", "2024-03-01_13", "2024-03-01_12",
			"2024-03-01_11", "2024-03-01_10", "2024-03-01_09", "2024-03-01_08", "2024-03-01_07", "2024-03-01_06",
			"2024-03-01_05", "2024-03-01_04", "2024-03-01_03", "2024-03-01_02", "2024-03-01_01", "2024-03-01_00",
		},
		filter_dates_yesterday: []string{
			"2024-02-29",
		},
	},
	{
		name:      "Test Case 6 - 00 o'clock",
		test_time: time.Date(2024, 3, 2, 0, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			"2024-03-02_00-13",
			"2024-03-01_23-13", "2024-03-01_22-13", "2024-03-01_21-13", "2024-03-01_20-13", "2024-03-01_19-13", "2024-03-01_18-13",
			"2024-03-01_17-13", "2024-03-01_16-13", "2024-03-01_15-13", "2024-03-01_14-13", "2024-03-01_13-13", "2024-03-01_12-13",
			"2024-03-01_11-13", "2024-03-01_10-13", "2024-03-01_09-13", "2024-03-01_08-13", "2024-03-01_07-13", "2024-03-01_06-13",
			"2024-03-01_05-13", "2024-03-01_04-13", "2024-03-01_03-13", "2024-03-01_02-13", "2024-03-01_01-13", "2024-03-01_00-13",
			// extra:
			"2024-02-29_23-13", "2024-02-29_22-13",
		},
		extra_daily_needed: false,
		extra_daily:        "2024-03-01",
		filter_dates_today: []string{
			"2024-03-02_00",
		},
		filter_dates_yesterday: []string{
			"2024-03-01_23", "2024-03-01_22", "2024-03-01_21", "2024-03-01_20", "2024-03-01_19", "2024-03-01_18",
			"2024-03-01_17", "2024-03-01_16", "2024-03-01_15", "2024-03-01_14", "2024-03-01_13", "2024-03-01_12",
			"2024-03-01_11", "2024-03-01_10", "2024-03-01_09", "2024-03-01_08", "2024-03-01_07", "2024-03-01_06",
			"2024-03-01_05", "2024-03-01_04", "2024-03-01_03", "2024-03-01_02", "2024-03-01_01",
		},
	},
}

var testsForAllFilters = []struct {
	name             string
	test_time        time.Time
	existing_dirs    []string
	expected_filters []string
}{
	{
		name:      "Test Case 1 - all there",
		test_time: time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			// 24 hourlys
			"2024-03-01_20-13", "2024-03-01_19-13", "2024-03-01_18-13", "2024-03-01_17-13", "2024-03-01_16-13", "2024-03-01_15-13",
			"2024-03-01_14-13", "2024-03-01_13-13", "2024-03-01_12-13", "2024-03-01_11-13", "2024-03-01_10-13", "2024-03-01_09-13",
			"2024-03-01_08-13", "2024-03-01_07-13", "2024-03-01_06-13", "2024-03-01_05-13", "2024-03-01_04-13", "2024-03-01_03-13",
			"2024-03-01_02-13", "2024-03-01_01-13", "2024-03-01_00-13", "2024-02-29_23-13", "2024-02-29_22-13", "2024-02-29_21-13",
			// 32 dailys
			"2024-02-28_23-13", "2024-02-27_23-13", "2024-02-26_23-13", "2024-02-25_23-13", "2024-02-24_23-13", "2024-02-23_23-13",
			"2024-02-22_23-13", "2024-02-21_23-13", "2024-02-20_23-13", "2024-02-19_23-13", "2024-02-18_23-13", "2024-02-17_23-13",
			"2024-02-16_23-13", "2024-02-15_23-13", "2024-02-14_23-13", "2024-02-13_23-13", "2024-02-12_23-13", "2024-02-11_23-13",
			"2024-02-10_23-13", "2024-02-09_23-13", "2024-02-08_23-13", "2024-02-07_23-13", "2024-02-06_23-13", "2024-02-05_23-13",
			"2024-02-04_23-13", "2024-02-03_23-13", "2024-02-02_23-13", "2024-02-01_23-13", "2024-01-31_23-13", "2024-01-30_23-13",
			"2024-01-29_23-13", "2024-01-28_23-13",
			// 6 monthlys
			"2023-12-28_23-13", "2023-11-28_23-13", "2023-10-28_23-13", "2023-09-28_23-13", "2023-08-28_23-13", "2023-07-28_23-13",
		},
		expected_filters: []string{
			// 24 for the hours
			"2024-03-01_20", "2024-03-01_19", "2024-03-01_18", "2024-03-01_17", "2024-03-01_16", "2024-03-01_15",
			"2024-03-01_14", "2024-03-01_13", "2024-03-01_12", "2024-03-01_11", "2024-03-01_10", "2024-03-01_09",
			"2024-03-01_08", "2024-03-01_07", "2024-03-01_06", "2024-03-01_05", "2024-03-01_04", "2024-03-01_03",
			"2024-03-01_02", "2024-03-01_01", "2024-03-01_00", "2024-02-29_23", "2024-02-29_22", "2024-02-29_21",
			// 30 for the days
			"2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25", "2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21", "2024-02-20", "2024-02-19",
			"2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13", "2024-02-12", "2024-02-11", "2024-02-10", "2024-02-09",
			"2024-02-08", "2024-02-07", "2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01", "2024-01-31", "2024-01-30",
			// 119 for the normal monthlys
			"2023-12", "2023-11", "2023-10", "2023-09", "2023-08", "2023-07", "2023-06", "2023-05", "2023-04", "2023-03", "2023-02", "2023-01",
			"2022-12", "2022-11", "2022-10", "2022-09", "2022-08", "2022-07", "2022-06", "2022-05", "2022-04", "2022-03", "2022-02", "2022-01",
			"2021-12", "2021-11", "2021-10", "2021-09", "2021-08", "2021-07", "2021-06", "2021-05", "2021-04", "2021-03", "2021-02", "2021-01",
			"2020-12", "2020-11", "2020-10", "2020-09", "2020-08", "2020-07", "2020-06", "2020-05", "2020-04", "2020-03", "2020-02", "2020-01",
			"2019-12", "2019-11", "2019-10", "2019-09", "2019-08", "2019-07", "2019-06", "2019-05", "2019-04", "2019-03", "2019-02", "2019-01",
			"2018-12", "2018-11", "2018-10", "2018-09", "2018-08", "2018-07", "2018-06", "2018-05", "2018-04", "2018-03", "2018-02", "2018-01",
			"2017-12", "2017-11", "2017-10", "2017-09", "2017-08", "2017-07", "2017-06", "2017-05", "2017-04", "2017-03", "2017-02", "2017-01",
			"2016-12", "2016-11", "2016-10", "2016-09", "2016-08", "2016-07", "2016-06", "2016-05", "2016-04", "2016-03", "2016-02", "2016-01",
			"2015-12", "2015-11", "2015-10", "2015-09", "2015-08", "2015-07", "2015-06", "2015-05", "2015-04", "2015-03", "2015-02", "2015-01",
			"2014-12", "2014-11", "2014-10", "2014-09", "2014-08", "2014-07", "2014-06", "2014-05", "2014-04", "2014-03", "2014-02",
		},
	},
	{
		name:      "Test Case 2 - no Hourlys for the 29th of February, no January Dailys",
		test_time: time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			// hourlys - none for the 29th
			"2024-03-01_20-13", "2024-03-01_19-13", "2024-03-01_18-13", "2024-03-01_17-13", "2024-03-01_16-13", "2024-03-01_15-13",
			"2024-03-01_14-13", "2024-03-01_13-13", "2024-03-01_12-13", "2024-03-01_11-13", "2024-03-01_10-13", "2024-03-01_09-13",
			"2024-03-01_08-13", "2024-03-01_07-13", "2024-03-01_06-13", "2024-03-01_05-13", "2024-03-01_04-13", "2024-03-01_03-13",
			"2024-03-01_02-13", "2024-03-01_01-13", "2024-03-01_00-13",
			// dailys - none in January but 1.1.
			"2024-02-28_23-13", "2024-02-27_23-13", "2024-02-26_23-13", "2024-02-25_23-13", "2024-02-24_23-13", "2024-02-23_23-13",
			"2024-02-22_23-13", "2024-02-21_23-13", "2024-02-20_23-13", "2024-02-19_23-13", "2024-02-18_23-13", "2024-02-17_23-13",
			"2024-02-16_23-13", "2024-02-15_23-13", "2024-02-14_23-13", "2024-02-13_23-13", "2024-02-12_23-13", "2024-02-11_23-13",
			"2024-02-10_23-13", "2024-02-09_23-13", "2024-02-08_23-13", "2024-02-07_23-13", "2024-02-06_23-13", "2024-02-05_23-13",
			"2024-02-04_23-13", "2024-02-03_23-13", "2024-02-02_23-13", "2024-02-01_23-13",
			"2024-01-01_23-13",
			// 6 monthlys
			"2023-12-28_23-13", "2023-11-28_23-13", "2023-10-28_23-13", "2023-09-28_23-13", "2023-08-28_23-13", "2023-07-28_23-13",
		},
		expected_filters: []string{
			// 21 Hourlys for the 3.1.
			"2024-03-01_20", "2024-03-01_19", "2024-03-01_18", "2024-03-01_17", "2024-03-01_16", "2024-03-01_15",
			"2024-03-01_14", "2024-03-01_13", "2024-03-01_12", "2024-03-01_11", "2024-03-01_10", "2024-03-01_09",
			"2024-03-01_08", "2024-03-01_07", "2024-03-01_06", "2024-03-01_05", "2024-03-01_04", "2024-03-01_03",
			"2024-03-01_02", "2024-03-01_01", "2024-03-01_00",
			// 1 Daily for the 29th
			"2024-02-29",
			// 28 for the days
			"2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25", "2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21", "2024-02-20", "2024-02-19",
			"2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13", "2024-02-12", "2024-02-11", "2024-02-10", "2024-02-09",
			"2024-02-08", "2024-02-07", "2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01",
			// 1 monthly for January
			"2024-01",
			// 119 for the normal monthlys
			"2023-12", "2023-11", "2023-10", "2023-09", "2023-08", "2023-07", "2023-06", "2023-05", "2023-04", "2023-03", "2023-02", "2023-01",
			"2022-12", "2022-11", "2022-10", "2022-09", "2022-08", "2022-07", "2022-06", "2022-05", "2022-04", "2022-03", "2022-02", "2022-01",
			"2021-12", "2021-11", "2021-10", "2021-09", "2021-08", "2021-07", "2021-06", "2021-05", "2021-04", "2021-03", "2021-02", "2021-01",
			"2020-12", "2020-11", "2020-10", "2020-09", "2020-08", "2020-07", "2020-06", "2020-05", "2020-04", "2020-03", "2020-02", "2020-01",
			"2019-12", "2019-11", "2019-10", "2019-09", "2019-08", "2019-07", "2019-06", "2019-05", "2019-04", "2019-03", "2019-02", "2019-01",
			"2018-12", "2018-11", "2018-10", "2018-09", "2018-08", "2018-07", "2018-06", "2018-05", "2018-04", "2018-03", "2018-02", "2018-01",
			"2017-12", "2017-11", "2017-10", "2017-09", "2017-08", "2017-07", "2017-06", "2017-05", "2017-04", "2017-03", "2017-02", "2017-01",
			"2016-12", "2016-11", "2016-10", "2016-09", "2016-08", "2016-07", "2016-06", "2016-05", "2016-04", "2016-03", "2016-02", "2016-01",
			"2015-12", "2015-11", "2015-10", "2015-09", "2015-08", "2015-07", "2015-06", "2015-05", "2015-04", "2015-03", "2015-02", "2015-01",
			"2014-12", "2014-11", "2014-10", "2014-09", "2014-08", "2014-07", "2014-06", "2014-05", "2014-04", "2014-03", "2014-02",
		},
	},
}

func Test_getAllFilters(t *testing.T) {
	for _, tt := range testsForAllFilters {
		t.Run(tt.name, func(t *testing.T) {
			got := getAllFilters(tt.test_time, tt.existing_dirs)
			if !reflect.DeepEqual(got, tt.expected_filters) {
				compareArrays(got, tt.expected_filters, t)
				t.Errorf("getAllFilters() result not as expected!")
			}
		})
	}
}

func Test_getDateDirectoriesNotMatchingAnyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		allDirs  []string
		prefixes []string
		want     []string
	}{
		{
			name:     "Test 1: No directory matches the prefix",
			allDirs:  []string{"2024-06-16test", "2024-06-17test", "2024-06-18test", "to_delete"},
			prefixes: []string{"2025"},
			want:     []string{"2024-06-16test", "2024-06-17test", "2024-06-18test"},
		},
		{
			name:     "Test 2: Some directories match the prefix",
			allDirs:  []string{"2024-06-16test", "2024-06-17test", "2024-06-18test"},
			prefixes: []string{"2024-06-16"},
			want:     []string{"2024-06-17test", "2024-06-18test"},
		},
		{
			name:     "Test 3: All directories match the prefix",
			allDirs:  []string{"2024-06-16test", "2024-06-17test", "2024-06-18test"},
			prefixes: []string{"2024"},
			want:     []string{},
		},
		{
			name:     "Test 4: No directory matches any prefix",
			allDirs:  []string{"2024-06-16test", "2024-06-17test", "2024-06-18test", "to_delete"},
			prefixes: []string{"2025", "2027"},
			want:     []string{"2024-06-16test", "2024-06-17test", "2024-06-18test"},
		},
		{
			name:     "Test 5: Some directories match some prefixes",
			allDirs:  []string{"2024-06-16test", "2024-06-17test", "2024-06-18test", "hello"},
			prefixes: []string{"2027", "2024-06-16", "2024-06-18"},
			want:     []string{"2024-06-17test"},
		},
		{
			name:     "Test 6: All directories match a prefix",
			allDirs:  []string{"2024-06-16test", "2024-06-17test", "2024-06-18test"},
			prefixes: []string{"2024-06-18", "2027", "2024-06-16", "2024-06-17"},
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDateDirectoriesNotMatchingAnyPrefix(tt.allDirs, tt.prefixes, 0); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDateDirectoriesNotMatchingAnyPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDateDirectoriesNotMatchingAnyPrefix_Verbosity(t *testing.T) {
	// Save the original stdout
	originalStdout := os.Stdout

	// Create a pipe to capture the output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// no output for verbosity 1
	getDateDirectoriesNotMatchingAnyPrefix([]string{"a", "b", "c"}, []string{}, 1)

	// output for verbosity 2
	getDateDirectoriesNotMatchingAnyPrefix([]string{"1", "2", "3"}, []string{}, 2)

	// Close the writer and restore the original stdout
	w.Close()
	os.Stdout = originalStdout

	// Read the captured output
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Errorf("Error reading back from stdout: %v", err)
	}
	capturedOutput := buf.String()

	// Check if the output is as expected
	expectedOutput := "Skipping 1 as it is not in date format.\nSkipping 2 as it is not in date format.\nSkipping 3 as it is not in date format.\n"
	if capturedOutput != expectedOutput {
		t.Errorf("Expected %q but got %q", expectedOutput, capturedOutput)
	}
}

func Test_printNiceNumbr(t *testing.T) {
	// Save the original stdout
	originalStdout := os.Stdout

	// Create a pipe to capture the output
	r, w, _ := os.Pipe()
	os.Stdout = w

	printNiceNumbr("a", 1)
	printNiceNumbr("b", 10)
	printNiceNumbr("c", 100)
	printNiceNumbr("d", 1000)
	printNiceNumbr("e", 10000)
	printNiceNumbr("f", 100000)
	printNiceNumbr("g", 1000000)

	// Close the writer and restore the original stdout
	w.Close()
	os.Stdout = originalStdout

	// Read the captured output
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Errorf("Error reading back from stdout: %v", err)
	}
	capturedOutput := buf.String()

	// Check if the output is as expected
	expectedOutput := "a : 1\nb : 10\nc : 100\nd : 1000 (i.e. 1.0 k)\ne : 10000 (i.e. 10.0 k)\nf : 100000 (i.e. 100.0 k)\ng : 1000000 (i.e. 1.0 M)\n"
	if capturedOutput != expectedOutput {
		t.Errorf("Expected %q but got %q", expectedOutput, capturedOutput)
	}
}

func Test_printNiceBytes(t *testing.T) {
	// Save the original stdout
	originalStdout := os.Stdout

	// Create a pipe to capture the output
	r, w, _ := os.Pipe()
	os.Stdout = w

	printNiceBytes("a", 1)
	printNiceBytes("b", 10)
	printNiceBytes("c", 100)
	printNiceBytes("d", 1000)
	printNiceBytes("e", 10000)
	printNiceBytes("f", 100000)
	printNiceBytes("g", 1000000)

	// Close the writer and restore the original stdout
	w.Close()
	os.Stdout = originalStdout

	// Read the captured output
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Errorf("Error reading back from stdout: %v", err)
	}
	capturedOutput := buf.String()

	// Check if the output is as expected
	expectedOutput := "a : 1 Bytes\nb : 10 Bytes\nc : 100 Bytes\nd : 1000 Bytes (i.e. 1.0 kBytes)\ne : 10000 Bytes (i.e. 10.0 kBytes)\nf : 100000 Bytes (i.e. 100.0 kBytes)\ng : 1000000 Bytes (i.e. 1.0 MBytes)\n"
	if capturedOutput != expectedOutput {
		t.Errorf("Expected %q but got %q", expectedOutput, capturedOutput)
	}
}
