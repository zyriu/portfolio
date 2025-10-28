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
	"github.com/zyriu/portfolio/backend/jobs/backup_grist"
	"github.com/zyriu/portfolio/backend/jobs/update_evm_balances"
	"github.com/zyriu/portfolio/backend/jobs/update_hyperliquid"
	"github.com/zyriu/portfolio/backend/jobs/update_kraken"
	"github.com/zyriu/portfolio/backend/jobs/update_lighter"
	"github.com/zyriu/portfolio/backend/jobs/update_non_evm_balances"
	"github.com/zyriu/portfolio/backend/jobs/update_pendle"
	"github.com/zyriu/portfolio/backend/jobs/update_pendle_positions"
	"github.com/zyriu/portfolio/backend/jobs/update_prices"
	"github.com/zyriu/portfolio/backend/jobs/update_stocks"
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

// createJobFromSettings creates a job based on current settings
func (m *Manager) createJobFromSettings(name string) error {
	return m.createJobFromSettingsWithLastRun(name, time.Time{})
}

// createJobFromSettingsWithLastRun creates a job based on current settings with a specific last run time
func (m *Manager) createJobFromSettingsWithLastRun(name string, lastRun time.Time) error {
	settingsData, err := settings.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %v", err)
	}

	switch name {
	case "backup_grist":
		if !settingsData.Grist.Enabled {
			return fmt.Errorf("grist job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Grist.Interval)*time.Second, backup_grist.Run, lastRun)
	case "update_kraken":
		if !settingsData.Exchanges.Kraken.Enabled {
			return fmt.Errorf("kraken job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Exchanges.Kraken.Interval)*time.Second, update_kraken.Run, lastRun)
	case "update_hyperliquid":
		if !settingsData.Exchanges.Hyperliquid.Enabled {
			return fmt.Errorf("hyperliquid job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Exchanges.Hyperliquid.Interval)*time.Second, update_hyperliquid.Run, lastRun)
	case "update_lighter":
		if !settingsData.Exchanges.Lighter.Enabled {
			return fmt.Errorf("lighter job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Exchanges.Lighter.Interval)*time.Second, update_lighter.Run, lastRun)
	case "update_evm_balances":
		if !settingsData.OnChain.EVM.Enabled {
			return fmt.Errorf("evm job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.OnChain.EVM.Interval)*time.Second, update_evm_balances.Run, lastRun)
	case "update_bitcoin_balances":
		if !settingsData.OnChain.Bitcoin.Enabled {
			return fmt.Errorf("bitcoin job is not enabled in settings")
		}
		// Check if there are any bitcoin wallets
		hasBitcoinWallet := false
		nonEvmWallets := settings.GetNonEVMWallets(settingsData)
		for _, w := range nonEvmWallets {
			if w.Type == "bitcoin" {
				hasBitcoinWallet = true
				break
			}
		}
		if !hasBitcoinWallet {
			return fmt.Errorf("no bitcoin wallets configured")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.OnChain.Bitcoin.Interval)*time.Second, update_non_evm_balances.Run, lastRun, "bitcoin")
	case "update_solana_balances":
		if !settingsData.OnChain.Solana.Enabled {
			return fmt.Errorf("solana job is not enabled in settings")
		}
		// Check if there are any solana wallets
		hasSolanaWallet := false
		nonEvmWallets := settings.GetNonEVMWallets(settingsData)
		for _, w := range nonEvmWallets {
			if w.Type == "solana" {
				hasSolanaWallet = true
				break
			}
		}
		if !hasSolanaWallet {
			return fmt.Errorf("no solana wallets configured")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.OnChain.Solana.Interval)*time.Second, update_non_evm_balances.Run, lastRun, "solana")
	case "update_prices":
		if !settingsData.Settings.Prices.Enabled {
			return fmt.Errorf("prices job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Settings.Prices.Interval)*time.Second, update_prices.Run, lastRun)
	case "update_stocks":
		if !settingsData.Settings.Stocks.Enabled {
			return fmt.Errorf("stocks job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Settings.Stocks.Interval)*time.Second, update_stocks.Run, lastRun)
	case "update_pendle":
		if !settingsData.Pendle.Markets.Enabled {
			return fmt.Errorf("pendle markets job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Pendle.Markets.Interval)*time.Second, update_pendle.Run, lastRun)
	case "update_pendle_positions":
		if !settingsData.Pendle.Positions.Enabled {
			return fmt.Errorf("pendle positions job is not enabled in settings")
		}
		m.AddAndStartWithLastRun(name, time.Duration(settingsData.Pendle.Positions.Interval)*time.Second, update_pendle_positions.Run, lastRun)
	default:
		return fmt.Errorf("unknown job name: %s", name)
	}

	return nil
}

// isJobEnabledInSettings checks if a job is enabled in the current settings
func (m *Manager) isJobEnabledInSettings(jobName string) bool {
	settingsData, err := settings.LoadSettings()
	if err != nil {
		// If we can't load settings, assume job is enabled to be safe
		return true
	}

	switch jobName {
	case "backup_grist":
		return settingsData.Grist.Enabled
	case "update_kraken":
		return settingsData.Exchanges.Kraken.Enabled
	case "update_hyperliquid":
		return settingsData.Exchanges.Hyperliquid.Enabled
	case "update_lighter":
		return settingsData.Exchanges.Lighter.Enabled
	case "update_evm_balances":
		return settingsData.OnChain.EVM.Enabled
	case "update_bitcoin_balances":
		if !settingsData.OnChain.Bitcoin.Enabled {
			return false
		}
		// Check if there are any bitcoin wallets
		nonEvmWallets := settings.GetNonEVMWallets(settingsData)
		for _, w := range nonEvmWallets {
			if w.Type == "bitcoin" {
				return true
			}
		}
		return false
	case "update_solana_balances":
		if !settingsData.OnChain.Solana.Enabled {
			return false
		}
		// Check if there are any solana wallets
		nonEvmWallets := settings.GetNonEVMWallets(settingsData)
		for _, w := range nonEvmWallets {
			if w.Type == "solana" {
				return true
			}
		}
		return false
	case "update_prices":
		return settingsData.Settings.Prices.Enabled
	case "update_stocks":
		return settingsData.Settings.Stocks.Enabled
	case "update_pendle":
		return settingsData.Pendle.Markets.Enabled
	case "update_pendle_positions":
		return settingsData.Pendle.Positions.Enabled
	default:
		// For unknown jobs, assume they are enabled
		return true
	}
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
		"backup_grist",
		"update_kraken",
		"update_hyperliquid",
		"update_lighter",
		"update_evm_balances",
		"update_bitcoin_balances",
		"update_solana_balances",
		"update_prices",
		"update_stocks",
		"update_pendle",
		"update_pendle_positions",
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
