//go:build e2e

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/TomTonic/rtcompare"
)

// Breadth/retention E2E test: generates tens of thousands of deterministic,
// date-named backup directories (same DPRNG seed and fixed reference time
// every run) and runs the real pruneDirectory retention logic over them.
//
// Per-slot exact-match verification (which single directory survives inside
// each hourly/daily/monthly retention slot) would effectively require a
// second, independently-written implementation of the layered retention
// policy in prune_backups.go (today/yesterday exempted from the daily
// window, the daily window exempted from the monthly window, etc.) - exactly
// the kind of shadow implementation we decided against. Instead:
//
//  1. The full list of directories pruneDirectory actually moved is diffed
//     against a golden file (testdata_e2e/e2e_breadth_golden.txt), regenerable
//     via `go test -tags e2e -run Test_E2E_PruneDirectory_Breadth -update .`
//     after a manual review of the change. Because generation is fully
//     deterministic (fixed DPRNG seed, fixed reference time - never
//     time.Now()), this is a real regression check, not a flaky snapshot.
//  2. A couple of policy-agnostic structural invariants are checked in
//     addition, to catch gross bugs even if the golden file itself were
//     wrong: every generated directory ends up in exactly one of
//     {survivors, moved}, and the number of survivors stays in the small
//     range the retention policy's slot count implies.
const (
	e2ePruneSeed     = 0xB2EAD7C0
	e2eSpanDays      = 3650 // ~10 years back from referenceNow
	e2eMinDirsPerDay = 1
	e2eMaxDirsPerDay = 20
	e2eGapDayChance  = 10 // 1-in-N days beyond day 1 get no backups at all
	// Deliberately not under testdata/: that directory is itself a fixture
	// scanned by TestShowStatusOf (prune_backups_test.go) and Test_du1
	// (stats_test.go), both asserting exact file/byte counts against it.
	e2eGoldenFile         = "testdata_e2e/e2e_breadth_golden.txt"
	e2eMaxPlausibleSurviv = 300 // generously above hourly+daily+monthly slot counts
)

var updateGolden = flag.Bool("update", false, "regenerate e2e golden files instead of comparing against them")

func Test_E2E_PruneDirectory_Breadth(t *testing.T) {
	root := t.TempDir()
	referenceNow := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	dprng := rtcompare.NewDPRNG(e2ePruneSeed)

	created := map[string]bool{}
	mkBackupDir := func(ts time.Time) {
		name := ts.Format("2006-01-02_15-04")
		if created[name] {
			return // DPRNG picked a timestamp we already have; skip rather than error
		}
		created[name] = true
		mustMkdirAllE2E(t, filepath.Join(root, name))
	}

	for dayOffset := 0; dayOffset <= e2eSpanDays; dayOffset++ {
		day := referenceNow.AddDate(0, 0, -dayOffset)
		if dayOffset > 1 && dprng.UInt32N(e2eGapDayChance) == 0 {
			continue // simulate an occasional day with no backups at all
		}
		numDirs := e2eMinDirsPerDay + int(dprng.UInt32N(e2eMaxDirsPerDay-e2eMinDirsPerDay+1))
		for range numDirs {
			hour := int(dprng.UInt32N(24))
			minute := int(dprng.UInt32N(60))
			mkBackupDir(time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, time.UTC))
		}
	}

	// Guarantee (rather than merely hope via DPRNG luck) that today has
	// several backups within one hour, and yesterday has backups in more
	// than one hour, so the "keep only the latest per hour" and the
	// "yesterday has hourly backups" branches are exercised deterministically.
	for _, minute := range []int{5, 20, 45} {
		mkBackupDir(time.Date(referenceNow.Year(), referenceNow.Month(), referenceNow.Day(), referenceNow.Hour(), minute, 0, 0, time.UTC))
	}
	yesterday := referenceNow.AddDate(0, 0, -1)
	for _, hour := range []int{9, 14} {
		mkBackupDir(time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), hour, 30, 0, 0, time.UTC))
	}

	allGenerated := make([]string, 0, len(created))
	for name := range created {
		allGenerated = append(allGenerated, name)
	}
	t.Logf("generated %d distinct backup directories", len(allGenerated))

	if err := pruneDirectory(root, referenceNow, "to_delete", 0, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	moved := readDirNames(t, filepath.Join(root, "to_delete"))
	survivors := readDirNames(t, root)
	survivors = removeName(survivors, "to_delete")

	if *updateGolden {
		writeGolden(t, e2eGoldenFile, moved)
		t.Logf("updated %s with %d entries", e2eGoldenFile, len(moved))
		return
	}

	want := readGolden(t, e2eGoldenFile)
	if diff := diffSortedStringSlices(want, moved); diff != "" {
		t.Errorf("directories moved to to_delete differ from golden file %s:\n%s", e2eGoldenFile, diff)
	}

	// Structural invariants, independent of the golden file.
	if len(survivors) == 0 || len(survivors) > e2eMaxPlausibleSurviv {
		t.Errorf("survivor count = %d, want a small number (<= %d) but > 0 - retention policy looks broken, not just changed", len(survivors), e2eMaxPlausibleSurviv)
	}
	if partitionErr := checkPartition(allGenerated, survivors, moved); partitionErr != "" {
		t.Errorf("every generated directory should end up in exactly one of {survivors, moved}: %s", partitionErr)
	}
}

