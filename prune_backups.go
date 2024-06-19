package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

var commitInfo = func() string {
	//var version = "<unknown>"
	var vcs_revision = "<unknown>"
	var vcs_time = "<unknown>"
	var vcs_modified = "<unknown>"
	if info, ok := debug.ReadBuildInfo(); ok {
		/*
			if info.Main.Version != "" {
				version = info.Main.Version // currently (Go 1.22.*) always returns "(devel)" - so ignore it. wait for https://github.com/golang/go/issues/50603 (ETA Go 1.24)
			}
		*/
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				vcs_revision = setting.Value
			}
			if setting.Key == "vcs.time" {
				vcs_time = setting.Value
			}
			if setting.Key == "vcs.modified" {
				vcs_modified = setting.Value
			}
		}
	}
	return "rev " + vcs_revision + " from " + vcs_time + ", modified=" + vcs_modified
}()

func main() {

	/* Processing command line parameters */

	pruneDirName := flag.String("dir", "<none>", "REQUIRED. The name of the directory that shall be pruned.")
	toDeleteDirName := flag.String("to_directory", "to_delete", "OPTIONAL. The name of the directory where the pruned directories shall be moved.")
	showVersion := flag.Bool("version", false, "OPTIONAL. Show version/build information and exit if `true`. (default false)") // caution: go will neither print the type nor the default for bool flags with default false. see https://github.com/golang/go/issues/63150
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

	pruneDirectory(*pruneDirName, now, *toDeleteDirName, *verbosity)
}

