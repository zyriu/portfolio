import React from "react";

interface SettingRowProps {
    children: React.ReactNode;
}

export function SettingRow({ children }: SettingRowProps) {
    return (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div style={{ display: 'flex', gap: '8px', flex: 1 }}>
                {children}
            </div>
        </div>
    );
}
