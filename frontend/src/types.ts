export type UnifiedWallet = {
  label: string;
  address: string;
  type: 'evm' | 'solana' | 'bitcoin';
  jobs: {
    hyperliquid: boolean;
    lighter: boolean;
    pendle: boolean;
  };
};

export type Settings = {
  wallets: UnifiedWallet[];

  pendle: { 
    markets: { enabled: boolean; interval: number };
    positions: { enabled: boolean; interval: number };
  };

  exchanges: {
    kraken: { enabled: boolean; interval: number; apiKey: string; apiSecret: string };
    hyperliquid: { enabled: boolean; interval: number };
    lighter: { enabled: boolean; interval: number };
  };

  onchain: {
    coinstatsApiKey: string;
    evm: { enabled: boolean; interval: number };
    bitcoin: { enabled: boolean; interval: number };
    solana: { enabled: boolean; interval: number };
  };

  grist: {
    enabled: boolean;
    interval: number;
    apiKey: string;
    documentId: string;
    backupPath: string;
  };

  settings: {
    prices: { enabled: boolean; interval: number; coingeckoApiKey: string };
    stocks: { enabled: boolean; interval: number; twelveDataApiKey: string };
  };
};

// sensible defaults for first run
export const defaultSettings: Settings = {
  wallets: [],

  pendle: { 
    markets: { enabled: true, interval: 600 }, // 10 minutes
    positions: { enabled: false, interval: 600 }, // 10 minutes
  },

  exchanges: {
    kraken: { enabled: false, interval: 600, apiKey: "", apiSecret: "" }, // 10 minutes
    hyperliquid: { enabled: false, interval: 300 }, // 5 minutes
    lighter: { enabled: false, interval: 300 }, // 5 minutes
  },

  onchain: {
    coinstatsApiKey: "",
    evm: { enabled: false, interval: 1800 }, // 30 minutes
    bitcoin: { enabled: false, interval: 10800 }, // 3 hours
    solana: { enabled: false, interval: 10800 }, // 3 hours
  },

  grist: {
    enabled: false,
    interval: 7200, // 2 hours
    apiKey: "",
    documentId: "",
    backupPath: "",
  },

  settings: {
    prices: { enabled: false, interval: 600, coingeckoApiKey: "" }, // 10 minutes
    stocks: { enabled: false, interval: 600, twelveDataApiKey: "" }, // 10 minutes
  },
};

// Helper functions to convert between seconds and minutes/seconds
export function secondsToMinutesSeconds(totalSeconds: number): { minutes: number; seconds: number } {
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return { minutes, seconds };
}

export function minutesSecondsToSeconds(minutes: number, seconds: number): number {
  return minutes * 60 + seconds;
}
