export function formatJobName(name: string): string {
  // Custom names for specific jobs
  const customNames: { [key: string]: string } = {
    'update_prices': 'Crypto Feed',
    'update_stocks': 'Stocks Feed',
    'update_evm_balances': 'EVM Balances',
    'update_bitcoin_balances': 'Bitcoin Balances',
    'update_solana_balances': 'Solana Balances',
    'update_kraken': 'Kraken',
    'update_hyperliquid': 'Hyperliquid',
    'update_lighter': 'Lighter',
    'update_pendle': 'Pendle Markets',
    'update_pendle_positions': 'Pendle Positions',
    'backup_grist': 'Grist Backup'
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

