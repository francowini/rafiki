"use client";

import { useEffect, useState } from "react";
import { ThinkListResponse } from "@/lib/types";
import { api, APIError } from "@/lib/api";
import { ThinkCard } from "./ThinkCard";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface ThinkListProps {
  refresh?: number; // Increment to trigger refresh
}

export function ThinkList({ refresh = 0 }: ThinkListProps) {
  const [data, setData] = useState<ThinkListResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);

  const fetchThinks = async (pageNum: number) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.thinks.getAll({
        page: pageNum,
        rows: 10,
        orderBy: "date_created",
      });
      setData(response);
    } catch (err) {
      if (err instanceof APIError) {
        setError(`Failed to load thinks: ${err.message}`);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchThinks(page);
  }, [page, refresh]);

  if (isLoading && !data) {
    return (
      <div className="space-y-4">
        {[...Array(3)].map((_, i) => (
          <Skeleton key={i} className="h-32 w-full" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (!data || data.items.length === 0) {
    return (
      <Alert>
        <AlertDescription>
          No thinks yet. Create your first think above!
        </AlertDescription>
      </Alert>
    );
  }

  const totalPages = Math.ceil(data.total / data.rowsPerPage);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">
          All Thinks ({data.total})
        </h2>
        <span className="text-sm text-muted-foreground">
          Page {data.page} of {totalPages}
        </span>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data.items.map((think) => (
          <ThinkCard key={think.id} think={think} />
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1 || isLoading}
          >
            Previous
          </Button>
          <span className="text-sm">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages || isLoading}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}