import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface WelcomeCardProps {
  userName: string;
}

export function WelcomeCard({ userName }: WelcomeCardProps) {
  const currentHour = new Date().getHours();
  let greeting = 'Good evening';

  if (currentHour < 12) {
    greeting = 'Good morning';
  } else if (currentHour < 18) {
    greeting = 'Good afternoon';
  }

  return (
    <Card className="bg-gradient-to-r from-blue-500 to-indigo-600 text-white border-0">
      <CardHeader>
        <CardTitle className="text-2xl">
          {greeting}, {userName}!
        </CardTitle>
        <CardDescription className="text-blue-100">
          Welcome back to your personal development journey
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-blue-50">
          Track your thoughts, define your values, and achieve your goals all in one place.
        </p>
      </CardContent>
    </Card>
  );
}
