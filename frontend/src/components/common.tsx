import React from "react";

export function Card({
  title,
  children,
  actions,
  variant = "main"
}: {
  title: string;
  children: React.ReactNode;
  actions?: React.ReactNode;
  variant?: "main" | "sub";
}) {
  if (variant === "sub") {
    return (
      <div style={{
        backgroundColor: 'var(--bg-secondary)',
        border: '1px solid var(--border)',
        borderRadius: '6px',
        padding: '16px'
      }}>
        <div style={{
          fontSize: '0.875rem',
          fontWeight: '600',
          color: 'var(--text-primary)',
          marginBottom: '12px',
          paddingBottom: '8px',
          borderBottom: '1px solid var(--border)'
        }}>
          {title}
        </div>
        <div>
          {children}
        </div>
      </div>
    );
  }

  return (
    <div className="panel-card">
      <div className="panel-card-header">
        <div className="panel-card-title">{title}</div>
        {actions}
      </div>
      <div className="panel-card-content">{children}</div>
    </div>
  );
}

export function Row({ children }: { children: React.ReactNode }) {
  return <div className="panel-row">{children}</div>;
}

export function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <label className="panel-field">
      <div className="panel-field-label">{label}</div>
      {children}
    </label>
  );
}

export function Input(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input {...props} className={`panel-input ${props.className || ""}`} />;
}

export function Switch({
  checked,
  onChange,
  label,
  disabled = false
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  label?: string;
  disabled?: boolean;
}) {
  const switchComponent = (
    <button
      onClick={() => !disabled && onChange(!checked)}
      className={`panel-switch ${checked ? "panel-switch-enabled" : ""}`}
      disabled={disabled}
    >
      {checked ? "Enabled" : "Disabled"}
    </button>
  );

  if (!label) {
    return switchComponent;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
      <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>
        {label}
      </span>
      <div style={{
        opacity: disabled ? 0.5 : 1,
        cursor: disabled ? 'default' : 'pointer',
        display: 'flex',
        alignItems: 'center'
      }}>
        {switchComponent}
      </div>
    </div>
  );
}

export function Button(props: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} className={`panel-button ${props.className || ""}`} />;
}
