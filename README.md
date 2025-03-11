# Prune Backup Directories

[![Tests](https://github.com/TomTonic/prune_backups/actions/workflows/coverage.yml/badge.svg?branch=main)](https://github.com/TomTonic/prune_backups/actions/workflows/coverage.yml)
![Coverage](https://raw.githubusercontent.com/TomTonic/prune_backups/badges/.badges/main/coverage.svg)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9890/badge)](https://www.bestpractices.dev/projects/9890)

`prune_backups` is a lightweight tool to elegantly prune (cull, weed, thin, trim, purge - you name it) incremental backups created with rsync or other backup tools. It follows a typical pattern: one backup per hour for a day, one per day for a month, and one per month thereafter.

## What does the tool do?

The tool `prune_backups` takes a directory name as a command-line argument. It searches for subdirectories matching the naming pattern YYYY-MM-DD_HH-mm, interprets these names as dates, and retains only one directory per hour for the last day, one per day for the last month, and one per month beyond that. The latest directory within each matching time slot is kept. All other directories are moved to a subdirectory named 'to_delete'. The tool **MOVES** directories and **DOES NOT** actually delete anything!

## Installation

* Install the golang-compiler, available for various operating systems and architectures (x86, x64, ARM32, ARM64, Windows, Linux, MacOS, etc.). See <https://go.dev/dl/>.
* Download source code: `git clone https://github.com/TomTonic/prune_backups.git`.
* Compile the source code by navigating to the cloned directory and running `go build`.
* The executable is named `prune_backups` or `prune_backups.exe`, depending on your system. You can place it anywhere you like, as it is fully self-contained.
* You are ready to go!

## How do I run it?

On the command line: `prune_backups from /mnt/backups`

In a script (with context):

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

# uncomment the following line if you really want to delete the old backups
# rm -rf $my_backup_storage_dir/to_delete
```

You would run this script hourly via cron on your backup server to backup your web server.

## What will a pruned directory look like?

Scenario: This example assumes you have a cron job running hourly at the 49th minute, creating a separate backup directory each time (for example with `rsync --link-dest`). On June 17, 2024, at 09:54 in the morning, running `prune_backups` will leave your backup directory with the following structure:

| <small>Directory name</small>      | <small>Directory name (cntd.)</small>      | <small>Directory name (cntd.)</small>      | <small>Directory name (cntd.)</small>      |
|---------------------|---------------------|---------------------|---------------------|
| <small><small>🟨 2024-06-17_09-49/</small></small> | <small><small>🟨 2024-06-16_18-49/</small></small> | <small><small>🟦 2024-06-09_23-49/</small></small> | <small><small>🟦 2024-05-25_23-49/</small></small> |
| <small><small>🟨 2024-06-17_08-49/</small></small> | <small><small>🟨 2024-06-16_17-49/</small></small> | <small><small>🟦 2024-06-08_23-49/</small></small> | <small><small>🟦 2024-05-24_23-49/</small></small> |
| <small><small>🟨 2024-06-17_07-49/</small></small> | <small><small>🟨 2024-06-16_16-49/</small></small> | <small><small>🟦 2024-06-07_23-49/</small></small> | <small><small>🟦 2024-05-23_23-49/</small></small> |
| <small><small>🟨 2024-06-17_06-49/</small></small> | <small><small>🟨 2024-06-16_15-49/</small></small> | <small><small>🟦 2024-06-06_23-49/</small></small> | <small><small>🟦 2024-05-22_23-49/</small></small> |
| <small><small>🟨 2024-06-17_05-49/</small></small> | <small><small>🟨 2024-06-16_14-49/</small></small> | <small><small>🟦 2024-06-05_23-49/</small></small> | <small><small>🟦 2024-05-21_23-49/</small></small> |
| <small><small>🟨 2024-06-17_04-49/</small></small> | <small><small>🟨 2024-06-16_13-49/</small></small> | <small><small>🟦 2024-06-04_23-49/</small></small> | <small><small>🟦 2024-05-20_23-49/</small></small> |
| <small><small>🟨 2024-06-17_03-49/</small></small> | <small><small>🟨 2024-06-16_12-49/</small></small> | <small><small>🟦 2024-06-03_23-49/</small></small> | <small><small>🟦 2024-05-19_23-49/</small></small> |
| <small><small>🟨 2024-06-17_02-49/</small></small> | <small><small>🟨 2024-06-16_11-49/</small></small> | <small><small>🟦 2024-06-02_23-49/</small></small> | <small><small>🟦 2024-05-18_23-49/</small></small> |
| <small><small>🟨 2024-06-17_01-49/</small></small> | <small><small>🟨 2024-06-16_10-49/</small></small> | <small><small>🟦 2024-06-01_23-49/</small></small> | <small><small>🟦 2024-05-17_23-49/</small></small> |
| <small><small>🟨 2024-06-17_00-49/</small></small> | <small><small>🟦 2024-06-15_23-49/</small></small> | <small><small>🟦 2024-05-31_23-49/</small></small> | <small><small>🟩 2024-04-30_23-49/</small></small> |
| <small><small>🟨 2024-06-16_23-49/</small></small> | <small><small>🟦 2024-06-14_23-49/</small></small> | <small><small>🟦 2024-05-30_23-49/</small></small> | <small><small>🟩 2024-03-31_23-49/</small></small> |
| <small><small>🟨 2024-06-16_22-49/</small></small> | <small><small>🟦 2024-06-13_23-49/</small></small> | <small><small>🟦 2024-05-29_23-49/</small></small> | <small><small>🟩 2024-02-29_23-49/</small></small> |
| <small><small>🟨 2024-06-16_21-49/</small></small> | <small><small>🟦 2024-06-12_23-49/</small></small> | <small><small>🟦 2024-05-28_23-49/</small></small> | <small><small>🟪 to_delete/</small></small> |
| <small><small>🟨 2024-06-16_20-49/</small></small> | <small><small>🟦 2024-06-11_23-49/</small></small> | <small><small>🟦 2024-05-27_23-49/</small></small> | <small><small>🟫 some_other_directory/</small></small>|
| <small><small>🟨 2024-06-16_19-49/</small></small> | <small><small>🟦 2024-06-10_23-49/</small></small> | <small><small>🟦 2024-05-26_23-49/</small></small> | <small><small>🟫 **latest** -> 2024-06-17_09-49/</small></small>|

* 🟨 **Hourly:** Your backup directory will contain (up to) 24 directories for the last 24 hours. If multiple directories exist for a certain hour, `prune_backups` keeps the latest directory (determined by name, not by metadata) and moves the rest. If no directory exists for a certain hour, it is skipped. Please note that no extra hourly backups will be kept to compensate for missing hourly backups.
* 🟦 **Daily:** Your backup directory will contain (up to) 30 directories for the last 30 days. If multiple directories exist for a certain day, `prune_backups` keeps the latest directory (determined by name, not by metadata) and moves the rest. If no directory exists for a certain day, that day will be skipped. Please note that no extra daily backups will be kept to compensate for missing daily backups.
* 🟩 **Monthly:** Your backup directory will contain directories for each month beyond the last 30 days. If multiple directories exist for a certain month, `prune_backups` will keeps the latest directory (determined by name, not by metadata) and moves the rest. If no directory exists for a certain month, that month will be skipped.
* 🟪 **Pruned directories:** The directory `to_delete` is created by `prune_backups` in the backup directory; and it moves all pruned directories here. You can change the name of this directory with the `--to` parameter. Please note that this directory should reside in the same filesystem as your backup directory for performance reasons.
* 🟫 **Other:** Files, symlinks, or directories with other naming schemes will remain untouched.

## What is the exact naming pattern? And how do I change this?

The exact naming pattern is YYYY-MM-DD_HH-mm, where

* YYYY is the 4-digit year,
* MM the 2-digit&dagger; month,
* DD the 2-digit&dagger; day,
* HH the 2-digit&dagger; hour (24h format), and
* mm the 2-digit&dagger; minute of the time the backup was created.

You cannot change this pattern unless you change the golang code. However, the tool will also work when you don't have the minutes or hours in your directory names, i.e. **a naming pattern of YYYY-MM-DD is sufficient**. The tool will simply will not prune hourly backups in this case.

<sup>&dagger; Please note that the tool needs a trailing zero.</sup>
