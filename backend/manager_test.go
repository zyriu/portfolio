package backend

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/settings"
)

func setupTestHome(t *testing.T) {
	t.Helper()
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	if err := settings.SaveSettings(settings.GetDefaultSettings()); err != nil {
		t.Fatalf("failed to prime settings: %v", err)
	}
}

func TestManagerAddAndStartRecordsExecution(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Startup(ctx)

	var runCount atomic.Int32
	m.AddAndStart("test_job", 20*time.Millisecond, func(ctx context.Context, args ...any) error {
		jobstatus.GetStatusUpdater(ctx)("running")
		runCount.Add(1)
		return nil
	})

	if !waitForCondition(t, func() bool { return runCount.Load() >= 1 }, time.Second) {
		t.Fatalf("expected job to run at least once")
	}

	if !waitForCondition(t, func() bool { return len(m.GetExecutions()) >= 1 }, time.Second) {
		t.Fatalf("expected execution history to be recorded")
	}

	jobs := m.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("expected exactly one job, got %d", len(jobs))
	}

	if err := m.StopAndRemove("test_job"); err != nil {
		t.Fatalf("failed to stop job: %v", err)
	}
}

func TestManagerSetIntervalValidation(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx := context.Background()
	m.Startup(ctx)

	m.AddAndStart("interval_job", 10*time.Second, func(ctx context.Context, args ...any) error { return nil })
	defer m.StopAndRemove("interval_job")

	if err := m.SetInterval("interval_job", 4); err == nil {
		t.Fatalf("expected error for interval less than 5 seconds")
	}
}

func TestManagerResumeCreateErrorWhenDisabled(t *testing.T) {
	setupTestHome(t)

	s, err := settings.LoadSettings()
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}
	s.Exchanges.Kraken.Enabled = false
	if err := settings.SaveSettings(s); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	m := NewManager()
	ctx := context.Background()
	m.Startup(ctx)

	err = m.Resume("update_kraken")
	if err == nil {
		t.Fatalf("expected error when creating disabled job")
	}
}

func TestManagerSyncJobsWithSettingsStopsDisabled(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	m.jobs["update_kraken"] = newJobController("update_kraken", time.Minute, func(ctx context.Context, args ...any) error {
		return nil
	})

	if err := m.SyncJobsWithSettings(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if _, exists := m.jobs["update_kraken"]; exists {
		t.Fatalf("expected disabled job to be removed")
	}
}

func TestManagerGetExecutionsReturnsCopy(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Startup(ctx)

	// Create a job that will generate an execution
	var runCount atomic.Int32
	m.AddAndStart("test_job", 20*time.Millisecond, func(ctx context.Context, args ...any) error {
		jobstatus.GetStatusUpdater(ctx)("running")
		runCount.Add(1)
		return nil
	})

	// Wait for the job to run and generate an execution
	if !waitForCondition(t, func() bool { return runCount.Load() >= 1 }, time.Second) {
		t.Fatalf("expected job to run at least once")
	}

	executions := m.GetExecutions()
	if len(executions) != 1 {
		t.Fatalf("expected one execution, got %d", len(executions))
	}

	// Ensure the returned slice points to a different underlying array
	m.executionsMu.RLock()
	originalFirst := m.executions[0]
	m.executionsMu.RUnlock()

	if &executions[0] == &originalFirst {
		t.Fatalf("expected executions slice to be a copy")
	}

	executions[0].ID = "modified"

	m.executionsMu.RLock()
	if m.executions[0].ID == "modified" {
		m.executionsMu.RUnlock()
		t.Fatalf("expected modifying copy not to change stored executions")
	}
	m.executionsMu.RUnlock()
}

func TestManagerJobsFiltersDisabled(t *testing.T) {
	setupTestHome(t)

	s, err := settings.LoadSettings()
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}
	s.Exchanges.Kraken.Enabled = true
	s.Settings.Prices.Enabled = false
	if err := settings.SaveSettings(s); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	m := NewManager()
	m.jobs["update_kraken"] = newJobController("update_kraken", time.Minute, func(ctx context.Context, args ...any) error {
		return nil
	})
	m.jobs["update_prices"] = newJobController("update_prices", time.Minute, func(ctx context.Context, args ...any) error {
		return nil
	})

	jobs := m.Jobs()
	if len(jobs) != 1 {
		t.Fatalf("expected only enabled job to be returned, got %d", len(jobs))
	}
	if jobs[0].Name != "update_kraken" {
		t.Fatalf("expected update_kraken job, got %s", jobs[0].Name)
	}
}

