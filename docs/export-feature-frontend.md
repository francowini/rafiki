# Export Feature - Frontend Implementation

## Overview

Client-side export dialog that fetches data from the new `/v1/export` endpoint and generates markdown for download. Users can select date ranges (default: last 7 days) and export moments + thinks for sharing with their psychologist.

**Backend API Compatibility**: This implementation is aligned with the Query Domain pattern used in the backend. The API supports pagination (`page`, `rows`, `orderBy`) but for export use cases, we default to a high row limit (1000) to fetch all data in the date range without pagination UI.

## File Structure

```
frontend/
├── lib/
│   ├── api.ts                    # Update: add export endpoint
│   ├── types.ts                  # Update: add export types
│   ├── date-utils.ts             # NEW: date utilities
│   └── file-utils.ts             # NEW: file download utilities
├── components/
│   └── features/
│       └── export/
│           ├── index.ts          # NEW: barrel export
│           ├── ExportDialog.tsx  # NEW: main dialog
│           ├── DateRangeSelector.tsx  # NEW: date picker
│           ├── ExportPreview.tsx # NEW: preview card
│           └── markdown-generator.ts  # NEW: markdown gen
└── app/
    └── (dashboard)/
        └── momentos/
            └── page.tsx          # Update: add export button
```

## TypeScript Types

Add to `frontend/lib/types.ts`:

```typescript
// ============================================================================
// Export Types
// ============================================================================

export interface ExportItem {
  id: string;
  itemType: 'moment' | 'think';
  itemDate: string; // ISO 8601 timestamp

  // Moment-specific fields (optional)
  situation?: string;
  thoughts?: string;
  physicalSymptoms?: string;
  behavior?: string;
  consequences?: string;
  valuesReflection?: string;
  intensity?: number;

  // Think-specific fields (optional)
  category?: ThinkCategory;
  content?: string;

  dateCreated: string;
}

export interface ExportParams {
  startDate: string;    // ISO 8601 date string
  endDate: string;      // ISO 8601 date string
  page?: number;        // Page number (1-based, default: 1)
  rows?: number;        // Items per page (default: 1000 for exports)
  orderBy?: string;     // Sort order (default: "item_date,DESC")
}

// Standard Query Domain response format
export interface ExportResponse {
  items: ExportItem[];
  total: number;
  page: number;         // Current page number
  rowsPerPage: number;  // Items per page
}

export type DateRangePreset = '7d' | '14d' | '30d' | '90d' | 'custom';

export interface DateRange {
  startDate: Date;
  endDate: Date;
}
```

## API Integration

Add to `frontend/lib/api.ts`:

```typescript
export const api = {
  // ... existing endpoints

  export: {
    getItems: async (params: ExportParams): Promise<ExportResponse> => {
      const queryParams = new URLSearchParams({
        start_date: params.startDate,
        end_date: params.endDate,
        page: (params.page || 1).toString(),
        rows: (params.rows || 1000).toString(),  // High default for exports
      });

      // Add orderBy if specified (default: item_date,DESC on backend)
      if (params.orderBy) {
        queryParams.set('orderBy', params.orderBy);
      }

      return fetchAPI<ExportResponse>(`/v1/export?${queryParams.toString()}`);
    },
  },
};
```

## Date Utilities

Create `frontend/lib/date-utils.ts`:

```typescript
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
  const diffDays = Math.ceil(
    (endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24)
  );

  if (diffDays > maxDays) {
    return `Date range cannot exceed ${maxDays} days`;
  }

  return null;
}
```

## File Download Utilities

Create `frontend/lib/file-utils.ts`:

```typescript
export function downloadMarkdown(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/markdown;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

export function generateExportFilename(startDate: Date, endDate: Date): string {
  const formatDate = (date: Date) => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  return `journal-export-${formatDate(startDate)}-to-${formatDate(endDate)}.md`;
}
```

## Markdown Generator

Create `frontend/components/features/export/markdown-generator.ts`:

