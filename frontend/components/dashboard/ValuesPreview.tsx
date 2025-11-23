'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Value } from '@/lib/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { getFacetConfig } from '@/lib/value-utils';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

export function ValuesPreview() {
  const [values, setValues] = useState<Value[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchValues = async () => {
      try {
        const response = await api.values.getAll();
        // Show top 3 values
        setValues(response.items.slice(0, 3));
      } catch (err) {
        console.error('Failed to load values preview:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchValues();
  }, []);

  if (isLoading || values.length === 0) {
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
            View all
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {values.map((value, index) => {
          const facetConfig = getFacetConfig(value.facet);
          return (
            <div key={value.id} className="flex items-start gap-3 p-3 rounded-lg bg-muted/50">
              <Badge
                variant="outline"
                className={
                  index === 0
                    ? 'bg-rose-100 text-rose-800 border-rose-300 font-semibold shrink-0'
                    : 'bg-gray-100 text-gray-700 border-gray-300 shrink-0'
                }
              >
                #{index + 1}
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
