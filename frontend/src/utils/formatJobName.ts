export function formatJobName(name: string): string {
  // Custom names for specific jobs
  const customNames: { [key: string]: string } = {
    'balances_bitcoin': 'Bitcoin Balances',
    'balances_evm_chains': 'EVM Balances',
    'balances_solana': 'Solana Balances',
    'exchange_hyperliquid': 'Hyperliquid',
    'exchange_kraken': 'Kraken',
    'exchange_lighter': 'Lighter',
    'grist_backup': 'Grist Backup',
    'pendle_markets': 'Pendle Markets',
    'pendle_user_positions': 'Pendle User Positions',
    'prices_cryptocurrencies': 'Cryptocurrencies Prices',
    'prices_stocks': 'Stocks Prices'
  };

  // Return custom name if available, otherwise use default formatting
  if (customNames[name]) {
    return customNames[name];
  }

  return name
    .split("_")
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

