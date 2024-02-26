package crong

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Ticker is a cron ticker that sends the current time
// on the Ticker.C channel when the schedule is triggered
type Ticker struct {
	schedule *Schedule
	C        chan time.Time
	tickCh   chan time.Time
	stop     chan struct{}
	// sendTimeout is the maximum time to wait for a receiver
	// to send a tick on the Ticker.C channel
	sendTimeout time.Duration

	firstTick time.Time
	lastTick  time.Time

	ticksSeen    atomic.Int64
	ticksSent    atomic.Int64
	ticksDropped atomic.Int64
	mu           sync.Mutex

	// cronTicker *time.Ticker
}

// NewTicker creates a new Ticker from a cron expression,
// sending the current time on Ticker.C when the schedule
// is triggered.
// It works similarly to [time.Ticker](https://golang.org/pkg/time/#Ticker),
// but is granular only to the minute. sendTimeout is the maximum time to wait
// for a receiver to send a tick on the Ticker.C channel (this differs from
// [time.Ticker], allowing some wiggle room for slow receivers).
// If the provided context is canceled, the ticker will stop automatically.
func NewTicker(
	ctx context.Context,
	schedule *Schedule,
	sendTimeout time.Duration,
) *Ticker {
	t := &Ticker{
		schedule:    schedule,
		C:           make(chan time.Time),
		stop:        make(chan struct{}, 1),
		tickCh:      make(chan time.Time),
		mu:          sync.Mutex{},
		sendTimeout: sendTimeout,
	}

	ctx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			select {
			case <-t.stop:
				slog.Info("ticker stopped, canceling")
				cancel()
				return
			case <-ctx.Done():
				t.Stop()
			}
		}
	}()

	wg.Add(1)
	go func() {
		wg.Done()
		t.tickOnSchedule(ctx)
	}()

	slog.Info("waiting for initial tick")
	init := <-t.tickCh
	slog.Info("initial tick", "time", init)
	wg.Add(1)
	go func() {
		wg.Done()
		t.run(ctx)
	}()
	wg.Wait()

	return t
}

func (t *Ticker) Stop() {
	select {
	case t.stop <- struct{}{}:
		//
	default:
		//
	}
}

// tickOnSchedule sends a tick when the current time matches
// the next scheduled time. The time is checked every minute.
// This is used instead of a [time.Ticker] to avoid drift.
func (t *Ticker) tickOnSchedule(ctx context.Context) {
	loc := t.schedule.loc
	t.tickCh <- time.Now().In(t.schedule.loc)
	nextTime := t.schedule.nextNoTruncate(time.Now().In(loc).Truncate(time.Minute))
	sleepDone := make(chan struct{}, 1)
	slog.Info("starting tick on schedule", "next_time", nextTime)
	for ctx.Err() == nil {
		now := time.Now().In(t.schedule.loc)
		if timesEqualToMinute(now, nextTime) {
			slog.Info("saw tick", "next_time", nextTime, "now", now)
			t.tick(ctx)
			nextTime = t.schedule.nextNoTruncate(
				time.Now().In(loc).Truncate(time.Minute),
			)
		}

		nextMinute := time.Now().Add(time.Minute).Truncate(time.Minute)
		untilNextMinute := nextMinute.Sub(time.Now())
		sleepDuration := untilNextMinute + (1 * time.Second)

		slog.Info(
			"sleeping",
			"duration",
			sleepDuration,
			"next_time",
			nextTime,
			"now",
			now,
			"until_next_minute",
			untilNextMinute,
		)
		go func() {
			time.Sleep(sleepDuration)
			sleepDone <- struct{}{}
		}()
		select {
		case <-ctx.Done():
			return
		case <-sleepDone:
			//
		}
	}
}

// run waits for ticks on the tick channel and sends
// them on the Ticker.C channel, then schedules the
// next tick
func (t *Ticker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("ticker stopped, breaking")
			return
		case currentTick := <-t.tickCh:
			slog.Info(
				"schedule triggered",
				"current_tick",
				currentTick,
				"schedule",
				t.schedule,
			)
			tctx, tcancel := context.WithTimeout(ctx, t.sendTimeout)
			select {
			case t.C <- currentTick:
				t.ticksSent.Add(1)
				slog.Info("sent tick")
			case <-tctx.Done():
				slog.Warn("dropped tick")
				t.ticksDropped.Add(1)
			}
			tcancel()
		}
	}
}

// tick sends a tick on the tick channel
func (t *Ticker) tick(ctx context.Context) bool {
	nt := time.Now().In(t.schedule.loc)
	select {
	case <-ctx.Done():
		return false
	case t.tickCh <- nt:
		slog.Info("sent tick", "tick", nt)
		t.ticksSeen.Add(1)

		t.mu.Lock()
		defer t.mu.Unlock()
		t.lastTick = nt
		if t.firstTick.IsZero() {
			t.firstTick = nt
		}
		return true
	}
}

func timesEqualToMinute(t1, t2 time.Time) bool {
	return t1.Truncate(time.Minute).Equal(t2.Truncate(time.Minute))
}
