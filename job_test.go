package crong

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduledJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	results := make(chan time.Time, 10)
	runCt := atomic.Int64{}
	sf := ScheduleFunc(
		ctx,
		s,
		ScheduledJobOptions{
			MaxConcurrent:        10,
			TickerReceiveTimeout: 5 * time.Second,
		},
		func(dt time.Time) error {
			runCt.Add(1)
			t.Logf("sending result: %s", dt)
			results <- dt
			return nil
		},
	)

	// push some ticks instead of waiting 60 seconds
	sf.ticker.tick(ctx)
	firstResult := <-results

	sf.ticker.tick(ctx)
	secondResult := <-results

	assertEqual(t, runCt.Load(), int64(2))

	// suspend the job, then resume and send a single tick,
	// to validate jobs weren't executed while suspended
	if suspended := sf.Suspend(); !suspended {
		t.Fatalf("expected to be suspended")
	}

	assertEqual(t, sf.State(), ScheduleSuspended)

	go sf.ticker.tick(ctx)
	go sf.ticker.tick(ctx)
	go sf.ticker.tick(ctx)

	time.Sleep(2 * time.Second)
	if resumed := sf.Resume(); !resumed {
		t.Fatalf("expected to be resumed")
	}

	sf.ticker.tick(ctx)
	thirdResult := <-results

	assertEqual(t, runCt.Load(), int64(3))
	assertEqual(t, sf.Runs.Load(), int64(3))

	stopped := sf.Stop(ctx)
	if !stopped {
		t.Fatalf("expected to be stopped")
	}

	rt := sf.Runtimes()
	if len(rt) != 3 {
		t.Fatalf("expected 3 runtimes, got %d", len(rt))
	}
	if !rt[0].Start.Equal(firstResult) {
		t.Fatalf(
			"expected Start time to be %s, got %s",
			firstResult,
			rt[0].Start,
		)
	}
	if !rt[1].Start.Equal(secondResult) {
		t.Fatalf(
			"expected Start time to be %s, got %s",
			secondResult,
			rt[1].Start,
		)
	}
	if !rt[2].Start.Equal(thirdResult) {
		t.Fatalf(
			"expected Start time to be %s, got %s",
			secondResult,
			rt[2].Start,
		)
	}

}

func TestScheduledContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	runCt := atomic.Int64{}

	ranCh := make(chan struct{}, 1)
	sj := NewScheduledJob(
		s,
		ScheduledJobOptions{
			MaxConcurrent:        1,
			TickerReceiveTimeout: 5 * time.Second,
		},
		func(dt time.Time) error {
			defer func() {
				ranCh <- struct{}{}
			}()
			runCt.Add(1)

			return nil
		},
	)

	sctx, scancel := context.WithCancel(ctx)
	defer scancel()

	go func() {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Fatalf("expected results")
			}
		case <-ranCh:
			sj.Stop(sctx)
		}
	}()
	go sj.ticker.tick(ctx)
	e := sj.Start(sctx)
	if e != nil {
		t.Fatalf("expected nil error")
	}
	assertEqual(t, sj.Runs.Load(), int64(1))
	assertEqual(t, runCt.Load(), int64(1))
	assertEqual(t, sj.State(), ScheduleStopped)

}

func TestJobFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	doneCh := make(chan struct{}, 1)
	sj := ScheduleFunc(
		ctx,
		s,
		ScheduledJobOptions{
			MaxConcurrent:        0,
			TickerReceiveTimeout: 5 * time.Second,
		},
		func(dt time.Time) error {
			defer func() {
				doneCh <- struct{}{}
			}()
			return errors.New("job failed")
		},
	)

	sj.ticker.tick(ctx)

	select {
	case <-ctx.Done():
		t.Fatalf("expected results")
	case <-doneCh:
		t.Logf("finished")
	}

	assertEqual(t, sj.Failures.Load(), int64(1))
	runtime := sj.Runtimes()
	if len(runtime) != 1 {
		t.Fatalf("expected 1 runtime, got %d", len(runtime))
	}
	if runtime[0].Error == nil {
		t.Fatalf("expected error, got nil")
	}

}

