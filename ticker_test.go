package crong

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	ticker := NewTicker(ctx, s, 5*time.Second)
	if ticker == nil {
		t.Fatalf("expected ticker")
	}
	defer ticker.Stop()
	nextTick := s.Next(time.Now())

	select {
	case <-ctx.Done():
		t.Fatalf("expected tick")
	case tick := <-ticker.C:
		tickMin := tick.Truncate(time.Minute)
		nextMin := nextTick.Truncate(time.Minute)

		if !tickMin.Equal(nextMin) {
			t.Fatalf("expected tick to be %s, got %s", tickMin, nextMin)
		}
	}
}

func TestEarlyTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	ticker := NewTicker(ctx, s, 5*time.Second)
	if ticker == nil {
		t.Fatalf("expected ticker")
	}
	defer ticker.Stop()

	nextTick := s.Next(time.Now())
	go func() {
		ticker.tick(ctx)
	}()

	select {
	case <-ctx.Done():
		t.Fatalf("expected tick")
	case tick := <-ticker.C:
		tickSecs := tick.Unix()
		nextSecs := nextTick.Unix()
		if nextSecs <= tickSecs {
			t.Fatalf("expected tick to be %d, got %d", tickSecs, nextSecs)
		}
	}
}

func TestTickerCanceled(t *testing.T) {
	// verify that we no longer receive ticks after canceling
	// the tick context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	ticker := NewTicker(ctx, s, 5*time.Second)
	if ticker == nil {
		t.Fatalf("expected ticker")
	}
	defer ticker.Stop()

	tctx, tcancel := context.WithCancel(context.Background())
	defer tcancel()

	sawTick := make(chan time.Time, 1)
	go func() {
		select {
		case <-tctx.Done():
			return
		case <-ticker.C:
			sawTick <- time.Now()
		}
	}()

	// cancel the context, which should prevent the tick from being emitted
	cancel()
	go func() {
		time.Sleep(500 * time.Millisecond)
		if sent := ticker.tick(ctx); sent {
			t.Errorf("expected no tick to be sent")
		}
	}()

	cctx, ccancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer ccancel()
	select {
	case <-sawTick:
		t.Fatalf("shouldn't have received tick")
	case <-cctx.Done():
		tcancel()
		return
	}
}

func TestTickerSendTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s, err := New("* * * * *", nil) // every minute
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	ticker := NewTicker(ctx, s, 3*time.Second)
	if ticker == nil {
		t.Fatalf("expected ticker")
	}
	defer ticker.Stop()
	ticker.tick(ctx)
	time.Sleep(5 * time.Second)
	assertEqual(t, ticker.ticksDropped.Load(), int64(1))
}
