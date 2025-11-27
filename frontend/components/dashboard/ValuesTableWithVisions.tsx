'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api';
import { Value, LifeVision } from '@/lib/types';
import { getFacetConfig } from '@/lib/value-utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Sparkles, ArrowRight } from 'lucide-react';

export function ValuesTableWithVisions() {
  const [values, setValues] = useState<Value[]>([]);
  const [visions, setVisions] = useState<LifeVision[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [valuesRes, visionsRes] = await Promise.all([
          api.values.getAll(),
          api.lifeVisions.getAll(),
        ]);
        setValues(valuesRes.items);
        setVisions(visionsRes.items);
      } catch (err: any) {
        setError(err.message || 'Failed to load data');
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  const getVisionsForValue = (valueId: string) => {
    return visions.filter((v) => v.valueId === valueId);
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-9 w-32" />
        </div>
        <div className="border rounded-lg">
          <div className="p-4 space-y-3">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <p>{error}</p>
      </div>
    );
  }

  if (values.length === 0) {
    return (
      <div className="border rounded-lg p-8 text-center">
        <Sparkles className="h-8 w-8 text-rose-400 mx-auto mb-3" />
        <h3 className="font-medium mb-1">No values defined yet</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Start by defining your core values, then add life visions.
        </p>
        <Link href="/values">
          <Button className="bg-rose-600 hover:bg-rose-700">Define your values</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-rose-600" />
          <h2 className="text-xl font-semibold text-rose-900">Values & Life Visions</h2>
        </div>
        <Link href="/life-visions">
          <Button variant="outline" size="sm" className="gap-2">
            Manage visions
            <ArrowRight className="h-4 w-4" />
          </Button>
        </Link>
      </div>

      <div className="border rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">#</TableHead>
              <TableHead>Value</TableHead>
              <TableHead>Life Visions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {values.map((value) => {
              const valueVisions = getVisionsForValue(value.id);
              const facetConfig = getFacetConfig(value.facet);

              return (
                <TableRow key={value.id}>
                  <TableCell className="font-medium text-muted-foreground">
                    {value.displayOrder}
                  </TableCell>
                  <TableCell>
                    <div className="space-y-1">
                      <p className="text-sm font-medium line-clamp-2">{value.content}</p>
                      <Badge
                        variant="outline"
                        className={`${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor}`}
                      >
                        <span className="mr-1">{facetConfig.icon}</span>
                        {facetConfig.label}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell>
                    {valueVisions.length > 0 ? (
                      <ul className="space-y-1">
                        {valueVisions.map((vision) => (
                          <li key={vision.id} className="text-sm text-muted-foreground">
                            <span className="text-rose-400 mr-2">•</span>
                            <span className="line-clamp-1">{vision.content}</span>
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <span className="text-sm text-muted-foreground italic">No visions yet</span>
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
