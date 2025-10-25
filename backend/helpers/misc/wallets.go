package misc

import (
	"encoding/json"
	"fmt"
	"os"
)

type Wallet struct {
	Label   string `json:"label"`
	Address string `json:"address"`
	Type    string `json:"type"`
}

const walletsListFilePath = "./wallets.json"

func RetrieveWallet(label string) (Wallet, error) {
	wallets, err := RetrieveWalletsList()
	if err != nil {
		return Wallet{}, err
	}

	for _, wallet := range wallets {
		if wallet.Label == label {
			return wallet, nil
		}
	}

	return Wallet{}, fmt.Errorf("wallet not found: %s", label)
}

func RetrieveWalletsList() ([]Wallet, error) {
	file, err := os.Open(walletsListFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var walletsList []Wallet
	if err := json.NewDecoder(file).Decode(&walletsList); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return walletsList, nil
}

func RetrieveWalletsForType(t string) ([]Wallet, error) {
	walletsList, err := RetrieveWalletsList()
	if err != nil {
		return nil, err
	}

	var wallets []Wallet
	for _, wallet := range walletsList {
		if wallet.Type == t {
			wallets = append(wallets, wallet)
		}
	}

	return wallets, nil
}
