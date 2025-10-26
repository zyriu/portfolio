# Portfolio

A desktop application for tracking cryptocurrency and stock portfolio positions across multiple exchanges, DeFi protocols, and blockchains. Built with [Wails](https://wails.io/) (Go + React/TypeScript) for a native desktop experience.

## Overview

Portfolio Tracker automatically syncs and monitors your portfolio across:

- **Centralized Exchanges**: Kraken
- **Decentralized Exchanges**: Hyperliquid, Lighter
- **DeFi Protocols**: Pendle (markets and positions)
- **On-Chain Balances**: EVM chains (Ethereum, Arbitrum, BSC, Base, HyperEVM), Solana, Bitcoin
- **Traditional Assets**: Stocks via TwelveData
- **Price Feeds**: CoinGecko for crypto prices

All data is stored and managed in [Grist](https://www.getgrist.com/), a spreadsheet-database hybrid that allows you to organize, analyze, and visualize your portfolio data.

### Key Features

- **Automated Data Sync**: Configurable jobs that run at specified intervals to keep your portfolio up to date
- **Multi-Wallet Support**: Track multiple wallets across different blockchains
- **Real-Time Monitoring**: Live job execution status and detailed logging
- **Backup System**: Automatic Grist document backups
- **Native Desktop App**: Fast, lightweight, and cross-platform

## Installation

### Option 1: Run from Binary (Recommended)

1. Download the pre-built binary for your platform from the releases page (or find it in the `build/bin/` directory)
2. On macOS:
   ```bash
   # Navigate to the binary location
   cd build/bin/Portfolio.app/Contents/MacOS
   
   # Run the application
   ./Portfolio
   ```
   
   Alternatively, you can double-click `Portfolio.app` from Finder.

3. On first launch, you may need to allow the app in System Preferences > Security & Privacy (macOS) or equivalent on other platforms.

### Option 2: Build from Source

#### Prerequisites

- **Go**: Version 1.25 or higher ([download](https://go.dev/dl/))
- **Node.js**: Version 16 or higher ([download](https://nodejs.org/))
- **Wails CLI**: Install with `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Platform-specific requirements**:
  - **macOS**: Xcode command line tools (`xcode-select --install`)
  - **Linux**: `gcc` and GTK3/WebKit2GTK development packages
  - **Windows**: MinGW-w64 or similar

#### Build Steps

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd portfolio
   ```

2. **Install frontend dependencies**:
   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. **Build the application**:
   ```bash
   # Development build
   wails dev

   # Production build (creates optimized binary)
   wails build
   ```

4. **Run the built application**:
   - The production binary will be in `build/bin/`
   - On macOS: `build/bin/Portfolio.app`
   - On Linux: `build/bin/Portfolio`
   - On Windows: `build/bin/Portfolio.exe`

#### Using Build Scripts

The repository includes convenience build scripts:

```bash
# Build for current platform
./build.sh

# Build for all platforms (macOS, Linux, Windows)
./build-all.sh
```

## Configuration

All configuration is done through the app's Settings panel. Settings are stored in `~/.portfolio/settings.json` and persist across application launches.

### Initial Setup

1. Launch the application
2. Navigate to the **Settings** panel
3. Configure the services and jobs you want to use (see below)

### Grist Setup (Required)

Grist is used as the database backend for storing all portfolio data. You must set this up first.

#### Step 1: Import the Grist Template

A pre-configured Grist template is included in this repository (`template.grist`) with all the necessary tables and columns already set up.

1. **Create a Grist account** at [https://www.getgrist.com/](https://www.getgrist.com/)
2. **Import the template**:
   - Click "Add New" in Grist
   - Select "Import document"
   - Upload the `template.grist` file from this repository
   - Give your document a name (e.g., "My Portfolio")

The template includes the following tables:
- `Positions_Crypto_`: Stores cryptocurrency positions across all exchanges and wallets
- `Prices`: Stores current prices for tokens and coins
- `Trades`: Stores trade history from exchanges
- Additional tables for specific protocols and assets

#### Step 2: Get Your API Credentials

3. **Get your API key**:
   - Go to your profile settings in Grist
   - Navigate to the API section
   - Generate a new API key
4. **Get your document ID**:
   - Open your imported Grist document
   - The document ID is in the URL: `https://docs.getgrist.com/DOC_ID_HERE/...`

#### Step 3: Configure the Application

5. **Configure in the app**:
   - Launch the Portfolio application
   - Navigate to the Settings panel
   - Enter your **Grist API Key**
   - Enter your **Grist Document ID**
   - Optionally set a backup path for periodic backups
   - Enable the Grist backup job if desired

**Note**: If you prefer to create your own Grist document structure, you can do so, but you may need to modify the code to match your schema. Using the provided template is highly recommended.

## Job Configuration

Each job can be individually enabled/disabled and configured with different update intervals. All intervals are in minutes:seconds.

### 1. Backup Grist

**Purpose**: Creates periodic backups of your Grist document.

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to backup (default: 7200s / 2 hours)
- **Backup Path**: Where to save backups locally

### 2. Update Prices

**Purpose**: Fetches current cryptocurrency prices from CoinGecko.

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to update prices (default: 600s / 10 minutes)
- **CoinGecko API Key**: Required for API access
  - Get a free API key at [https://www.coingecko.com/en/api](https://www.coingecko.com/en/api)
  - Free tier allows 10-30 calls/minute

**How it works**: Reads CoinGecko IDs from your Grist `Prices` table and updates the current price for each token.

### 3. Update Stocks

**Purpose**: Fetches stock prices from TwelveData API.

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to update (default: 600s / 10 minutes)
- **TwelveData API Key**: Required for stock data
  - Get a free API key at [https://twelvedata.com/](https://twelvedata.com/)
  - Free tier allows 800 API credits/day

### 4. Update EVM Balances

**Purpose**: Syncs token balances from EVM-compatible blockchains (Ethereum, Arbitrum, BSC, Base, HyperEVM).

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to sync (default: 1800s / 30 minutes)
- **CoinStats API Key**: Required for balance data
  - Get an API key at [https://coinstats.app/](https://coinstats.app/)
- **Wallets**: Add EVM wallet addresses in the Wallets section
  - Each wallet needs a label and address
  - Set wallet type to "evm"

**How it works**: Queries CoinStats API for all token balances across specified chains and updates your Grist database.

### 5. Update Non-EVM Balances

**Purpose**: Syncs balances from non-EVM chains (Solana, Bitcoin).

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to sync (default: 10800s / 3 hours)
- **CoinStats API Key**: Same as above
- **Wallets**: Add Solana/Bitcoin wallet addresses
  - Set wallet type to "solana" or "bitcoin"

**Note**: Bitcoin and Solana jobs run as separate sub-jobs. The job will only run for chains that have configured wallets.

### 6. Update Kraken

**Purpose**: Syncs balances and trade history from Kraken exchange.

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to sync (default: 600s / 10 minutes)
- **API Key**: Your Kraken API key
- **API Secret**: Your Kraken API secret

**Getting Kraken API credentials**:
1. Log in to [Kraken](https://www.kraken.com/)
2. Go to Settings > API
3. Create a new API key with permissions:
   - Query Funds
   - Query Open Orders & Trades
   - Query Closed Orders & Trades
4. Save the API Key and Private Key (API Secret)

### 7. Update Hyperliquid

**Purpose**: Syncs balances and trade history from Hyperliquid DEX.

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to sync (default: 300s / 5 minutes)
- **Wallets**: Add wallet addresses in the Wallets section
  - Enable the "Hyperliquid" job toggle for each wallet you want to track

**Note**: Hyperliquid is a permissionless DEX, so no API keys are needed. Just provide your wallet address.

### 8. Update Lighter

**Purpose**: Syncs data from Lighter DEX.

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to sync (default: 300s / 5 minutes)
- **Wallets**: Add wallet addresses with the "Lighter" job toggle enabled

### 9. Update Pendle Markets

**Purpose**: Syncs data about Pendle markets (yields, TVL, etc.).

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to sync (default: 600s / 10 minutes)

### 10. Update Pendle Positions

**Purpose**: Syncs your personal positions in Pendle protocol.

**Configuration**:
- **Enabled**: Toggle on/off
- **Interval**: How often to sync (default: 600s / 10 minutes)
- **Wallets**: Add wallet addresses with the "Pendle" job toggle enabled

## Wallet Management

Wallets are configured in the Settings panel under the "Wallets" section.

**Wallet Fields**:
- **Label**: A friendly name for your wallet (e.g., "Main Wallet", "Trading Wallet")
- **Address**: The wallet address (public key)
- **Type**: Select from "evm", "solana", or "bitcoin"
- **Jobs**: Toggle which jobs should track this wallet:
  - Hyperliquid
  - Lighter
  - Pendle

**Examples**:
```json
{
  "label": "Main Wallet",
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "type": "evm",
  "jobs": {
    "hyperliquid": true,
    "lighter": false,
    "pendle": true
  }
}
```

## Using the Application

### Dashboard

The main dashboard shows:
- **Active Jobs**: All configured and enabled jobs
- **Job Status**: Running, Paused, or Stopped
- **Last Execution**: When each job last ran
- **Next Execution**: Countdown to next scheduled run

**Actions**:
- **Play/Pause**: Start or pause a job
- **Trigger**: Manually run a job immediately
- **Settings Icon**: Adjust job interval

### Logs Panel

View detailed execution logs for all jobs:
- Real-time log updates during job execution
- Success/error indicators
- Filterable by job name
- Shows step-by-step progress of each job

### Settings Panel

Configure all aspects of the application:
- API keys and credentials
- Enable/disable jobs
- Set update intervals
- Manage wallets
- Configure backup settings

**Important**: After changing settings, the app will automatically sync the running jobs. Disabled jobs will stop, and newly enabled jobs will start.

## Troubleshooting

### Jobs Not Running

1. Check that the job is **enabled** in Settings
2. Verify that all required API keys are configured
3. For wallet-based jobs, ensure you have wallets configured with the correct job toggles
4. Check the Logs panel for error messages

### API Rate Limits

If you're hitting rate limits:
1. Increase the job intervals in Settings
2. Upgrade your API plan (for CoinGecko, CoinStats, TwelveData)
3. Consider using multiple API keys and implementing rotation (requires code changes)

### Grist Connection Issues

1. Verify your API key is correct
2. Verify your Document ID is correct
3. Ensure your Grist document has the required tables
4. Check that your API key has read/write permissions

### Build Issues

- **Wails not found**: Run `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Go version**: Ensure you have Go 1.25 or higher
- **Frontend build fails**: Try deleting `frontend/node_modules` and running `npm install` again
- **Platform-specific issues**: See [Wails documentation](https://wails.io/docs/gettingstarted/installation)

## Architecture

### Backend (Go)

- **Main Entry**: `main.go` - Wails application setup
- **Manager**: `backend/manager.go` - Job orchestration and state management
- **Job Controller**: `backend/job_controller.go` - Individual job scheduling and execution
- **Jobs**: `backend/jobs/*/run.go` - Specific job implementations
- **Helpers**: `backend/helpers/` - API clients and utilities

### Frontend (React + TypeScript)

- **App**: `frontend/src/App.tsx` - Main application component
- **Panels**: `frontend/src/panels/` - Dashboard, Settings, Logs views
- **Components**: `frontend/src/components/` - Reusable UI components
- **Backend Bridge**: `frontend/wailsjs/` - Generated Wails bindings

### Data Flow

1. **Job Scheduler** triggers job at configured interval
2. **Job** fetches data from external API (exchange, blockchain, price feed)
3. **Data Processing** normalizes and validates the data
4. **Grist Client** upserts records to Grist database
5. **Status Updates** sent to frontend for real-time display

## Security Considerations

- **API Keys**: Stored in `~/.portfolio/settings.json` - ensure this file has appropriate permissions (`chmod 600`)
- **No Cloud Storage**: All data is local or in your own Grist account
- **No Telemetry**: No tracking or analytics
- **Open Source**: Audit the code yourself

## Development

### Running in Development Mode

```bash
wails dev
```

This starts the app with hot-reload enabled for both frontend and backend changes.

### Project Structure

```
portfolio/
├── main.go                 # Application entry point
├── backend/
│   ├── manager.go          # Job manager
│   ├── job_controller.go   # Job scheduling logic
│   ├── jobs/               # Individual job implementations
│   └── helpers/            # Utility packages and API clients
├── frontend/
│   ├── src/
│   │   ├── App.tsx         # Main React component
│   │   ├── panels/         # Main views (Dashboard, Settings, Logs)
│   │   ├── components/     # Reusable components
│   │   └── wailsjs/        # Generated Go<->JS bindings
│   └── package.json
├── build/                  # Build output directory
└── wails.json             # Wails configuration
```

### Adding a New Job

1. Create a new package in `backend/jobs/your_job/`
2. Implement a `Run(ctx context.Context, args ...any) error` function
3. Add the job to `manager.go`:
   - Update `createJobFromSettings()`
   - Update `isJobEnabledInSettings()`
   - Add to `allJobNames` in `SyncJobsWithSettings()`
4. Add configuration fields to `backend/helpers/settings/settings.go`
5. Add UI controls in `frontend/src/panels/GlobalSettings.tsx`

### Local Git Hooks

To ensure commits always include passing tests, the repository provides a pre-commit hook in `.githooks/pre-commit` that runs `go test ./...`.

Enable it in your local clone with:

```bash
git config core.hooksPath .githooks
```

Git will now execute the hook automatically before each commit.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](http://creativecommons.org/licenses/by-nc-sa/4.0/).

**TL;DR**: 
- ✓ Free to use for personal purposes
- ✓ Free to modify and fork
- ✓ Can contribute back to the project
- ✗ Cannot use commercially or for profit

See the [LICENSE](LICENSE) file for complete details.

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

## Changelog

### Current Version
- Multi-wallet support
- Unified wallet configuration
- Job-specific wallet toggles
- Improved error handling and logging
- Automatic job synchronization with settings

---

**Note**: This is a personal portfolio tracker designed for technical users. A pre-configured Grist template is included to get you started quickly. Simply import the `template.grist` file into your Grist account and configure your API credentials in the application.
