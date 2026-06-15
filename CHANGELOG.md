# Changelog

All notable changes to this project will be documented in this file.
Versions follow [Semantic Versioning](https://semver.org/). Only minor version
releases (vX.Y) are listed here; patch releases (vX.Y.Z) contain dependency
updates only and are not listed separately.

## [v0.10] - 2026-06-15

### Changed Behavior

- Move failures are now reported as an error instead of being silently ignored.
  All moves are still attempted; a summary error is returned when one or more fail.
- Exit code corrected: the process now exits with code `1` on error (previously `-1`,
  which is undefined on some platforms).
- Stats errors are now printed when `DiskUsage` cannot determine usage for a path.
- Goroutine and OS-thread limits for parallel directory scanning were raised to prevent
  "too many threads" panics on very large directory trees.
- The internal concurrency limiter was rewritten from a busy-wait (CAS spin loop) to a
  channel-based semaphore, eliminating CPU overhead while goroutines wait for a free slot.
- CLI help text and command structure were refined for clarity.

### Source Code Updates (Supply Chain)

| Dependency | Previous | Current |
|---|---|---|
| `github.com/alecthomas/kong` | v1.9.0 | v1.15.0 |
| `golang.org/x/sys` | v0.30.0 | v0.46.0 |

### CI Environment Updates

| Component | Previous | Current |
|---|---|---|
| Go toolchain | 1.24.1 | 1.26 |
| `golangci/golangci-lint-action` | v6.5.0 | v9.0.0 |
| `github/codeql-action` | v3.28.10 | v4.35.1 |
| `actions/checkout` | v4.2.2 | v6.0.2 |
| `actions/setup-go` | v5.3.0 | v6.0.0 |
| `actions/upload-artifact` | v4.6.1 | v7 |
| `actions/dependency-review-action` | v4.7.0 | v5 |
| `softprops/action-gh-release` | v2.2.1 | v3 |
| `vladopajic/go-test-coverage` | v2.12.1 | v2.18.3 |

- Migrated from Dependabot to Renovate for automated dependency updates.
- Applied StepSecurity GitHub Actions hardening across all workflows.
- golangci-lint configuration migrated to v2 syntax.
- Test coverage tool now runs as a Go source tool instead of a Docker image.

---

## [v0.9] - 2025-03-03

### New Features

- New `stats` subcommand: `prune_backups stats <path>` reports detailed disk usage
  for a directory or file, including file counts, sizes, link counts, and permission
  errors. Supported on Linux, macOS, and Windows.

### Changed Behavior

- Switched to [Kong](https://github.com/alecthomas/kong) for command-line argument
  processing, replacing the previous manual flag parsing.
- `--version` now reports the Go build revision via Go 1.24 build information.
- Error output when the deletion target directory cannot be created now includes the
  list of directories that would have been moved.
- Significantly expanded test suite; coverage for all filter functions now complete.

### CI Environment Updates

- Go toolchain updated from 1.22.x to 1.24.0.
- Added CI/CD pipeline for automated multi-platform binary builds and GitHub Releases.
- Added OpenSSF Scorecard analysis and StepSecurity GitHub Actions hardening.
- Added `vladopajic/go-test-coverage` for enforcing minimum coverage thresholds in CI.

---

## [v0.8] - 2024-06-21

### New Features

- New `stats` functionality (predecessor of the `stats` subcommand): reports disk
  usage statistics including file counts, directory counts, and total sizes.
- Parallel directory scanning with a semaphore-limited goroutine pool (~6× speedup
  on 8-core machines).
- Support for Linux and Windows; macOS not yet supported in this version.

### Changed Behavior

- Improved formatting of statistics output.
- Fixed Linux implementation of hard-link-aware file size tracking.

---

## [v0.7] - 2024-06-20

### Changed Behavior

- Rewrote the monthly backup filter algorithm to fix incorrect behaviour for
  edge cases around month boundaries
  ([issue #1](https://github.com/TomTonic/prune_backups/issues/1)).
- Removed dead code from the previous algorithm.

---

## [v0.6] - 2024-06-18

### New Features

- First automated test suite.
- `--version` flag reporting the build version.
- Multi-platform build targets (Linux, macOS, Windows, ARM).

### Changed Behavior

- Moved all test code out of the main program into separate `_test.go` files.
- Restructured internal code for better testability.

---

## [v0.5] - 2024-06-17

Initial release.

- Backup directory pruner: retains 24 hourly, 30 daily, and 119 monthly snapshots
  and moves older directories to a configurable deletion folder.
- Cross-platform support (Linux, macOS, Windows).
- Verbosity levels (`-v`, `-vv`).
- GitHub Actions CI workflow.
