'use client';

import { useState, useCallback } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Loader2 } from 'lucide-react';
import { useUpdateProgress } from '@/lib/hooks/use-objectives';
import type { Objective } from '@/lib/types';

interface ProgressUpdateDialogProps {
  objective: Objective | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ProgressUpdateDialog({
  objective,
  open,
  onOpenChange,
}: ProgressUpdateDialogProps) {
  const [value, setValue] = useState('');
  const [error, setError] = useState('');
  const [initialized, setInitialized] = useState(false);

  const { mutate: updateProgress, isPending } = useUpdateProgress();

  // Initialize value when dialog opens - using onOpenChange callback pattern
  const handleOpenChange = useCallback(
    (newOpen: boolean) => {
      if (newOpen && objective && !initialized) {
        setValue(objective.metricaActual?.toString() || '0');
        setError('');
        setInitialized(true);
      }
      if (!newOpen) {
        setInitialized(false);
      }
      onOpenChange(newOpen);
    },
    [objective, onOpenChange, initialized],
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!objective) return;

    // Validate input
    const numValue = parseFloat(value);
    if (isNaN(numValue)) {
      setError('Por favor ingresa un número válido');
      return;
    }

    if (numValue < 0) {
      setError('El valor no puede ser negativo');
      return;
    }

    if (objective.metricaObjetivo && numValue > objective.metricaObjetivo) {
      setError(`El valor no puede ser mayor que la meta (${objective.metricaObjetivo})`);
      return;
    }

    // Clear error and submit
    setError('');
    updateProgress(
      { id: objective.id, value: numValue },
      {
        onSuccess: () => {
          setValue('');
          onOpenChange(false);
        },
      },
    );
  };

  const handleCancel = () => {
    setValue('');
    setError('');
    onOpenChange(false);
  };

  if (!objective) return null;

  const currentProgress = objective.metricaActual || 0;
  const target = objective.metricaObjetivo || 1;
  const percentage = Math.round((currentProgress / target) * 100);
  const parsedValue = value ? parseFloat(value) : NaN;
  const isValidNumber = !isNaN(parsedValue);
  const newPercentage = isValidNumber ? Math.round((parsedValue / target) * 100) : percentage;

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>Actualizar progreso</DialogTitle>
          <DialogDescription className="line-clamp-2">{objective.titulo}</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Current progress display */}
          <div className="bg-blue-50 rounded-lg p-4 space-y-3">
            <div className="flex justify-between text-sm">
              <span className="text-gray-600">Progreso actual:</span>
              <span className="font-semibold text-gray-900">
                {currentProgress} / {target}
              </span>
            </div>
            <Progress value={percentage} className="h-2" />
            <div className="text-center">
              <span className="text-2xl font-bold text-blue-700">{percentage}%</span>
              <span className="text-sm text-gray-500 ml-2">completado</span>
            </div>
          </div>

          {/* Value input */}
          <div className="space-y-2">
            <Label htmlFor="progress-value">Nuevo valor</Label>
            <Input
              id="progress-value"
              type="number"
              step="any"
              min="0"
              max={objective.metricaObjetivo}
              value={value}
              onChange={(e) => {
                setValue(e.target.value);
                setError('');
              }}
              placeholder={`0 - ${objective.metricaObjetivo}`}
              autoFocus
            />
            {error && <p className="text-sm text-red-600">{error}</p>}
            <p className="text-xs text-muted-foreground">
              Ingresa el valor actual de tu progreso (no el incremento).
            </p>
          </div>

          {/* Preview of new progress */}
          {value && !error && isValidNumber && parsedValue !== currentProgress && (
            <div className="bg-green-50 rounded-lg p-3 text-sm">
              <span className="text-green-700">
                Nuevo progreso: <strong>{newPercentage}%</strong>
                {newPercentage >= 100 && ' - ¡Meta alcanzada!'}
                {newPercentage >= 50 && newPercentage < 100 && ' - ¡Más de la mitad!'}
              </span>
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleCancel}>
              Cancelar
            </Button>
            <Button type="submit" disabled={!value || isPending}>
              {isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Guardando...
                </>
              ) : (
                'Actualizar'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
