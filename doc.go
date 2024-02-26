/*
Package crong is a library for parsing cron expressions, calculating
the next run times of the expressions, and validating the expressions.

# Syntax

Supports standard cron syntax (see https://en.wikipedia.org/wiki/Cron),
as well as less standard expressions. For example, `5/10 4,5 * *` means
"every 10 minutes starting at the 5th minute of the hour, for hours 4 and 5."

Days of the week are indexed 0-6, with 0 being Sunday, and can be
referenced by name (SUN, MON, TUE, WED, THU, FRI, SAT) or by number.

Months are indexed 1-12, and can be referenced by
name (JAN, FEB, MAR, APR, MAY, JUN, JUL, AUG, SEP, OCT, NOV, DEC) or by number.

Cron macros supported:

	@yearly (or @annually) - Run once a year, midnight, Jan. 1
	@monthly - Run once a month, midnight, first of month
	@weekly - Run once a week, midnight between Saturday and Sunday
	@daily (or @midnight) - Run once a day, midnight
	@hourly - Run once an hour, beginning of hour

Other characters supported:

  - - any value
    , - value list separator
  - - range of values
    / - step values
    ? - no specific value (month, day of month, day of week only)
    L - last day of month (when used, must be used alone)
*/
package crong
