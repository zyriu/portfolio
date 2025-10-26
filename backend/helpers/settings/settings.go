package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	Wallets []UnifiedWallet `json:"wallets"`

	Pendle struct {
		Markets struct {
			Enabled  bool `json:"enabled"`
			Interval int  `json:"interval"`
		} `json:"markets"`
		Positions struct {
			Enabled  bool `json:"enabled"`
			Interval int  `json:"interval"`
		} `json:"positions"`
	} `json:"pendle"`

	Exchanges struct {
		Kraken struct {
			Enabled   bool   `json:"enabled"`
			Interval  int    `json:"interval"`
			APIKey    string `json:"apiKey"`
			APISecret string `json:"apiSecret"`
		} `json:"kraken"`
		Hyperliquid struct {
			Enabled  bool `json:"enabled"`
			Interval int  `json:"interval"`
		} `json:"hyperliquid"`
		Lighter struct {
			Enabled  bool `json:"enabled"`
			Interval int  `json:"interval"`
		} `json:"lighter"`
	} `json:"exchanges"`

	OnChain struct {
		CoinStatsAPIKey string `json:"coinstatsApiKey"`
		EVM             struct {
			Enabled  bool `json:"enabled"`
			Interval int  `json:"interval"`
		} `json:"evm"`
		Bitcoin struct {
			Enabled  bool `json:"enabled"`
			Interval int  `json:"interval"`
		} `json:"bitcoin"`
		Solana struct {
			Enabled  bool `json:"enabled"`
			Interval int  `json:"interval"`
		} `json:"solana"`
	} `json:"onchain"`

	Grist struct {
		Enabled    bool   `json:"enabled"`
		Interval   int    `json:"interval"`
		APIKey     string `json:"apiKey"`
		DocumentID string `json:"documentId"`
		BackupPath string `json:"backupPath"`
	} `json:"grist"`

	Settings struct {
		Prices struct {
			Enabled         bool   `json:"enabled"`
			Interval        int    `json:"interval"`
			CoinGeckoAPIKey string `json:"coingeckoApiKey"`
		} `json:"prices"`
		Stocks struct {
			Enabled          bool   `json:"enabled"`
			Interval         int    `json:"interval"`
			TwelveDataAPIKey string `json:"twelveDataApiKey"`
		} `json:"stocks"`
	} `json:"settings"`
}

type UnifiedWallet struct {
	Label   string `json:"label"`
	Address string `json:"address"`
	Type    string `json:"type"` // "evm", "solana", "bitcoin"
	Jobs    struct {
		Hyperliquid bool `json:"hyperliquid"`
		Lighter     bool `json:"lighter"`
		Pendle      bool `json:"pendle"`
	} `json:"jobs"`
}

// Legacy types for backwards compatibility
type Wallet struct {
	Label   string `json:"label"`
	Address string `json:"address"`
	Type    string `json:"type"`
}

