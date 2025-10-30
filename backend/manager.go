package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/settings"
	"github.com/zyriu/portfolio/backend/jobs/balances_evm_chains"
	"github.com/zyriu/portfolio/backend/jobs/balances_other_chains"
	"github.com/zyriu/portfolio/backend/jobs/exchange_hyperliquid"
	"github.com/zyriu/portfolio/backend/jobs/exchange_kraken"
	"github.com/zyriu/portfolio/backend/jobs/exchange_lighter"
	"github.com/zyriu/portfolio/backend/jobs/grist_backup"
	"github.com/zyriu/portfolio/backend/jobs/pendle_markets"
	"github.com/zyriu/portfolio/backend/jobs/pendle_user_positions"
	"github.com/zyriu/portfolio/backend/jobs/prices_cryptocurrencies"
	"github.com/zyriu/portfolio/backend/jobs/prices_stocks"
)

type JobLog struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"` // "info", "error", "success"
}

type JobExecution struct {
	ID        string   `json:"id"`
	JobName   string   `json:"jobName"`
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime,omitempty"`
	Status    string   `json:"status"` // "running", "completed", "failed"
	Logs      []JobLog `json:"logs"`
}

type Manager struct {
	ctx        context.Context
	mu         sync.RWMutex
	jobs       map[string]*jobController
	emitEvents bool // Control whether to emit events (prevents focus grabbing)

	executionsMu  sync.RWMutex
	executions    []JobExecution
	maxExecutions int // Maximum number of executions to keep in history
}

func NewManager() *Manager {
	return &Manager{
		jobs:          make(map[string]*jobController),
		emitEvents:    false, // Disable events by default to prevent focus grabbing
		executions:    make([]JobExecution, 0),
		maxExecutions: 100, // Keep last 100 executions
	}
}

func (m *Manager) AddAndStart(name string, interval time.Duration, fn JobRunnerFunc, args ...any) {
	m.AddAndStartWithLastRun(name, interval, fn, time.Time{}, args...)
}

func (m *Manager) AddAndStartWithLastRun(name string, interval time.Duration, fn JobRunnerFunc, lastRun time.Time, args ...any) {
	j := newJobController(name, interval, fn, args...)

	// Set the callbacks to track executions
	j.onCompleteCallback = m.recordExecution
	j.onStartCallback = m.recordExecutionStart

	m.mu.Lock()
	m.jobs[name] = j
	m.mu.Unlock()
	j.StartWithLastRun(m.ctx, lastRun)
}

// recordExecutionStart creates a running job execution entry
func (m *Manager) recordExecutionStart(jobName string, startTime time.Time) {
	execution := JobExecution{
		ID:        fmt.Sprintf("%s-%d", jobName, startTime.Unix()),
		JobName:   jobName,
		StartTime: startTime.Format(time.RFC3339),
		Status:    "running",
		Logs:      []JobLog{},
	}

	m.executionsMu.Lock()
	defer m.executionsMu.Unlock()

	// Add to the front of the list
	m.executions = append([]JobExecution{execution}, m.executions...)

	// Trim to max executions
	if len(m.executions) > m.maxExecutions {
		m.executions = m.executions[:m.maxExecutions]
	}
}

// recordExecution stores a completed job execution in history
func (m *Manager) recordExecution(jobName string, startTime, endTime time.Time, logs []JobLog, err error) {
	status := "completed"
	if err != nil {
		status = "failed"
	}

	// Make a copy of logs
	logsCopy := make([]JobLog, len(logs))
	copy(logsCopy, logs)

	if len(logsCopy) > 127 {
		logsCopy = logsCopy[len(logsCopy)-127:]
	}

	executionID := fmt.Sprintf("%s-%d", jobName, startTime.Unix())

	m.executionsMu.Lock()
	defer m.executionsMu.Unlock()

	// Find and update the existing running execution
	for i, exec := range m.executions {
		if exec.ID == executionID && exec.Status == "running" {
			// Update the existing execution
			m.executions[i].EndTime = endTime.Format(time.RFC3339)
			m.executions[i].Status = status
			m.executions[i].Logs = logsCopy
			break
		}
	}

	// Store execution time in Grist (async to avoid blocking)
	go func() {
		if err := m.storeExecutionTimeInGrist(jobName, endTime); err != nil {
			fmt.Printf("Warning: Failed to store execution time for job %q in Grist: %v\n", jobName, err)
		}
	}()
}

