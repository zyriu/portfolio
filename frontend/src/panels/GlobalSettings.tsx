import React, { useEffect, useState } from "react";
import { loadSettings, saveSettings } from "../backend";
import { Settings, UnifiedWallet } from "../types";
import { Card, Input, Button, SettingRow, InputField, ErrorMessage, WalletRow, SaveButton, Switch, IntervalInput } from "../components";

export default function GlobalSettings() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [originalSettings, setOriginalSettings] = useState<Settings | null>(null);
  const [showGristMessage, setShowGristMessage] = useState<boolean>(false);
  const [showKrakenMessage, setShowKrakenMessage] = useState<boolean>(false);
  const [showCoinStatsMessage, setShowCoinStatsMessage] = useState<boolean>(false);
  const [showStocksMessage, setShowStocksMessage] = useState<boolean>(false);

  useEffect(() => {
    loadSettings().then(loadedSettings => {
      setSettings(loadedSettings);
      setOriginalSettings(loadedSettings);
    });
  }, []);

  if (!settings || !originalSettings) return null;

  const hasChanges = JSON.stringify(settings) !== JSON.stringify(originalSettings);

  const saveSettingsAndUpdate = async (updatedSettings: Settings) => {
    await saveSettings(updatedSettings);
    setOriginalSettings(updatedSettings);
  };

  const updateInterval = (section: string, subsection: string, newInterval: number) => {
    if (section === 'settings') {
      setSettings({
        ...settings,
        settings: {
          ...settings.settings,
          [subsection]: {
            ...(settings.settings as any)[subsection],
            interval: newInterval
          }
        }
      });
    } else if (section === 'grist' && subsection === '') {
      setSettings({
        ...settings,
        grist: {
          ...settings.grist,
          interval: newInterval
        }
      });
    } else {
      setSettings({
        ...settings,
        [section]: {
          ...(settings as any)[section],
          [subsection]: {
            ...(settings as any)[section][subsection],
            interval: newInterval
          }
        }
      });
    }
  };

  const updateApiKey = (section: string, subsection: string, key: string, value: string) => {
    const updated = { ...settings };
    if (section === 'settings') {
      (updated.settings as any)[subsection][key] = value;
    } else {
      (updated as any)[section][subsection][key] = value;
    }
    setSettings(updated);
  };

  const updateWallet = (index: number, field: keyof UnifiedWallet, value: any) => {
    const updated = { ...settings };
    updated.wallets = [...updated.wallets];
    updated.wallets[index] = { ...updated.wallets[index], [field]: value };

    if (field === 'type' && value !== 'evm') {
      updated.wallets[index].jobs = { hyperliquid: false, lighter: false, pendle: false };
    }
    setSettings(updated);
  };

  const toggleWalletJob = (index: number, job: keyof UnifiedWallet['jobs']) => {
    const updated = { ...settings };
    updated.wallets = [...updated.wallets];
    updated.wallets[index] = {
      ...updated.wallets[index],
      jobs: { ...updated.wallets[index].jobs, [job]: !updated.wallets[index].jobs[job] }
    };
    setSettings(updated);
  };

  const addWallet = () => {
    const newWallet: UnifiedWallet = {
      label: '',
      address: '',
      type: 'evm',
      jobs: { hyperliquid: false, lighter: false, pendle: false }
    };
    const updated = { ...settings };
    updated.wallets = [...updated.wallets, newWallet];
    setSettings(updated);
  };

  const removeWallet = (index: number) => {
    const updated = { ...settings };
    updated.wallets = [...updated.wallets];
    updated.wallets.splice(index, 1);
    setSettings(updated);
  };

  const toggleEnabled = (section: string, subsection: string, enabled: boolean) => {
    const updated = { ...settings };
    if (section === 'grist' && subsection === '') {
      if (enabled) {
        const backupPath = updated.grist.backupPath.trim();
        if (!backupPath) {
          setShowGristMessage(true);
          return;
        }
        if (backupPath.length < 3 || !backupPath.includes('.')) {
          setShowGristMessage(true);
          return;
        }
      }
      updated.grist.enabled = enabled;
      setShowGristMessage(false);
    } else if (section === 'exchanges' && subsection === 'kraken') {
      if (enabled) {
        const apiKey = updated.exchanges.kraken.apiKey.trim();
        const apiSecret = updated.exchanges.kraken.apiSecret.trim();
        if (!apiKey || !apiSecret) {
          setShowKrakenMessage(true);
          return;
        }
      }
      updated.exchanges.kraken.enabled = enabled;
      setShowKrakenMessage(false);
    } else if (section === 'onchain' && (subsection === 'evm' || subsection === 'bitcoin' || subsection === 'solana')) {
      if (enabled) {
        const apiKey = updated.onchain.coinstatsApiKey.trim();
        if (!apiKey) {
          setShowCoinStatsMessage(true);
          return;
        }
      }
      (updated.onchain as any)[subsection].enabled = enabled;
      setShowCoinStatsMessage(false);
    } else if (section === 'settings' && subsection === 'stocks') {
      if (enabled) {
        const apiKey = updated.settings.stocks.twelveDataApiKey.trim();
        if (!apiKey) {
          setShowStocksMessage(true);
          return;
        }
      }
      updated.settings.stocks.enabled = enabled;
      setShowStocksMessage(false);
    } else if (section === 'settings' && subsection === 'prices') {
      updated.settings.prices.enabled = enabled;
    } else {
      (updated as any)[section][subsection].enabled = enabled;
    }
    setSettings(updated);
    saveSettingsAndUpdate(updated);
  };

  return (
    <div className="settings-single-column">
      <Card title="Grist">
        <div style={{ display: 'flex', gap: '8px' }}>
          <InputField
            label="API Key"
            value={settings.grist.apiKey}
            onChange={(value) => setSettings({ ...settings, grist: { ...settings.grist, apiKey: value } })}
            placeholder="API Key"
          />
          <InputField
            label="Document ID"
            value={settings.grist.documentId}
            onChange={(value) => setSettings({ ...settings, grist: { ...settings.grist, documentId: value } })}
            placeholder="Document ID"
          />
        </div>

        <SettingRow>
          <Switch
            checked={settings.grist.enabled}
            onChange={(enabled) => toggleEnabled('grist', '', enabled)}
            label="Backup"
            disabled={!settings.grist.backupPath.trim() || settings.grist.backupPath.trim().length < 3 || !settings.grist.backupPath.trim().includes('.')}
          />
          <IntervalInput
            value={settings.grist.interval}
            onChange={(interval) => updateInterval('grist', '', interval)}
            disabled={!settings.grist.backupPath.trim() || settings.grist.backupPath.trim().length < 3 || !settings.grist.backupPath.trim().includes('.')}
          />
          <InputField
            label="Backup Path"
            value={settings.grist.backupPath}
            onChange={(value) => {
              setSettings({ ...settings, grist: { ...settings.grist, backupPath: value } });
              setShowGristMessage(false);
            }}
            placeholder="Backup Path"
            style={{ height: '35px', flex: 2 }}
          />
        </SettingRow>

        <ErrorMessage
          message="Enter a valid file path (e.g., /path/to/backup.json) to enable Grist backup"
          show={showGristMessage}
        />
      </Card>

      <Card title="Exchanges">
        <SettingRow>
          <Switch
            checked={settings.exchanges.kraken.enabled}
            onChange={(enabled) => toggleEnabled('exchanges', 'kraken', enabled)}
            label="Kraken"
            disabled={!settings.exchanges.kraken.apiKey.trim() || !settings.exchanges.kraken.apiSecret.trim()}
          />
          <IntervalInput
            value={settings.exchanges.kraken.interval}
            onChange={(interval) => updateInterval('exchanges', 'kraken', interval)}
            disabled={!settings.exchanges.kraken.apiKey.trim() || !settings.exchanges.kraken.apiSecret.trim()}
          />
          <InputField
            label="API Key"
            value={settings.exchanges.kraken.apiKey}
            onChange={(value) => setSettings({ ...settings, exchanges: { ...settings.exchanges, kraken: { ...settings.exchanges.kraken, apiKey: value } } })}
            placeholder="API Key"
            style={{ height: '35px', flex: 1 }}
          />
          <InputField
            label="API Secret"
            value={settings.exchanges.kraken.apiSecret}
            onChange={(value) => setSettings({ ...settings, exchanges: { ...settings.exchanges, kraken: { ...settings.exchanges.kraken, apiSecret: value } } })}
            placeholder="API Secret"
            type="password"
            style={{ height: '35px', flex: 1 }}
          />
        </SettingRow>

        <ErrorMessage
          message="Enter both API Key and API Secret to enable Kraken"
          show={showKrakenMessage}
        />

        <SettingRow>
          <Switch
            checked={settings.exchanges.hyperliquid.enabled}
            onChange={(enabled) => toggleEnabled('exchanges', 'hyperliquid', enabled)}
            label="Hyperliquid"
          />
          <IntervalInput
            value={settings.exchanges.hyperliquid.interval}
            onChange={(interval) => updateInterval('exchanges', 'hyperliquid', interval)}
          />
        </SettingRow>

        <SettingRow>
          <Switch
            checked={settings.exchanges.lighter.enabled}
            onChange={(enabled) => toggleEnabled('exchanges', 'lighter', enabled)}
            label="Lighter"
          />
          <IntervalInput
            value={settings.exchanges.lighter.interval}
            onChange={(interval) => updateInterval('exchanges', 'lighter', interval)}
          />
        </SettingRow>
      </Card>

      <Card title="Pendle">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <SettingRow>
              <Switch
                checked={settings.pendle.markets.enabled}
                onChange={(enabled) => toggleEnabled('pendle', 'markets', enabled)}
                label="Markets"
              />
              <IntervalInput
                value={settings.pendle.markets.interval}
                onChange={(interval) => updateInterval('pendle', 'markets', interval)}
              />
            </SettingRow>
          </div>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <SettingRow>
              <Switch
                checked={settings.pendle.positions.enabled}
                onChange={(enabled) => toggleEnabled('pendle', 'positions', enabled)}
                label="Positions"
              />
              <IntervalInput
                value={settings.pendle.positions.interval}
                onChange={(interval) => updateInterval('pendle', 'positions', interval)}
              />
            </SettingRow>
          </div>
        </div>
      </Card>

      <Card title="On-Chain">
        <InputField
          label="CoinStats API Key"
          value={settings.onchain.coinstatsApiKey}
          onChange={(value) => {
            setSettings({ ...settings, onchain: { ...settings.onchain, coinstatsApiKey: value } });
            setShowCoinStatsMessage(false);
          }}
          placeholder="CoinStats API Key"
          style={{ width: '450px' }}
        />

        <Card title="Balances" variant="sub">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
              <SettingRow>
                <Switch
                  checked={settings.onchain.evm.enabled}
                  onChange={(enabled) => toggleEnabled('onchain', 'evm', enabled)}
                  label="EVM"
                  disabled={!settings.onchain.coinstatsApiKey.trim()}
                />
                <IntervalInput
                  value={settings.onchain.evm.interval}
                  onChange={(interval) => updateInterval('onchain', 'evm', interval)}
                  disabled={!settings.onchain.coinstatsApiKey.trim()}
                />
              </SettingRow>
            </div>
            <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
              <SettingRow>
                <Switch
                  checked={settings.onchain.bitcoin.enabled}
                  onChange={(enabled) => toggleEnabled('onchain', 'bitcoin', enabled)}
                  label="Bitcoin"
                  disabled={!settings.onchain.coinstatsApiKey.trim()}
                />
                <IntervalInput
                  value={settings.onchain.bitcoin.interval}
                  onChange={(interval) => updateInterval('onchain', 'bitcoin', interval)}
                  disabled={!settings.onchain.coinstatsApiKey.trim()}
                />
              </SettingRow>
            </div>
            <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
              <SettingRow>
                <Switch
                  checked={settings.onchain.solana.enabled}
                  onChange={(enabled) => toggleEnabled('onchain', 'solana', enabled)}
                  label="Solana"
                  disabled={!settings.onchain.coinstatsApiKey.trim()}
                />
                <IntervalInput
                  value={settings.onchain.solana.interval}
                  onChange={(interval) => updateInterval('onchain', 'solana', interval)}
                  disabled={!settings.onchain.coinstatsApiKey.trim()}
                />
              </SettingRow>
            </div>
          </div>

          <ErrorMessage
            message="Enter CoinStats API Key to enable balance fetching"
            show={showCoinStatsMessage}
          />
        </Card>
      </Card>

      <Card title="Price Feeds">
        <SettingRow>
          <Switch
            checked={settings.settings.prices.enabled}
            onChange={(enabled) => toggleEnabled('settings', 'prices', enabled)}
            label="Crypto"
          />
          <IntervalInput
            value={settings.settings.prices.interval}
            onChange={(interval) => updateInterval('settings', 'prices', interval)}
          />
          <InputField
            label="CoinGecko API Key"
            value={settings.settings.prices.coingeckoApiKey}
            onChange={(value) => setSettings({ ...settings, settings: { ...settings.settings, prices: { ...settings.settings.prices, coingeckoApiKey: value } } })}
            placeholder="CoinGecko API Key"
            style={{ height: '35px', flex: 2 }}
          />
        </SettingRow>

        <SettingRow>
          <Switch
            checked={settings.settings.stocks.enabled}
            onChange={(enabled) => toggleEnabled('settings', 'stocks', enabled)}
            label="Stocks"
            disabled={!settings.settings.stocks.twelveDataApiKey.trim()}
          />
          <IntervalInput
            value={settings.settings.stocks.interval}
            onChange={(interval) => updateInterval('settings', 'stocks', interval)}
            disabled={!settings.settings.stocks.twelveDataApiKey.trim()}
          />
          <InputField
            label="TwelveData API Key"
            value={settings.settings.stocks.twelveDataApiKey}
            onChange={(value) => {
              setSettings({ ...settings, settings: { ...settings.settings, stocks: { ...settings.settings.stocks, twelveDataApiKey: value } } });
              setShowStocksMessage(false);
            }}
            placeholder="TwelveData API Key"
            style={{ height: '35px', flex: 2 }}
          />
        </SettingRow>

        <ErrorMessage
          message="Enter TwelveData API Key to enable stocks"
          show={showStocksMessage}
        />
      </Card>


      <Card title="Wallets">
        {settings.wallets.map((wallet, index) => (
          <WalletRow
            key={index}
            wallet={wallet}
            index={index}
            onUpdate={updateWallet}
            onToggleJob={toggleWalletJob}
            onRemove={removeWallet}
          />
        ))}
        <Button onClick={addWallet} style={{ width: '120px', minWidth: '120px' }}>
          Add Wallet
        </Button>
      </Card>

      <div style={{ height: '50px' }} />
      <SaveButton
        hasChanges={hasChanges}
        onSave={() => saveSettingsAndUpdate(settings)}
      />
    </div>
  );
}

