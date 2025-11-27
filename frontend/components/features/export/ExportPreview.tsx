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
              <Badge variant="outline" className="border-purple-300 bg-purple-50">
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