func (m *Manager) emit(event string, payload any) {
	// Only emit events if enabled (disabled by default to prevent focus grabbing)
	if m.emitEvents && m.ctx != nil {
		runtime.EventsEmit(m.ctx, event, payload)
	}
}

func (m *Manager) get(name string) (*jobController, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[name]
	if !ok {
		return nil, fmt.Errorf("unknown job %q", name)
	}
	return j, nil
}

func (m *Manager) Jobs() []JobState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]JobState, 0, len(m.jobs))
	for _, j := range m.jobs {
		// Only include jobs that are enabled in settings
		if m.isJobEnabledInSettings(j.name) {
			out = append(out, j.State())
		}
	}
	return out
}

func (m *Manager) Pause(name string) error {
	j, err := m.get(name)
	if err != nil {
		return err
	}
	j.Pause()
	m.emit("job_paused", map[string]any{"name": name})
	return nil
}

func (m *Manager) Resume(name string) error {
	j, err := m.get(name)
	if err != nil {
		// Job doesn't exist, try to create it based on current settings
		if err := m.createJobFromSettings(name); err != nil {
			return fmt.Errorf("job %q not found and could not be created: %v", name, err)
		}
		// Get the newly created job
		j, err = m.get(name)
		if err != nil {
			return err
		}
		// Set callbacks for the newly created job
		j.onCompleteCallback = m.recordExecution
		j.onStartCallback = m.recordExecutionStart
	}
	j.Resume()
	m.emit("job_resumed", map[string]any{"name": name})
	return nil
}

func (m *Manager) SetInterval(name string, interval int64) error {
	j, err := m.get(name)
	if err != nil {
		// Job doesn't exist, try to create it based on current settings
		if err := m.createJobFromSettings(name); err != nil {
			return fmt.Errorf("job %q not found and could not be created: %v", name, err)
		}
		// Get the newly created job
		j, err = m.get(name)
		if err != nil {
			return err
		}
		// Set callbacks for the newly created job
		j.onCompleteCallback = m.recordExecution
		j.onStartCallback = m.recordExecutionStart
	}

	if interval < 5 {
		return errors.New("interval must be at least 5 seconds")
	}

	j.SetInterval(time.Duration(interval) * time.Second)
	m.emit("job_interval_changed", map[string]any{"name": name, "interval": interval})
	return nil
}

func (m *Manager) Startup(ctx context.Context) {
	m.ctx = ctx
}

// SetEventEmissions controls whether events are emitted (disabled by default to prevent focus grabbing)
func (m *Manager) SetEventEmissions(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emitEvents = enabled
}

func (m *Manager) Trigger(name string) error {
	j, err := m.get(name)
	if err != nil {
		return err
	}
	j.Trigger()
	m.emit("job_triggered", map[string]any{"name": name})
	return nil
}

func (m *Manager) ClearError(name string) error {
	j, err := m.get(name)
	if err != nil {
		return err
	}
	j.ClearError()
	m.emit("job_error_cleared", map[string]any{"name": name})
	return nil
}

// StopAndRemove stops a job and removes it from the manager
func (m *Manager) StopAndRemove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	j, ok := m.jobs[name]
	if !ok {
		return fmt.Errorf("job %q not found", name)
	}

	// Stop the job
	j.Stop()

	// Remove from the jobs map
	delete(m.jobs, name)

	m.emit("job_stopped", map[string]any{"name": name})
	return nil
}

// Settings management methods
func (m *Manager) LoadSettings() (string, error) {
	settingsData, err := settings.LoadSettings()
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(settingsData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %v", err)
	}

	return string(jsonData), nil
}

func (m *Manager) SaveSettings(settingsJSON string) error {
	var settingsData settings.Settings
	if err := json.Unmarshal([]byte(settingsJSON), &settingsData); err != nil {
		return fmt.Errorf("failed to parse settings JSON: %v", err)
	}

	if err := settings.SaveSettings(settingsData); err != nil {
		return fmt.Errorf("failed to save settings: %v", err)
	}

	m.emit("settings_updated", map[string]any{"success": true})
	return nil
}