func TestManagerSetEventEmissions(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	if m.emitEvents {
		t.Fatalf("expected events disabled by default")
	}
	m.SetEventEmissions(true)
	if !m.emitEvents {
		t.Fatalf("expected events to be enabled")
	}
	m.SetEventEmissions(false)
	if m.emitEvents {
		t.Fatalf("expected events to be disabled")
	}
}

func TestManagerStopAndRemoveErrorsWhenMissing(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	err := m.StopAndRemove("missing")
	if err == nil || !strings.Contains(err.Error(), "job \"missing\" not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestManagerPauseAndTrigger(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Startup(ctx)

	// Test pause on non-existent job
	err := m.Pause("missing")
	if err == nil || !strings.Contains(err.Error(), "unknown job") {
		t.Fatalf("expected unknown job error, got %v", err)
	}

	// Test trigger on non-existent job
	err = m.Trigger("missing")
	if err == nil || !strings.Contains(err.Error(), "unknown job") {
		t.Fatalf("expected unknown job error, got %v", err)
	}

	// Add a job and test pause/trigger
	var runCount atomic.Int32
	m.AddAndStart("test_job", 100*time.Millisecond, func(ctx context.Context, args ...any) error {
		runCount.Add(1)
		return nil
	})

	// Wait for initial run
	if !waitForCondition(t, func() bool { return runCount.Load() >= 1 }, time.Second) {
		t.Fatalf("expected job to run initially")
	}

	// Test pause functionality
	if err := m.Pause("test_job"); err != nil {
		t.Fatalf("failed to pause job: %v", err)
	}

	// Test trigger functionality (this should work even if paused)
	if err := m.Trigger("test_job"); err != nil {
		t.Fatalf("failed to trigger job: %v", err)
	}
}

func TestManagerLoadAndSaveSettings(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Startup(ctx)

	// Test LoadSettings
	settingsJSON, err := m.LoadSettings()
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}

	// Verify it's valid JSON
	var settingsData map[string]interface{}
	if err := json.Unmarshal([]byte(settingsJSON), &settingsData); err != nil {
		t.Fatalf("settings JSON is invalid: %v", err)
	}

	// Test SaveSettings with valid JSON
	modifiedSettings := `{"grist":{"enabled":true,"apiKey":"test","interval":300},"exchanges":{"kraken":{"enabled":true,"interval":600}}}`
	if err := m.SaveSettings(modifiedSettings); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	// Test SaveSettings with invalid JSON
	if err := m.SaveSettings("invalid json"); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestManagerCreateJobFromSettings(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Startup(ctx)

	// Test creating a job that doesn't exist in settings
	err := m.createJobFromSettings("unknown_job")
	if err == nil || !strings.Contains(err.Error(), "unknown job name") {
		t.Fatalf("expected unknown job error, got %v", err)
	}

	// Test creating a job that's disabled in settings
	err = m.createJobFromSettings("update_kraken")
	if err == nil || !strings.Contains(err.Error(), "kraken job is not enabled") {
		t.Fatalf("expected disabled job error, got %v", err)
	}
}

func TestManagerIsJobEnabledInSettings(t *testing.T) {
	setupTestHome(t)

	m := NewManager()

	// Test with default settings (most jobs disabled)
	if m.isJobEnabledInSettings("update_kraken") {
		t.Fatalf("expected kraken job to be disabled by default")
	}

	if m.isJobEnabledInSettings("update_prices") {
		t.Fatalf("expected prices job to be disabled by default")
	}

	// Test unknown job (should return true for safety)
	if !m.isJobEnabledInSettings("unknown_job") {
		t.Fatalf("expected unknown job to be enabled for safety")
	}
}

func TestManagerSyncJobsWithSettings(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Startup(ctx)

	// Test sync with default settings (should not create any jobs)
	if err := m.SyncJobsWithSettings(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	jobs := m.Jobs()
	if len(jobs) != 0 {
		t.Fatalf("expected no jobs with default settings, got %d", len(jobs))
	}
}

func TestManagerSyncJobsWithSettingsPublic(t *testing.T) {
	setupTestHome(t)

	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Startup(ctx)

	// Test the public method
	if err := m.SyncJobsWithSettingsPublic(); err != nil {
		t.Fatalf("public sync failed: %v", err)
	}
}
