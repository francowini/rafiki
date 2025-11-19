import { Think } from '@/lib/types';
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface ThinkCardProps {
  think: Think;
}

const categoryColors: Record<Think['category'], string> = {
  personal: 'bg-blue-500',
  work: 'bg-green-500',
  ideas: 'bg-purple-500',
  learning: 'bg-yellow-500',
  reflection: 'bg-pink-500',
};

export function ThinkCard({ think }: ThinkCardProps) {
  const formattedDate = new Date(think.dateCreated).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <Badge className={categoryColors[think.category]}>{think.category}</Badge>
          <span className="text-xs text-muted-foreground">{formattedDate}</span>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm whitespace-pre-wrap">{think.content}</p>
      </CardContent>
      <CardFooter className="pt-2">
        <span className="text-xs text-muted-foreground">ID: {think.id.slice(0, 8)}</span>
      </CardFooter>
    </Card>
  );
}
