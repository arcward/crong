package crong

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type ScheduleState int64

const (
	ScheduleStarted ScheduleState = iota + 1
	ScheduleSuspended
	ScheduleStopped
)

type ScheduledJobOptions struct {
	// MaxConcurrent is the maximum number of concurrent job executions.
	// If 0, there is no limit
	MaxConcurrent int

	// TickerReceiveTimeout is the maximum time the job's ticker will
	// wait for the job to receive a tick on the Ticker.C channel
	TickerReceiveTimeout time.Duration

	// MaxFailures is the maximum number of times the job can fail
	// before it is stopped. 0=no limit
	MaxFailures int

	// MaxConsecutiveFailures is the maximum number of consecutive
	// times the job can fail before it is stopped. 0=no limit
	MaxConsecutiveFailures int
}

func (s ScheduledJobOptions) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("max_concurrent", s.MaxConcurrent),
		slog.Int("max_failures", s.MaxFailures),
		slog.Int("max_consecutive_failures", s.MaxConsecutiveFailures),
		slog.Duration("ticker_receive_timeout", s.TickerReceiveTimeout),
	)
}

// ScheduledJob is a function that runs on Ticker ticks
// for a Schedule
type ScheduledJob struct {
	schedule *Schedule
	ticker   *Ticker
	f        func(t time.Time) error
	runtimes []*JobRuntime
	mu       sync.RWMutex
	stopCh   chan struct{}

	// Failures is the number of times the job has failed
	Failures atomic.Int64

	// ConsecutiveFailures is the number of times the job has failed in a row
	ConsecutiveFailures atomic.Int64

	// Runs is the number of times the job has run
	Runs atomic.Int64

	// Running is the number of times the job is currently running
	Running atomic.Int64

	state             atomic.Int64
	previouslyStarted atomic.Bool
	startMu           sync.Mutex
	options           ScheduledJobOptions
}

func NewScheduledJob(
	schedule *Schedule,
	opts ScheduledJobOptions,
	f func(t time.Time) error,
) *ScheduledJob {
	job := &ScheduledJob{
		schedule: schedule,
		ticker: NewTicker(
			context.Background(),
			schedule,
			opts.TickerReceiveTimeout,
		),
		f:        f,
		runtimes: make([]*JobRuntime, 0),
		stopCh:   make(chan struct{}, 1),
		options:  opts,
	}

	return job
}

func (s ScheduledJob) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("schedule", s.schedule.String()),
		slog.Group(
			"options", slog.Int("max_concurrent", s.options.MaxConcurrent),
			slog.Int("max_failures", s.options.MaxFailures),
			slog.Int(
				"max_consecutive_failures",
				s.options.MaxConsecutiveFailures,
			),
			slog.Duration(
				"ticker_receive_timeout",
				s.options.TickerReceiveTimeout,
			),
		),
		slog.Int64("failures", s.Failures.Load()),
		slog.Int64("consecutive_failures", s.ConsecutiveFailures.Load()),
		slog.Int64("runs", s.Runs.Load()),
		slog.Int64("running", s.Running.Load()),
	)
}

// ScheduleFunc creates and starts a new ScheduledJob with the given schedule and options.
// It immediately begins executing the provided function according to the schedule.
//
// The function f will be called with the current time whenever the schedule is triggered.
// If f returns an error, it will be recorded in the job's runtime history.
//
// Parameters:
//   - ctx: A context.Context for cancellation and timeout control.
//   - schedule: A *Schedule that determines when the job should run.
//   - opts: ScheduledJobOptions to configure the job's behavior.
//   - f: A function to be executed on each scheduled tick, with the signature func(time.Time) error.
//
// Returns:
//   - *ScheduledJob: A pointer to the newly created and started ScheduledJob.
//
// The returned ScheduledJob is already running and does not need to be started manually.
// Use the returned ScheduledJob's methods (e.g., Stop, Suspend, Resume) to control its execution.
//
// Example:
//
//	schedule, _ := crong.New("*/5 * * * *", nil)
//	job := crong.ScheduleFunc(ctx, schedule, crong.ScheduledJobOptions{
//		MaxConcurrent: 1,
//		TickerReceiveTimeout: 5 * time.Second,
//	}, func(t time.Time) error {
//		fmt.Printf("Job ran at %v\n", t)
//		return nil
//	})
//
//	// ... later
//	job.Stop(context.Background())
func ScheduleFunc(
	ctx context.Context,
	schedule *Schedule,
	opts ScheduledJobOptions,
	f func(t time.Time) error,
) *ScheduledJob {
	s := &ScheduledJob{
		schedule:          schedule,
		ticker:            NewTicker(ctx, schedule, opts.TickerReceiveTimeout),
		f:                 f,
		runtimes:          make([]*JobRuntime, 0),
		stopCh:            make(chan struct{}, 1),
		state:             atomic.Int64{},
		previouslyStarted: atomic.Bool{},
		options:           opts,
	}
	s.state.Store(int64(ScheduleStarted))
	s.previouslyStarted.Store(true)

	go func() {
		_ = s.start(ctx)
	}()
	return s
}

