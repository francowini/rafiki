'use client';

import type { ReactNode } from 'react';
import { useState, useEffect, useRef } from 'react';
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

const EXPORT_ROWS_DEFAULT = 100;

interface ExportDialogProps {
  children?: ReactNode;
}

export function ExportDialog({ children }: ExportDialogProps) {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [dateRange, setDateRange] = useState<DateRange>(getDateRangeFromPreset('7d'));
  const [exportData, setExportData] = useState<ExportResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Guard against stale responses overwriting state
  const requestIdRef = useRef(0);

  useEffect(() => {
    if (!open) return;

    const currentRequestId = ++requestIdRef.current;

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

        // Only update state if this is still the latest request
        if (currentRequestId === requestIdRef.current) {
          setExportData(response);
        }
      } catch (err: unknown) {
        // Only update state if this is still the latest request
        if (currentRequestId === requestIdRef.current) {
          const message = err instanceof Error ? err.message : 'Failed to load export data';
          setError(message);
          setExportData(null);
        }
      } finally {
        // Only update state if this is still the latest request
        if (currentRequestId === requestIdRef.current) {
          setIsLoading(false);
        }
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
      const markdown = generateMarkdown(exportData.items, dateRange.startDate, dateRange.endDate);

      const filename = generateExportFilename(dateRange.startDate, dateRange.endDate);

      downloadMarkdown(markdown, filename);

      toast({
        title: 'Export successful',
        description: `Downloaded ${exportData.items.length} entries to ${filename}`,
      });

      setOpen(false);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'An unexpected error occurred';
      toast({
        variant: 'destructive',
        title: 'Export failed',
        description: message,
      });
    } finally {
      setIsExporting(false);
    }
  };

  const momentsCount = exportData?.items.filter((i) => i.itemType === 'moment').length || 0;
  const thinksCount = exportData?.items.filter((i) => i.itemType === 'think').length || 0;
  const hasMoreItems = !!exportData && exportData.items.length < exportData.total;

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
            Export your moments and thinks to a markdown file. Select a date range to get started.
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
          <Button variant="outline" onClick={() => setOpen(false)} disabled={isExporting}>
            Cancel
          </Button>
          <Button
            onClick={handleExport}
            disabled={
              isLoading || isExporting || !exportData || exportData.total === 0 || hasMoreItems
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
