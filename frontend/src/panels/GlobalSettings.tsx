import React, { useEffect, useState } from "react";
import { loadSettings, saveSettings } from "../backend";
import { Settings, UnifiedWallet, secondsToMinutesSeconds, minutesSecondsToSeconds } from "../types";
import { Card, Input, Switch, Button } from "../components/common";

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

  // Check if there are unsaved changes
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
    
    // If changing type from EVM to non-EVM, clear all job toggles
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
      // Don't allow enabling Grist if backup path is empty or invalid
      if (enabled) {
        const backupPath = updated.grist.backupPath.trim();
        if (!backupPath) {
          setShowGristMessage(true);
          return; // Don't enable if backup path is empty
        }
        // Basic validation for path format - should have at least a dot and some extension
        if (backupPath.length < 3 || !backupPath.includes('.')) {
          setShowGristMessage(true);
          return; // Don't enable if path doesn't look like a file
        }
      }
      updated.grist.enabled = enabled;
      setShowGristMessage(false); // Hide message when successfully toggling
    } else if (section === 'exchanges' && subsection === 'kraken') {
      // Don't allow enabling Kraken if either API key or secret is empty
      if (enabled) {
        const apiKey = updated.exchanges.kraken.apiKey.trim();
        const apiSecret = updated.exchanges.kraken.apiSecret.trim();
        if (!apiKey || !apiSecret) {
          setShowKrakenMessage(true);
          return; // Don't enable if either field is empty
        }
      }
      updated.exchanges.kraken.enabled = enabled;
      setShowKrakenMessage(false); // Hide message when successfully toggling
    } else if (section === 'onchain' && (subsection === 'evm' || subsection === 'bitcoin' || subsection === 'solana')) {
      // Don't allow enabling EVM/Non-EVM balance jobs if CoinStats API key is empty
      if (enabled) {
        const apiKey = updated.onchain.coinstatsApiKey.trim();
        if (!apiKey) {
          setShowCoinStatsMessage(true);
          return; // Don't enable if API key is empty
        }
      }
      (updated.onchain as any)[subsection].enabled = enabled;
      setShowCoinStatsMessage(false); // Hide message when successfully toggling
    } else if (section === 'settings' && subsection === 'stocks') {
      // Don't allow enabling Stocks if TwelveData API key is empty
      if (enabled) {
        const apiKey = updated.settings.stocks.twelveDataApiKey.trim();
        if (!apiKey) {
          setShowStocksMessage(true);
          return; // Don't enable if API key is empty
        }
      }
      updated.settings.stocks.enabled = enabled;
      setShowStocksMessage(false); // Hide message when successfully toggling
    } else if (section === 'settings' && subsection === 'prices') {
      // Allow enabling Crypto prices without API key
      updated.settings.prices.enabled = enabled;
    } else {
      (updated as any)[section][subsection].enabled = enabled;
    }
    setSettings(updated);
    saveSettingsAndUpdate(updated);
  };

  return (
    <div className="settings-single-column">
      {/* Grist Section */}
      <Card title="Grist">
        <div style={{ display: 'flex', alignItems: 'center', marginBottom: '-0.5rem' }}>
          <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>API Key</span>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Document ID</span>
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px' }}>
          <Input 
            value={settings.grist.apiKey} 
            onChange={e => setSettings({ ...settings, grist: { ...settings.grist, apiKey: e.target.value }})} 
            placeholder="API Key"
            style={{ flex: 1 }}
          />
          <Input 
            value={settings.grist.documentId} 
            onChange={e => setSettings({ ...settings, grist: { ...settings.grist, documentId: e.target.value }})} 
            placeholder="Document ID"
            style={{ flex: 1 }}
          />
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Backup</span>
              <div style={{ 
                opacity: (!settings.grist.backupPath.trim() || settings.grist.backupPath.trim().length < 3 || !settings.grist.backupPath.trim().includes('.')) ? 0.5 : 1,
                cursor: (!settings.grist.backupPath.trim() || settings.grist.backupPath.trim().length < 3 || !settings.grist.backupPath.trim().includes('.')) ? 'default' : 'pointer',
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.grist.enabled} 
                  onChange={v => toggleEnabled('grist', '', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.grist.interval).minutes}
                  onChange={e => updateInterval('grist', '', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.grist.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.grist.interval).seconds}
                  onChange={e => updateInterval('grist', '', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.grist.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', flex: 2 }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Backup Path</span>
              <Input 
                value={settings.grist.backupPath} 
                onChange={e => {
                  setSettings({ ...settings, grist: { ...settings.grist, backupPath: e.target.value }});
                  setShowGristMessage(false); // Hide message when user starts typing
                }} 
                placeholder="Backup Path"
                style={{ height: '35px', width: '300px' }}
              />
            </div>
          </div>
        </div>
        {showGristMessage && (
          <div style={{ 
            marginTop: '8px', 
            fontSize: '12px', 
            color: 'var(--error-text, #c33)', 
            fontStyle: 'italic' 
          }}>
            Enter a valid file path (e.g., /path/to/backup.json) to enable Grist backup
          </div>
        )}
      </Card>

      {/* Exchanges Section */}
      <Card title="Exchanges">
        {/* Kraken Settings */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Kraken</span>
              <div style={{ 
                opacity: (!settings.exchanges.kraken.apiKey.trim() || !settings.exchanges.kraken.apiSecret.trim()) ? 0.5 : 1,
                cursor: (!settings.exchanges.kraken.apiKey.trim() || !settings.exchanges.kraken.apiSecret.trim()) ? 'default' : 'pointer',
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.exchanges.kraken.enabled} 
                  onChange={v => toggleEnabled('exchanges', 'kraken', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.exchanges.kraken.interval).minutes}
                  onChange={e => updateInterval('exchanges', 'kraken', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.exchanges.kraken.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.exchanges.kraken.interval).seconds}
                  onChange={e => updateInterval('exchanges', 'kraken', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.exchanges.kraken.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', flex: 2 }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>API Key</span>
              <Input 
                value={settings.exchanges.kraken.apiKey} 
                onChange={e => setSettings({ ...settings, exchanges: { ...settings.exchanges, kraken: { ...settings.exchanges.kraken, apiKey: e.target.value }}})} 
                placeholder="API Key"
                style={{ height: '35px' }}
              />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', flex: 2 }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>API Secret</span>
              <Input 
                type="password"
                value={settings.exchanges.kraken.apiSecret} 
                onChange={e => {
                  setSettings({ ...settings, exchanges: { ...settings.exchanges, kraken: { ...settings.exchanges.kraken, apiSecret: e.target.value }}});
                  setShowKrakenMessage(false); // Hide message when user starts typing
                }} 
                placeholder="API Secret"
                style={{ height: '35px' }}
              />
            </div>
          </div>
        </div>
        {showKrakenMessage && (
          <div style={{ 
            marginTop: '8px', 
            fontSize: '12px', 
            color: 'var(--error-text, #c33)', 
            fontStyle: 'italic' 
          }}>
            Enter both API Key and API Secret to enable Kraken
          </div>
        )}
        {/* Hyperliquid Settings */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Hyperliquid</span>
              <div style={{ 
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.exchanges.hyperliquid.enabled} 
                  onChange={v => toggleEnabled('exchanges', 'hyperliquid', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.exchanges.hyperliquid.interval).minutes}
                  onChange={e => updateInterval('exchanges', 'hyperliquid', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.exchanges.hyperliquid.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.exchanges.hyperliquid.interval).seconds}
                  onChange={e => updateInterval('exchanges', 'hyperliquid', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.exchanges.hyperliquid.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
          </div>
        </div>
        {/* Lighter Settings */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Lighter</span>
              <div style={{ 
                display: 'flex',
                alignItems: 'center',
              }}>
                <Switch 
                  checked={settings.exchanges.lighter.enabled} 
                  onChange={v => toggleEnabled('exchanges', 'lighter', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.exchanges.lighter.interval).minutes}
                  onChange={e => updateInterval('exchanges', 'lighter', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.exchanges.lighter.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.exchanges.lighter.interval).seconds}
                  onChange={e => updateInterval('exchanges', 'lighter', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.exchanges.lighter.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Pendle Section */}
      <Card title="Pendle">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          {/* Markets Group - Left Side */}
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Markets</span>
              <div style={{ 
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.pendle.markets.enabled} 
                  onChange={v => toggleEnabled('pendle', 'markets', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.pendle.markets.interval).minutes}
                  onChange={e => updateInterval('pendle', 'markets', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.pendle.markets.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.pendle.markets.interval).seconds}
                  onChange={e => updateInterval('pendle', 'markets', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.pendle.markets.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
          </div>

          {/* Positions Group - Right Side */}
          <div style={{ display: 'flex', gap: '1rem', flex: 1}}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Positions</span>
              <div style={{ 
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.pendle.positions.enabled} 
                  onChange={v => toggleEnabled('pendle', 'positions', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.pendle.positions.interval).minutes}
                  onChange={e => updateInterval('pendle', 'positions', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.pendle.positions.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.pendle.positions.interval).seconds}
                  onChange={e => updateInterval('pendle', 'positions', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.pendle.positions.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* On-Chain Section */}
      <Card title="On-Chain">
        <div style={{ display: 'flex', alignItems: 'center', marginBottom: '-0.5rem' }}>
          <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>CoinStats API Key</span>
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px', width: '450px' }}>
          <Input 
            value={settings.onchain.coinstatsApiKey} 
            onChange={e => {
              setSettings({ ...settings, onchain: { ...settings.onchain, coinstatsApiKey: e.target.value }});
              setShowCoinStatsMessage(false); // Hide message when user starts typing
            }} 
            placeholder="CoinStats API Key"
            style={{ flex: 1 }}
          />
        </div>
        {/* On-Chain Balances Row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          {/* EVM Balances Group - Left Side */}
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>EVM Balances</span>
              <div style={{ 
                opacity: (!settings.onchain.coinstatsApiKey.trim()) ? 0.5 : 1,
                cursor: (!settings.onchain.coinstatsApiKey.trim()) ? 'default' : 'pointer',
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.onchain.evm.enabled} 
                  onChange={v => toggleEnabled('onchain', 'evm', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.onchain.evm.interval).minutes}
                  onChange={e => updateInterval('onchain', 'evm', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.onchain.evm.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.onchain.evm.interval).seconds}
                  onChange={e => updateInterval('onchain', 'evm', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.onchain.evm.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
          </div>

          {/* Bitcoin Balances Group - Center */}
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Bitcoin Balances</span>
              <div style={{ 
                opacity: (!settings.onchain.coinstatsApiKey.trim()) ? 0.5 : 1,
                cursor: (!settings.onchain.coinstatsApiKey.trim()) ? 'default' : 'pointer',
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.onchain.bitcoin.enabled} 
                  onChange={v => toggleEnabled('onchain', 'bitcoin', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.onchain.bitcoin.interval).minutes}
                  onChange={e => updateInterval('onchain', 'bitcoin', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.onchain.bitcoin.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.onchain.bitcoin.interval).seconds}
                  onChange={e => updateInterval('onchain', 'bitcoin', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.onchain.bitcoin.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
          </div>

          {/* Solana Balances Group - Right Side */}
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Solana Balances</span>
              <div style={{ 
                opacity: (!settings.onchain.coinstatsApiKey.trim()) ? 0.5 : 1,
                cursor: (!settings.onchain.coinstatsApiKey.trim()) ? 'default' : 'pointer',
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.onchain.solana.enabled} 
                  onChange={v => toggleEnabled('onchain', 'solana', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Input
                  type="number"
                  min={0}
                  value={secondsToMinutesSeconds(settings.onchain.solana.interval).minutes}
                  onChange={e => updateInterval('onchain', 'solana', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.onchain.solana.interval).seconds))}
                  style={{ width: '60px' }}
                  placeholder="min"
                />
                <span>:</span>
                <Input
                  type="number"
                  min={0}
                  max={59}
                  value={secondsToMinutesSeconds(settings.onchain.solana.interval).seconds}
                  onChange={e => updateInterval('onchain', 'solana', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.onchain.solana.interval).minutes, Number(e.target.value)))}
                  style={{ width: '60px' }}
                  placeholder="sec"
                />
              </div>
            </div>
          </div>
        </div>
        {showCoinStatsMessage && (
          <div style={{ 
            marginTop: '8px', 
            fontSize: '12px', 
            color: 'var(--error-text, #c33)', 
            fontStyle: 'italic' 
          }}>
            Enter CoinStats API Key to enable balance fetching
          </div>
        )}
      </Card>

      <Card title="Feed">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Crypto</span>
              <div style={{ 
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.settings.prices.enabled} 
                  onChange={v => toggleEnabled('settings', 'prices', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
              <Input
                type="number"
                min={0}
                value={secondsToMinutesSeconds(settings.settings.prices.interval).minutes}
                onChange={e => updateInterval('settings', 'prices', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.settings.prices.interval).seconds))}
                style={{ width: '60px' }}
                placeholder="min"
              />
              <span>:</span>
              <Input
                type="number"
                min={0}
                max={59}
                value={secondsToMinutesSeconds(settings.settings.prices.interval).seconds}
                onChange={e => updateInterval('settings', 'prices', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.settings.prices.interval).minutes, Number(e.target.value)))}
                style={{ width: '60px' }}
                placeholder="sec"
              />
            </div>
          </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', flex: 2 }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>CoinGecko API Key</span>
              <Input 
                value={settings.settings.prices.coingeckoApiKey} 
                onChange={e => setSettings({ ...settings, settings: { ...settings.settings, prices: { ...settings.settings.prices, coingeckoApiKey: e.target.value }}})} 
                placeholder="CoinGecko API Key"
                style={{ height: '35px', width: '300px' }}
              />
            </div>
          </div>
        </div>

        {/* Stocks Settings */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', gap: '1rem', flex: 1 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Stocks</span>
              <div style={{ 
                opacity: (!settings.settings.stocks.twelveDataApiKey.trim()) ? 0.5 : 1,
                cursor: (!settings.settings.stocks.twelveDataApiKey.trim()) ? 'default' : 'pointer',
                display: 'flex',
                alignItems: 'center'
              }}>
                <Switch 
                  checked={settings.settings.stocks.enabled} 
                  onChange={v => toggleEnabled('settings', 'stocks', v)}
                />
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>Interval (m:s)</span>
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
              <Input
                type="number"
                min={0}
                value={secondsToMinutesSeconds(settings.settings.stocks.interval).minutes}
                onChange={e => updateInterval('settings', 'stocks', minutesSecondsToSeconds(Number(e.target.value), secondsToMinutesSeconds(settings.settings.stocks.interval).seconds))}
                style={{ width: '60px' }}
                placeholder="min"
              />
              <span>:</span>
              <Input
                type="number"
                min={0}
                max={59}
                value={secondsToMinutesSeconds(settings.settings.stocks.interval).seconds}
                onChange={e => updateInterval('settings', 'stocks', minutesSecondsToSeconds(secondsToMinutesSeconds(settings.settings.stocks.interval).minutes, Number(e.target.value)))}
                style={{ width: '60px' }}
                placeholder="sec"
              />
            </div>
          </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', flex: 2 }}>
              <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>TwelveData API Key</span>
              <Input 
                value={settings.settings.stocks.twelveDataApiKey} 
                onChange={e => {
                  setSettings({ ...settings, settings: { ...settings.settings, stocks: { ...settings.settings.stocks, twelveDataApiKey: e.target.value }}});
                  setShowStocksMessage(false); // Hide message when user starts typing
                }} 
                placeholder="TwelveData API Key"
                style={{ height: '35px', width: '300px' }}
              />
            </div>
          </div>
        </div>
        {showStocksMessage && (
          <div style={{ 
            marginTop: '8px', 
            fontSize: '12px', 
            color: 'var(--error-text, #c33)', 
            fontStyle: 'italic' 
          }}>
            Enter TwelveData API Key to enable stocks
          </div>
        )}
      </Card>


      {/* Wallets Section */}
      <Card title="Wallets">
        {settings.wallets.map((wallet, index) => (
          <div key={index} style={{ 
            display: 'flex', 
            gap: '8px', 
            alignItems: 'center',
            justifyContent: 'space-between',
            backgroundColor: 'var(--bg-secondary)'
          }}>
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center', flex: '0 1 auto', minWidth: 0 }}>
              <Input
                value={wallet.label}
                onChange={e => updateWallet(index, 'label', e.target.value)}
                placeholder="Label"
                style={{ width: '120px', minWidth: '120px' }}
              />
              <Input
                value={wallet.address}
                onChange={e => updateWallet(index, 'address', e.target.value)}
                placeholder="Address"
                style={{ width: '400px', minWidth: '300px' }}
              />
              <select
                value={wallet.type}
                onChange={e => updateWallet(index, 'type', e.target.value as 'evm' | 'solana' | 'bitcoin')}
                style={{ 
                  width: '100px',
                  minWidth: '100px',
                  padding: '0.5rem 0.625rem',
                  backgroundColor: 'var(--bg-secondary)',
                  border: '1px solid var(--border)',
                  borderRadius: '4px',
                  color: 'var(--text-primary)',
                  fontSize: '0.875rem',
                  cursor: 'pointer',
                  transition: 'border-color 0.15s',
                  appearance: 'none',
                  WebkitAppearance: 'none',
                  MozAppearance: 'none',
                  backgroundImage: 'url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns=\'http://www.w3.org/2000/svg\' viewBox=\'0 0 24 24\' fill=\'none\' stroke=\'%2394a3b8\' stroke-width=\'2\' stroke-linecap=\'round\' stroke-linejoin=\'round\'%3e%3cpolyline points=\'6,9 12,15 18,9\'%3e%3c/polyline%3e%3c/svg%3e")',
                  backgroundRepeat: 'no-repeat',
                  backgroundPosition: 'right 0.625rem center',
                  backgroundSize: '1rem',
                  paddingRight: '2rem'
                }}
              >
                <option value="evm">EVM</option>
                <option value="solana">Solana</option>
                <option value="bitcoin">Bitcoin</option>
              </select>
            </div>
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center', flexShrink: 0 }}>
              {wallet.type === 'evm' && (
                <>
                  <Button 
                    onClick={() => toggleWalletJob(index, 'hyperliquid')}
                    style={{ 
                      padding: '0.5rem 0.75rem',
                      backgroundColor: wallet.jobs.hyperliquid ? 'rgba(16, 185, 129, 0.1)' : 'var(--bg-tertiary)',
                      color: wallet.jobs.hyperliquid ? 'var(--success)' : 'var(--text-secondary)',
                      border: wallet.jobs.hyperliquid ? '1px solid var(--success)' : '1px solid var(--border)',
                      borderRadius: '4px',
                      fontSize: '0.8125rem',
                      fontWeight: '500',
                      cursor: 'pointer',
                      transition: 'all 0.15s',
                      whiteSpace: 'nowrap',
                      minWidth: '60px'
                    }}
                  >
                    Hyperliquid
                  </Button>
                  <Button 
                    onClick={() => toggleWalletJob(index, 'lighter')}
                    style={{ 
                      padding: '0.5rem 0.75rem',
                      backgroundColor: wallet.jobs.lighter ? 'rgba(16, 185, 129, 0.1)' : 'var(--bg-tertiary)',
                      color: wallet.jobs.lighter ? 'var(--success)' : 'var(--text-secondary)',
                      border: wallet.jobs.lighter ? '1px solid var(--success)' : '1px solid var(--border)',
                      borderRadius: '4px',
                      fontSize: '0.8125rem',
                      fontWeight: '500',
                      cursor: 'pointer',
                      transition: 'all 0.15s',
                      whiteSpace: 'nowrap',
                      minWidth: '70px'
                    }}
                  >
                    Lighter
                  </Button>
                  <Button 
                    onClick={() => toggleWalletJob(index, 'pendle')}
                    style={{ 
                      padding: '0.5rem 0.75rem',
                      backgroundColor: wallet.jobs.pendle ? 'rgba(16, 185, 129, 0.1)' : 'var(--bg-tertiary)',
                      color: wallet.jobs.pendle ? 'var(--success)' : 'var(--text-secondary)',
                      border: wallet.jobs.pendle ? '1px solid var(--success)' : '1px solid var(--border)',
                      borderRadius: '4px',
                      fontSize: '0.8125rem',
                      fontWeight: '500',
                      cursor: 'pointer',
                      transition: 'all 0.15s',
                      whiteSpace: 'nowrap',
                      minWidth: '70px'
                    }}
                  >
                    Pendle
                  </Button>
                </>
              )}
              <Button 
                onClick={() => removeWallet(index)}
                style={{ 
                  padding: '0.5rem 1rem',
                  backgroundColor: 'var(--accent-primary)',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  fontSize: '0.875rem',
                  fontWeight: '500',
                  cursor: 'pointer',
                  transition: 'background-color 0.15s',
                  whiteSpace: 'nowrap'
                }}
              >
                Remove
              </Button>
            </div>
          </div>
        ))}
        <Button onClick={addWallet} style={{ marginTop: '8px', width: '120px', minWidth: '120px' }}>
          Add Wallet
        </Button>
      </Card>

      {/* Bottom padding to account for fixed save button */}
      <div style={{ height: '50px' }} />

      {/* Fixed Save Button */}
      <div style={{ 
        position: 'fixed',
        bottom: 0,
        left: 0,
        right: 0,
        padding: '16px',
        backgroundColor: 'var(--bg-secondary)',
        borderTop: '1px solid var(--border)',
        display: 'flex',
        justifyContent: 'center',
        zIndex: 100
      }}>
        <Button 
          onClick={() => saveSettingsAndUpdate(settings)}
          disabled={!hasChanges}
          style={{ 
            padding: '12px 24px', 
            fontSize: '16px',
            backgroundColor: hasChanges ? 'var(--accent-primary)' : 'var(--bg-tertiary)',
            color: hasChanges ? 'white' : 'var(--text-muted)',
            cursor: hasChanges ? 'pointer' : 'not-allowed',
            opacity: hasChanges ? 1 : 0.5,
            transition: 'all 0.2s ease'
          }}
        >
          Save Settings
        </Button>
      </div>
    </div>
  );
}

