# Portfolio Management Application

A comprehensive portfolio management application built with Wails (Go backend + React frontend) that tracks cryptocurrency, stock, and DeFi positions across multiple exchanges and blockchains.

## 🚀 Features

### Multi-Exchange Support
- **Kraken** - Spot and futures trading data
- **Hyperliquid** - Perpetual futures and spot trading
- **Lighter** - DEX trading data

### Blockchain Integration
- **EVM Chains** - Ethereum, Polygon, Arbitrum, Optimism, Base, etc.
- **Bitcoin** - Native Bitcoin wallet tracking
- **Solana** - Solana blockchain wallet support

### DeFi Protocols
- **Pendle** - Yield trading and position tracking
- **Price Data** - Real-time cryptocurrency prices via CoinGecko
- **Stock Data** - Traditional stock market integration

### Data Management
- **Grist Integration** - Cloud-based spreadsheet storage
- **Automated Backups** - Regular data synchronization
- **Job Scheduling** - Configurable update intervals
- **Real-time Monitoring** - Live job execution tracking

## 🏗️ Architecture

### Backend (Go)
- **Job Controller** - Manages scheduled tasks and execution
- **Exchange APIs** - Integration with various trading platforms
- **Blockchain APIs** - On-chain data retrieval
- **Settings Management** - Configuration persistence
- **Grist Client** - Cloud data synchronization

### Frontend (React + TypeScript)
- **Job Dashboard** - Monitor and control scheduled tasks
- **Logs Viewer** - Real-time execution logs and error tracking
- **Settings Panel** - Configure exchanges, wallets, and intervals
- **Responsive UI** - Modern, clean interface

## 📋 Prerequisites

- **Go 1.25+** - Backend runtime
- **Node.js 18+** - Frontend build tools
- **Wails CLI** - Desktop app framework
- **API Keys** - For exchanges and data providers

## 🛠️ Installation

### 1. Clone the Repository
```bash
git clone https://github.com/zyriu/portfolio.git
cd portfolio
```

### 2. Install Dependencies

#### Backend Dependencies
```bash
go mod download
```

#### Frontend Dependencies
```bash
cd frontend
npm install
cd ..
```

#### Install Wails CLI
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 3. Build the Application
```bash
# Build for current platform
./compile.sh

# Or build manually
wails build
```

## ⚙️ Setup Guide

### Step 1: Required Setup (Only Grist)

