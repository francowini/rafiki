'use client';

import { useCallback, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { quickTaskSchema, type QuickTaskFormData } from '@/lib/schemas/quick-task-schema';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Loader2 } from 'lucide-react';
import { useCreateTask } from '@/lib/hooks/use-tasks';
import { useObjectives } from '@/lib/hooks/use-objectives';
import { useToast } from '@/hooks/use-toast';

function getContributionLabel(value: number): string {
  if (value <= 3) return 'Pequeño paso';
  if (value <= 6) return 'Avance sólido';
  if (value <= 8) return 'Gran contribución';
  return 'Contribución máxima';
}

// Sentinel value for inbox (Radix Select doesn't handle empty strings well)
const INBOX_VALUE = '__INBOX__';

interface QuickTaskSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function QuickTaskSheet({ open, onOpenChange }: QuickTaskSheetProps) {
  const { toast } = useToast();
  const createMutation = useCreateTask();
  const { data: objectivesData } = useObjectives({ status: 'active' });
  const [contribution, setContribution] = useState(5);

  const form = useForm<QuickTaskFormData>({
    resolver: zodResolver(quickTaskSchema),
    defaultValues: {
      title: '',
      objectiveId: null,
    },
  });

  const selectedObjectiveId = form.watch('objectiveId');

  // Get selected objective to check tracking type
  const selectedObjective = objectivesData?.items.find((o) => o.id === selectedObjectiveId);
  const isResultObjective = selectedObjective?.trackingType === 'result';

  // Convert null to sentinel value for Select display
  const selectValue = selectedObjectiveId ?? INBOX_VALUE;

  // Reset form when dialog closes (F9: State Management)
  const handleOpenChange = useCallback(
    (newOpen: boolean) => {
      if (!newOpen) {
        form.reset();
        setContribution(5);
      }
      onOpenChange(newOpen);
    },
    [onOpenChange, form],
  );

  const handleSelectChange = (value: string) => {
    // Convert sentinel value back to null
    form.setValue('objectiveId', value === INBOX_VALUE ? null : value);
  };

  const onSubmit = async (data: QuickTaskFormData) => {
    try {
      // Determine contribution: required for result objectives, null otherwise
      const contributionValue = isResultObjective ? contribution : null;

      await createMutation.mutateAsync({
        title: data.title,
        objectiveId: data.objectiveId ?? null,
        contribution: contributionValue,
      });

      toast({
        title: 'Tarea creada',
        description: data.objectiveId
          ? 'La tarea ha sido anadida al objetivo'
          : 'La tarea ha sido anadida al inbox',
      });

      form.reset();
      onOpenChange(false);
    } catch {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: 'No se pudo crear la tarea',
      });
    }
  };

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetContent side="bottom" className="h-auto max-h-[85vh]">
        <SheetHeader>
          <SheetTitle>Nueva tarea rapida</SheetTitle>
          <SheetDescription>
            Crea una tarea en el inbox o asignala directamente a un objetivo
          </SheetDescription>
        </SheetHeader>

        <form onSubmit={form.handleSubmit(onSubmit)} className="mt-6 space-y-4">
          <div className="space-y-2">
            <Label htmlFor="quick-task-title">Titulo</Label>
            <Input
              id="quick-task-title"
              {...form.register('title')}
              placeholder="Ej: Llamar al dentista"
              autoFocus
            />
            {form.formState.errors.title && (
              <p className="text-sm text-destructive">{form.formState.errors.title.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="quick-task-objective">Objetivo (opcional)</Label>
            <Select value={selectValue} onValueChange={handleSelectChange}>
              <SelectTrigger id="quick-task-objective">
                <SelectValue placeholder="Sin objetivo (Inbox)" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={INBOX_VALUE}>Sin objetivo (Inbox)</SelectItem>
                {objectivesData?.items.map((obj) => (
                  <SelectItem key={obj.id} value={obj.id}>
                    <div className="flex flex-col">
                      <span>{obj.title}</span>
                      <span className="text-xs text-muted-foreground">
                        {obj.trackingType === 'result'
                          ? `${obj.currentMetric ?? 0}/${obj.targetMetric}`
                          : 'Frecuencia'}
                      </span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Si no seleccionas un objetivo, la tarea ira a tu Inbox.
            </p>
          </div>

          {/* Contribution Slider - Only for result objectives */}
          {isResultObjective && (
            <div className="space-y-4 bg-blue-50 rounded-lg p-4">
              <Label className="text-base font-medium flex items-center gap-2">
                Contribución al objetivo
                <span className="text-destructive">*</span>
              </Label>

              <div className="text-center py-2">
                <span className="text-4xl font-bold text-blue-600">+{contribution}</span>
                <p className="text-sm text-muted-foreground mt-1">
                  {getContributionLabel(contribution)}
                </p>
              </div>

              <Slider
                min={1}
                max={10}
                step={1}
                value={[contribution]}
                onValueChange={(value) => setContribution(value[0])}
                className="py-4"
              />

              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Pequeño</span>
                <span>Moderado</span>
                <span>Grande</span>
              </div>
            </div>
          )}

          <div className="flex gap-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              className="flex-1"
            >
              Cancelar
            </Button>
            <Button type="submit" disabled={createMutation.isPending} className="flex-1">
              {createMutation.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Creando...
                </>
              ) : (
                'Crear tarea'
              )}
            </Button>
          </div>
        </form>
      </SheetContent>
    </Sheet>
  );
}