func (s *ScheduledJob) Start(ctx context.Context) error {
	if ScheduleState(s.state.Load()) == ScheduleStopped {
		return errors.New("cannot start a job that has been stopped")
	}

	if s.previouslyStarted.Load() {
		return errors.New("job has already been started")
	}

	return s.start(ctx)
}

// Stop stops job execution. After Stop is called, the job cannot be
// restarted.
func (s *ScheduledJob) Stop(ctx context.Context) bool {
	select {
	case <-ctx.Done():
	case s.stopCh <- struct{}{}:
		//
	}
	old := s.state.Swap(int64(ScheduleStopped))
	if old == int64(ScheduleStopped) {
		return false
	}
	return true
}

// Suspend pauses job execution until Resume is called
func (s *ScheduledJob) Suspend() bool {
	return s.state.CompareAndSwap(
		int64(ScheduleStarted),
		int64(ScheduleSuspended),
	)
}

// Resume resumes job execution after a call to Suspend
func (s *ScheduledJob) Resume() bool {
	return s.state.CompareAndSwap(
		int64(ScheduleSuspended),
		int64(ScheduleStarted),
	)
}

// Runtimes returns a slice of the job's runtimes
func (s *ScheduledJob) Runtimes() []*JobRuntime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runtimes[:]
}

func (s *ScheduledJob) State() ScheduleState {
	return ScheduleState(s.state.Load())
}

// Start starts the job. If the job has already been started,
// it returns an error. If the job has been stopped, it returns an error.
func (s *ScheduledJob) start(ctx context.Context) error {
	s.mu.Lock()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.state.Store(int64(ScheduleStarted))

	defer s.ticker.Stop()
	s.previouslyStarted.Store(true)
	s.mu.Unlock()
	wg := sync.WaitGroup{}

	// Waits for a stop signal, then cancels the context
	wg.Add(1)
	go func() {
		defer s.state.Store(int64(ScheduleStopped))
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			cancel()
			return
		}
	}()

	var jobCh chan time.Time

	if s.options.MaxConcurrent > 0 {
		jobCh = make(chan time.Time)
		defer close(jobCh)
		for i := 0; i < s.options.MaxConcurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case rt := <-jobCh:
						s.execute(rt)
					}
				}
			}()
		}
	}

	// Waits for ticks on the Ticker.C channel, then
	// executes the job
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case rt := <-s.ticker.C:
				switch {
				case ScheduleState(s.state.Load()) == ScheduleSuspended:
					Logger.Debug(
						"execution suspended, skipping tick",
						"scheduled_job", s,
						"tick", rt,
					)
				case jobCh == nil:
					wg.Add(1)
					go func() {
						defer wg.Done()
						s.execute(rt)
					}()
				default:
					jobCh <- rt
				}
			}

		}
	}()
	wg.Wait()
	return nil
}

func (s *ScheduledJob) execute(rt time.Time) {
	s.Runs.Add(1)

	s.Running.Add(1)
	defer s.Running.Add(-1)

	s.mu.Lock()
	defer s.mu.Unlock()

	runtime := &JobRuntime{Start: rt}

	Logger.Info("running scheduled job", "scheduled_job", s)

	runtime.Error = s.f(rt)
	if runtime.Error == nil {
		s.ConsecutiveFailures.Store(0)
	} else {
		failures := s.Failures.Add(1)
		consecutiveFailures := s.ConsecutiveFailures.Add(1)

		if s.options.MaxFailures > 0 && failures >= int64(s.options.MaxFailures) {
			Logger.Warn(
				"max failures reached, stopping job",
				"scheduled_job", s,
			)
			select {
			case s.stopCh <- struct{}{}:
			default:
			}
		} else if s.options.MaxConsecutiveFailures > 0 &&
			consecutiveFailures >= int64(s.options.MaxConsecutiveFailures) {
			Logger.Warn(
				"max consecutive failures reached, stopping job",
				"scheduled_job", s,
			)
			select {
			case s.stopCh <- struct{}{}:
			default:
			}
		}
	}

	runtime.End = time.Now()
	Logger.Info(
		"job finished",
		"start", runtime.Start,
		"end", runtime.End,
		"scheduled_job", s,
	)
	s.runtimes = append(s.runtimes, runtime)
}

// JobRuntime is a record of a job's runtime and any error
type JobRuntime struct {
	// Start is the time the job started
	Start time.Time

	// End is the time the job ended
	End time.Time

	// Error is any error that occurred during the job
	Error error
}
