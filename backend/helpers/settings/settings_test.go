package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultSettings(t *testing.T) {
	defaults := GetDefaultSettings()
	if defaults.Pendle.Markets.Interval != 600 {
		t.Fatalf("expected pendle markets interval 600, got %d", defaults.Pendle.Markets.Interval)
	}
	if defaults.OnChain.EVM.Enabled {
		t.Fatalf("expected evm job disabled by default")
	}
	if defaults.Settings.Prices.Interval != 600 {
		t.Fatalf("expected prices interval 600, got %d", defaults.Settings.Prices.Interval)
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	original := GetDefaultSettings()
	original.Grist.Enabled = true
	original.Grist.APIKey = "abc"
	original.Wallets = append(original.Wallets, UnifiedWallet{Label: "wallet", Address: "0x1", Type: "evm"})

	if err := SaveSettings(original); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}

	if len(loaded.Wallets) != 1 || loaded.Wallets[0].Address != "0x1" {
		t.Fatalf("expected wallet to round trip")
	}
	if !loaded.Grist.Enabled || loaded.Grist.APIKey != "abc" {
		t.Fatalf("expected grist settings to persist")
	}

	settingsPath := filepath.Join(tempHome, ".portfolio", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		t.Fatalf("expected settings file to exist: %v", err)
	}
}

func TestLoadSettingsCreatesDefault(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}
	if loaded.Pendle.Markets.Enabled {
		t.Fatalf("expected pendle markets disabled by default")
	}
}

func TestWalletHelpers(t *testing.T) {
	settings := Settings{
		Wallets: []UnifiedWallet{
			{Label: "evm", Type: "evm", Jobs: struct {
				Hyperliquid bool `json:"hyperliquid"`
				Lighter     bool `json:"lighter"`
				Pendle      bool `json:"pendle"`
			}{Hyperliquid: true, Pendle: true}},
			{Label: "sol", Type: "solana"},
			{Label: "btc", Type: "bitcoin"},
			{Label: "lighter", Type: "evm", Jobs: struct {
				Hyperliquid bool `json:"hyperliquid"`
				Lighter     bool `json:"lighter"`
				Pendle      bool `json:"pendle"`
			}{Lighter: true}},
		},
	}

	if len(GetHyperliquidWallets(settings)) != 1 {
		t.Fatalf("expected one hyperliquid wallet")
	}
	if len(GetPendleWallets(settings)) != 1 {
		t.Fatalf("expected one pendle wallet")
	}
	if len(GetLighterWallets(settings)) != 1 {
		t.Fatalf("expected one lighter wallet")
	}
	nonEvm := GetNonEVMWallets(settings)
	if len(nonEvm) != 2 {
		t.Fatalf("expected two non-evm wallets, got %d", len(nonEvm))
	}
}
