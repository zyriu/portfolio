package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
)

type JobRunnerFunc func(ctx context.Context, args ...any) error

type ExecutionCompleteCallback func(jobName string, startTime, endTime time.Time, logs []JobLog, err error)

type jobController struct {
	name string
	fn   JobRunnerFunc
	args []any

	mu       sync.RWMutex
	interval time.Duration
	paused   bool

	// control chans
	triggerCh chan struct{}
	updateCh  chan time.Duration
	pauseCh   chan bool
	stopCh    chan struct{}

	// runtime
	ctx                context.Context
	cancel             context.CancelFunc
	lastRun            time.Time
	nextRun            time.Time
	lastErr            error
	started            bool
	isExecuting        bool
	logBuffer          []JobLog
	currentStatus      string // Keep for backwards compatibility
	onCompleteCallback ExecutionCompleteCallback
}

type JobState struct {
	Name          string   `json:"name"`
	Interval      int64    `json:"interval"`
	Running       bool     `json:"running"`
	LastRunUnix   int64    `json:"lastRunUnix"`
	NextRunUnix   int64    `json:"nextRunUnix"`
	Err           string   `json:"err"`
	IsExecuting   bool     `json:"isExecuting"`
	CurrentStatus string   `json:"currentStatus"` // Keep for backwards compatibility
	Logs          []JobLog `json:"logs"`
}

func newJobController(name string, interval time.Duration, fn JobRunnerFunc, args ...any) *jobController {
	return &jobController{
		name:     name,
		fn:       fn,
		args:     args,
		interval: interval,

		triggerCh: make(chan struct{}, 1),
		updateCh:  make(chan time.Duration, 1),
		pauseCh:   make(chan bool, 1),
		stopCh:    make(chan struct{}),
	}
}

func (j *jobController) Start(parent context.Context) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.started {
		return
	}
	j.ctx, j.cancel = context.WithCancel(parent)
	j.started = true

	go j.loop()
}

func (j *jobController) loop() {
	// Run immediately on startup
	j.runOnce()

	timer := time.NewTimer(j.interval)
	defer timer.Stop()

	for {
		var next time.Duration
		j.mu.RLock()
		if j.paused {
			next = time.Hour * 24 * 365 * 100 // effectively never
		} else {
			next = j.interval
		}
		j.mu.RUnlock()

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(next)

		select {
		case <-j.ctx.Done():
			return
		case <-timer.C:
			j.runOnce()
		case <-j.triggerCh:
			j.runOnce()
		case newInterval := <-j.updateCh:
			// Handle interval update intelligently
			j.mu.Lock()
			j.interval = newInterval

			// Calculate elapsed time since last run
			elapsed := time.Since(j.lastRun)

			// Drain the timer first
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			// If the new interval is less than elapsed time, trigger immediately
			if newInterval <= elapsed {
				// Set timer to expire immediately (1 nanosecond)
				timer.Reset(1 * time.Nanosecond)
				j.mu.Unlock()
			} else {
				// New interval is greater than elapsed time
				// Set timer to remaining time: newInterval - elapsed
				remaining := newInterval - elapsed
				j.nextRun = j.lastRun.Add(newInterval)
				timer.Reset(remaining)
				j.mu.Unlock()
			}
			continue // Skip the normal timer reset at the top of the loop
		case p := <-j.pauseCh:
			j.mu.Lock()
			j.paused = p
			j.mu.Unlock()
		case <-j.stopCh:
			if j.cancel != nil {
				j.cancel()
			}
			return
		}
	}
}

func (j *jobController) Pause() {
	select {
	case j.pauseCh <- true:
	default:
	}
}

func (j *jobController) Resume() {
	select {
	case j.pauseCh <- false:
	default:
	}
}

func (j *jobController) updateStatus(status string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Add to log buffer with default level "info"
	log := JobLog{
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   status,
		Level:     "info",
	}
	j.logBuffer = append(j.logBuffer, log)

	// Also update currentStatus for backwards compatibility
	j.currentStatus = status
}

func (j *jobController) runOnce() {
	j.mu.RLock()
	paused := j.paused
	j.mu.RUnlock()
	if paused {
		return
	}

	startTime := time.Now()

	j.mu.Lock()
	j.isExecuting = true
	j.nextRun = time.Now().Add(j.interval)
	// Clear log buffer at the start of each run
	j.logBuffer = []JobLog{}
	j.mu.Unlock()

	// Create context with status updater
	ctx := jobstatus.WithStatusUpdater(j.ctx, j.updateStatus)

	err := j.fn(ctx, j.args...)
	endTime := time.Now()

	j.mu.Lock()
	if err != nil {
		j.lastErr = err
		j.currentStatus = "Job failed with error"
		// Add error to log buffer
		j.logBuffer = append(j.logBuffer, JobLog{
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   fmt.Sprintf("❌ Job failed: %v", err),
			Level:     "error",
		})
	} else {
		j.lastErr = nil
		j.currentStatus = "Job completed successfully"
		// Add success to log buffer
		j.logBuffer = append(j.logBuffer, JobLog{
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "✅ Job completed successfully",
			Level:     "success",
		})
	}

	j.isExecuting = false
	j.lastRun = time.Now()

	// Make a copy of the logs for the callback
	logsCopy := make([]JobLog, len(j.logBuffer))
	copy(logsCopy, j.logBuffer)
	callback := j.onCompleteCallback
	j.mu.Unlock()

	// Call the completion callback if set (outside the lock to avoid deadlock)
	if callback != nil {
		callback(j.name, startTime, endTime, logsCopy, err)
	}
}

func (j *jobController) SetInterval(d time.Duration) {
	select {
	case j.updateCh <- d:
	default:
	}
}

func (j *jobController) State() JobState {
	j.mu.RLock()
	defer j.mu.RUnlock()
	var errStr string
	if j.lastErr != nil {
		errStr = j.lastErr.Error()
	}

	// Make a copy of the log buffer to avoid data races
	logsCopy := make([]JobLog, len(j.logBuffer))
	copy(logsCopy, j.logBuffer)

	return JobState{
		Name:          j.name,
		Interval:      j.interval.Milliseconds() / 1000,
		Running:       !j.paused,
		LastRunUnix:   j.lastRun.Unix(),
		NextRunUnix:   j.nextRun.Unix(),
		Err:           errStr,
		IsExecuting:   j.isExecuting,
		CurrentStatus: j.currentStatus,
		Logs:          logsCopy,
	}
}

func (j *jobController) Stop() {
	close(j.stopCh)
}

func (j *jobController) Trigger() {
	select {
	case j.triggerCh <- struct{}{}:
	default:
	}
}
