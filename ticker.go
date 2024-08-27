package crong

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Logger used by [Ticker] and [ScheduledJob]. By default, it discards all logs.
var Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

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
				Logger.Debug("ticker stopped, canceling", "ticker", t)
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

	Logger.Debug("waiting for initial tick", "ticker", t)
	init := <-t.tickCh
	Logger.Debug("initial tick", "time", init, "ticker", t)
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
	Logger.Debug(
		"starting tick on schedule",
		"next_time", nextTime,
		"ticker", t,
	)
	for ctx.Err() == nil {
		now := time.Now().In(t.schedule.loc)
		if timesEqualToMinute(now, nextTime) {
			Logger.Debug(
				"saw tick",
				"next_time", nextTime,
				"now", now,
				"ticker", t,
			)
			t.tick(ctx)
			nextTime = t.schedule.nextNoTruncate(
				time.Now().In(loc).Truncate(time.Minute),
			)
		}

		nextMinute := time.Now().Add(time.Minute).Truncate(time.Minute)
		untilNextMinute := nextMinute.Sub(time.Now())
		sleepDuration := untilNextMinute + (1 * time.Second)

		Logger.Info(
			"sleeping",
			"duration", sleepDuration,
			"next_time", nextTime,
			"now", now,
			"until_next_minute", untilNextMinute,
			"ticker", t,
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
			Logger.Debug("ticker stopped, breaking", "ticker", t)
			return
		case currentTick := <-t.tickCh:
			Logger.Debug(
				"schedule triggered",
				"current_tick", currentTick,
				"ticker", t,
			)
			tctx, tcancel := context.WithTimeout(ctx, t.sendTimeout)
			select {
			case t.C <- currentTick:
				t.ticksSent.Add(1)
				Logger.Debug("sent tick", "ticker", t)
			case <-tctx.Done():
				Logger.Debug("dropped tick", "ticker", t)
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
		Logger.Info("sent tick", "tick", nt, "ticker", t)
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

func (t Ticker) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("schedule", t.schedule.String()),
		slog.Group(
			"ticks",
			"seen", t.ticksSeen.Load(),
			"sent", t.ticksSent.Load(),
			"dropped", t.ticksDropped.Load(),
		),
	)
}

func timesEqualToMinute(t1, t2 time.Time) bool {
	return t1.Truncate(time.Minute).Equal(t2.Truncate(time.Minute))
}
