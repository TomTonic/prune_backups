# Prune Backup Directories

[![Tests](https://github.com/TomTonic/prune_backups/actions/workflows/coverage.yml/badge.svg?branch=main)](https://github.com/TomTonic/prune_backups/actions/workflows/coverage.yml)
![Coverage](https://raw.githubusercontent.com/TomTonic/prune_backups/badges/.badges/main/coverage.svg)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9890/badge)](https://www.bestpractices.dev/projects/9890)

`prune_backups` is a small tool to prune (cull, weed, thin, trim, purge - you name it) incremental backups created with rsync or other backup tools. It follows a typical pattern: one backup per hour for a day, one per day for a month, and one per month thereafter.

## What does the tool do?

`prune_backups` takes a directory and searches for subdirectories matching the naming pattern YYYY-MM-DD_HH-mm. It interprets these names as dates, and retains only one directory per hour for the last day, one per day for the last month, and one per month beyond that. The latest directory within each matching time slot is kept. All other directories are moved into a subdirectory; usually named 'to_delete'. For safety and speed, the tool only **MOVES** directories and **DOES NOT** actually delete anything.

## Installation

### Linux Packages (deb, rpm, apk, Arch)

Every [release](https://github.com/TomTonic/prune_backups/releases) also ships ready-to-install packages for `amd64` and `arm64`, built directly from the same static, dependency-free Go binary described below. Since the binary has no C library dependency (no cgo), one package per architecture works across all versions of the corresponding distro family that are still supported by the Go toolchain used to build it - there is no need to pick a package per Ubuntu/Debian/Fedora version.

* **Debian / Ubuntu** (and derivatives such as Linux Mint, Pop!_OS, Raspberry Pi OS): download the `.deb` asset, then run:

  ```Shell
  sudo apt install ./prune_backups_<version>_<amd64|arm64>.deb
  ```

* **Fedora / RHEL / CentOS / Rocky / AlmaLinux**: download the `.rpm` asset, then run:

  ```Shell
  sudo dnf install ./prune_backups-<version>-1.<x86_64|aarch64>.rpm
  ```

* **openSUSE / SLES**: same `.rpm` asset, installed with:

  ```Shell
  sudo zypper install ./prune_backups-<version>-1.<x86_64|aarch64>.rpm
  ```

* **Alpine Linux**: download the `.apk` asset, then run:

  ```Shell
  sudo apk add --allow-untrusted ./prune_backups_<version>_<x86_64|aarch64>.apk
  ```

* **Arch Linux** (and derivatives such as Manjaro, EndeavourOS): download the `.pkg.tar.zst` asset, then run:

  ```Shell
  sudo pacman -U ./prune_backups-<version>-1-<x86_64|aarch64>.pkg.tar.zst
  ```

All of these install the `prune_backups` binary to `/usr/bin/prune_backups`, so it is immediately available on your `PATH`.

### Binary Distribution

You can download fully self-contained executable binaries of `prune_backups` for various platforms from the [releases page of this project](https://github.com/TomTonic/prune_backups/releases). Just navigate to the 'Assets' section of the latest release and download the appropriate binary. To run it just **rename it to 'prune_backups' and mark it as executable** (use [`chmod u+x`](https://man7.org/linux/man-pages/man1/chmod.1p.html) on Linux and MacOS). On Windows and MacOS you will have to **mark the binaries as save for execution**, as they are not digitally signed. These binaries are build on the GitHub servers by the GitHub Actions code in this repository. If you don't trust them either, you can easily build your own binaries from the source code (see next section).

### Build your own Executables

* Install the Go compiler, if you do not yet have it (check with `go version` on the command line). The Go compiler is available for free for many operating systems and architectures (x86, x64, ARM32, ARM64, Windows, Linux, MacOS, etc.), see <https://go.dev/dl/>.
* Download source code: `git clone https://github.com/TomTonic/prune_backups.git`
* Compile the source: `go build`
* Optionally run the tests: `go test -cover`
* The executable is named `prune_backups` or `prune_backups.exe`, depending on your system. You can place it anywhere you like, it is fully self-contained.
* You are ready to go!

## How do I run it?

On the command line, run `prune_backups from /mnt/backups` to prune your backups folder. Alternatively, call `prune_backups stats /mnt/backups` to get statistics about its contents (it may take a while to collect the information though!).

### Planned Execution

A typical backup script would look as follows. You would, for example, run this script hourly using a cron job on your backup server to backup your web server.

```Shell
#!/bin/sh

# this is the target directory where all backups shall be stored. adapt to your needs
my_backup_storage_dir=/srv/backup/mywebserver

# create the directory name for the current backup. this must match the naming scheme of prune_backups
current_snapshot_dir=$(date +%Y-%m-%d_%H-%M)

# the rsync command will do the actual backup of the directory /var/www from the server mywebserver.example.com, logging into this machine with the user backupuser.
# --link-dest=$my_backup_storage_dir/latest will ensure rsync creates hardlinks for identical files, so more diskspace is only needed for new/changed files
# see rsync documentation.
# to make sure we can identify incomplete backups by their directory name, we start the directory name with an underscore character (_).
rsync -avR --checksum --delete --link-dest=$my_backup_storage_dir/latest backupuser@mywebserver.example.com:/var/www $my_backup_storage_dir/_$current_snapshot_dir

cd $my_backup_storage_dir

# to make sure we can identify incomplete backups by their directory name, we started the directory name with an underscore character (_). now rename it to indicate it was complete.
mv _$current_snapshot_dir $current_snapshot_dir

# make the latest snapshot easily referable for next incremental backup. see above
ln -nsf $current_snapshot_dir latest

# prune old backups
prune_backups from $my_backup_storage_dir

# uncomment the following line if you REALLY want to DELETE the old backups
# rm -rf $my_backup_storage_dir/to_delete
```

## What will a pruned directory look like?

Scenario: This example assumes you have a cron job running hourly at the 49th minute, creating a separate backup directory each time (for example with `rsync --link-dest`). On June 17, 2024, at 09:54 in the morning, running `prune_backups` will leave your backup directory with the following structure:

| Directory name      | Directory name (cntd.)      | Directory name (cntd.)      | Directory name (cntd.)      |
|---------------------|---------------------|---------------------|---------------------|
| 🟨 2024-06-17_09-49/ | 🟨 2024-06-16_18-49/ | 🟦 2024-06-09_23-49/ | 🟦 2024-05-25_23-49/ |
| 🟨 2024-06-17_08-49/ | 🟨 2024-06-16_17-49/ | 🟦 2024-06-08_23-49/ | 🟦 2024-05-24_23-49/ |
| 🟨 2024-06-17_07-49/ | 🟨 2024-06-16_16-49/ | 🟦 2024-06-07_23-49/ | 🟦 2024-05-23_23-49/ |
| 🟨 2024-06-17_06-49/ | 🟨 2024-06-16_15-49/ | 🟦 2024-06-06_23-49/ | 🟦 2024-05-22_23-49/ |
| 🟨 2024-06-17_05-49/ | 🟨 2024-06-16_14-49/ | 🟦 2024-06-05_23-49/ | 🟦 2024-05-21_23-49/ |
| 🟨 2024-06-17_04-49/ | 🟨 2024-06-16_13-49/ | 🟦 2024-06-04_23-49/ | 🟦 2024-05-20_23-49/ |
| 🟨 2024-06-17_03-49/ | 🟨 2024-06-16_12-49/ | 🟦 2024-06-03_23-49/ | 🟦 2024-05-19_23-49/ |
| 🟨 2024-06-17_02-49/ | 🟨 2024-06-16_11-49/ | 🟦 2024-06-02_23-49/ | 🟦 2024-05-18_23-49/ |
| 🟨 2024-06-17_01-49/ | 🟨 2024-06-16_10-49/ | 🟦 2024-06-01_23-49/ | 🟦 2024-05-17_23-49/ |
| 🟨 2024-06-17_00-49/ | 🟦 2024-06-15_23-49/ | 🟦 2024-05-31_23-49/ | 🟩 2024-04-30_23-49/ |
| 🟨 2024-06-16_23-49/ | 🟦 2024-06-14_23-49/ | 🟦 2024-05-30_23-49/ | 🟩 2024-03-31_23-49/ |
| 🟨 2024-06-16_22-49/ | 🟦 2024-06-13_23-49/ | 🟦 2024-05-29_23-49/ | 🟩 2024-02-29_23-49/ |
| 🟨 2024-06-16_21-49/ | 🟦 2024-06-12_23-49/ | 🟦 2024-05-28_23-49/ | 🟪 to_delete/ |
| 🟨 2024-06-16_20-49/ | 🟦 2024-06-11_23-49/ | 🟦 2024-05-27_23-49/ | 🟫 some_other_directory/|
| 🟨 2024-06-16_19-49/ | 🟦 2024-06-10_23-49/ | 🟦 2024-05-26_23-49/ | 🟫 **latest** -> 2024-06-17_09-49/|

* 🟨 **Hourly:** Your backup directory will contain (up to) 24 directories for the last 24 hours. If multiple directories exist for a certain hour, `prune_backups` keeps the latest directory (determined by name, not by metadata) and moves the rest. If no directory exists for a certain hour, it is skipped. Please note that no extra hourly backups will be kept to compensate for missing hourly backups.
* 🟦 **Daily:** Your backup directory will contain (up to) 30 directories for the last 30 days. If multiple directories exist for a certain day, `prune_backups` keeps the latest directory (determined by name, not by metadata) and moves the rest. If no directory exists for a certain day, that day will be skipped. Please note that no extra daily backups will be kept to compensate for missing daily backups.
* 🟩 **Monthly:** Your backup directory will contain directories for each month beyond the last 30 days. If multiple directories exist for a certain month, `prune_backups` will keeps the latest directory (determined by name, not by metadata) and moves the rest. If no directory exists for a certain month, that month will be skipped.
* 🟪 **Pruned directories:** The directory `to_delete` is created by `prune_backups` in the backup directory; and it moves all pruned directories here. You can change the name of this directory with the `--to` parameter. Please note that this directory should reside in the same filesystem as your backup directory for performance reasons.
* 🟫 **Other:** Files, symlinks, or directories with other naming schemes will remain untouched.

## What is the exact naming pattern? And how do I change this?

The exact naming pattern is YYYY-MM-DD_HH-mm, where

* YYYY is the 4-digit year,
* MM the 2-digit month,
* DD the 2-digit day,
* HH the 2-digit hour (24h format), and
* mm the 2-digit minute of the time the backup was created.

Use the [`date`](https://man7.org/linux/man-pages/man1/date.1.html) command in Linux to create according directory names:

```Shell
date +%Y-%m-%d_%H-%M
```

You cannot change this pattern unless you change the golang code. However, the tool will also work when you don't have the minutes or hours in your directory names, i.e. **a naming pattern of YYYY-MM-DD is sufficient**. The tool will simply will not prune hourly backups in this case.