```typescript
import { ExportItem, ThinkCategory } from '@/lib/types';
import { formatDateForMarkdown, formatDateTimeForMarkdown } from '@/lib/date-utils';

export function generateMarkdown(
  items: ExportItem[],
  startDate: Date,
  endDate: Date
): string {
  const lines: string[] = [];

  // Header
  lines.push('# Weekly Journal Export');
  lines.push('');
  lines.push(
    `**Period:** ${formatDateForMarkdown(startDate)} - ${formatDateForMarkdown(endDate)}`
  );
  lines.push('');
  lines.push(`**Total entries:** ${items.length}`);
  lines.push('');
  lines.push('---');
  lines.push('');

  // Separate moments and thinks
  const moments = items.filter((item) => item.itemType === 'moment');
  const thinks = items.filter((item) => item.itemType === 'think');

  // Moments section
  if (moments.length > 0) {
    lines.push('## Moments');
    lines.push('');
    lines.push(
      `Recorded difficult moments for reflection and pattern recognition. Total: ${moments.length}`
    );
    lines.push('');

    const sortedMoments = [...moments].sort((a, b) => {
      const dateA = new Date(a.itemDate);
      const dateB = new Date(b.itemDate);
      return dateB.getTime() - dateA.getTime();
    });

    sortedMoments.forEach((moment, index) => {
      lines.push(...formatMoment(moment, index + 1));
    });
  }

  // Thinks section
  if (thinks.length > 0) {
    lines.push('## Thoughts & Notes');
    lines.push('');
    lines.push(
      `Captured thoughts, ideas, and reflections. Total: ${thinks.length}`
    );
    lines.push('');

    const thinksByCategory = groupThinksByCategory(thinks);

    Object.entries(thinksByCategory).forEach(([category, categoryThinks]) => {
      lines.push(`### ${formatCategory(category as ThinkCategory)}`);
      lines.push('');

      const sortedThinks = [...categoryThinks].sort((a, b) => {
        const dateA = new Date(a.dateCreated);
        const dateB = new Date(b.dateCreated);
        return dateB.getTime() - dateA.getTime();
      });

      sortedThinks.forEach((think, index) => {
        lines.push(...formatThink(think, index + 1));
      });
    });
  }

  // Footer
  lines.push('---');
  lines.push('');
  lines.push(`*Exported on ${formatDateTimeForMarkdown(new Date())}*`);
  lines.push('');

  return lines.join('\n');
}

function formatMoment(moment: ExportItem, number: number): string[] {
  const lines: string[] = [];
  const momentDate = new Date(moment.itemDate);

  lines.push(
    `### ${number}. ${formatDateTimeForMarkdown(momentDate)} - Intensity: ${moment.intensity}/10`
  );
  lines.push('');

  if (moment.situation) {
    lines.push('**Situation:**');
    lines.push(moment.situation);
    lines.push('');
  }

  if (moment.thoughts) {
    lines.push('**Thoughts:**');
    lines.push(moment.thoughts);
    lines.push('');
  }

  if (moment.physicalSymptoms) {
    lines.push('**Physical Symptoms:**');
    lines.push(moment.physicalSymptoms);
    lines.push('');
  }

  if (moment.behavior) {
    lines.push('**Behavior:**');
    lines.push(moment.behavior);
    lines.push('');
  }

  if (moment.consequences) {
    lines.push('**Consequences:**');
    lines.push(moment.consequences);
    lines.push('');
  }

  if (moment.valuesReflection) {
    lines.push('**Values Reflection:**');
    lines.push(moment.valuesReflection);
    lines.push('');
  }

  return lines;
}

function formatThink(think: ExportItem, number: number): string[] {
  const lines: string[] = [];
  const thinkDate = new Date(think.dateCreated);

  lines.push(`#### ${number}. ${formatDateTimeForMarkdown(thinkDate)}`);
  lines.push('');
  lines.push(think.content || '');
  lines.push('');

  return lines;
}

function groupThinksByCategory(
  thinks: ExportItem[]
): Record<string, ExportItem[]> {
  return thinks.reduce(
    (acc, think) => {
      const category = think.category || 'personal';
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category].push(think);
      return acc;
    },
    {} as Record<string, ExportItem[]>
  );
}

