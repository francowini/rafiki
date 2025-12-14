import { DateRange, DateRangePreset } from './types';

/**
 * Returns a DateRange for the given preset.
 *
 * The 'custom' preset is intentionally excluded from the type signature.
 * Callers must manage custom date ranges via their own state.
 */
export function getDateRangeFromPreset(preset: Exclude<DateRangePreset, 'custom'>): DateRange {
  const endDate = new Date();
  endDate.setHours(23, 59, 59, 999);

  const startDate = new Date();
  startDate.setHours(0, 0, 0, 0);

  // Presets are "last N days including today", so subtract N-1 to get N calendar days
  switch (preset) {
    case '7d':
      startDate.setDate(startDate.getDate() - 6);
      break;
    case '14d':
      startDate.setDate(startDate.getDate() - 13);
      break;
    case '30d':
      startDate.setDate(startDate.getDate() - 29);
      break;
    case '90d':
      startDate.setDate(startDate.getDate() - 89);
      break;
  }

  return { startDate, endDate };
}

/**
 * Converts a Date to ISO 8601 string for API requests.
 */
export function formatDateForAPI(date: Date): string {
  return date.toISOString();
}

/**
 * Formats a Date as YYYY-MM-DD for HTML date input elements.
 */
export function formatDateForInput(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * Parses a YYYY-MM-DD string (from HTML date input) into a Date at local midnight.
 *
 * Note: The resulting Date is at 00:00:00 in the LOCAL timezone.
 * When converted to ISO (e.g., via toISOString()), the time will be adjusted to UTC.
 * If the API expects UTC-day boundaries, consider constructing a UTC midnight Date instead.
 */
export function parseDateFromInput(value: string): Date {
  return new Date(`${value}T00:00:00`);
}

/**
 * Formats a Date for display in markdown exports (e.g., "lunes, 27 de noviembre de 2025").
 * Uses 'es-MX' locale. For i18n support, pass locale as a parameter.
 */
export function formatDateForMarkdown(date: Date | string, locale = 'es-MX'): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString(locale, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

/**
 * Formats a Date with time for display in markdown exports
 * (e.g., "lunes, 27 de noviembre de 2025, 03:45 p.m.").
 * Uses 'es-MX' locale. For i18n support, pass locale as a parameter.
 */
export function formatDateTimeForMarkdown(date: Date | string, locale = 'es-MX'): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleString(locale, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * Validates a date range and returns an error message if invalid, or null if valid.
 */
export function validateDateRange(startDate: Date, endDate: Date): string | null {
  if (startDate > endDate) {
    return 'La fecha de inicio debe ser antes de la fecha de fin';
  }

  const maxDays = 365;
  const diffDays = Math.ceil((endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24));

  if (diffDays > maxDays) {
    return `El rango de fechas no puede exceder ${maxDays} días`;
  }

  return null;
}
