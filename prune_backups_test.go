package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

/*
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
*/

func compareArrays(result []string, want []string, t *testing.T) {
	max := len(result)
	if len(want) > max {
		max = len(want)
	}
	for i := 0; i < max; i++ {
		if i < len(want) && i < len(result) {
			t.Logf("   wanted: " + want[i] + ", got: " + result[i])
		} else if i < len(want) {
			t.Logf("   wanted: " + want[i] + ", got: <no more values>")
		} else {
			t.Logf("   wanted: <no more values>, got: " + result[i])
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

const USE_DEFAULT_DIRECTORY_FOR_TEMP_FILES = "" // see https://pkg.go.dev/os#MkdirTemp

func generateHourlyTestDirectories(t *testing.T, test_time time.Time, number int) string {

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

/*
func Test_getFiltersFor30Dailys(t *testing.T) {
	for _, tt := range testsFor30Dailys {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFiltersFor30Dailys(tt.test_time); !reflect.DeepEqual(got, tt.filter_dates) {
				compareArrays(got, tt.filter_dates, t)
				t.Errorf("getFiltersFor30Dailys() result not as expected!")
			}
		})
	}
}
*/
/*
func Test_getMonthToLookForAnExtraMonthly(t *testing.T) {
	for _, tt := range testsFor30Dailys {
		t.Run(tt.name, func(t *testing.T) {
			got_needed, got_date := getMonthToLookForAnExtraMonthly(tt.test_time)
			if (got_needed != tt.extra_monthly_needed) || (got_date != tt.extra_monthly_date) {
				t.Errorf("Expected needed: %v, got needed: %v, expected date: %v, got date: %v", tt.extra_monthly_needed, got_needed, tt.extra_monthly_date, got_date)
			}
		})
	}
}
*/

var testsFor30Dailys = []struct {
	name                 string
	test_time            time.Time
	extra_monthly_needed bool
	extra_monthly_date   time.Time
	filter_dates         []string
}{
	{
		name:                 "Test Case 1 - middle of the month",
		test_time:            time.Date(2014, 7, 17, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2014, 6, 15, 0, 0, 0, 0, time.UTC),
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
		name:                 "Test Case 2 - today and yesterday in one month, the other 30 days in the month before",
		test_time:            time.Date(2014, 7, 2, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: false,
		extra_monthly_date:   time.Date(2014, 6, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			"2014-06-30", "2014-06-29", "2014-06-28", "2014-06-27", "2014-06-26", "2014-06-25", "2014-06-24", "2014-06-23", "2014-06-22", "2014-06-21",
			"2014-06-20", "2014-06-19", "2014-06-18", "2014-06-17", "2014-06-16", "2014-06-15", "2014-06-14", "2014-06-13", "2014-06-12", "2014-06-11",
			"2014-06-10", "2014-06-09", "2014-06-08", "2014-06-07", "2014-06-06", "2014-06-05", "2014-06-04", "2014-06-03", "2014-06-02", "2014-06-01",
		},
	},
	{
		name:                 "Test Case 3 - today in one month, yesterday and the other 30 days in a 31-day month before",
		test_time:            time.Date(2014, 6, 1, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: false,
		extra_monthly_date:   time.Date(2014, 5, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			"2014-05-30", "2014-05-29", "2014-05-28", "2014-05-27", "2014-05-26", "2014-05-25", "2014-05-24", "2014-05-23", "2014-05-22", "2014-05-21",
			"2014-05-20", "2014-05-19", "2014-05-18", "2014-05-17", "2014-05-16", "2014-05-15", "2014-05-14", "2014-05-13", "2014-05-12", "2014-05-11",
			"2014-05-10", "2014-05-09", "2014-05-08", "2014-05-07", "2014-05-06", "2014-05-05", "2014-05-04", "2014-05-03", "2014-05-02", "2014-05-01",
		},
	},
	{
		name:                 "Test Case 4 - today in one month, yesterday and 29 days in a 30-day month before, and 1 day in the month before that",
		test_time:            time.Date(2014, 7, 1, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2014, 5, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ "2014-06-29", "2014-06-28", "2014-06-27", "2014-06-26", "2014-06-25", "2014-06-24", "2014-06-23", "2014-06-22", "2014-06-21",
			"2014-06-20", "2014-06-19", "2014-06-18", "2014-06-17", "2014-06-16", "2014-06-15", "2014-06-14", "2014-06-13", "2014-06-12", "2014-06-11",
			"2014-06-10", "2014-06-09", "2014-06-08", "2014-06-07", "2014-06-06", "2014-06-05", "2014-06-04", "2014-06-03", "2014-06-02", "2014-06-01",
			// one day in May
			"2014-05-31",
		},
	},
	{
		name:                 "Test Case 5 - today and yesterday in a 30-day month, 28 days in the rest of the month, and 2 day in the month before that",
		test_time:            time.Date(2014, 6, 30, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2014, 5, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ "2014-06-28", "2014-06-27", "2014-06-26", "2014-06-25", "2014-06-24", "2014-06-23", "2014-06-22", "2014-06-21",
			"2014-06-20", "2014-06-19", "2014-06-18", "2014-06-17", "2014-06-16", "2014-06-15", "2014-06-14", "2014-06-13", "2014-06-12", "2014-06-11",
			"2014-06-10", "2014-06-09", "2014-06-08", "2014-06-07", "2014-06-06", "2014-06-05", "2014-06-04", "2014-06-03", "2014-06-02", "2014-06-01",
			// two days in May
			"2014-05-31", "2014-05-30",
		},
	},
	{
		name:                 "Test Case 6 - today and yesterday in a 29-day month, 27 days in the rest of the month, and 3 day in the month before that",
		test_time:            time.Date(2024, 2, 29, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ "2024-02-27", "2024-02-26", "2024-02-25", "2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21",
			"2024-02-20", "2024-02-19", "2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13", "2024-02-12", "2024-02-11",
			"2024-02-10", "2024-02-09", "2024-02-08", "2024-02-07", "2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01",
			// three days in January
			"2024-01-31", "2024-01-30", "2024-01-29",
		},
	},
	{
		name:                 "Test Case 7 - today and yesterday in a 28-day month, 26 days in the rest of the month, and 4 day in the month before that",
		test_time:            time.Date(2023, 2, 28, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ "2023-02-26", "2023-02-25", "2023-02-24", "2023-02-23", "2023-02-22", "2023-02-21",
			"2023-02-20", "2023-02-19", "2023-02-18", "2023-02-17", "2023-02-16", "2023-02-15", "2023-02-14", "2023-02-13", "2023-02-12", "2023-02-11",
			"2023-02-10", "2023-02-09", "2023-02-08", "2023-02-07", "2023-02-06", "2023-02-05", "2023-02-04", "2023-02-03", "2023-02-02", "2023-02-01",
			// three days in January
			"2023-01-31", "2023-01-30", "2023-01-29", "2023-01-28",
		},
	},
	{
		name:                 "Test Case 8 - today in a month before a 29-day month, yesterday and 28 days in the rest of the month, and 2 days in the month before that",
		test_time:            time.Date(2024, 3, 1, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ "2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25", "2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21",
			"2024-02-20", "2024-02-19", "2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13", "2024-02-12", "2024-02-11",
			"2024-02-10", "2024-02-09", "2024-02-08", "2024-02-07", "2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01",
			// three days in January
			"2024-01-31", "2024-01-30",
		},
	},
	{
		name:                 "Test Case 9 - today in a month before a 28-day month, yesterday and 27 days in the rest of the month, and 3 days in the month before that",
		test_time:            time.Date(2023, 3, 1, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ /*XXXXXXXXX*/ "2023-02-27", "2023-02-26", "2023-02-25", "2023-02-24", "2023-02-23", "2023-02-22", "2023-02-21",
			"2023-02-20", "2023-02-19", "2023-02-18", "2023-02-17", "2023-02-16", "2023-02-15", "2023-02-14", "2023-02-13", "2023-02-12", "2023-02-11",
			"2023-02-10", "2023-02-09", "2023-02-08", "2023-02-07", "2023-02-06", "2023-02-05", "2023-02-04", "2023-02-03", "2023-02-02", "2023-02-01",
			// three days in January
			"2023-01-31", "2023-01-30", "2023-01-29",
		},
	},

	{
		name:                 "Test Case 10 - today and yesterday in a month before a 29-day month, 29 days in the rest of the month, and 1 days in the month before that",
		test_time:            time.Date(2024, 3, 2, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ "2024-02-29", "2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25", "2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21",
			"2024-02-20", "2024-02-19", "2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13", "2024-02-12", "2024-02-11",
			"2024-02-10", "2024-02-09", "2024-02-08", "2024-02-07", "2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01",
			// three days in January
			"2024-01-31",
		},
	},
	{
		name:                 "Test Case 11 - today and yesterday in a month before a 28-day month, 28 days in the rest of the month, and 2 days in the month before that",
		test_time:            time.Date(2023, 3, 2, 9, 54, 21, 0, time.UTC),
		extra_monthly_needed: true,
		extra_monthly_date:   time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		filter_dates: []string{
			/*XXXXXXXXX*/ /*XXXXXXXXX*/ "2023-02-28", "2023-02-27", "2023-02-26", "2023-02-25", "2023-02-24", "2023-02-23", "2023-02-22", "2023-02-21",
			"2023-02-20", "2023-02-19", "2023-02-18", "2023-02-17", "2023-02-16", "2023-02-15", "2023-02-14", "2023-02-13", "2023-02-12", "2023-02-11",
			"2023-02-10", "2023-02-09", "2023-02-08", "2023-02-07", "2023-02-06", "2023-02-05", "2023-02-04", "2023-02-03", "2023-02-02", "2023-02-01",
			// three days in January
			"2023-01-31", "2023-01-30",
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
			"2024-03-01_20-13", "2024-03-01_19-13", "2024-03-01_18-13", "2024-03-01_17-13", "2024-03-01_16-13", "2024-03-01_15-13",
			"2024-03-01_14-13", "2024-03-01_13-13", "2024-03-01_12-13", "2024-03-01_11-13", "2024-03-01_10-13", "2024-03-01_09-13",
			"2024-03-01_08-13", "2024-03-01_07-13", "2024-03-01_06-13", "2024-03-01_05-13", "2024-03-01_04-13", "2024-03-01_03-13",
			"2024-03-01_02-13", "2024-03-01_01-13", "2024-03-01_00-13", "2024-02-29_23-13", "2024-02-29_22-13", "2024-02-29_21-13",
			"2024-02-28", "2024-02-27", "2024-02-26", "2024-02-25", "2024-02-24", "2024-02-23", "2024-02-22", "2024-02-21", "2024-02-20", "2024-02-19",
			"2024-02-18", "2024-02-17", "2024-02-16", "2024-02-15", "2024-02-14", "2024-02-13", "2024-02-12", "2024-02-11", "2024-02-10", "2024-02-09",
			"2024-02-08", "2024-02-07", "2024-02-06", "2024-02-05", "2024-02-04", "2024-02-03", "2024-02-02", "2024-02-01", "2024-01-31", "2024-01-30",
			"2024-01-29", "2024-01-28",
			"2024-01-02", "2024-01-01",
			"2023-12-31", "2023-12-30",
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
		name:      "Test Case 2 - no January dailys",
		test_time: time.Date(2024, 3, 1, 20, 34, 58, 0, time.UTC),
		existing_dirs: []string{
			"2024-03-01_20-13", "2024-03-01_19-13", "2024-03-01_18-13", "2024-03-01_17-13", "2024-03-01_16-13", "2024-03-01_15-13" /* extra: */, "2024-03-01_15-03",
			"2024-03-01_14-13", "2024-03-01_13-13", "2024-03-01_12-13", "2024-03-01_11-13", "2024-03-01_10-13", "2024-03-01_09-13",
			"2024-03-01_08-13", "2024-03-01_07-13", "2024-03-01_06-13", "2024-03-01_05-13", "2024-03-01_04-13", "2024-03-01_03-13",
			"2024-03-01_02-13", "2024-03-01_01-13", "2024-03-01_00-13", "2024-02-29_23-13", "2024-02-29_22-13", "2024-02-29_21-13",
			// extra:
			"2024-02-29_20-13", "2024-02-29_19-13",
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
			// as we don't have any January directories in the existing_dirs, the code must emit an extra filter to keep ANY January directories, not only those for the 30 days
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