func readDirNames(t *testing.T, dir string) []string {
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

func removeName(names []string, name string) []string {
	result := names[:0]
	for _, n := range names {
		if n != name {
			result = append(result, n)
		}
	}
	return result
}

// checkPartition verifies that generated is exactly the disjoint union of a
// and b, without assuming anything about which retention rule put a given
// name into which side.
func checkPartition(generated, a, b []string) string {
	inA := make(map[string]bool, len(a))
	for _, n := range a {
		inA[n] = true
	}
	inB := make(map[string]bool, len(b))
	for _, n := range b {
		inB[n] = true
	}

	var missing, duplicated []string
	for _, n := range generated {
		switch {
		case inA[n] && inB[n]:
			duplicated = append(duplicated, n)
		case !inA[n] && !inB[n]:
			missing = append(missing, n)
		}
	}
	extra := len(a) + len(b) - len(generated) - len(duplicated)

	var msg strings.Builder
	if len(missing) > 0 {
		fmt.Fprintf(&msg, "%d generated director(ies) missing from both sides, e.g. %v; ", len(missing), head(missing, 5))
	}
	if len(duplicated) > 0 {
		fmt.Fprintf(&msg, "%d director(ies) present on both sides, e.g. %v; ", len(duplicated), head(duplicated, 5))
	}
	if extra > 0 {
		fmt.Fprintf(&msg, "%d unexplained extra entr(ies) beyond what was generated; ", extra)
	}
	return msg.String()
}

func head(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func readGolden(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open golden file %s (run with -update to create it): %v", path, err)
	}
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	sort.Strings(lines)
	return lines
}

func writeGolden(t *testing.T, path string, lines []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create golden file dir for %s: %v", path, err)
	}
	sorted := append([]string(nil), lines...)
	sort.Strings(sorted)
	var buf strings.Builder
	for _, line := range sorted {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("failed to write golden file %s: %v", path, err)
	}
}

// diffSortedStringSlices returns a human-readable, length-capped description
// of the difference between two already-sorted slices, or "" if they match.
func diffSortedStringSlices(want, got []string) string {
	wantSet := make(map[string]bool, len(want))
	for _, w := range want {
		wantSet[w] = true
	}
	gotSet := make(map[string]bool, len(got))
	for _, g := range got {
		gotSet[g] = true
	}

	var onlyInWant, onlyInGot []string
	for _, w := range want {
		if !gotSet[w] {
			onlyInWant = append(onlyInWant, w)
		}
	}
	for _, g := range got {
		if !wantSet[g] {
			onlyInGot = append(onlyInGot, g)
		}
	}
	if len(onlyInWant) == 0 && len(onlyInGot) == 0 {
		return ""
	}

	const maxExamples = 20
	var b strings.Builder
	fmt.Fprintf(&b, "golden has %d entries, actual has %d entries\n", len(want), len(got))
	if len(onlyInWant) > 0 {
		fmt.Fprintf(&b, "missing from actual (%d), e.g.: %v\n", len(onlyInWant), head(onlyInWant, maxExamples))
	}
	if len(onlyInGot) > 0 {
		fmt.Fprintf(&b, "unexpected in actual (%d), e.g.: %v\n", len(onlyInGot), head(onlyInGot, maxExamples))
	}
	return b.String()
}