// checkJobSettingsAndCreate checks if a job is enabled in settings and optionally creates it
// Returns true if the job is enabled, false otherwise
// If createIfEnabled is true and the job is enabled, it will also create the job
func (m *Manager) checkJobSettingsAndCreate(name string, lastRun time.Time, createIfEnabled bool) (bool, error) {
	settingsData, err := settings.LoadSettings()
	if err != nil {
		if createIfEnabled {
			return false, fmt.Errorf("failed to load settings: %v", err)
		}
		// If we can't load settings and not creating, assume job is enabled to be safe
		return true, nil
	}

	var isEnabled bool
	var interval time.Duration
	var jobFunc JobRunnerFunc
	var args []any

	switch name {
	case "balances_bitcoin":
		if !settingsData.OnChain.Bitcoin.Enabled {
			isEnabled = false
		} else {
			hasBitcoinWallet := false
			nonEvmWallets := settings.GetNonEVMWallets(settingsData)
			for _, w := range nonEvmWallets {
				if w.Type == "bitcoin" {
					hasBitcoinWallet = true
					break
				}
			}
			isEnabled = hasBitcoinWallet
		}
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.OnChain.Bitcoin.Interval) * time.Second
			jobFunc = balances_other_chains.Run
			args = []any{"bitcoin"}
		}
	case "balances_evm_chains":
		isEnabled = settingsData.OnChain.EVM.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.OnChain.EVM.Interval) * time.Second
			jobFunc = balances_evm_chains.Run
			args = []any{}
		}
	case "balances_solana":
		if !settingsData.OnChain.Solana.Enabled {
			isEnabled = false
		} else {
			hasSolanaWallet := false
			nonEvmWallets := settings.GetNonEVMWallets(settingsData)
			for _, w := range nonEvmWallets {
				if w.Type == "solana" {
					hasSolanaWallet = true
					break
				}
			}
			isEnabled = hasSolanaWallet
		}
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.OnChain.Solana.Interval) * time.Second
			jobFunc = balances_other_chains.Run
			args = []any{"solana"}
		}
	case "exchange_hyperliquid":
		isEnabled = settingsData.Exchanges.Hyperliquid.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Exchanges.Hyperliquid.Interval) * time.Second
			jobFunc = exchange_hyperliquid.Run
			args = []any{}
		}
	case "exchange_kraken":
		isEnabled = settingsData.Exchanges.Kraken.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Exchanges.Kraken.Interval) * time.Second
			jobFunc = exchange_kraken.Run
			args = []any{}
		}
	case "exchange_lighter":
		isEnabled = settingsData.Exchanges.Lighter.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Exchanges.Lighter.Interval) * time.Second
			jobFunc = exchange_lighter.Run
			args = []any{}
		}
	case "grist_backup":
		isEnabled = settingsData.Grist.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Grist.Interval) * time.Second
			jobFunc = grist_backup.Run
			args = []any{}
		}
	case "pendle_markets":
		isEnabled = settingsData.Pendle.Markets.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Pendle.Markets.Interval) * time.Second
			jobFunc = pendle_markets.Run
			args = []any{}
		}
	case "pendle_user_positions":
		isEnabled = settingsData.Pendle.Positions.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Pendle.Positions.Interval) * time.Second
			jobFunc = pendle_user_positions.Run
			args = []any{}
		}
	case "prices_cryptocurrencies":
		isEnabled = settingsData.Settings.Prices.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Settings.Prices.Interval) * time.Second
			jobFunc = prices_cryptocurrencies.Run
			args = []any{}
		}
	case "prices_stocks":
		isEnabled = settingsData.Settings.Stocks.Enabled
		if isEnabled && createIfEnabled {
			interval = time.Duration(settingsData.Settings.Stocks.Interval) * time.Second
			jobFunc = prices_stocks.Run
			args = []any{}
		}
	default:
		if createIfEnabled {
			return false, fmt.Errorf("unknown job name: %s", name)
		}
		return true, nil
	}

	// Create the job if requested and enabled
	if isEnabled && createIfEnabled {
		m.AddAndStartWithLastRun(name, interval, jobFunc, lastRun, args...)
	}

	return isEnabled, nil
}

