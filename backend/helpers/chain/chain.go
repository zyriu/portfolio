package chain

func IDToName(chainID int) string {
	switch chainID {
	case 1:
		return "Ethereum"
	case 56:
		return "BSC"
	case 146:
		return "Sonic"
	case 999:
		return "Hyperliquid"
	case 8453:
		return "Base"
	case 9745:
		return "Plasma"
	case 42161:
		return "Arbitrum"
	case 80094:
		return "Berachain"
	}

	return "Unknown"
}

func NameToID(name string) int {
	switch name {
	case "Ethereum":
		return 1
	case "BSC":
		return 56
	case "Sonic":
		return 146
	case "Hyperliquid":
		return 999
	case "Base":
		return 8453
	case "Plasma":
		return 9745
	case "Arbitrum":
		return 42161
	case "Berachain":
		return 88094
	}

	return 0
}
