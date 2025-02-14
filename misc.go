package main

import (
	"flag"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var commitInfo = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" {
			version := info.Main.Version
			return version
		}
	}
	return "untagged"
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

func getAnyMatchingAnyPrefixes(search_in []string, prefixes []string) bool {
	for _, s := range search_in {
		for _, prefix := range prefixes {
			if strings.HasPrefix(s, prefix) {
				return true
			}
		}
	}
	return false
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
	return getUltimo(year, month).Day()
}

func getUltimo(year int, month time.Month) time.Time {
	// Start with the first day of the next month
	t := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	// Subtract a day to get the last day of the original month
	t = t.AddDate(0, 0, -1)
	return t
}

func get15thOfMonthBefore(current_time time.Time) time.Time {
	t := time.Date(current_time.Year(), current_time.Month(), 15, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, -1, 0)
	return t
}

func formatSI(b uint64) string {
	const basis = 1000
	if b < basis {
		return fmt.Sprintf("%d ", b)
	}
	div := int64(basis)
	exp := 0
	for n := b / basis; n >= basis; n /= basis {
		div *= basis
		exp++
	}
	return fmt.Sprintf(
		"%.1f %c",
		float64(b)/float64(div),
		"kMGTPE"[exp],
	)
}
