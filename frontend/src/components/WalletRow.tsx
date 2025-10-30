import React from "react";
import { Input, Button } from "./common";
import { UnifiedWallet } from "../types";

interface WalletRowProps {
    wallet: UnifiedWallet;
    index: number;
    onUpdate: (index: number, field: keyof UnifiedWallet, value: any) => void;
    onToggleJob: (index: number, job: keyof UnifiedWallet['jobs']) => void;
    onRemove: (index: number) => void;
}

export function WalletRow({ wallet, index, onUpdate, onToggleJob, onRemove }: WalletRowProps) {
    return (
        <div style={{
            display: 'flex',
            gap: '8px',
            alignItems: 'center',
            justifyContent: 'space-between',
            backgroundColor: 'var(--bg-secondary)'
        }}>
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center', flex: '0 1 auto', minWidth: 0 }}>
                <Input
                    value={wallet.label}
                    onChange={e => onUpdate(index, 'label', e.target.value)}
                    placeholder="Label"
                    style={{ width: '120px', minWidth: '120px' }}
                />
                <Input
                    value={wallet.address}
                    onChange={e => onUpdate(index, 'address', e.target.value)}
                    placeholder="Address"
                    style={{ width: '400px', minWidth: '300px' }}
                />
                <select
                    value={wallet.type}
                    onChange={e => onUpdate(index, 'type', e.target.value as 'evm' | 'solana' | 'bitcoin')}
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
                            onClick={() => onToggleJob(index, 'hyperliquid')}
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
                            onClick={() => onToggleJob(index, 'lighter')}
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
                            onClick={() => onToggleJob(index, 'pendle')}
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
                    onClick={() => onRemove(index)}
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
    );
}
