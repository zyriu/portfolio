import React from "react";
import { Button } from "./common";

interface SaveButtonProps {
    hasChanges: boolean;
    onSave: () => void;
}

export function SaveButton({ hasChanges, onSave }: SaveButtonProps) {
    return (
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
                onClick={onSave}
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
    );
}
