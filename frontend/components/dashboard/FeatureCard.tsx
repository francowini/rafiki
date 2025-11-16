import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { LucideIcon, ArrowRight, Lock } from 'lucide-react';

interface FeatureCardProps {
  title: string;
  description: string;
  icon: LucideIcon;
  href: string;
  color: 'blue' | 'red' | 'green' | 'purple' | 'indigo' | 'teal';
  available: boolean;
}

const colorClasses = {
  blue: 'bg-blue-100 text-blue-600 hover:bg-blue-200',
  red: 'bg-red-100 text-red-600 hover:bg-red-200',
  green: 'bg-green-100 text-green-600 hover:bg-green-200',
  purple: 'bg-purple-100 text-purple-600 hover:bg-purple-200',
  indigo: 'bg-indigo-100 text-indigo-600 hover:bg-indigo-200',
  teal: 'bg-teal-100 text-teal-600 hover:bg-teal-200',
};

export function FeatureCard({ title, description, icon: Icon, href, color, available }: FeatureCardProps) {
  const content = (
    <Card className={`transition-all hover:shadow-lg ${!available ? 'opacity-60' : ''}`}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className={`p-3 rounded-lg ${colorClasses[color]}`}>
            <Icon className="h-6 w-6" />
          </div>
          {!available && (
            <Lock className="h-4 w-4 text-muted-foreground" />
          )}
        </div>
        <CardTitle className="mt-4">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        {available ? (
          <Button variant="ghost" className="w-full justify-between group">
            Explore
            <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
          </Button>
        ) : (
          <Button variant="ghost" className="w-full" disabled>
            Coming Soon
          </Button>
        )}
      </CardContent>
    </Card>
  );

  if (available) {
    return <Link href={href}>{content}</Link>;
  }

  return content;
}