func TestPreviouslyStarted(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	doneCh := make(chan struct{}, 1)

	sj := ScheduleFunc(
		ctx,
		s,
		ScheduledJobOptions{
			MaxConcurrent:        0,
			TickerReceiveTimeout: 5 * time.Second,
		}, func(dt time.Time) error {
			defer func() {
				doneCh <- struct{}{}
			}()
			return errors.New("job failed")
		},
	)
	defer sj.Stop(context.Background())
	err = sj.Start(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}

}

func TestAlreadyStopped(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	doneCh := make(chan struct{}, 1)

	sj := ScheduleFunc(
		ctx,
		s,
		ScheduledJobOptions{
			MaxConcurrent:        0,
			TickerReceiveTimeout: 5 * time.Second,
		},
		func(dt time.Time) error {
			defer func() {
				doneCh <- struct{}{}
			}()
			return errors.New("job failed")
		},
	)
	assertEqual(t, sj.Stop(ctx), true)
	assertEqual(t, sj.state.Load(), int64(ScheduleStopped))
	err = sj.Start(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}

	assertEqual(t, sj.Stop(ctx), false)
	cancel()
	assertEqual(t, sj.Stop(ctx), false)

}

func TestJobMaxFailures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	sj := NewScheduledJob(
		s,
		ScheduledJobOptions{
			MaxConcurrent:        3,
			TickerReceiveTimeout: 5 * time.Second,
			MaxFailures:          3,
		},
		func(dt time.Time) error {
			return errors.New("job failed")
		},
	)

	go func() {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Fatalf("expected results")
			}
		}
	}()
	go sj.ticker.tick(ctx)
	go sj.ticker.tick(ctx)
	go sj.ticker.tick(ctx)

	err = sj.Start(ctx)
	if err != nil {
		t.Fatalf("expected error")
	}
	assertEqual(t, sj.Failures.Load(), int64(3))

	assertEqual(t, sj.State(), ScheduleStopped)
}

func TestJobConsecutiveFailures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	fail := atomic.Bool{}
	fail.Store(true)
	doneCh := make(chan struct{}, 10)
	sj := NewScheduledJob(
		s,
		ScheduledJobOptions{
			MaxConcurrent:          3,
			TickerReceiveTimeout:   5 * time.Second,
			MaxConsecutiveFailures: 3,
		},
		func(dt time.Time) error {
			defer func() {
				doneCh <- struct{}{}
			}()
			if fail.Load() {
				return errors.New("job failed")
			}
			return nil
		},
	)

	go func() {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Fatalf("expected results")
			}
		}
	}()

	stoppedCh := make(chan struct{}, 1)
	go func() {
		_ = sj.Start(ctx)
		stoppedCh <- struct{}{}

	}()

	sj.ticker.tick(ctx)
	sj.ticker.tick(ctx)

	select {
	case <-ctx.Done():
		t.Fatalf("expected results")
	case <-doneCh:
		t.Logf("finished")
	}

	select {
	case <-ctx.Done():
		t.Fatalf("expected results")
	case <-doneCh:
		t.Logf("finished")
	}

	assertEqual(t, sj.Failures.Load(), int64(2))

	assertEqual(t, sj.State(), ScheduleStarted)

	fail.Store(false)

	sj.ticker.tick(ctx)

	select {
	case <-ctx.Done():
		t.Fatalf("expected results")
	case <-doneCh:
		t.Logf("finished")
	}

	assertEqual(t, sj.Failures.Load(), int64(2))
	assertEqual(t, sj.Runs.Load(), int64(3))
	assertEqual(t, sj.State(), ScheduleStarted)

	fail.Store(true)
	sj.ticker.tick(ctx)
	sj.ticker.tick(ctx)
	sj.ticker.tick(ctx)

	select {
	case <-ctx.Done():
		t.Fatalf("expected results")
	case <-doneCh:
		t.Logf("finished")
	}

	select {
	case <-ctx.Done():
		t.Fatalf("expected results")
	case <-doneCh:
		t.Logf("finished")
	}

	select {
	case <-ctx.Done():
		t.Fatalf("expected results")
	case <-doneCh:
		t.Logf("finished")
	}
	<-stoppedCh
	assertEqual(t, sj.Failures.Load(), int64(5))
	assertEqual(t, sj.Runs.Load(), int64(6))
	assertEqual(t, sj.State(), ScheduleStopped)
}
