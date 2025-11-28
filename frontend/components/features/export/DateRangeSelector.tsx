'use client';

import { useState, useMemo } from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { DateRangePreset, DateRange } from '@/lib/types';
import {
  getDateRangeFromPreset,
  formatDateForInput,
  parseDateFromInput,
  validateDateRange,
} from '@/lib/date-utils';
import { Alert, AlertDescription } from '@/components/ui/alert';

interface DateRangeSelectorProps {
  value: DateRange;
  onChange: (range: DateRange) => void;
}

const PRESET_OPTIONS: { value: DateRangePreset; label: string }[] = [
  { value: '7d', label: 'Last 7 days' },
  { value: '14d', label: 'Last 14 days' },
  { value: '30d', label: 'Last 30 days' },
  { value: 'custom', label: 'Custom range' },
];

/**
 * Detects which preset matches the given date range, or 'custom' if none match.
 */
function detectPresetFromRange(range: DateRange): DateRangePreset {
  const presets: Exclude<DateRangePreset, 'custom'>[] = ['7d', '14d', '30d'];

  for (const preset of presets) {
    const presetRange = getDateRangeFromPreset(preset);
    // Compare dates by day (ignore time differences)
    const sameStart =
      formatDateForInput(range.startDate) === formatDateForInput(presetRange.startDate);
    const sameEnd = formatDateForInput(range.endDate) === formatDateForInput(presetRange.endDate);

    if (sameStart && sameEnd) {
      return preset;
    }
  }

  return 'custom';
}

export function DateRangeSelector({ value, onChange }: DateRangeSelectorProps) {
  // Track if user explicitly selected custom mode (even if dates happen to match a preset)
  const [isCustomMode, setIsCustomMode] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);

  // Derive the active preset from the current date range value
  const detectedPreset = useMemo(() => detectPresetFromRange(value), [value]);

  // Show custom mode if user explicitly selected it OR if dates don't match any preset
  const activePreset = isCustomMode ? 'custom' : detectedPreset;

  const handlePresetChange = (preset: DateRangePreset) => {
    setValidationError(null);

    if (preset === 'custom') {
      setIsCustomMode(true);
    } else {
      setIsCustomMode(false);
      const range = getDateRangeFromPreset(preset);
      onChange(range);
    }
  };

  const handleCustomDateChange = (type: 'start' | 'end', dateString: string) => {
    const newDate = parseDateFromInput(dateString);
    const newRange =
      type === 'start'
        ? { startDate: newDate, endDate: value.endDate }
        : { startDate: value.startDate, endDate: newDate };

    const error = validateDateRange(newRange.startDate, newRange.endDate);
    setValidationError(error);

    if (!error) {
      onChange(newRange);
    }
  };

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>Quick Select</Label>
        <div className="flex flex-wrap gap-2">
          {PRESET_OPTIONS.map((option) => (
            <Button
              key={option.value}
              type="button"
              variant={activePreset === option.value ? 'default' : 'outline'}
              size="sm"
              onClick={() => handlePresetChange(option.value)}
            >
              {option.label}
            </Button>
          ))}
        </div>
      </div>

      {activePreset === 'custom' && (
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="startDate">Start Date</Label>
            <Input
              id="startDate"
              type="date"
              value={formatDateForInput(value.startDate)}
              onChange={(e) => handleCustomDateChange('start', e.target.value)}
              max={formatDateForInput(value.endDate)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="endDate">End Date</Label>
            <Input
              id="endDate"
              type="date"
              value={formatDateForInput(value.endDate)}
              onChange={(e) => handleCustomDateChange('end', e.target.value)}
              min={formatDateForInput(value.startDate)}
              max={formatDateForInput(new Date())}
            />
          </div>
        </div>
      )}

      {validationError && (
        <Alert variant="destructive">
          <AlertDescription>{validationError}</AlertDescription>
        </Alert>
      )}
    </div>
  );
}
