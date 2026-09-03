//go:build e2e

package main

import (
	"os"
	"testing"
)

// mustMkdirAllE2E creates dir (and any missing parents) or fails the test.
func mustMkdirAllE2E(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}
}

// createE2EFile writes a file of the given size (zero-filled). Unlike
// stats_test.go's createTestfile, it skips file.Sync(): these fixtures are
// thrown away at the end of the test, so durability against a crash is
// irrelevant, and at the file counts these E2E tests generate, fsync's
// per-call latency (particularly on a real disk rather than the tmpfs CI
// uses) dominates runtime for no benefit.
func createE2EFile(name string, size int) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	_, err = file.Write(make([]byte, size))
	return err
}