type NonEvmWallet struct {
	Label   string `json:"label"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

// getSettingsFilePath returns the path to the settings file
// Uses user's home directory to ensure persistence across app launches
func getSettingsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	// Use ~/.portfolio/settings.json
	configDir := filepath.Join(homeDir, ".portfolio")

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	return filepath.Join(configDir, "settings.json"), nil
}

// GetDefaultSettings returns sensible defaults for first run
func GetDefaultSettings() Settings {
	settings := Settings{}

	// Set defaults
	settings.Wallets = []UnifiedWallet{}

	settings.Pendle.Markets.Enabled = false
	settings.Pendle.Markets.Interval = 600 // 10 minutes
	settings.Pendle.Positions.Enabled = false
	settings.Pendle.Positions.Interval = 600 // 10 minutes

	settings.Exchanges.Kraken.Enabled = false
	settings.Exchanges.Kraken.Interval = 600 // 10 minutes
	settings.Exchanges.Kraken.APIKey = ""
	settings.Exchanges.Kraken.APISecret = ""

	settings.Exchanges.Hyperliquid.Enabled = false
	settings.Exchanges.Hyperliquid.Interval = 300 // 5 minutes

	settings.Exchanges.Lighter.Enabled = false
	settings.Exchanges.Lighter.Interval = 300 // 5 minutes

	settings.OnChain.CoinStatsAPIKey = ""
	settings.OnChain.EVM.Enabled = false
	settings.OnChain.EVM.Interval = 1800 // 30 minutes

	settings.OnChain.Bitcoin.Enabled = false
	settings.OnChain.Bitcoin.Interval = 10800 // 3 hours

	settings.OnChain.Solana.Enabled = false
	settings.OnChain.Solana.Interval = 1800 // 30 minutes

	settings.Grist.Enabled = false
	settings.Grist.Interval = 7200 // 2 hours
	settings.Grist.APIKey = ""
	settings.Grist.DocumentID = ""
	settings.Grist.BackupPath = ""

	settings.Settings.Prices.Enabled = false
	settings.Settings.Prices.Interval = 600 // 10 minutes
	settings.Settings.Prices.CoinGeckoAPIKey = ""

	settings.Settings.Stocks.Enabled = false
	settings.Settings.Stocks.Interval = 600 // 10 minutes
	settings.Settings.Stocks.TwelveDataAPIKey = ""

	return settings
}

// LoadSettings loads settings from settings.json file
func LoadSettings() (Settings, error) {
	var settings Settings

	settingsFilePath, err := getSettingsFilePath()
	if err != nil {
		return settings, err
	}

	// Check if settings file exists
	if _, err := os.Stat(settingsFilePath); os.IsNotExist(err) {
		// Create default settings file if it doesn't exist
		settings = GetDefaultSettings()
		if err := SaveSettings(settings); err != nil {
			return settings, fmt.Errorf("failed to create default settings: %v", err)
		}
		return settings, nil
	}

	// Read and parse settings file
	data, err := os.ReadFile(settingsFilePath)
	if err != nil {
		return settings, fmt.Errorf("failed to read settings file: %v", err)
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, fmt.Errorf("failed to parse settings file: %v", err)
	}

	return settings, nil
}

// SaveSettings saves settings to settings.json file
func SaveSettings(settings Settings) error {
	settingsFilePath, err := getSettingsFilePath()
	if err != nil {
		return err
	}

	// Ensure directory exists (done in getSettingsFilePath, but being extra safe)
	dir := filepath.Dir(settingsFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %v", err)
	}

	// Marshal settings to JSON
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %v", err)
	}

	// Write to file
	if err := os.WriteFile(settingsFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	return nil
}

// GetCurrentSettings returns the current settings from file
func GetCurrentSettings() (Settings, error) {
	return LoadSettings()
}

// Helper functions to get wallets for specific services

// GetHyperliquidWallets returns all wallets with Hyperliquid job enabled
func GetHyperliquidWallets(settings Settings) []UnifiedWallet {
	var wallets []UnifiedWallet
	for _, w := range settings.Wallets {
		if w.Jobs.Hyperliquid {
			wallets = append(wallets, w)
		}
	}
	return wallets
}

// GetLighterWallets returns all wallets with Lighter job enabled
func GetLighterWallets(settings Settings) []UnifiedWallet {
	var wallets []UnifiedWallet
	for _, w := range settings.Wallets {
		if w.Jobs.Lighter {
			wallets = append(wallets, w)
		}
	}
	return wallets
}

// GetPendleWallets returns all wallets with Pendle job enabled
func GetPendleWallets(settings Settings) []UnifiedWallet {
	var wallets []UnifiedWallet
	for _, w := range settings.Wallets {
		if w.Jobs.Pendle {
			wallets = append(wallets, w)
		}
	}
	return wallets
}

// GetEVMWallets returns all EVM wallets
func GetEVMWallets(settings Settings) []UnifiedWallet {
	var wallets []UnifiedWallet
	for _, w := range settings.Wallets {
		if w.Type == "evm" {
			wallets = append(wallets, w)
		}
	}
	return wallets
}

// GetNonEVMWallets returns all non-EVM wallets (Solana and Bitcoin)
func GetNonEVMWallets(settings Settings) []UnifiedWallet {
	var wallets []UnifiedWallet
	for _, w := range settings.Wallets {
		if w.Type == "solana" || w.Type == "bitcoin" {
			wallets = append(wallets, w)
		}
	}
	return wallets
}
