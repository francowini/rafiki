'use client';

import { useState } from 'react';
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

export function DateRangeSelector({ value, onChange }: DateRangeSelectorProps) {
  const [selectedPreset, setSelectedPreset] = useState<DateRangePreset>('7d');
  const [validationError, setValidationError] = useState<string | null>(null);

  const handlePresetChange = (preset: DateRangePreset) => {
    setSelectedPreset(preset);
    setValidationError(null);

    if (preset !== 'custom') {
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
              variant={selectedPreset === option.value ? 'default' : 'outline'}
              size="sm"
              onClick={() => handlePresetChange(option.value)}
            >
              {option.label}
            </Button>
          ))}
        </div>
      </div>

      {selectedPreset === 'custom' && (
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
