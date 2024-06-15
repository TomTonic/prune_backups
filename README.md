# prune_backups
A small tool to prune a bunch of backup directories to the typical pattern of one per hour for a day, one per day for a month, and then one per month for 10 years.

## What does the tool do?
The tool takes one directory name as command line argument. The tool will look for subdirectories in the given directory that start with a certain naming pattern: YYYY-MM-DD_HH-mm. The tool interpretes these directory names as dates and keeps exactly one of these directories for the current hour, one for the last hour and so on. The tool will always keep the latest and **move** all other directories to a newly created subdirectory 'to_delete'. **The tool does not actually delete anything!**

## What do I need to run it?
You only need the golang-compiler once to build the executable. It is available on a wide variety of operating systems and processor architectures, including x86, x64, ARM32, ARM64, Windows, Linux, MacOS, etc.. See https://go.dev/dl/

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

<sup><sub>&dagger; Please be aware that the tool needs a trailing zero.</sub></sup>
