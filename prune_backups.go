package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

var commit_info = func() string {
	//var version = "<unknown>"
	var vcs_revision = "<unknown>"
	var vcs_time = "<unknown>"
	var vcs_modified = "<unknown>"
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" {
			// version = info.Main.Version // currently (Go 1.22.*) always returns "(devel)" - so ignore it. wait for https://github.com/golang/go/issues/50603 (ETA Go 1.24)
		}
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
	doTest := flag.Bool("test", false, "OPTIONAL. Generate test directories and prune them afterwards if `true`. (default false)") // caution: go will neither print the type nor the default for bool flags with default false. see https://github.com/golang/go/issues/63150
	showVersion := flag.Bool("version", false, "OPTIONAL. Show version/build information and exit if `true`. (default false)")     // caution: go will neither print the type nor the default for bool flags with default false. see https://github.com/golang/go/issues/63150

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if *showVersion {
		fmt.Println("prune_backups", commit_info)
		os.Exit(0)
	}

	// workaroud as REQUIRED parameters are not supported by the flag package
	if !isFlagPassed("dir") {
		fmt.Println("Pruning directory missing (-dir).")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// generates 2800 hourly directories (24h x 30d x 4months)
	if *doTest {
		generateTestDirectories(*pruneDirName, 2800)
	}

	/* Reading and filtering directories */

	files, err := os.ReadDir(*pruneDirName)
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
	fmt.Println("I found", len(dirs), "directories in", *pruneDirName)

	// Sort in descending order - caution: this is important for the algorithm!
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	now := time.Now()
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

	/* now we have collected all directory names that need to be moved in to_delete. next we will create the target directory and actually move them */

	delPath := filepath.Join(*pruneDirName, *toDeleteDirName)
	err2 := os.MkdirAll(delPath, 0755)
	if err2 != nil {
		fmt.Print("Error creating directory \"", delPath, "\": ")
		fmt.Println(err)
		fmt.Println("I woud have moved the following directories there:")
		for _, dir := range to_delete {
			fmt.Println(" -", dir)
		}
		os.Exit(1)
	}

	var successful_move_counter = 0
	for _, dirname := range to_delete {
		fromPath := filepath.Join(*pruneDirName, dirname)
		toPath := filepath.Join(delPath, dirname)
		fmt.Print("Moving ", fromPath, " to ", toPath, "... ")
		err3 := os.Rename(fromPath, toPath)
		if err3 != nil {
			fmt.Println(err)
		} else {
			fmt.Println("done.")
			successful_move_counter++
		}
	}
	fmt.Println("I moved", successful_move_counter, "directories to", delPath)
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
	var result []string
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
	var result []string
	for _, s := range from {
		if strings.HasPrefix(s, prefix) {
			result = append(result, s)
		}
	}
	return result
}

func createPrefixesForTimeSlotsToKeepOne(current time.Time) []string {
	// Create an array to hold the prefixes
	prefixes := make([]string, 24+30+120)

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

	// Subtract one month from the current timestamp
	current = current.AddDate(0, -1, 0)

	for i := 24 + 30; i < 24+30+120; i++ {
		// Format the time in the format YYYY-MM
		prefixes[i] = current.Format("2006-01")

		// Subtract one month from the current timestamp
		current = current.AddDate(0, -1, 0)
	}

	/*
		for _, prefix := range prefixes {
			fmt.Println(prefix)
		}
	*/

	return prefixes
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

func generateTestDirectories(dirPath string, number int) {

	_, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Trying to create it...")
		// The second argument is the permission mode.
		// 0755 commonly used for directories.
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	now := time.Now()
	for i := 0; i < number; i++ {
		t := now.Add(time.Duration(-i) * time.Hour)
		subDir := t.Format("2006-01-02_15-04")
		fullPath := filepath.Join(dirPath, subDir)
		err := os.MkdirAll(fullPath, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}
}
