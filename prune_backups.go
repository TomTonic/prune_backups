package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func main() {

	/* Processing command line parameters */

	pruneDirName := flag.String("dir", "<none>", "REQUIRED. The name of the directory that shall be pruned.")
	toDeleteDirName := flag.String("to_directory", "to_delete", "OPTIONAL. The name of the directory where the pruned directories shall be moved.")
	showStats := flag.Bool("stats", false, "OPTIONAL. Show total size of linked and unlinked files in the pruned directories if `true`. (default false)") // caution: go will neither print the type nor the default for bool flags with default false. see https://github.com/golang/go/issues/63150
	showVersion := flag.Bool("version", false, "OPTIONAL. Show version/build information and exit if `true`. (default false)")                            // caution: go will neither print the type nor the default for bool flags with default false. see https://github.com/golang/go/issues/63150
	verbosity := flag.Int("v", 1, "OPTIONAL. Set verbosity. O - mute, 1 - some, 2 - a lot.")

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if *showVersion {
		fmt.Println("prune_backups", commitInfo)
		os.Exit(0)
	}

	// workaroud as REQUIRED parameters are not supported by the flag package
	if !isFlagPassed("dir") {
		fmt.Println("Pruning directory missing (-dir).")
		flag.PrintDefaults()
		os.Exit(1)
	}

	now := time.Now()

	pruneDirectory(*pruneDirName, now, *toDeleteDirName, *verbosity, *showStats)
}

