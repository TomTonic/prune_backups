# Special Cases

This page gives additional details about the behavior of `prune_backups`.

The simple rule is that `prune_backups` looks for directories in the format `YYYY-MM-DD_HH` and moves those to the `to_delete`-directory that do not match a certain temporal pattern. The pattern is:

1) Keep the newest directory for each of the preceeding 24 hours (affecting backups from today and yesterday)
2) Keep the newest directory for each of the 30 days preceeding the days of today and yesterday (affecting backups from this month and/or the last month and/or the month before that)
3) Keep the newest directory for each of the 119 months preceeding the 30 days preceeding the days of today and yesterday

## Special case of additional 'keep-the-newest-of-the-day'

Let's assume the actual directory to be pruned contains fully valid backups directories for yesterday in the format `YYYY-MM-DD_HH`. It is possible that none of those directories matches any of the filters for rule 1) from above. **All** backup directories for yesterday would be pruned. Even though this is conforming with the above rules, it is counter-intuitive and would punch a whole into the sequence of daily backups.

This scenario can happen if there are only backups yesterday's folder that are older than 24h, i.e. there are some, but none matching the hourly filters for the past 24h. In this case, `prune_backups` will add an extra *keep-the-newest-of-the-day*-filter with the same semantics as rule 2) above for the day of yesterday. The result is that `prune_backups` will keep one backup for yesterday.

## Special case of additional 'keep-the-newest-of-the-month'

Let's assume the actual directory to be pruned contains fully valid backups directories for the past month in the format `YYYY-MM-DD`. It is possible that none of those directories matches any of the filters for rule 2) from above. **All** backup directories for the past month would be pruned. Even though this is conforming with the above rules, it is counter-intuitive and would punch a whole into the sequence of monthly backups.

The 24h-logic affects two days (today and yesterday). The 30 daily backups affect 30 days before that. This sums up to 32 days. There are two cases:

1) The 32 days **affect two months M0 and M1**. M0 is the month that contains today. M1 is the month where the 30st daily backup lies. Month M2 denotes the month before M1.
    * Assertion: The necessary part of M0 is completely covered by hourly and/or daily backups.
    * Case: **M1 is completely covered with hourly and/or daily backups.**
        - As only two months are covered, this is equivalent to the case where the end of the 30 days exactly falls together with the 'end' (i.e., the 1st) of the month.
        - This is the case iff
            * day(today) = 4 && daycount(M1) = 28, or
            * day(today) = 3 && daycount(M1) = 29, or
            * day(today) = 2 && daycount(M1) = 30, or
            * day(today) = 1 && daycount(M1) = 31
            * Alternative representation: day(today) + daycount(M1) = 32
        - In this case we **may not** use an extra "(only-)keep-the-newest-of-the-month"-filter.
        - The monthly filters start from M2.
    * Case: **M1 is not completely covered with daily backups.**
        - This is the case iff
            * day(today) > 4, or
            * day(today) = 4 && daycount(M1) > 28, or
            * day(today) = 3 && daycount(M1) > 29, or
            * day(today) = 2 && daycount(M1) > 30, or
            * day(today) = 1 && daycount(M1) > 31 (impossible)
            * Alternative representation: day(today) + daycount(M1) > 32
        - Assertion: M1 is only affected by daily filters: M1 is not completely covered with daily backups -> M1 has at least 1 day left uncovered -> M1 consumes at most 30 daily filters.
        - In this case we **need** an extra "(only-)keep-the-newest-of-the-month"-filter <=> (if and only if) there are no actual matches for daily filters in M1.
        - The monthly filters start from M2.
2) The 32 days **affect three months M0, M1, and M2**. M0 is the month that contains today. M2 is the month where the 30st daily backup lies. Month M3 denotes the month before M1.
    * This is the case iff
        - day(today) = 3 && daycount(M1) <= 28, or
        - day(today) = 2 && daycount(M1) <= 29, or
        - day(today) = 1 && daycount(M1) <= 30
        - Alternative representation: day(today) + daycount(M1) < 32
    * Assertion: The necessary part of M0 is completely covered by hourly and/or daily backups.
    * Assertion: M1 is a month with less than 31 days.
    * Assertion: M1 is completely covered by hourly and/or daily backups.
    * Assertion: M2 is not completely covered with hourly and/or daily backups.
    * Only one case: **M2 is not completely covered with daily backups.**
        - In this case we **nee**d an extra "(only-)keep-the-newest-of-the-month"-filter <=> (if and only if) there are no actual matches for daily filters in M2.
        - The monthly filters start from M3
