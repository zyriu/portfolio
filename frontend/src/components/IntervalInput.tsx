import React from "react";
import { Input } from "./common";
import { secondsToMinutesSeconds, minutesSecondsToSeconds } from "../types";

interface IntervalInputProps {
    value: number;
    onChange: (newInterval: number) => void;
    disabled?: boolean;
    showLabel?: boolean;
}

export function IntervalInput({ value, onChange, disabled = false, showLabel = true }: IntervalInputProps) {
    const time = secondsToMinutesSeconds(value);

    const handleMinutesChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newMinutes = Number(e.target.value);
        onChange(minutesSecondsToSeconds(newMinutes, time.seconds));
    };

    const handleSecondsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newSeconds = Number(e.target.value);
        onChange(minutesSecondsToSeconds(time.minutes, newSeconds));
    };

    const inputComponent = (
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
            <Input
                type="number"
                min={0}
                value={time.minutes}
                onChange={handleMinutesChange}
                style={{ width: '60px' }}
                placeholder="min"
                disabled={disabled}
            />
            <span>:</span>
            <Input
                type="number"
                min={0}
                max={59}
                value={time.seconds}
                onChange={handleSecondsChange}
                style={{ width: '60px' }}
                placeholder="sec"
                disabled={disabled}
            />
        </div>
    );

    if (!showLabel) {
        return inputComponent;
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: '500', color: 'var(--text-secondary)' }}>
                Interval (m:s)
            </span>
            {inputComponent}
        </div>
    );
}
