'use client';

import { useAuth } from '@/lib/auth-context';
import { WelcomeCard } from '@/components/dashboard/WelcomeCard';
import { FeatureCard } from '@/components/dashboard/FeatureCard';
import { ValuesTableWithVisions } from '@/components/dashboard/ValuesTableWithVisions';
import { Brain, Target, Heart, Compass, Activity, Clock, Sparkles } from 'lucide-react';

export default function DashboardPage() {
  const { user } = useAuth();

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <WelcomeCard userName={user?.name || 'there'} />

      {/* Values & Life Visions Section */}
      <ValuesTableWithVisions />

      {/* Feature Navigation Section */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Your Journey</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <FeatureCard
            title="Values"
            description="Define and track your core values and principles"
            icon={Heart}
            href="/values"
            color="red"
            available
          />
          <FeatureCard
            title="Life Visions"
            description="Define how you want to live each of your values"
            icon={Sparkles}
            href="/life-visions"
            color="red"
            available
          />
          <FeatureCard
            title="Thinks"
            description="Capture and organize your thoughts, ideas, and insights"
            icon={Brain}
            href="/thinks"
            color="blue"
            available
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
            title="Moments"
            description="Record and reflect on your difficult moments"
            icon={Clock}
            href="/momentos"
            color="teal"
            available
          />
        </div>
      </div>
    </div>
  );
}
