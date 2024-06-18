package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

func Test_getAllMatchingPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		from   []string
		prefix string
		want   []string
	}{
		{
			name:   "Test no matches",
			from:   []string{"cat", "cap", "car"},
			prefix: "df",
			want:   []string{},
		},
		{
			name:   "Test empty input",
			from:   []string{},
			prefix: "df",
			want:   []string{},
		},
		{
			name:   "Test empty prefix",
			from:   []string{"cat", "cap", "car"},
			prefix: "",
			want:   []string{"cat", "cap", "car"},
		},
		{
			name:   "Test 2 out of 4",
			from:   []string{"apple", "banana", "apricot", "grape"},
			prefix: "ap",
			want:   []string{"apple", "apricot"},
		},
		{
			name:   "Test all 3",
			from:   []string{"dog", "deer", "duck"},
			prefix: "d",
			want:   []string{"dog", "deer", "duck"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getAllMatchingPrefix(tc.from, tc.prefix)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getAllMatchingPrefix() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_getAllButFirstMatchingPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		from   []string
		prefix string
		want   []string
	}{
		{
			name:   "Test no matches",
			from:   []string{"cat", "cap", "car"},
			prefix: "df",
			want:   []string{},
		},
		{
			name:   "Test empty input",
			from:   []string{},
			prefix: "df",
			want:   []string{},
		},
		{
			name:   "Test empty prefix",
			from:   []string{"cat", "cap", "car"},
			prefix: "",
			want:   []string{"cap", "car"},
		},
		{
			name:   "Test 2 out of 4",
			from:   []string{"apple", "banana", "apricot", "grape"},
			prefix: "ap",
			want:   []string{"apricot"},
		},
		{
			name:   "Test all 3",
			from:   []string{"dog", "deer", "duck"},
			prefix: "d",
			want:   []string{"deer", "duck"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getAllButFirstMatchingPrefix(tc.from, tc.prefix)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getAllButFirstMatchingPrefix() = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_DateAdd(t *testing.T) {
	test_time := time.Date(2023, 3, 29, 20, 34, 58, 0, time.UTC)
	expect := time.Date(2023, 3, 1, 20, 34, 58, 0, time.UTC)
	got := test_time.AddDate(0, -1, 0)

	if got != expect {
		t.Errorf("DateAdd-Test: expected=%v, got=%v", expect, got)
	}

	test_time2 := time.Date(2023, 5, 31, 20, 34, 58, 0, time.UTC)
	expect2 := time.Date(2023, 5, 1, 20, 34, 58, 0, time.UTC)
	got2 := test_time2.AddDate(0, -1, 0)

	if got2 != expect2 {
		t.Errorf("DateAdd-Test: expected=%v, got=%v", expect2, got2)
	}
}

func Test_prevMonth(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		year          int
		month         int
		expectedYear  int
		expectedMonth int
	}{
		{"January to December", 2024, 1, 2023, 12},
		{"February to January", 2024, 2, 2024, 1},
		{"December to November", 2024, 12, 2024, 11},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function with the current test case's year and month
			prevMonth(&tc.year, &tc.month)

			// Check if the year and month match the expected values
			if tc.year != tc.expectedYear || tc.month != tc.expectedMonth {
				t.Errorf("For %s, expected year:month to be %d:%d, but got %d:%d", tc.name, tc.expectedYear, tc.expectedMonth, tc.year, tc.month)
			}
		})
	}
}

func TestToDateStr(t *testing.T) {
	tests := []struct {
		year   int
		month  int
		expect string
	}{
		{2024, 1, "2024-01"},
		{1912, 9, "1912-09"},
		{2024, 10, "2024-10"},
		{2024, 12, "2024-12"},
	}

	for _, test := range tests {
		result := toDateStr(test.year, test.month)
		if result != test.expect {
			t.Errorf("toDateStr(%d, %d) = %s; want %s", test.year, test.month, result, test.expect)
		}
	}
}

func Test_createPrefixesForTimeSlotsToKeepOne(t *testing.T) {
	test_time := time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC)
	want := []string{
		// 24 for the hours
		"2024-03-01_20", "2024-03-01_19", "2024-03-01_18", "2024-03-01_17", "2024-03-01_16", "2024-03-01_15",
		"2024-03-01_14", "2024-03-01_13", "2024-03-01_12", "2024-03-01_11", "2024-03-01_10", "2024-03-01_09",
		"2024-03-01_08", "2024-03-01_07", "2024-03-01_06", "2024-03-01_05", "2024-03-01_04", "2024-03-01_03",
		"2024-03-01_02", "2024-03-01_01", "2024-03-01_00", "2024-02-29_23", "2024-02-29_22", "2024-02-29_21",
		// 30 for the days
		"2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25", "2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21", "2024-02-20", "2024-02-19",
		"2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13", "2024-02-12", "2024-02-11", "2024-02-10", "2024-02-09",
		"2024-02-08", "2024-02-07", "2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01", "2024-01-31", "2024-01-30",
		// 119 for the months
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
	}

	result := createPrefixesForTimeSlotsToKeepOne(test_time)

	if !reflect.DeepEqual(result, want) {
		t.Errorf("getAllButFirstMatchingPrefix() result not as expected!")
		compareArrays(result, want, t)
	}

}

func compareArrays(result []string, want []string, t *testing.T) {
	max := len(result)
	if len(want) > max {
		max = len(want)
	}
	for i := 0; i < max; i++ {
		if i < len(want) && i < len(result) {
			t.Errorf("   wanted: " + want[i] + ", got: " + result[i])
		} else if i < len(want) {
			t.Errorf("   wanted: " + want[i] + ", got: <no more values>")
		} else {
			t.Errorf("   wanted: <no more values>, got: " + result[i])
		}
	}
}

func Test_pruneDirectory(t *testing.T) {
	test_time_gen := time.Date(2024, 6, 17, 9, 49, 33, 0, time.UTC)
	test_time_prune := time.Date(2024, 6, 17, 9, 54, 21, 0, time.UTC)

	test_dir := generateTestDirectories(t, test_time_gen, 2800)

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

func pruneAndCheck(t *testing.T, test_dir string, test_time_pruning time.Time, expect_remaining []string, number_expect_deleted int) {
	pruneDirectory(test_dir, test_time_pruning, "to_delete", 0)

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

func generateTestDirectories(t *testing.T, test_time time.Time, number int) string {

	/*
		_, err := os.ReadDir(dirPath)
		if err != nil {
			// The second argument is the permission mode.
			// 0755 commonly used for directories.
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	*/

	const USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES = "" // see https://pkg.go.dev/os#MkdirTemp
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
