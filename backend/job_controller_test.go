package backend

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
)

func waitForCondition(t *testing.T, cond func() bool, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

func TestJobControllerStartAndRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runCount atomic.Int32

	job := newJobController("test", 20*time.Millisecond, func(ctx context.Context, args ...any) error {
		updater := jobstatus.GetStatusUpdater(ctx)
		updater("running")
		runCount.Add(1)
		return nil
	})

	job.Start(ctx)
	defer job.Stop()

	if !waitForCondition(t, func() bool { return runCount.Load() >= 1 }, 500*time.Millisecond) {
		t.Fatalf("expected job to run at least once, got %d", runCount.Load())
	}

	state := job.State()
	if !state.Running {
		t.Fatalf("expected job to be running")
	}
	if len(state.Logs) == 0 {
		t.Fatalf("expected logs to be recorded")
	}
	if state.CurrentStatus == "" {
		t.Fatalf("expected current status to be set")
	}
}

func TestJobControllerSetIntervalTriggersImmediate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runCount atomic.Int32
	job := newJobController("interval", 100*time.Millisecond, func(ctx context.Context, args ...any) error {
		runCount.Add(1)
		return nil
	})

	job.Start(ctx)
	defer job.Stop()

	if !waitForCondition(t, func() bool { return runCount.Load() >= 1 }, 500*time.Millisecond) {
		t.Fatalf("expected job to run initial execution")
	}

	time.Sleep(30 * time.Millisecond)
	job.SetInterval(10 * time.Millisecond)

	if !waitForCondition(t, func() bool { return runCount.Load() >= 2 }, 500*time.Millisecond) {
		t.Fatalf("expected job to execute after interval change, got %d runs", runCount.Load())
	}
}

func TestJobControllerPauseAndResume(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runCount atomic.Int32
	job := newJobController("pause", 40*time.Millisecond, func(ctx context.Context, args ...any) error {
		runCount.Add(1)
		return nil
	})

	job.Start(ctx)
	defer job.Stop()

	if !waitForCondition(t, func() bool { return runCount.Load() >= 1 }, 500*time.Millisecond) {
		t.Fatalf("expected job to run initially")
	}

	job.Pause()
	pausedRuns := runCount.Load()
	time.Sleep(100 * time.Millisecond)
	if runCount.Load() != pausedRuns {
		t.Fatalf("expected run count to remain %d while paused, got %d", pausedRuns, runCount.Load())
	}

	job.Resume()
	job.Trigger()

	if !waitForCondition(t, func() bool { return runCount.Load() >= pausedRuns+1 }, 500*time.Millisecond) {
		t.Fatalf("expected job to run after resume, got %d runs", runCount.Load())
	}
}