function formatCategory(category: ThinkCategory): string {
  const categoryMap: Record<ThinkCategory, string> = {
    personal: 'Personal',
    work: 'Work',
    ideas: 'Ideas',
    learning: 'Learning',
    reflection: 'Reflection',
  };
  return categoryMap[category] || category;
}
```

## DateRangeSelector Component

Create `frontend/components/features/export/DateRangeSelector.tsx`:

```typescript
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

  const handleCustomDateChange = (
    type: 'start' | 'end',
    dateString: string
  ) => {
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
```

## ExportPreview Component

Create `frontend/components/features/export/ExportPreview.tsx`:

```typescript
'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { FileText, Calendar, Download } from 'lucide-react';

interface ExportPreviewProps {
  totalItems: number;
  momentsCount: number;
  thinksCount: number;
  startDate: Date;
  endDate: Date;
  // Optional: show if there are more items than fetched
  fetchedCount?: number;
}

export function ExportPreview({
  totalItems,
  momentsCount,
  thinksCount,
  startDate,
  endDate,
  fetchedCount,
}: ExportPreviewProps) {
  const formatDate = (date: Date) => {
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const hasMoreItems = fetchedCount !== undefined && fetchedCount < totalItems;

  return (
    <Card className="bg-muted/50">
      <CardContent className="pt-6">
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-muted-foreground" />
            <h3 className="font-medium">Export Preview</h3>
          </div>

          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Calendar className="h-4 w-4" />
            <span>
              {formatDate(startDate)} - {formatDate(endDate)}
            </span>
          </div>

          <div className="flex flex-wrap gap-2">
            <Badge variant="secondary" className="gap-1">
              <Download className="h-3 w-3" />
              {totalItems} total entries
            </Badge>
            {momentsCount > 0 && (
              <Badge
                variant="outline"
                className="border-purple-300 bg-purple-50"
              >
                {momentsCount} moments
              </Badge>
            )}
            {thinksCount > 0 && (
              <Badge variant="outline" className="border-blue-300 bg-blue-50">
                {thinksCount} thinks
              </Badge>
            )}
          </div>

          {hasMoreItems && (
            <p className="text-sm text-amber-600">
              Showing {fetchedCount} of {totalItems} entries. Consider using a smaller date range.
            </p>
          )}

          {totalItems === 0 && (
            <p className="text-sm text-muted-foreground italic">
              No entries found in this date range
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
```

## ExportDialog Component

Create `frontend/components/features/export/ExportDialog.tsx`:

```typescript
'use client';

import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Download, Loader2 } from 'lucide-react';
import { DateRangeSelector } from './DateRangeSelector';
import { ExportPreview } from './ExportPreview';
import { api } from '@/lib/api';
import { ExportResponse, DateRange } from '@/lib/types';
import { getDateRangeFromPreset, formatDateForAPI } from '@/lib/date-utils';
import { generateMarkdown } from './markdown-generator';
import { downloadMarkdown, generateExportFilename } from '@/lib/file-utils';
import { useToast } from '@/hooks/use-toast';

// Default rows for export (high limit to fetch all data)
const EXPORT_ROWS_DEFAULT = 1000;

interface ExportDialogProps {
  children?: React.ReactNode;
}

export function ExportDialog({ children }: ExportDialogProps) {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [dateRange, setDateRange] = useState<DateRange>(
    getDateRangeFromPreset('7d')
  );
  const [exportData, setExportData] = useState<ExportResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;

    const fetchExportData = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await api.export.getItems({
          startDate: formatDateForAPI(dateRange.startDate),
          endDate: formatDateForAPI(dateRange.endDate),
          page: 1,
          rows: EXPORT_ROWS_DEFAULT,
          orderBy: 'item_date,DESC',
        });
        setExportData(response);
      } catch (err: unknown) {
        const message =
          err instanceof Error ? err.message : 'Failed to load export data';
        setError(message);
        setExportData(null);
      } finally {
        setIsLoading(false);
      }
    };

    fetchExportData();
  }, [dateRange, open]);

  const handleExport = async () => {
    if (!exportData || exportData.items.length === 0) {
      toast({
        variant: 'destructive',
        title: 'No data to export',
        description: 'There are no entries in the selected date range.',
      });
      return;
    }

    setIsExporting(true);

    try {
      const markdown = generateMarkdown(
        exportData.items,
        dateRange.startDate,
        dateRange.endDate
      );

      const filename = generateExportFilename(
        dateRange.startDate,
        dateRange.endDate
      );

      downloadMarkdown(markdown, filename);

      toast({
        title: 'Export successful',
        description: `Downloaded ${exportData.items.length} entries to ${filename}`,
      });

      setOpen(false);
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'An unexpected error occurred';
      toast({
        variant: 'destructive',
        title: 'Export failed',
        description: message,
      });
    } finally {
      setIsExporting(false);
    }
  };

  const momentsCount =
    exportData?.items.filter((i) => i.itemType === 'moment').length || 0;
  const thinksCount =
    exportData?.items.filter((i) => i.itemType === 'think').length || 0;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children || (
          <Button variant="outline" size="sm">
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Export Journal</DialogTitle>
          <DialogDescription>
            Export your moments and thoughts to a markdown file. Select a date
            range to get started.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          <DateRangeSelector value={dateRange} onChange={setDateRange} />

          {isLoading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          )}

          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {!isLoading && !error && exportData && (
            <ExportPreview
              totalItems={exportData.total}
              momentsCount={momentsCount}
              thinksCount={thinksCount}
              startDate={dateRange.startDate}
              endDate={dateRange.endDate}
              fetchedCount={exportData.items.length}
            />
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={isExporting}
          >
            Cancel
          </Button>
          <Button
            onClick={handleExport}
            disabled={
              isLoading || isExporting || !exportData || exportData.total === 0
            }
            className="gap-2"
          >
            {isExporting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Exporting...
              </>
            ) : (
              <>
                <Download className="h-4 w-4" />
                Export to Markdown
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

## Index Export

Create `frontend/components/features/export/index.ts`:

```typescript
export { ExportDialog } from './ExportDialog';
export { DateRangeSelector } from './DateRangeSelector';
export { ExportPreview } from './ExportPreview';
export { generateMarkdown } from './markdown-generator';
```

## Integration into Moments Page

Update `frontend/app/(dashboard)/momentos/page.tsx`:

Add import:

```typescript
import { ExportDialog } from '@/components/features/export';
import { Download } from 'lucide-react';
```

Update header section (add Export button before New moment):

```typescript
<div className="flex items-center justify-between mb-8">
  <div>
    <h1 className="text-3xl font-bold tracking-tight">Moments</h1>
    <p className="text-muted-foreground mt-1">
      Record and reflect on your difficult moments
    </p>
  </div>

  <div className="flex gap-3">
    {/* Export Button */}
    <ExportDialog>
      <Button variant="outline" size="lg">
        <Download className="h-5 w-5 mr-2" />
        Export
      </Button>
    </ExportDialog>

    {/* New Moment Button */}
    <Sheet open={isFormOpen} onOpenChange={setIsFormOpen}>
      <SheetTrigger asChild>
        <Button size="lg" className="bg-purple-600 hover:bg-purple-700">
          <Plus className="h-5 w-5 mr-2" />
          New moment
        </Button>
      </SheetTrigger>
      {/* ... rest of sheet content ... */}
    </Sheet>
  </div>
</div>
```

## API Specification

### Endpoint

```
GET /v1/export?start_date=2024-11-20T00:00:00Z&end_date=2024-11-27T23:59:59Z&page=1&rows=1000
```

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| start_date | string (RFC3339) | No | - | Filter items >= this date |
| end_date | string (RFC3339) | No | - | Filter items <= this date |
| page | int | No | 1 | Page number (1-based) |
| rows | int | No | 10 | Items per page (use 1000 for exports) |
| orderBy | string | No | item_date,DESC | Sort field and direction |

### Response Format

```json
{
  "items": [...],
  "total": 45,
  "page": 1,
  "rowsPerPage": 1000
}
```

## Example Markdown Output

```markdown
# Weekly Journal Export

**Period:** Wednesday, November 20, 2025 - Wednesday, November 27, 2025

**Total entries:** 12

---

## Moments

Recorded difficult moments for reflection and pattern recognition. Total: 8

### 1. Wednesday, November 27, 2025, 03:45 PM - Intensity: 7/10

**Situation:**
I was at home, alone, after finishing lunch...

**Thoughts:**
I'm wasting my time, I don't know what I'm going to do with my life...

**Physical Symptoms:**
Heart palpitations, sweaty palms, anxiety...

**Behavior:**
I started cleaning the kitchen to distract myself...

**Consequences:**
I felt a bit better while cleaning, but the anxiety returned...

**Values Reflection:**
I avoided sitting with my discomfort and didn't work on my goals...

### 2. Tuesday, November 26, 2025, 10:30 AM - Intensity: 5/10

...

## Thoughts & Notes

Captured thoughts, ideas, and reflections. Total: 4

### Personal

#### 1. Wednesday, November 27, 2025, 06:12 PM

Feeling grateful for the small wins today. Made progress on my project...

### Reflection

#### 1. Tuesday, November 26, 2025, 09:23 AM

Need to remember: it's okay to take breaks...

---

*Exported on Wednesday, November 27, 2025, 08:30 PM*
```

## Implementation Checklist

- [ ] Add types to `frontend/lib/types.ts`
- [ ] Add API method to `frontend/lib/api.ts`
- [ ] Create `frontend/lib/date-utils.ts`
- [ ] Create `frontend/lib/file-utils.ts`
- [ ] Create `frontend/components/features/export/markdown-generator.ts`
- [ ] Create `frontend/components/features/export/DateRangeSelector.tsx`
- [ ] Create `frontend/components/features/export/ExportPreview.tsx`
- [ ] Create `frontend/components/features/export/ExportDialog.tsx`
- [ ] Create `frontend/components/features/export/index.ts`
- [ ] Update `frontend/app/(dashboard)/momentos/page.tsx`
- [ ] Test export with various date ranges
- [ ] Test markdown output format
- [ ] Test file download in different browsers