#### Grist (Data Storage) - REQUIRED
1. Go to [grist.io](https://grist.io) and create a free account
2. Import `template.grist` (provided in the repository) as your document
3. Go to Document Settings → API
4. Copy the document ID
5. Go to Profile Settings → API
6. Copy the API Key

### Step 2: Optional APIs (Choose what you want to track)

#### Kraken (Exchange Data) - Optional
1. Go to [kraken.com](https://kraken.com)
2. Go to Settings → API
3. Create a new API key with these permissions:
   - Query Funds
   - Query Open Orders
   - Query Closed Orders
   - Query Trades History
4. Copy both the API Key and Secret

#### CoinStats (Enhanced Blockchain Data) - Optional
1. Go to [coinstats.app](https://coinstats.app)
2. Create a free account 
3. Go to API section in your profile
4. Generate an API key
5. Copy the API key

#### CoinGecko (Crypto Price Data) - Optional
1. Go to [coingecko.com](https://coingecko.com)
2. Create a free account
3. Go to your profile → API section
4. Generate an API key (free tier available)
5. Copy the API key

#### TwelveData (Stock Price Data) - Optional
1. Go to [twelvedata.com](https://twelvedata.com)
2. Create a free account
3. Go to your dashboard → API Keys
4. Generate an API key (free tier: 800 requests/day)
5. Copy the API key

### Step 3: Configure the Application

1. **Open Settings Panel** in the application
2. **Required - Grist Setup**:
   - Paste your Grist API key
   - Paste your Grist document ID
   - Enable backup job
3. **Optional - Exchange Setup** (only if you want to track exchanges):
   - Enable Kraken and add your API credentials
   - Enable other exchanges (no API keys needed)
4. **Optional - Blockchain Setup** (only if you want to track wallets):
   - Add your wallet addresses (Ethereum, Bitcoin, Solana)
   - Add your CoinStats API key for enhanced data
5. **Optional - Price Data Setup** (only if you want price tracking):
   - Add your CoinGecko API key for crypto prices
   - Add your TwelveData API key for stock prices
6. **Optional - Set Update Intervals**:
   - Prices: 5-15 minutes
   - Balances: 10-30 minutes
   - Trades: 1-5 minutes

### Step 4: Add Your Wallets (Optional)

#### For EVM Chains (Ethereum, Polygon, etc.)
- Add your wallet addresses (0x...)
- The app will automatically detect tokens

#### For Bitcoin
- Add your Bitcoin addresses (1... or 3... or bc1...)
- Supports legacy, P2SH, and Bech32 formats

#### For Solana
- Add your Solana wallet addresses (base58 format)
- Supports all SPL tokens

### Step 5: Enable Optional Features

- **Price Updates**: Enable to get real-time crypto prices
- **Stock Data**: Enable for traditional stock tracking
- **Pendle**: Enable for DeFi yield tracking
- **Exchanges**: Enable to track trading data
- **Blockchain**: Enable to track wallet balances

> **Note**: Only Grist is required. Everything else is optional based on what you want to track!

## 🚀 How to Use

### Starting the Application
1. **Run the app**: Double-click `Portfolio.app` (macOS) or `Portfolio.exe` (Windows)
2. **First time setup**: Follow the setup guide above to configure your accounts

### Managing Your Portfolio

#### View Your Data
- **Jobs Panel**: See all your data sources and their status
- **Logs Panel**: Check if everything is working correctly
- **Settings Panel**: Configure your accounts and wallets

#### Control Data Updates
- **Pause/Resume**: Stop or start data collection for any source
- **Manual Update**: Click "Trigger" to update data immediately
- **Adjust Frequency**: Change how often data updates (every 5 minutes to 1 hour)

#### Troubleshooting
- **Red Error Button**: Click to clear errors and see what went wrong
- **Check Logs**: View detailed information about data collection
- **Restart Jobs**: Pause and resume jobs to reset them

## 🔧 Development

### Backend Development
```bash
# Run tests
go test ./...

# Build backend only
go build -o portfolio main.go
```

### Frontend Development
```bash
cd frontend
npm run dev
```

### Full Development Mode
```bash
wails dev
```

## 📊 Supported Data Sources

### Exchanges
- **Kraken** - Spot, futures, staking
- **Hyperliquid** - Perpetuals, spot
- **Lighter** - DEX trading

### Blockchains
- **Ethereum** - ERC-20 tokens, NFTs
- **Polygon** - Layer 2 scaling
- **Arbitrum** - L2 rollup
- **Optimism** - L2 rollup
- **Base** - Coinbase L2
- **Bitcoin** - Native BTC
- **Solana** - SPL tokens

### DeFi Protocols
- **Pendle** - Yield trading
- **CoinGecko** - Price data
- **TwelveData** - Stock prices

## 🔒 Security

- API keys stored locally in settings
- No sensitive data transmitted
- Local-first architecture
- Optional cloud backup via Grist

## 🐛 Common Issues & Solutions

### "Jobs not starting"
1. **Check your settings**: Make sure you've entered all required API keys
2. **Verify API keys**: Test your keys on the provider's website
3. **Check logs**: Look at the Logs panel to see specific error messages

### "No data showing up"
1. **Wait a few minutes**: Data collection takes time to start
2. **Check wallet addresses**: Make sure they're correct and have transactions
3. **Verify exchange accounts**: Ensure your exchange accounts have trading history

### "Errors keep appearing"
1. **Click the red error button**: This clears the error and shows you what went wrong
2. **Check your internet connection**: The app needs internet to fetch data
3. **Restart the job**: Pause and resume the failing job
4. **Check API limits**: Some services have daily limits on free accounts

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License - see the [LICENSE](LICENSE) file for details.

### What this means:
- ✅ **You CAN**: Use for personal, non-commercial purposes, modify the code, fork and contribute
- ❌ **You CANNOT**: Use for commercial purposes, sell the software, or incorporate into commercial products

## 📞 Support

For issues and questions:
- Create an issue on GitHub
- Check the logs for error details
- Review the configuration settings
