'use client';

import { useEffect, useState, useRef } from 'react';
import { useAuth } from '@/lib/auth-context';
import { api } from '@/lib/api';
import type { MomentStats as MomentStatsType } from '@/lib/types';
import { WelcomeCard } from '@/components/dashboard/WelcomeCard';
import { ValuesTableWithVisions } from '@/components/dashboard/ValuesTableWithVisions';
import { MomentStats } from '@/components/features/MomentStats';

export default function DashboardPage() {
  const { user } = useAuth();
  const [stats, setStats] = useState<MomentStatsType | null>(null);
  const [isLoadingStats, setIsLoadingStats] = useState(true);
  const requestIdRef = useRef(0);

  useEffect(() => {
    const currentRequestId = ++requestIdRef.current;

    const fetchStats = async () => {
      try {
        const data = await api.moments.getStats();
        if (currentRequestId === requestIdRef.current) {
          setStats(data);
        }
      } catch (err) {
        if (currentRequestId === requestIdRef.current) {
          console.error('Error loading stats:', err);
        }
      } finally {
        if (currentRequestId === requestIdRef.current) {
          setIsLoadingStats(false);
        }
      }
    };

    fetchStats();
  }, []);

  return (
    <div className="space-y-8">
      <WelcomeCard userName={user?.name || 'there'} />

      <MomentStats
        thisWeek={stats?.thisWeek ?? 0}
        thisMonth={stats?.thisMonth ?? 0}
        last30Days={stats?.last30Days ?? 0}
        isLoading={isLoadingStats}
      />

      <ValuesTableWithVisions />
    </div>
  );
}
