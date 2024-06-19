package main

import (
	"flag"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
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

func daysInMonth(year int, month time.Month) int {
	// Start with the first day of the next month
	t := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	// Subtract a day to get the last day of the original month
	t = t.AddDate(0, 0, -1)
	return t.Day()
}

func get15thOfMonthBefore(current_time time.Time) time.Time {
	t := time.Date(current_time.Year(), current_time.Month(), 15, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, -1, 0)
	return t
}

func Test_toDateStr3(t *testing.T) {
	type args struct {
		year  int
		month int
		day   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test Case 1",
			args: args{year: 2024, month: 6, day: 16},
			want: "2024-06-16",
		},
		{
			name: "Test Case 2",
			args: args{year: 2024, month: 11, day: 5},
			want: "2024-11-05",
		},
		{
			name: "Test Case 3",
			args: args{year: 2024, month: 10, day: 15},
			want: "2024-10-15",
		},
		{
			name: "Test Case 4",
			args: args{year: 2024, month: 3, day: 5},
			want: "2024-03-05",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toDateStr3(tt.args.year, tt.args.month, tt.args.day); got != tt.want {
				t.Errorf("toDateStr3() = %v, want %v", got, tt.want)
			}
		})
	}
}
