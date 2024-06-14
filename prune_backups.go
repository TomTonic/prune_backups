package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func createTimeStampPatterns() []string {
	// Create an array to hold the timestamps
	timestamps := make([]string, 24+30+120)

	// Get the current time
	t := time.Now()

	// Generate the timestamps
	for i := 0; i < 24; i++ {
		// Format the time in the format YYYY-MM-DD_hh
		timestamps[i] = t.Format("2006-01-02_15")

		// Subtract one hour from the current timestamp
		t = t.Add(-1 * time.Hour)
	}

	// Subtract one day from the current timestamp
	t = t.Add(-24 * time.Hour)

	for i := 24; i < 24+30; i++ {
		// Format the time in the format YYYY-MM-DD
		timestamps[i] = t.Format("2006-01-02")

		// Subtract one day from the current timestamp
		t = t.Add(-24 * time.Hour)
	}

	// Subtract one month from the current timestamp
	t = t.AddDate(0, -1, 0)

	for i := 24 + 30; i < 24+30+120; i++ {
		// Format the time in the format YYYY-MM
		timestamps[i] = t.Format("2006-01")

		// Subtract one month from the current timestamp
		t = t.AddDate(0, -1, 0)
	}

	return timestamps
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please provide a directory path")
		os.Exit(1)
	}
	dirPath := os.Args[1]

	/*
		// generate random test directories

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
		} else {
			fmt.Println(err)
			os.Exit(1)
		}

		now := time.Now()
		for i := 0; i < 100; i++ {
			randomNumber := rand.Intn(438) + 1
			t := now.Add(time.Duration(-randomNumber) * time.Hour)
			subDir := t.Format("2006-01-02_15-04")
			fullPath := filepath.Join(dirPath, subDir)
			err := os.MkdirAll(fullPath, 0755)
			if err != nil {
				fmt.Println(err)
			}
		}
	*/

	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dirs := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	fmt.Println("I found", len(dirs), "directories in", dirPath)

	// Sort in descending order
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	/*
		fmt.Println("Found directories:")
		for _, dir := range dirs {
			fmt.Println(dir)
		}
	*/

	prefixes := createTimeStampPatterns()

	var to_delete []string
	for _, prefix := range prefixes {
		add_to_delete := all_but_first_matching_prefix(dirs, prefix)
		to_delete = append(to_delete, add_to_delete...)
	}

	delPath := filepath.Join(dirPath, "to_delete")
	err2 := os.MkdirAll(delPath, 0755)
	if err2 != nil {
		fmt.Print("Error creating 'to_delete'-directory: ")
		fmt.Println(err)
		fmt.Println("I woud have moved the following directories there:")
		for _, dir := range to_delete {
			fmt.Println(" -", dir)
		}
		os.Exit(1)
	}

	var successful_move_counter = 0
	for _, dirname := range to_delete {
		fromPath := filepath.Join(dirPath, dirname)
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

func all_but_first_matching_prefix(from []string, prefix string) []string {
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
