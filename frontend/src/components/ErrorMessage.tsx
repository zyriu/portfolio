import React from "react";

interface ErrorMessageProps {
    message: string;
    show: boolean;
}

export function ErrorMessage({ message, show }: ErrorMessageProps) {
    if (!show) return null;

    return (
        <div style={{
            marginTop: '8px',
            fontSize: '12px',
            color: 'var(--error-text, #c33)',
            fontStyle: 'italic'
        }}>
            {message}
        </div>
    );
}
