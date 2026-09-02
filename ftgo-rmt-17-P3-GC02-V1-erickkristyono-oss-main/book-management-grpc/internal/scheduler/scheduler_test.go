package scheduler

import (
	"context"
	"io"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_RunsPeriodically(t *testing.T) {
	var calls int32
	job := func(ctx context.Context) (int, error) {
		atomic.AddInt32(&calls, 1)
		return 0, nil
	}

	s := New("test", 20*time.Millisecond, job,
		WithLogger(log.New(io.Discard, "", 0)),
		WithRunOnStart(true),
	)

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	// Beri waktu untuk beberapa tick.
	time.Sleep(90 * time.Millisecond)
	cancel()
	s.Stop()

	got := atomic.LoadInt32(&calls)
	// 1x saat start + minimal beberapa tick berikutnya.
	if got < 3 {
		t.Fatalf("jumlah eksekusi = %d, want >= 3", got)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	var calls int32
	job := func(ctx context.Context) (int, error) {
		atomic.AddInt32(&calls, 1)
		return 0, nil
	}
	s := New("test", 10*time.Millisecond, job,
		WithLogger(log.New(io.Discard, "", 0)),
		WithRunOnStart(false),
	)
	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)
	time.Sleep(35 * time.Millisecond)
	cancel()
	s.Stop()

	before := atomic.LoadInt32(&calls)
	time.Sleep(30 * time.Millisecond)
	after := atomic.LoadInt32(&calls)
	if before != after {
		t.Fatalf("job masih berjalan setelah cancel: before=%d after=%d", before, after)
	}
}

func TestScheduler_RecoversFromPanic(t *testing.T) {
	s := New("panic", time.Hour, func(ctx context.Context) (int, error) {
		panic("boom")
	}, WithLogger(log.New(io.Discard, "", 0)), WithRunOnStart(false))

	// runOnce harus memulihkan panic tanpa meng-crash test.
	s.runOnce(context.Background())
}
