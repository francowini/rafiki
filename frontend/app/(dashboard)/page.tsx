'use client';

import { useAuth } from '@/lib/auth-context';
import { WelcomeCard } from '@/components/dashboard/WelcomeCard';
import { QuickStatsCard } from '@/components/dashboard/QuickStatsCard';
import { FeatureCard } from '@/components/dashboard/FeatureCard';
import { Brain, Target, Heart, Compass, Activity, Clock } from 'lucide-react';

export default function DashboardPage() {
  const { user } = useAuth();

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <WelcomeCard userName={user?.name || 'there'} />

      {/* Quick Stats Section */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <QuickStatsCard
          title="Thinks"
          value={0}
          description="Ideas captured"
          trend="+2 this week"
          icon={Brain}
          color="blue"
        />
        <QuickStatsCard
          title="Values"
          value={0}
          description="Core values"
          icon={Heart}
          color="red"
        />
        <QuickStatsCard
          title="Goals"
          value={0}
          description="Active goals"
          icon={Target}
          color="green"
        />
        <QuickStatsCard
          title="Habits"
          value={0}
          description="Daily habits"
          trend="5 day streak"
          icon={Activity}
          color="purple"
        />
      </div>

      {/* Feature Navigation Section */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Your Journey</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <FeatureCard
            title="Thinks"
            description="Capture and organize your thoughts, ideas, and insights"
            icon={Brain}
            href="/thinks"
            color="blue"
            available={true}
          />
          <FeatureCard
            title="Values"
            description="Define and track your core values and principles"
            icon={Heart}
            href="/values"
            color="red"
            available={false}
          />
          <FeatureCard
            title="Goals"
            description="Set meaningful goals and track your progress"
            icon={Target}
            href="/goals"
            color="green"
            available={false}
          />
          <FeatureCard
            title="Purpose"
            description="Discover and clarify your life's purpose"
            icon={Compass}
            href="/purpose"
            color="indigo"
            available={false}
          />
          <FeatureCard
            title="Habits"
            description="Build positive habits and break negative ones"
            icon={Activity}
            href="/habits"
            color="purple"
            available={false}
          />
          <FeatureCard
            title="Momentos"
            description="Registra y reflexiona sobre tus momentos difíciles"
            icon={Clock}
            href="/momentos"
            color="purple"
            available={true}
          />
        </div>
      </div>
    </div>
  );
}
