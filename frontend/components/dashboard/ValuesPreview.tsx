'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Value } from '@/lib/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { getFacetConfig } from '@/lib/value-utils';
import { ArrowRight, Loader2 } from 'lucide-react';
import Link from 'next/link';

export function ValuesPreview() {
  const [values, setValues] = useState<Value[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchValues = async () => {
      try {
        const response = await api.values.getAll();
        // Sort by displayOrder and show all values
        const sortedValues = [...response.items].sort(
          (a, b) => a.displayOrder - b.displayOrder,
        );
        setValues(sortedValues);
        setError(null);
      } catch (err: unknown) {
        const errorMessage =
          err instanceof Error ? err.message : 'Failed to load values preview';
        setError(errorMessage);
      } finally {
        setIsLoading(false);
      }
    };

    fetchValues();
  }, []);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Your Core Values</CardTitle>
          <CardDescription>What matters most to you</CardDescription>
        </CardHeader>
        <CardContent className="flex items-center justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin text-rose-600" />
          <span className="ml-2 text-sm text-muted-foreground">Loading values...</span>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Your Core Values</CardTitle>
          <CardDescription>What matters most to you</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  if (values.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">Your Core Values</CardTitle>
            <CardDescription>What matters most to you</CardDescription>
          </div>
          <Link
            href="/values"
            className="text-rose-600 hover:text-rose-700 flex items-center gap-1 text-sm font-medium"
          >
            Manage
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {values.map((value) => {
          const facetConfig = getFacetConfig(value.facet);
          return (
            <div key={value.id} className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
              <Badge
                variant="outline"
                className="bg-gray-100 text-gray-700 border-gray-300 shrink-0"
              >
                #{value.displayOrder}
              </Badge>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium line-clamp-2">{value.content}</p>
                <Badge
                  variant="outline"
                  className={`mt-1 text-xs ${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor}`}
                >
                  {facetConfig.icon} {facetConfig.label}
                </Badge>
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
