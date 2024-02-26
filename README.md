# crong

**Cron**g: Lightweight, straightforward cron expression parser, ticker
and task scheduler for your Golang projects.

## Usage

Create a schedule from a cron expression, calculate future/past schedules:

```go
package main

import (
	"fmt"
	"log"
	"time"
	"github.com/arcward/crong"
)

func main() {
	schedule, err := crong.New("0 0 * * *", time.UTC)
	if err != nil {
		log.Fatal(err)
    	}
	
	// Get the next (or most recent) scheduled time relative to the given time
	fmt.Println("Next scheduled time:", schedule.Next(time.Now()))
	fmt.Println("Previous scheduled time:", schedule.Prev(time.Now()))
    
	// Check if the given time satisfies the schedule
	if schedule.Matches(time.Now()) {
		fmt.Println("It's time!")
    	}
}
```

Create a ticker that sends a tick on a channel whenever the cron
schedule fires, similar to `time.Ticker`:

```go
package main

import (
	"context"
	"log"
	"time"
	"github.com/arcward/crong"
)

func main() {
	schedule, err := crong.New("@hourly", time.UTC)
	if err != nil {
		log.Fatal(err)
    	}
	
	ticker := crong.NewTicker(context.Background(), schedule, 1 * time.Minute)
	defer ticker.Stop()
	
	select {
	case t := <-ticker.C:
		log.Printf("%s: Tick!", t)
    	}
}
```

Schedule a function to run whenever the schedule fires:

```go
package main

import (
	"context"
	"log"
	"time"
	"github.com/arcward/crong"
)

func main() {
	schedule, err := crong.New("* * * * *", time.UTC)
	if err != nil {
		log.Fatal(err)
    	}
	
	// MaxConcurrent=0 only allows the job to run sequentially, while
	// increasing TickerReceiveTimeout can accommodate potentially long-running
	// jobs, where you may not want the next tick to be dropped immediately.
	// MaxConsecutiveFailures=10 will stop executing the given function if it
	// returns a non-nil error ten times in a row.
	opts := &crong.ScheduledJobOptions{
		MaxConcurrent:        0,
		TickerReceiveTimeout: 30 * time.Second,
		MaxConsecutiveFailures: 10,
	}
	scheduledJob := crong.ScheduleFunc(
		context.Background(),
		schedule,
		opts,
		func(t time.Time) error {
			log.Printf("Scheduled run for %s started at %s", t, time.Now())
			return nil
        	},
	)
	defer scheduledJob.Stop(context.Background())
}
```

## Syntax

Supports standard cron syntax (see https://en.wikipedia.org/wiki/Cron), as well as less standard expressions. 
For example, `5/10 4,5 * *` means "every 10 minutes starting at the 5th minute of the hour, for hours 4 and 5."

Days of the week are indexed 0-6, with 0 being Sunday, and can be referenced by 
name (SUN, MON, TUE, WED, THU, FRI, SAT) or by number.

Months are indexed 1-12, and can be referenced by name 
(JAN, FEB, MAR, APR, MAY, JUN, JUL, AUG, SEP, OCT, NOV, DEC) or by number.

Cron macros supported:

  - `@yearly` (or `@annually`) - Run once a year, midnight, Jan. 1
  - `@monthly` - Run once a month, midnight, first of month
  - `@weekly` - Run once a week, midnight between Saturday and Sunday
  - `@daily` (or `@midnight`) - Run once a day, midnight
  - `@hourly` - Run once an hour, beginning of hour

Other characters supported:

  - `*`: Wildcard/Any value (ex: Every minute: `* * * * *`)
  - `,`: Value list separator (ex: Minute 0 and 30 of every hour: `0,30 * * * *`)
  - `-`: Range of values (ex: Minute 0-15 of every hour: `0-15 * * * *`)
  - `/`: Step values (ex: Every 2nd minute from minute 10-20 of every hour: `10-20/2 * * * *`)
  - `?`: No specific value (month, day of month, day of week only)
  - `L`: Last day of month. When used, must be used alone in the day
    field (ex: 12:30 on the last day of every month: `30 12 L * *`)
