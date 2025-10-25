import React from "react";

export function Card({ title, children, actions }: { title: string; children: React.ReactNode; actions?: React.ReactNode }) {
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

export function Switch({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      onClick={() => onChange(!checked)}
      className={`panel-switch ${checked ? "panel-switch-enabled" : ""}`}
    >
      {checked ? "Enabled" : "Disabled"}
    </button>
  );
}

export function Button(props: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} className={`panel-button ${props.className || ""}`} />;
}
