import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface WelcomeCardProps {
  userName: string;
}

export function WelcomeCard({ userName }: WelcomeCardProps) {
  const currentHour = new Date().getHours();
  let greeting = 'Buenas noches';

  if (currentHour < 12) {
    greeting = 'Buenos días';
  } else if (currentHour < 18) {
    greeting = 'Buenas tardes';
  }

  return (
    <Card className="bg-gradient-to-r from-blue-500 to-indigo-600 text-white border-0">
      <CardHeader>
        <CardTitle className="text-2xl">
          {greeting}, {userName}!
        </CardTitle>
        <CardDescription className="text-blue-100">
          Bienvenido de vuelta a tu viaje de desarrollo personal
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-blue-50">
          Registra tus pensamientos, define tus valores y alcanza tus metas en un solo lugar.
        </p>
      </CardContent>
    </Card>
  );
}
