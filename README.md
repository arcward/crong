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
