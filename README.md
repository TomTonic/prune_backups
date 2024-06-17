# prune_backups

A small tool to prune a bunch of backup directories to the typical pattern of one per hour for a day, one per day for a month, and then one per month for 10 years.

## What does the tool do?

The tool takes one directory name as command line argument. The tool will look for subdirectories in the given directory that start with a certain naming pattern: YYYY-MM-DD_HH-mm. The tool interpretes these directory names as dates and keeps exactly one of these directories for the current hour, one for the last hour and so on. The tool will always keep the latest and **move** all other directories to a newly created subdirectory 'to_delete'. **The tool does not actually delete anything!**

## What will a pruned directory look like?

| <small>Directory name</small>      | <small>Directory name (cntd.)</small>      | <small>Directory name (cntd.)</small>      | <small>Directory name (cntd.)</small>      |
|---------------------|---------------------|---------------------|---------------------|
| <small><small>🟨 2024-06-17_09-49</small></small> | <small><small>🟨 2024-06-16_18-49</small></small> | <small><small>🟦 2024-06-09_23-49</small></small> | <small><small>🟦 2024-05-25_23-49</small></small> |
| <small><small>🟨 2024-06-17_08-49</small></small> | <small><small>🟨 2024-06-16_17-49</small></small> | <small><small>🟦 2024-06-08_23-49</small></small> | <small><small>🟦 2024-05-24_23-49</small></small> |
| <small><small>🟨 2024-06-17_07-49</small></small> | <small><small>🟨 2024-06-16_16-49</small></small> | <small><small>🟦 2024-06-07_23-49</small></small> | <small><small>🟦 2024-05-23_23-49</small></small> |
| <small><small>🟨 2024-06-17_06-49</small></small> | <small><small>🟨 2024-06-16_15-49</small></small> | <small><small>🟦 2024-06-06_23-49</small></small> | <small><small>🟦 2024-05-22_23-49</small></small> |
| <small><small>🟨 2024-06-17_05-49</small></small> | <small><small>🟨 2024-06-16_14-49</small></small> | <small><small>🟦 2024-06-05_23-49</small></small> | <small><small>🟦 2024-05-21_23-49</small></small> |
| <small><small>🟨 2024-06-17_04-49</small></small> | <small><small>🟨 2024-06-16_13-49</small></small> | <small><small>🟦 2024-06-04_23-49</small></small> | <small><small>🟦 2024-05-20_23-49</small></small> |
| <small><small>🟨 2024-06-17_03-49</small></small> | <small><small>🟨 2024-06-16_12-49</small></small> | <small><small>🟦 2024-06-03_23-49</small></small> | <small><small>🟦 2024-05-19_23-4</small></small> |
| <small><small>🟨 2024-06-17_02-49</small></small> | <small><small>🟨 2024-06-16_11-49</small></small> | <small><small>🟦 2024-06-02_23-49</small></small> | <small><small>🟦 2024-05-18_23-49</small></small> |
| <small><small>🟨 2024-06-17_01-49</small></small> | <small><small>🟨 2024-06-16_10-49</small></small> | <small><small>🟦 2024-06-01_23-49</small></small> | <small><small>🟦 2024-05-17_23-49</small></small> |
| <small><small>🟨 2024-06-17_00-49</small></small> | <small><small>🟦 2024-06-15_23-49</small></small> | <small><small>🟦 2024-05-31_23-49</small></small> | <small><small>🟩 2024-04-30_23-49</small></small> |
| <small><small>🟨 2024-06-16_23-49</small></small> | <small><small>🟦 2024-06-14_23-49</small></small> | <small><small>🟦 2024-05-30_23-49</small></small> | <small><small>🟩 2024-03-31_23-49</small></small> |
| <small><small>🟨 2024-06-16_22-49</small></small> | <small><small>🟦 2024-06-13_23-49</small></small> | <small><small>🟦 2024-05-29_23-49</small></small> | <small><small>🟩 2024-02-29_23-49</small></small> |
| <small><small>🟨 2024-06-16_21-49</small></small> | <small><small>🟦 2024-06-12_23-49</small></small> | <small><small>🟦 2024-05-28_23-49</small></small> | <small><small>🟪 to_delete</small></small> |
| <small><small>🟨 2024-06-16_20-49</small></small> | <small><small>🟦 2024-06-11_23-49</small></small> | <small><small>🟦 2024-05-27_23-49</small></small> | <small><small>🟫 some_other_filename</small></small>|
| <small><small>🟨 2024-06-16_19-49</small></small> | <small><small>🟦 2024-06-10_23-49</small></small> | <small><small>🟦 2024-05-26_23-49</small></small> | <small><small>🟫 **latest** -> 2024-06-17_09-49</small></small>|

## What do I need to run it?

You only need the golang-compiler once to build the executable. It is available on a wide variety of operating systems and processor architectures, including x86, x64, ARM32, ARM64, Windows, Linux, MacOS, etc.. See <https://go.dev/dl/>

## How do I run it?

Download the code and run "go build prune_backups.go" once. This will build a platform specific command line executable for you. You can run this executable on the shell or in scrips or in cron jobs... as you like.

## What is the exact naming pattern? And how do I change this?

The exact naming pattern is YYYY-MM-DD_HH-mm, where

* YYYY is the 4-digit year,
* MM the 2-digit&dagger; month,
* DD the 2-digit&dagger; day,
* HH the 2-digit&dagger; hour (24h format), and
* mm the 2-digit&dagger; minute of the time the backup was created.

You cannot change this pattern unless you change the code. However, the tool will also work when you don't have the minutes or hours in your directory names, i.e. **a naming pattern of YYYY-MM-DD is sufficient**. The tool will simply will not prune hourly backups for you in this case.

<sup>&dagger; Please be aware that the tool needs a trailing zero.</sup>
