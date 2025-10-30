import React from "react";
import { Input } from "./common";

interface InputFieldProps {
    label: string;
    value: string;
    onChange: (value: string) => void;
    placeholder: string;
    type?: "text" | "password";
    style?: React.CSSProperties;
}

export function InputField({ 
    label, 
    value, 
    onChange, 
    placeholder, 
    type = "text",
    style = { flex: 1 }
}: InputFieldProps) {
    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>
                {label}
            </span>
            <Input
                type={type}
                value={value}
                onChange={e => onChange(e.target.value)}
                placeholder={placeholder}
                style={style}
            />
        </div>
    );
}