func pruneDirectory(pruneDirName string, now time.Time, toDeleteDirName string, verbosity int, showStats bool) {
	files, err := os.ReadDir(pruneDirName)
	if err != nil {
		fmt.Print("Could not read pruning directory (-dir): ")
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	dirs := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	if verbosity > 0 {
		fmt.Println("I found", len(dirs), "directories in", pruneDirName)
	}

	// Sort in descending order - caution: this is important for the algorithm!
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	var to_delete []string // in this array we will collect all directories that we will move to the to_delete-directory

	filters := getAllFilters(now, dirs)
	for _, filter := range filters {
		add_to_delete := getAllButFirstMatchingPrefix(dirs, filter)
		to_delete = append(to_delete, add_to_delete...)
	}

	cleaup_others := getDateDirectoriesNotMatchingAnyPrefix(dirs, filters, verbosity)
	to_delete = append(to_delete, cleaup_others...)

	delPath := filepath.Join(pruneDirName, toDeleteDirName)
	err2 := os.MkdirAll(delPath, 0755)
	if err2 != nil {
		fmt.Print("Error creating directory \"", delPath, "\": ")
		fmt.Println(err)
		if verbosity > 0 {
			fmt.Println("I woud have moved the following directories there:")
			for _, dir := range to_delete {
				fmt.Println(" -", dir)
			}
		}
		os.Exit(1)
	}

	/* now we have collected all directory names that need to be moved in to_delete. next we will create the target directory and actually move them */
	var successful_move_counter = 0
	for _, dirname := range to_delete {
		fromPath := filepath.Join(pruneDirName, dirname)
		toPath := filepath.Join(delPath, dirname)
		if verbosity > 1 {
			fmt.Print("Moving ", fromPath, " to ", toPath, "... ")
		}
		err3 := os.Rename(fromPath, toPath)
		if err3 != nil {
			if verbosity > 1 {
				fmt.Println(err)
			} else {
				fmt.Println("Error moving ", fromPath, " to ", toPath, ": ", err)
			}
		} else {
			if verbosity > 1 {
				fmt.Println("done.")
			}
			successful_move_counter++
		}
	}
	if verbosity > 0 {
		fmt.Println("I moved", successful_move_counter, "directories to", delPath)
	}
	if showStats {
		number_of_unlinked_files, size_of_unlinked_files, number_of_linked_files, size_of_linked_files, number_of_subdirs := du(delPath)
		fmt.Printf("The directory %v now contains %v bytes in %v unlinked files and %v bytes in %v hard-linked files in a total of %v directories.\n", delPath, size_of_unlinked_files, number_of_unlinked_files, size_of_linked_files, number_of_linked_files, number_of_subdirs)
	}
}

func getAllFilters(start_time time.Time, existingDirs []string) []string {
	var result = []string{}

	// append hourly filters

	result = append(result, getFiltersForHourlys(start_time, existingDirs)...)

	// append daily filters

	filters_for_dailys, first_month_for_the_monthlys := getFiltersForDailys(start_time.AddDate(0, 0, -2), existingDirs)
	result = append(result, filters_for_dailys...)

	// append monthly filters

	result = append(result, getFiltersForMonthlys(first_month_for_the_monthlys, 119)...)

	return result
}

func getFiltersForHourlys(start_time time.Time, existingDirs []string) []string {
	var result = []string{}
	filters_today := getFiltersForToday(start_time)
	result = append(result, filters_today...)
	remaining_hourlys := 24 - len(filters_today)
	filters_yesterday := getFiltersForYesterday(start_time, remaining_hourlys, existingDirs)
	result = append(result, filters_yesterday...)
	return result
}

func getFiltersForToday(current_time time.Time) []string {
	var result = []string{}

	day := current_time.Day()

	for current_time.Day() == day {
		// Format the time in the format YYYY-MM-DD_hh
		prefix := current_time.Format("2006-01-02_15") // caution, this is a magic number in go!
		result = append(result, prefix)

		// Subtract one hour from the current timestamp
		current_time = current_time.Add(-1 * time.Hour)
	}
	return result
}

func getFiltersForYesterday(current_time time.Time, remaining_hourly_backups int, allDirs []string) []string {
	var hourlyFilters = []string{}

	yesterday := current_time.Add(-24 * time.Hour)

	var year int = yesterday.Year()
	var month int = (int)(yesterday.Month())
	var day int = yesterday.Day()
	yester_date_str := toDateStr3(year, month, day)

	for i := 0; i < remaining_hourly_backups; i++ {
		prefix := yester_date_str + "_" + twoDigit(23-i)
		hourlyFilters = append(hourlyFilters, prefix)
	}

	any_matches := getAnyMatchingAnyPrefixes(allDirs, hourlyFilters) // check what is actually there - the filter for yesterday will depend on it

	if any_matches {
		// we found some hourly backup folders for yesterday, so return the filter for the hourly backups for yesterday, i.e. some YYYY-MM-DD_HH filters
		return hourlyFilters
	} else {
		// we found at least one hourly backup folders for yesterday, so return the filter for the latest backup for yesterday, i.e. one YYYY-MM-DD filter
		return []string{yester_date_str}
	}
}

func getFiltersForDailys(start_date time.Time, existingDirs []string) ([]string, time.Time) {
	var result = []string{}
	var first_month_for_the_monthlys time.Time
	M1 := get15thOfMonthBefore(start_date)
	M2 := get15thOfMonthBefore(M1)
	M3 := get15thOfMonthBefore(M2)
	daysM0 := daysInMonth(start_date.Year(), start_date.Month())
	daysM1 := daysInMonth(M1.Year(), M1.Month())

	switch start_date.Day() {
	case 1:
		{
			switch daysM1 {
			case 28:
				// The 30 days affect THREE months M0, M1, and M2 and M2 is NOT completely covered with daily backups.
				// pin 29 normal dailys and test 1 daily in in M2. continue with M3
				result = append(result, getFiltersForDailysSimple(start_date, 29)...)
				result = append(result, getFiltersForDailysOrForMonth(getUltimo(M2.Year(), M2.Month()), 1, existingDirs)...)
				first_month_for_the_monthlys = M3
			case 29:
				// The 30 days affect TWO months M0 and M1 and M1 is completely covered with daily backups.
				// pin 30 normal dailys and test nothing. continue with M2
				result = append(result, getFiltersForDailysSimple(start_date, 30)...)
				first_month_for_the_monthlys = M2
			case 30, 31:
				// The 30 days affect TWO months M0 and M1 and M1 is NOT completely covered with daily backups.
				// pin 1 normal daily and test 29 dailys in in M1. continue with M2
				result = append(result, getFiltersForDailysSimple(start_date, 1)...)
				result = append(result, getFiltersForDailysOrForMonth(getUltimo(M1.Year(), M1.Month()), 29, existingDirs)...)
				first_month_for_the_monthlys = M2
			}
		}
	case 2:
		{
			switch daysM1 {
			case 28:
				// The 30 days affect TWO months M0 and M1 and M1 is completely covered with daily backups.
				// pin 30 normal dailys and test nothing. continue with M2
				result = append(result, getFiltersForDailysSimple(start_date, 30)...)
				first_month_for_the_monthlys = M2
			case 29, 30, 31:
				// The 30 days affect TWO months M0 and M1 and M1 is NOT completely covered with daily backups.
				// pin 2 normal dailys and test 28 dailys in in M1. continue with M2
				result = append(result, getFiltersForDailysSimple(start_date, 2)...)
				result = append(result, getFiltersForDailysOrForMonth(getUltimo(M1.Year(), M1.Month()), 28, existingDirs)...)
				first_month_for_the_monthlys = M2
			}
		}
	case 30:
		{
			switch daysM0 {
			case 28, 29:
				// illegal in a month with a 30st day
			case 30:
				// The 30 days affect ONE month M0 and M0 is completely covered with daily backups.
				// pin 30 normal dailys and test nothing. continue with M1
				result = append(result, getFiltersForDailysSimple(start_date, 30)...)
				first_month_for_the_monthlys = M1
			case 31:
				// The 30 days affect ONE month M0 and (the rest of) M0 is completely covered with daily backups.
				// Please note: the 31. will already be covered by the hourly backup filter logic, so the 30 daily filters will indeed cover the rest of the month
				// pin 30 normal dailys and test nothing. continue with M1
				result = append(result, getFiltersForDailysSimple(start_date, 30)...)
				first_month_for_the_monthlys = M1
			}
		}
	case 31:
		{
			switch daysM0 {
			case 28, 29, 30:
				// illegal in a month with a 31st day
			case 31:
				// The 30 days affect ONE month M0 and M0 is NOT completely covered with daily backups.
				// pin 0 normal dailys and test 30 dailys in in M0. continue with M1
				result = append(result, getFiltersForDailysOrForMonth(start_date, 30, existingDirs)...)
				first_month_for_the_monthlys = M1
			}
		}
	default:
		// The 30 days affect TWO months M0 and M1 and M1 is NOT completely covered with daily backups.
		// pin daysM0 normal dailys and test 30-daysM0 dailys in in M1. continue with M2
		result = append(result, getFiltersForDailysSimple(start_date, start_date.Day())...)
		result = append(result, getFiltersForDailysOrForMonth(getUltimo(M1.Year(), M1.Month()), 30-start_date.Day(), existingDirs)...)
		first_month_for_the_monthlys = M2
	}
	return result, first_month_for_the_monthlys
}

func getFiltersForDailysSimple(start_date time.Time, count int) []string {
	var result = []string{}
	for i := 0; i < count; i++ {
		// Format the time in the format YYYY-MM-DD
		prefix := start_date.Format("2006-01-02")
		result = append(result, prefix)
		start_date = start_date.AddDate(0, 0, -1)
	}
	return result
}

func getFiltersForDailysOrForMonth(start_date time.Time, remaining int, existingDirs []string) []string {
	filters_for_dailys := getFiltersForDailysSimple(start_date, remaining)
	any_matches := getAnyMatchingAnyPrefixes(existingDirs, filters_for_dailys) // check what is actually there
	if any_matches {
		// we found some daily backup folders, so return the filter for the daily backups, i.e. some YYYY-MM-DD filters
		return filters_for_dailys
	} else {
		// we found no daily backup folders within the specified range, so return a filter for month, i.e. one YYYY-MM filter
		filter := toDateStr(start_date.Year(), int(start_date.Month()))
		return []string{filter}
	}
}

func getFiltersForMonthlys(current time.Time, count int) []string {
	var result = []string{}
	// don't use AddDate(0, -1, 0) as this function does not work as expected when we're on a March, 29th in a non-leap-year, e.g.
	// use simpler and more robust approach, as from now on we don't need (leap-) days arithmetics anyhow
	var year int = current.Year()
	var month int = (int)(current.Month())

	for i := 0; i < count; i++ {
		// Format the time in the format YYYY-MM
		result = append(result, toDateStr(year, month))
		prevMonth(&year, &month)
	}
	return result
}

func getDateDirectoriesNotMatchingAnyPrefix(allDirs []string, prefixes []string, verbosity int) []string {
	var result = []string{}
	r, _ := regexp.Compile(`^[\d]{4}\-[\d]{2}\-[\d]{2}.*`)
	for _, dir := range allDirs {
		if r.Match([]byte(dir)) {
			found_match := false
			for _, prefix := range prefixes {
				if strings.HasPrefix(dir, prefix) {
					found_match = true
					break
				}
			}
			if !found_match {
				result = append(result, dir)
			}
		} else {
			if verbosity > 1 {
				fmt.Println("Skipping", dir, "as it is not in date format.")
			}
		}
	}
	return result
}