// createJobFromSettings creates a job based on current settings
func (m *Manager) createJobFromSettings(name string) error {
	_, err := m.checkJobSettingsAndCreate(name, time.Time{}, true)
	return err
}

// createJobFromSettingsWithLastRun creates a job based on current settings with a specific last run time
func (m *Manager) createJobFromSettingsWithLastRun(name string, lastRun time.Time) error {
	_, err := m.checkJobSettingsAndCreate(name, lastRun, true)
	return err
}

// isJobEnabledInSettings checks if a job is enabled in the current settings
func (m *Manager) isJobEnabledInSettings(jobName string) bool {
	enabled, _ := m.checkJobSettingsAndCreate(jobName, time.Time{}, false)
	return enabled
}

// SyncJobsWithSettings stops disabled jobs and starts enabled jobs based on current settings
func (m *Manager) SyncJobsWithSettings() error {
	_, err := settings.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %v", err)
	}

	// Fetch latest execution times from Grist
	executionTimes, err := m.getLatestExecutionTimes()
	if err != nil {
		// If we can't fetch from Grist, continue without lastRun times (jobs will start immediately)
		fmt.Printf("Warning: Could not fetch execution times from Grist: %v\n", err)
	}

	// Get list of all possible job names
	allJobNames := []string{
		"balances_bitcoin",
		"balances_evm_chains",
		"balances_solana",
		"exchange_kraken",
		"exchange_hyperliquid",
		"exchange_lighter",
		"grist_backup",
		"pendle_markets",
		"pendle_user_positions",
		"prices_cryptocurrencies",
		"prices_stocks",
	}

	// Stop and remove disabled jobs
	for _, jobName := range allJobNames {
		if !m.isJobEnabledInSettings(jobName) {
			// Job is disabled, stop and remove it if it exists
			if err := m.StopAndRemove(jobName); err != nil {
				// Job might not exist, which is fine
				if err.Error() != fmt.Sprintf("job %q not found", jobName) {
					return fmt.Errorf("failed to stop job %q: %v", jobName, err)
				}
			}
		}
	}

	// Start enabled jobs that aren't already running
	for _, jobName := range allJobNames {
		if m.isJobEnabledInSettings(jobName) {
			// Check if job already exists
			m.mu.RLock()
			_, exists := m.jobs[jobName]
			m.mu.RUnlock()

			if !exists {
				// Job is enabled but not running, create it with lastRun time if available
				lastRun := time.Time{}
				if executionTimes != nil {
					if lastRunTime, exists := executionTimes[jobName]; exists {
						lastRun = lastRunTime
					}
				}

				if err := m.createJobFromSettingsWithLastRun(jobName, lastRun); err != nil {
					// If job can't be created (e.g., missing config), that's okay
					// Just log it and continue
					fmt.Printf("Warning: Could not create job %q: %v\n", jobName, err)
				}
			}
		}
	}

	return nil
}

// getLatestExecutionTimes fetches the latest execution times from Grist
func (m *Manager) getLatestExecutionTimes() (map[string]time.Time, error) {
	// Initialize Grist client
	g, err := grist.InitiateClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Grist client: %v", err)
	}

	// Fetch execution times from Grist
	executionTimes, err := g.GetLatestExecutionTimes(m.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch execution times from Grist: %v", err)
	}

	return executionTimes, nil
}

// storeExecutionTimeInGrist stores a job execution time in Grist
func (m *Manager) storeExecutionTimeInGrist(jobName string, executionTime time.Time) error {
	// Initialize Grist client
	g, err := grist.InitiateClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Grist client: %v", err)
	}

	// Store execution time in Grist
	if err := g.StoreJobExecutionTime(m.ctx, jobName, executionTime); err != nil {
		return fmt.Errorf("failed to store execution time in Grist: %v", err)
	}

	return nil
}

// SyncJobsWithSettingsPublic is a public method that can be called from the frontend
func (m *Manager) SyncJobsWithSettingsPublic() error {
	return m.SyncJobsWithSettings()
}

// GetExecutions returns the execution history
func (m *Manager) GetExecutions() []JobExecution {
	m.executionsMu.RLock()
	defer m.executionsMu.RUnlock()

	// Return a copy to avoid data races
	executions := make([]JobExecution, len(m.executions))
	copy(executions, m.executions)
	return executions
}
