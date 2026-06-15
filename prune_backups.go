package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Version VersionCmd `cmd:"" help:"Show version/build information and exit."`
	From    PruneCmd   `cmd:"" help:"Prune subdirectories from <dir> and move them to a 'to_delete' subdirectory (default, will be created automatically in <dir>) or --to a given location."`
	Stats   StatsCmd   `cmd:"" help:"Show total size of linked and unlinked files in a given directory."`
}

type VersionCmd struct{}

type StatsCmd struct {
	Dir string `arg:"" help:"REQUIRED. The name of the directory for searching and aggregating file types and sizes." required:"true"`
}

type PruneCmd struct {
	To        string `help:"OPTIONAL. The name of the directory where the pruned directories will be moved." default:"to_delete" short:"t"`
	Stats     bool   `help:"OPTIONAL. Show total size of linked and unlinked files in the pruned directories." default:"false" short:"s"`
	Verbosity int    `help:"OPTIONAL. Set verbosity. 0 - mute, 1 - some, 2 - a lot." default:"1" short:"v"`
	Dir       string `arg:"" help:"REQUIRED. The name of the directory that will be pruned. Make sure the user running prune_backups has r/w access rights to it." required:"true"`
}

func (v *VersionCmd) Run(cli *CLI) error {
	fmt.Println("prune_backups", runtime.GOARCH, runtime.GOOS, commitInfo)
	return nil
}

func (p *PruneCmd) Run(cli *CLI) error {
	if p.Stats && !Stats_SupportedOS {
		return errors.New("stats flag not supported for your OS")
	}

	now := time.Now()

	err := pruneDirectory(p.Dir, now, p.To, p.Verbosity, p.Stats)
	return err
}

