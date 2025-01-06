# Tidy RSync Backup Directories

[![Automated Tests](https://github.com/TomTonic/prune_backups/actions/workflows/coverage.yml/badge.svg?branch=main)](https://github.com/TomTonic/prune_backups/actions/workflows/coverage.yml)
![Test Coverage](https://raw.githubusercontent.com/TomTonic/prune_backups/badges/.badges/main/coverage.svg)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9890/badge)](https://www.bestpractices.dev/projects/9890)

`prune_backups` is a small tool to tidy (prune) incremental backups created with rsync (and other backup tools) to the typical pattern of one per hour for a day, one per day for a month, and then one per month.

## What does the tool do?

The tool `prune_backups` takes one directory name as command line argument. It looks for subdirectories in that directory matching the naming pattern YYYY-MM-DD_HH-mm. The tool interpretes these directory names as dates and keeps exactly one of these directories for the current hour, one for the last hour and so on. The tool will always keep the latest and move all other directories into a subdirectory 'to_delete'. **The tool ONLY MOVES directories and DOES NOT ACTUALLY DELETE anything!**

## What will a pruned directory look like?

Scenario: This example assumes you have a cron-job running hourly in the 49th minute, each creating a separate backup directory (for example with `rsync --link-dest`). It is the 17th of June 2024 today, 09:54 in the morning when you run `prune_backups`. It will leave your backup directory (`-dir` parameter) with the following directory layout:

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

* 🟨 Your backup directory will contain (up to) 24 directories for the last 24h. If there are multiple directories for a certain hour, `prune_backups` will keep the latest directory (determined by name, not by metadata) and prune the other directories for this hour. If there is no backup directory for a certain hour, that hour will simply be skipped; i.e. there won't be any extra hourly backups appended for compensation after the 24h mark.
* 🟦 Your backup directory will contain (up to) 30 directories for the last 30 days. If there are multiple directories for a certain day, `prune_backups` will keep the last hourly directory (determined by name, not by metadata) and prune the other directories for this day. If there is no backup directory for a certain day, that day will simply be skipped; i.e. there won't be any extra daily backups appended for compensation after the 30 days mark.
* 🟩 Your backup directory will contain directories for each month before that. If there are multiple directories for a certain month, `prune_backups` will keep the last daily directory (determined by name, not by metadata) and prune the other directories for this month. If there is no backup directory for a certain month, that month will simply be skipped; i.e. there won't be any extra daily backups kept for compensation in neighboring months or any magic like that.
* 🟪 The directory `to_delete` is created by `prune_backups` in the backup directory; and it moves all pruned directories here. You can change the name of this directory with the `to_directory` parameter. Please note that this directory should reside in the same filesystem as your backup directory.
* 🟫 Files, softlinks, or directories with other naming schemes than `YYYY-MM-DD*` will remain untouched.

## What do I need to run it?

You only need the golang-compiler once to build the executable. It is available on a wide variety of operating systems and processor architectures, including x86, x64, ARM32, ARM64, Windows, Linux, MacOS, etc.. See <https://go.dev/dl/>

## How do I run it?

Download the code and run "go build prune_backups.go" once. This will build a platform specific command line executable for you. You can run this executable on the shell or in scrips or in cron jobs... as you like.

Example: `prune_backups -dir=/mnt/backups`

Longer Example (with context):

```Shell
#!/bin/sh

my_backup_storage_dir=/srv/backup/mywebserver
current_snapshot_dir=$(date +%Y-%m-%d_%H-%M)

rsync -avR --checksum --delete --link-dest=$my_backup_storage_dir/latest backupuser@mywebserver.example.com:/var/www $my_backup_storage_dir/_$current_snapshot_dir

cd $my_backup_storage_dir

# to make sure we can identify incomplete backups by their directory name, we started the directory name with an underscore character (_). now rename it to indicate it was complete.
mv _$current_snapshot_dir $current_snapshot_dir

# make the latest snapshot easily referable
ln -nsf $current_snapshot_dir latest

# prune old backups
prune_backups -dir=$my_backup_storage_dir

# uncomment the following line if you really want to delete the old backups
# rm -rf $my_backup_storage_dir/to_delete
```

(You would run this script hourly via cron on your backup server to backup your web server.)

## What is the exact naming pattern? And how do I change this?

The exact naming pattern is YYYY-MM-DD_HH-mm, where

* YYYY is the 4-digit year,
* MM the 2-digit&dagger; month,
* DD the 2-digit&dagger; day,
* HH the 2-digit&dagger; hour (24h format), and
* mm the 2-digit&dagger; minute of the time the backup was created.

You cannot change this pattern unless you change the code. However, the tool will also work when you don't have the minutes or hours in your directory names, i.e. **a naming pattern of YYYY-MM-DD is sufficient**. The tool will simply will not prune hourly backups for you in this case.

<sup>&dagger; Please be aware that the tool needs a trailing zero.</sup>