func pruneDirectory(pruneDirName string, now time.Time, toDeleteDirName string, verbosity int) {
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

	prefixesForTimeSlotsToKeepOne := createPrefixesForTimeSlotsToKeepOne(now)
	for _, prefix := range prefixesForTimeSlotsToKeepOne {
		add_to_delete := getAllButFirstMatchingPrefix(dirs, prefix)
		to_delete = append(to_delete, add_to_delete...)
	}

	prefixesForTimeSlotsToKeepNone := createPrefixesForTimeSlotsToKeepNone(now)
	for _, prefix := range prefixesForTimeSlotsToKeepNone {
		add_to_delete := getAllMatchingPrefix(dirs, prefix)
		to_delete = append(to_delete, add_to_delete...)
	}

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
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func getAllButFirstMatchingPrefix(from []string, prefix string) []string {
	var result = []string{} // make sure it's not nil
	var first = true
	for _, s := range from {
		if strings.HasPrefix(s, prefix) {
			if first {
				first = false
			} else {
				result = append(result, s)
			}
		}
	}
	return result
}

func getAllMatchingPrefix(from []string, prefix string) []string {
	var result = []string{} // make sure it's not nil
	for _, s := range from {
		if strings.HasPrefix(s, prefix) {
			result = append(result, s)
		}
	}
	return result
}

func getAllMatchingAllPrefixes(from []string, prefixes []string) []string {
	var result = []string{} // make sure it's not nil
	for _, s := range from {
		for _, prefix := range prefixes {
			if strings.HasPrefix(s, prefix) {
				result = append(result, s)
			}
		}
	}
	return result
}

func createPrefixesForTimeSlotsToKeepOne(current time.Time) []string {
	// Create an array to hold the prefixes
	prefixes := make([]string, 24+30+119)

	// Generate the timestamps
	for i := 0; i < 24; i++ {
		// Format the time in the format YYYY-MM-DD_hh
		prefixes[i] = current.Format("2006-01-02_15") // caution, this is a magic number in go!

		// Subtract one hour from the current timestamp
		current = current.Add(-1 * time.Hour)
	}

	// Subtract one day from the current timestamp
	current = current.Add(-24 * time.Hour)

	for i := 24; i < 24+30; i++ {
		// Format the time in the format YYYY-MM-DD
		prefixes[i] = current.Format("2006-01-02")

		// Subtract one day from the current timestamp
		current = current.Add(-24 * time.Hour)
	}

	// don't use AddDate(0, -1, 0) as this function does not work as expected when we're on a March, 29th in a non-leap-year, e.g.
	// use simpler and more robust approach, as from now on we don't need (leap-) days arithmetics anyhow

	var year int = current.Year()
	var month int = (int)(current.Month())

	// we already keep the days of the 30 days leaping into the current month, so we continue with the next month
	prevMonth(&year, &month)

	for i := 24 + 30; i < 24+30+119; i++ {
		// Format the time in the format YYYY-MM
		prefixes[i] = toDateStr(year, month)
		prevMonth(&year, &month)
	}

	/*
		for _, prefix := range prefixes {
			fmt.Println(prefix)
		}
	*/

	return prefixes
}

func prevMonth(year *int, month *int) {
	*month--
	if *month <= 0 {
		*month = 12
		*year--
	}
}

func toDateStr(year int, month int) string {
	if month < 10 {
		return strconv.Itoa(year) + "-0" + strconv.Itoa(month)
	} else {
		return strconv.Itoa(year) + "-" + strconv.Itoa(month)
	}
}

func toDateStr3(year int, month int, day int) string {
	strSep1 := "-"
	strSep2 := "-"
	if month < 10 {
		strSep1 = "-0"
	}
	if day < 10 {
		strSep2 = "-0"
	}
	return strconv.Itoa(year) + strSep1 + strconv.Itoa(month) + strSep2 + strconv.Itoa(day)
}

func twoDigit(i int) string {
	if i < 10 {
		return "0" + strconv.Itoa(i)
	}
	return strconv.Itoa(i)
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

	yesterday_hourly := getAllMatchingAllPrefixes(allDirs, hourlyFilters) // check what is actually there - the filter for yesterday will depend on it

	if len(yesterday_hourly) > 0 {
		// we found some hourly backup folders for yesterday, so return the filter for the hourly backups for yesterday, i.e. some YYYY-MM-DD_HH filters
		return hourlyFilters
	} else {
		// we found at least one hourly backup folders for yesterday, so return the filter for the latest backup for yesterday, i.e. one YYYY-MM-DD filter
		return []string{yester_date_str}
	}
}

func getFiltersFor30Dailys(current_time time.Time) []string {
	var result = []string{}
	current_time = current_time.Add(-48 * time.Hour)
	// now current_time is the day before yesterday - the first day of the 30 daily backups filter

	for i := 0; i < 30; i++ {
		// Format the time in the format YYYY-MM-DD
		prefix := current_time.Format("2006-01-02")
		result = append(result, prefix)

		// Subtract one day from the current timestamp
		current_time = current_time.Add(-24 * time.Hour)
	}
	return result
}

func daysInMonth(year int, month time.Month) int {
	// Start with the first day of the next month
	t := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	// Subtract a day to get the last day of the original month
	t = t.AddDate(0, 0, -1)
	return t.Day()
}

func getMonthToLookForAnExtraMonthly(current_time time.Time) (bool, time.Time) {
	day_of_today := current_time.Day()
	month_before := get15thOfMonthBefore(current_time) // we don't really care for the exact day, just make sure it's not the 29th-31st as substracting a month will mean a hastle

	if day_of_today > 4 {
		// the 30 days start on the day before yesterday (2 days). the shortest month has 28 days (2 days).
		// if the day is >4 it is impossible that the 30 days do not end up somewhere 'in the middle' of the month before, even if it has 31 days
		is_needed := true
		return is_needed, month_before
	}

	if day_of_today == 3 {
		num_days_in_month_before := daysInMonth(month_before.Year(), month_before.Month())
		if num_days_in_month_before <= 28 {
			// spill-over into the month even before that
			month_before = get15thOfMonthBefore(month_before)
			is_needed := true
			return is_needed, month_before
		}
		if num_days_in_month_before == 29 {
			// the 30 Dailys cover every day of the month before
			is_needed := false
			return is_needed, month_before
		}
		// the 30 Dailys DO NOT cover every day of the month before
		is_needed := false
		return is_needed, month_before
	}
	if day_of_today == 2 {
		num_days_in_month_before := daysInMonth(month_before.Year(), month_before.Month())
		if num_days_in_month_before <= 29 {
			// spill-over into the month even before that
			month_before = get15thOfMonthBefore(month_before)
			is_needed := true
			return is_needed, month_before
		}
		if num_days_in_month_before == 30 {
			// the 30 Dailys cover every day of the month before
			is_needed := false
			return is_needed, month_before
		}
		// the 30 Dailys DO NOT cover every day of the month before
		is_needed := false
		return is_needed, month_before
	}
	// must be the 1st of the month

	num_days_in_month_before := daysInMonth(month_before.Year(), month_before.Month())
	if num_days_in_month_before <= 30 {
		// spill-over into the month even before that
		month_before = get15thOfMonthBefore(month_before)
		is_needed := true
		return is_needed, month_before
	}

	// the 30 Dailys cover every day of the month before
	is_needed := false
	return is_needed, month_before
}

func get15thOfMonthBefore(current_time time.Time) time.Time {
	t := time.Date(current_time.Year(), current_time.Month(), 15, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, -1, 0)
	return t
}

func createPrefixesForTimeSlotsToKeepNone(current time.Time) []string {
	// Create an array to hold the prefixes
	var prefixes []string

	// keep first 24h
	current = current.Add(-24 * time.Hour)
	dayOfLastHourlyBackup := current.Day()

	// add all hours betwen dayOfLastHourlyBackup and the next day/date where we keep the first daily backup
	for current.Day() == dayOfLastHourlyBackup {
		prefixes = append(prefixes, current.Format("2006-01-02_15"))
		current = current.Add(-1 * time.Hour)
	}

	// keep 30 daily backups
	current = current.Add(30 * -24 * time.Hour)

	dayOfLastMonthlyBackup := current.Month()

	// add all days betwen +24h and the next day/date
	for current.Month() == dayOfLastMonthlyBackup {
		prefixes = append(prefixes, current.Format("2006-01-02"))
		current = current.Add(-24 * time.Hour)
	}

	/*
		for _, prefix := range prefixes {
			fmt.Println(prefix)
		}
	*/
	return prefixes
}
