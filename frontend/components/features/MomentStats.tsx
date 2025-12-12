'use client';

import { Clock } from 'lucide-react';

interface MomentStatsProps {
  thisWeek: number;
  thisMonth: number;
  last30Days: number;
  isLoading?: boolean;
}

function StatCard({ label, value, isLoading }: { label: string; value: number; isLoading?: boolean }) {
  return (
    <div className="bg-white border rounded-lg p-4">
      <p className="text-sm font-medium text-gray-600 mb-1">{label}</p>
      {isLoading ? (
        <div className="h-8 w-16 bg-gray-200 animate-pulse rounded" />
      ) : (
        <p className="text-3xl font-bold text-gray-900">{value}</p>
      )}
    </div>
  );
}

export function MomentStats({ thisWeek, thisMonth, last30Days, isLoading }: MomentStatsProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <Clock className="h-5 w-5 text-gray-600" />
        <h2 className="text-lg font-semibold text-gray-900">Your Reflections</h2>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <StatCard label="This Week" value={thisWeek} isLoading={isLoading} />
        <StatCard label="This Month" value={thisMonth} isLoading={isLoading} />
        <StatCard label="Last 30 Days" value={last30Days} isLoading={isLoading} />
      </div>
    </div>
  );
}
