import { DateRange, DateRangePreset } from './types';

export function getDateRangeFromPreset(preset: DateRangePreset): DateRange {
  const endDate = new Date();
  endDate.setHours(23, 59, 59, 999);

  const startDate = new Date();
  startDate.setHours(0, 0, 0, 0);

  switch (preset) {
    case '7d':
      startDate.setDate(startDate.getDate() - 7);
      break;
    case '14d':
      startDate.setDate(startDate.getDate() - 14);
      break;
    case '30d':
      startDate.setDate(startDate.getDate() - 30);
      break;
    case '90d':
      startDate.setDate(startDate.getDate() - 90);
      break;
    case 'custom':
      break;
  }

  return { startDate, endDate };
}

export function formatDateForAPI(date: Date): string {
  return date.toISOString();
}

export function formatDateForInput(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export function parseDateFromInput(value: string): Date {
  return new Date(`${value}T00:00:00`);
}

export function formatDateForMarkdown(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function formatDateTimeForMarkdown(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function validateDateRange(startDate: Date, endDate: Date): string | null {
  if (startDate > endDate) {
    return 'Start date must be before end date';
  }

  const maxDays = 365;
  const diffDays = Math.ceil((endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24));

  if (diffDays > maxDays) {
    return `Date range cannot exceed ${maxDays} days`;
  }

  return null;
}