func (p *StatsCmd) Run(cli *CLI) error {
	if !Stats_SupportedOS {
		return errors.New("stats command not supported for your OS")
	}
	err := showStatsOf(p.Dir)
	return err
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("prune_backups"),
		kong.Description("A lightweight tool designed to elegantly trim backup directories based on filename conventions, maintaining one per hour for a day, one per day for a month, and one per month thereafter. The pattern is YYYY-MM-DD_HH-mm. Within each time slot, the latest directory is retained."),
		// kong.UsageOnError(),
	)
	err := ctx.Run(&cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func pruneDirectory(pruneDirName string, now time.Time, toDeleteDirName string, verbosity int, showStats bool) error {
	files, err := os.ReadDir(pruneDirName)
	if err != nil {
		errorMessage := fmt.Sprintf("Could not read pruning directory: %s", err)
		return errors.New(errorMessage)
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

	var toDelete []string // in this array we will collect all directories that we will move to the to_delete-directory

	filters := getAllFilters(now, dirs)
	for _, filter := range filters {
		addToDelete := getAllButFirstMatchingPrefix(dirs, filter)
		toDelete = append(toDelete, addToDelete...)
	}

	cleanupOthers := getDateDirectoriesNotMatchingAnyPrefix(dirs, filters, verbosity)
	toDelete = append(toDelete, cleanupOthers...)

	delPath := filepath.Join(pruneDirName, toDeleteDirName)
	err2 := os.MkdirAll(delPath, 0755)
	if err2 != nil {
		errorMessage := fmt.Sprintf("Error creating directory \"%s\": %s", delPath, err2)
		if verbosity > 0 {
			movedDirs := "\nI would have moved the following directories there:\n"
			for _, dir := range toDelete {
				movedDirs += fmt.Sprintf(" - %s\n", dir)
			}
			errorMessage += movedDirs
		}
		return errors.New(errorMessage)
	}

	/* now we have collected all directory names that need to be moved in toDelete. next we will create the target directory and actually move them */
	var successfulMoveCounter, failedMoveCounter int
	for _, dirname := range toDelete {
		fromPath := filepath.Join(pruneDirName, dirname)
		toPath := filepath.Join(delPath, dirname)
		if verbosity > 1 {
			fmt.Print("Moving ", fromPath, " to ", toPath, "... ")
		}
		if moveErr := os.Rename(fromPath, toPath); moveErr != nil {
			failedMoveCounter++
			if verbosity > 1 {
				fmt.Println(moveErr)
			} else {
				fmt.Println("Error moving ", fromPath, " to ", toPath, ": ", moveErr)
			}
		} else {
			if verbosity > 1 {
				fmt.Println("done.")
			}
			successfulMoveCounter++
		}
	}
	if verbosity > 0 {
		fmt.Println("I moved", successfulMoveCounter, "directories to", delPath)
	}
	var result error
	if failedMoveCounter > 0 {
		result = fmt.Errorf("%d of %d directories could not be moved to %s", failedMoveCounter, len(toDelete), delPath)
	}
	if showStats {
		return errors.Join(result, showStatsOf(delPath))
	}
	return result
}

func showStatsOf(delPath string) error {
	info, err := DiskUsage(delPath)
	if err != nil {
		return err
	}
	fmt.Printf("Content of %v:\n", delPath)
	printNiceNumbr(" - unlinked files            ", uint64(info.number_of_unlinked_files))
	printNiceBytes(" - bytes in unlinked files   ", info.size_of_unlinked_files)
	printNiceNumbr(" - hard-linked files         ", uint64(info.number_of_linked_files))
	printNiceBytes(" - bytes in hard-linked files", info.size_of_linked_files)
	printNiceNumbr(" - directories               ", uint64(info.number_of_subdirs))
	printNiceNumbr(" - append-only-flagged files ", uint64(info.nr_apnd))
	printNiceNumbr(" - exclusive-flagged files   ", uint64(info.nr_excl))
	printNiceNumbr(" - temporary-flagged files   ", uint64(info.nr_tmp))
	printNiceNumbr(" - symlinks                  ", uint64(info.nr_sym))
	printNiceNumbr(" - device nodes              ", uint64(info.nr_dev))
	printNiceNumbr(" - named pipes               ", uint64(info.nr_pipe))
	printNiceNumbr(" - sockets                   ", uint64(info.nr_sock))
	if info.number_of_permission_errors_files+info.number_of_permission_errors_dirs+info.number_of_other_errors_files+info.number_of_other_errors_dirs == 0 {
		fmt.Println("No I/O errors occurred scanning the directory tree.")
	} else {
		fmt.Printf("%v errors occurred scanning the directory tree:\n", info.number_of_permission_errors_files+info.number_of_permission_errors_dirs+info.number_of_other_errors_files+info.number_of_other_errors_dirs)
		printNiceNumbr(" - permission denial accessing directories ", uint64(info.number_of_permission_errors_dirs))
		printNiceNumbr(" - permission denial accessing files       ", uint64(info.number_of_permission_errors_files))
		printNiceNumbr(" - other errors accessing directories      ", uint64(info.number_of_other_errors_dirs))
		printNiceNumbr(" - other errors accessing files            ", uint64(info.number_of_other_errors_files))
	}
	return nil
}

func printNiceNumbr(prefix string, val uint64) {
	if val > 999 {
		fmt.Printf("%s : %v (i.e. %v)\n", prefix, val, formatSI(val))
	} else {
		fmt.Printf("%s : %v\n", prefix, val)
	}
}

func printNiceBytes(prefix string, val uint64) {
	if val > 999 {
		fmt.Printf("%s : %v Bytes (i.e. %vBytes)\n", prefix, val, formatSI(val))
	} else {
		fmt.Printf("%s : %v Bytes\n", prefix, val)
	}
}

func getAllFilters(startTime time.Time, existingDirs []string) []string {
	var result = []string{}

	// append hourly filters

	result = append(result, getFiltersForHourlies(startTime, existingDirs)...)

	// append daily filters

	filtersForDailies, firstMonthForMonthlies := getFiltersForDailies(startTime.AddDate(0, 0, -2), existingDirs)
	result = append(result, filtersForDailies...)

	// append monthly filters

	result = append(result, getFiltersForMonthlies(firstMonthForMonthlies, 119)...)

	return result
}

func getFiltersForHourlies(startTime time.Time, existingDirs []string) []string {
	var result = []string{}
	filtersToday := getFiltersForToday(startTime)
	result = append(result, filtersToday...)
	remainingHourlies := 24 - len(filtersToday)
	filtersYesterday := getFiltersForYesterday(startTime, remainingHourlies, existingDirs)
	result = append(result, filtersYesterday...)
	return result
}

func getFiltersForToday(currentTime time.Time) []string {
	var result = []string{}

	day := currentTime.Day()

	for currentTime.Day() == day {
		// Format the time in the format YYYY-MM-DD_hh
		prefix := currentTime.Format("2006-01-02_15") // caution, this is a magic number in go!
		result = append(result, prefix)

		// Subtract one hour from the current timestamp
		currentTime = currentTime.Add(-1 * time.Hour)
	}
	return result
}

func getFiltersForYesterday(currentTime time.Time, remainingHourlyBackups int, allDirs []string) []string {
	var hourlyFilters = []string{}

	yesterday := currentTime.Add(-24 * time.Hour)

	var year = yesterday.Year()
	var month = (int)(yesterday.Month())
	var day = yesterday.Day()
	yesterDateStr := toDateStr3(year, month, day)

	for i := range remainingHourlyBackups {
		prefix := yesterDateStr + "_" + twoDigit(23-i)
		hourlyFilters = append(hourlyFilters, prefix)
	}

	anyMatches := getAnyMatchingAnyPrefixes(allDirs, hourlyFilters) // check what is actually there - the filter for yesterday will depend on it

	if anyMatches {
		// we found some hourly backup folders for yesterday, so return the filter for the hourly backups for yesterday, i.e. some YYYY-MM-DD_HH filters
		return hourlyFilters
	} else {
		// we found no hourly backup folders for yesterday, so return the filter for the latest backup for yesterday, i.e. one YYYY-MM-DD filter
		return []string{yesterDateStr}
	}
}

func getFiltersForDailies(startDate time.Time, existingDirs []string) ([]string, time.Time) {
	var result = []string{}
	var firstMonthForMonthlies time.Time
	M1 := get15thOfMonthBefore(startDate)
	M2 := get15thOfMonthBefore(M1)
	M3 := get15thOfMonthBefore(M2)
	daysM0 := daysInMonth(startDate.Year(), startDate.Month())
	daysM1 := daysInMonth(M1.Year(), M1.Month())

	switch startDate.Day() {
	case 1:
		{
			switch daysM1 {
			case 28:
				// The 30 days affect THREE months M0, M1, and M2 and M2 is NOT completely covered with daily backups.
				// pin 29 normal dailies and test 1 daily in M2. continue with M3
				result = append(result, getFiltersForDailiesSimple(startDate, 29)...)
				result = append(result, getFiltersForDailiesOrForMonth(getUltimo(M2.Year(), M2.Month()), 1, existingDirs)...)
				firstMonthForMonthlies = M3
			case 29:
				// The 30 days affect TWO months M0 and M1 and M1 is completely covered with daily backups.
				// pin 30 normal dailies and test nothing. continue with M2
				result = append(result, getFiltersForDailiesSimple(startDate, 30)...)
				firstMonthForMonthlies = M2
			case 30, 31:
				// The 30 days affect TWO months M0 and M1 and M1 is NOT completely covered with daily backups.
				// pin 1 normal daily and test 29 dailies in M1. continue with M2
				result = append(result, getFiltersForDailiesSimple(startDate, 1)...)
				result = append(result, getFiltersForDailiesOrForMonth(getUltimo(M1.Year(), M1.Month()), 29, existingDirs)...)
				firstMonthForMonthlies = M2
			}
		}
	case 2:
		{
			switch daysM1 {
			case 28:
				// The 30 days affect TWO months M0 and M1 and M1 is completely covered with daily backups.
				// pin 30 normal dailies and test nothing. continue with M2
				result = append(result, getFiltersForDailiesSimple(startDate, 30)...)
				firstMonthForMonthlies = M2
			case 29, 30, 31:
				// The 30 days affect TWO months M0 and M1 and M1 is NOT completely covered with daily backups.
				// pin 2 normal dailies and test 28 dailies in M1. continue with M2
				result = append(result, getFiltersForDailiesSimple(startDate, 2)...)
				result = append(result, getFiltersForDailiesOrForMonth(getUltimo(M1.Year(), M1.Month()), 28, existingDirs)...)
				firstMonthForMonthlies = M2
			}
		}
	case 30:
		{
			switch daysM0 {
			// case 28, 29:
			// impossible in a month with a 30st day if daysInMonth() works correctly
			case 30:
				// The 30 days affect ONE month M0 and M0 is completely covered with daily backups.
				// pin 30 normal dailies and test nothing. continue with M1
				result = append(result, getFiltersForDailiesSimple(startDate, 30)...)
				firstMonthForMonthlies = M1
			case 31:
				// The 30 days affect ONE month M0 and (the rest of) M0 is completely covered with daily backups.
				// Please note: the 31. will already be covered by the hourly backup filter logic, so the 30 daily filters will indeed cover the rest of the month
				// pin 30 normal dailies and test nothing. continue with M1
				result = append(result, getFiltersForDailiesSimple(startDate, 30)...)
				firstMonthForMonthlies = M1
			}
		}
	case 31:
		{
			switch daysM0 {
			// case 28, 29, 30:
			// impossible in a month with a 31st day if daysInMonth() works correctly
			case 31:
				// The 30 days affect ONE month M0 and M0 is NOT completely covered with daily backups.
				// pin 0 normal dailies and test 30 dailies in M0. continue with M1
				result = append(result, getFiltersForDailiesOrForMonth(startDate, 30, existingDirs)...)
				firstMonthForMonthlies = M1
			}
		}
	default:
		// The 30 days affect TWO months M0 and M1 and M1 is NOT completely covered with daily backups.
		// pin daysM0 normal dailies and test 30-daysM0 dailies in M1. continue with M2
		result = append(result, getFiltersForDailiesSimple(startDate, startDate.Day())...)
		result = append(result, getFiltersForDailiesOrForMonth(getUltimo(M1.Year(), M1.Month()), 30-startDate.Day(), existingDirs)...)
		firstMonthForMonthlies = M2
	}
	return result, firstMonthForMonthlies
}

func getFiltersForDailiesSimple(startDate time.Time, count int) []string {
	var result = []string{}
	for range count {
		// Format the time in the format YYYY-MM-DD
		prefix := startDate.Format("2006-01-02")
		result = append(result, prefix)
		startDate = startDate.AddDate(0, 0, -1)
	}
	return result
}

func getFiltersForDailiesOrForMonth(startDate time.Time, remaining int, existingDirs []string) []string {
	filtersForDailies := getFiltersForDailiesSimple(startDate, remaining)
	anyMatches := getAnyMatchingAnyPrefixes(existingDirs, filtersForDailies) // check what is actually there
	if anyMatches {
		// we found some daily backup folders, so return the filter for the daily backups, i.e. some YYYY-MM-DD filters
		return filtersForDailies
	} else {
		// we found no daily backup folders within the specified range, so return a filter for month, i.e. one YYYY-MM filter
		filter := toDateStr(startDate.Year(), int(startDate.Month()))
		return []string{filter}
	}
}

func getFiltersForMonthlies(current time.Time, count int) []string {
	var result = []string{}
	// don't use AddDate(0, -1, 0) as this function does not work as expected when we're on a March, 29th in a non-leap-year, e.g.
	// use simpler and more robust approach, as from now on we don't need (leap-) days arithmetics anyhow
	var year = current.Year()
	var month = (int)(current.Month())

	for range count {
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
			foundMatch := false
			for _, prefix := range prefixes {
				if strings.HasPrefix(dir, prefix) {
					foundMatch = true
					break
				}
			}
			if !foundMatch {
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
