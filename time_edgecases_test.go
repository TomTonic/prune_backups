package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	// Guarantees the IANA time zone database is available even on a minimal
	// system without OS-provided zoneinfo (e.g. some container images),
	// so the DST tests below don't depend on the test environment.
	_ "time/tzdata"
)

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("failed to load time zone %q: %v", name, err)
	}
	return loc
}

// Test_getFiltersForToday_DST documents how getFiltersForToday behaves on the
// two days per year where a DST-observing local clock doesn't have a regular
// 24-hour day. currentTime.Add(-1*time.Hour) always steps by exactly one real
// hour; formatting that back to "HH" in local wall-clock terms is where DST
// shows up.
func Test_getFiltersForToday_DST(t *testing.T) {
	berlin := mustLoadLocation(t, "Europe/Berlin")

	t.Run("SpringForward_HourIsSkipped", func(t *testing.T) {
		// 2024-03-31: at 02:00 CET clocks jump to 03:00 CEST - local 02:00
		// never happens that day, so it must not appear in the result, and
		// the day has one fewer distinct hourly slot than usual.
		got := getFiltersForToday(time.Date(2024, 3, 31, 20, 0, 0, 0, berlin))
		assertNoDuplicates(t, got)
		if containsString(got, "2024-03-31_02") {
			t.Errorf("getFiltersForToday(spring-forward day) = %v, must not contain the skipped local hour 02", got)
		}
		if !containsString(got, "2024-03-31_01") || !containsString(got, "2024-03-31_03") {
			t.Errorf("getFiltersForToday(spring-forward day) = %v, want both neighboring hours 01 and 03 present", got)
		}
	})

	t.Run("FallBack_HourIsDuplicated", func(t *testing.T) {
		// 2024-10-27: local 02:00-02:59 happens twice (once as CEST, once as
		// CET). getFiltersForToday formats by date+hour only, so it cannot
		// tell the two apart - the same "_02" prefix is appended twice.
		// This is the root cause of the pruneDirectory bug covered by
		// Test_pruneDirectory_DST_FallBack_DuplicateHourCausesSpuriousError
		// below: it is documented here, not asserted as desired behavior.
		got := getFiltersForToday(time.Date(2024, 10, 27, 20, 0, 0, 0, berlin))
		count := 0
		for _, f := range got {
			if f == "2024-10-27_02" {
				count++
			}
		}
		if count != 2 {
			t.Errorf("getFiltersForToday(fall-back day) contains %d entries for the repeated local hour 02, want exactly 2 (documenting the duplicate that causes the pruneDirectory bug) - got %v", count, got)
		}
	})
}

