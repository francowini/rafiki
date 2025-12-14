'use client';

import { Moment } from '@/lib/types';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';

interface MomentDetailProps {
  moment: Moment | null;
  open: boolean;
  onClose: () => void;
}

export function MomentDetail({ moment, open, onClose }: MomentDetailProps) {
  if (!moment) return null;

  const momentDate = new Date(moment.momentDate);
  const dateStr = momentDate.toLocaleDateString('es-MX', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  const getIntensityColor = (intensity: number) => {
    if (intensity >= 8) return 'bg-red-100 text-red-800 border-red-300';
    if (intensity >= 5) return 'bg-yellow-100 text-yellow-800 border-yellow-300';
    return 'bg-green-100 text-green-800 border-green-300';
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex items-start justify-between">
            <div>
              <DialogTitle className="text-xl">{dateStr}</DialogTitle>
              <DialogDescription className="text-sm mt-1">Registro del momento</DialogDescription>
            </div>
            <Badge variant="outline" className={getIntensityColor(moment.intensity)}>
              Intensidad: {moment.intensity}/10
            </Badge>
          </div>
        </DialogHeader>

        <div className="space-y-6 pt-4">
          {/* Situation */}
          <div className="space-y-2">
            <h4 className="font-medium text-sm text-muted-foreground">Situación</h4>
            <p className="text-base leading-relaxed">{moment.situation}</p>
          </div>

          <Separator />

          {/* Thoughts */}
          <div className="space-y-2">
            <h4 className="font-medium text-sm text-muted-foreground">Pensamientos</h4>
            <p className="text-base leading-relaxed">{moment.thoughts}</p>
          </div>

          <Separator />

          {/* Physical Symptoms */}
          <div className="space-y-2">
            <h4 className="font-medium text-sm text-muted-foreground">
              Síntomas físicos o emociones
            </h4>
            <p className="text-base leading-relaxed">{moment.physicalSymptoms}</p>
          </div>

          <Separator />

          {/* Behavior */}
          <div className="space-y-2">
            <h4 className="font-medium text-sm text-muted-foreground">Lo que hiciste</h4>
            <p className="text-base leading-relaxed">{moment.behavior}</p>
          </div>

          <Separator />

          {/* Consequences */}
          <div className="space-y-2">
            <h4 className="font-medium text-sm text-muted-foreground">Consecuencias</h4>
            <p className="text-base leading-relaxed">{moment.consequences}</p>
          </div>

          <Separator />

          {/* Values Reflection */}
          <div className="space-y-2">
            <h4 className="font-medium text-sm text-muted-foreground">Reflexión de valores</h4>
            <p className="text-base leading-relaxed">{moment.valuesReflection}</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