// Test_pruneDirectory_DST_FallBack_DuplicateHourCausesSpuriousError is a
// regression test for a real bug: on the DST fall-back day, the local hour
// 02:00-02:59 occurs twice, so getFiltersForToday (see above) returns its
// "YYYY-MM-DD_02" filter twice. pruneDirectory applies every filter in that
// list independently and appends whatever getAllButFirstMatchingPrefix
// returns to toDelete without deduplicating - so if there is more than one
// backup directory in that ambiguous hour, the one that isn't kept ends up
// in toDelete twice. Its first move succeeds; the second os.Rename then
// fails because the source no longer exists, and pruneDirectory reports a
// spurious "could not be moved" error even though the retention outcome
// itself (keep the latest, move the rest) was already correct.
//
// As of now this test fails, correctly reproducing the bug.
func Test_pruneDirectory_DST_FallBack_DuplicateHourCausesSpuriousError(t *testing.T) {
	berlin := mustLoadLocation(t, "Europe/Berlin")
	dir := t.TempDir()

	// Two backups within the ambiguous local hour 02:00-02:59 on 2024-10-27.
	dirsToCreate := []string{"2024-10-27_02-10", "2024-10-27_02-40"}
	for _, name := range dirsToCreate {
		if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	now := time.Date(2024, 10, 27, 20, 0, 0, 0, berlin)
	if err := pruneDirectory(dir, now, "to_delete", 0, false); err != nil {
		t.Fatalf("pruneDirectory returned an error on the DST fall-back day: %v", err)
	}

	moved := readDirNamesSorted(t, filepath.Join(dir, "to_delete"))
	if want := []string{"2024-10-27_02-10"}; !equalStringSlices(moved, want) {
		t.Errorf("moved = %v, want %v (the earlier of the two ambiguous-hour backups)", moved, want)
	}
	survivors := readDirNamesSorted(t, dir)
	survivors = removeStringEntry(survivors, "to_delete")
	if want := []string{"2024-10-27_02-40"}; !equalStringSlices(survivors, want) {
		t.Errorf("survivors = %v, want %v (the later of the two ambiguous-hour backups)", survivors, want)
	}
}

// Test_pruneDirectory_DST_SpringForward_NoError checks that the missing local
// hour on the spring-forward day (see Test_getFiltersForToday_DST above)
// does not, unlike the fall-back day, cause pruneDirectory itself to error -
// there's no duplicate filter here, just one fewer than usual.
func Test_pruneDirectory_DST_SpringForward_NoError(t *testing.T) {
	berlin := mustLoadLocation(t, "Europe/Berlin")
	dir := t.TempDir()

	dirsToCreate := []string{"2024-03-31_01-30", "2024-03-31_03-30"}
	for _, name := range dirsToCreate {
		if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	now := time.Date(2024, 3, 31, 20, 0, 0, 0, berlin)
	if err := pruneDirectory(dir, now, "to_delete", 0, false); err != nil {
		t.Errorf("pruneDirectory returned an unexpected error on the DST spring-forward day: %v", err)
	}
}

// Test_pruneDirectory_YearBoundary is a confidence/regression test for
// running the tool right at midnight on New Year's Day, with existing
// backups spanning back across that boundary (and, further back in the
// monthly-retention chain, an earlier year boundary too). Unlike the DST
// case above, this exercises correct, already-covered logic
// (prevMonth's Jan->Dec wrap is unit-tested in misc_test.go); this test adds
// end-to-end coverage that the daily and monthly windows both handle the
// wraparound correctly together, which no existing pruneDirectory-level test
// checks. Expectations below were verified against a manual trace of
// getAllFilters, not guessed.
func Test_pruneDirectory_YearBoundary(t *testing.T) {
	dir := t.TempDir()

	dirsToCreate := []string{
		"2025-01-01_09-30", // today, hour 9 (latest) -> survives
		"2025-01-01_09-15", // today, hour 9 (older)   -> moved (duplicate hour slot)
		"2024-12-31_23-45", // yesterday, hour 23      -> survives (yesterday has hourly coverage)
		"2024-12-31_22-10", // yesterday, hour 22      -> survives
		"2024-12-31_10-00", // yesterday, hour 10: outside both yesterday's hourly range and the daily window -> moved
		"2024-12-15_08-00", // within the ~30-day daily window (Dec 1-30)  -> survives
		"2024-11-20_08-00", // outside the daily window, but is the 2024-11 monthly slot -> survives
		"2023-12-10_08-00", // 2023-12 monthly slot, one year further back -> survives
		"2022-01-05_08-00", // 2022-01 monthly slot, crossing another year boundary in the monthly chain -> survives
		"2010-01-01_08-00", // far beyond the ~119-month monthly window -> moved
		"someothername",    // not date-formatted -> left alone entirely
	}
	for _, name := range dirsToCreate {
		if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	now := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	if err := pruneDirectory(dir, now, "to_delete", 0, false); err != nil {
		t.Fatalf("pruneDirectory returned an unexpected error at the year boundary: %v", err)
	}

	wantMoved := []string{"2010-01-01_08-00", "2024-12-31_10-00", "2025-01-01_09-15"}
	if moved := readDirNamesSorted(t, filepath.Join(dir, "to_delete")); !equalStringSlices(moved, wantMoved) {
		t.Errorf("moved = %v, want %v", moved, wantMoved)
	}

	wantSurvivors := []string{
		"2022-01-05_08-00", "2023-12-10_08-00", "2024-11-20_08-00",
		"2024-12-15_08-00", "2024-12-31_22-10", "2024-12-31_23-45",
		"2025-01-01_09-30", "someothername",
	}
	survivors := readDirNamesSorted(t, dir)
	survivors = removeStringEntry(survivors, "to_delete")
	if !equalStringSlices(survivors, wantSurvivors) {
		t.Errorf("survivors = %v, want %v", survivors, wantSurvivors)
	}
}

func assertNoDuplicates(t *testing.T, s []string) {
	t.Helper()
	seen := make(map[string]bool, len(s))
	for _, v := range s {
		if seen[v] {
			t.Errorf("unexpected duplicate entry %q in %v", v, s)
		}
		seen[v] = true
	}
}

func containsString(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

func readDirNamesSorted(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func removeStringEntry(names []string, name string) []string {
	result := names[:0]
	for _, n := range names {
		if n != name {
			result = append(result, n)
		}
	}
	return result
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
