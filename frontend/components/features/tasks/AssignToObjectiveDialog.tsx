'use client';

import { useState, useCallback } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Slider } from '@/components/ui/slider';
import { Loader2 } from 'lucide-react';
import { useMoveTask } from '@/lib/hooks/use-tasks';
import { useObjectives } from '@/lib/hooks/use-objectives';
import { useToast } from '@/hooks/use-toast';
import type { Task } from '@/lib/types';

interface AssignToObjectiveDialogProps {
  task: Task | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function getContributionLabel(value: number): string {
  if (value <= 3) return 'Pequeno paso';
  if (value <= 6) return 'Avance solido';
  if (value <= 8) return 'Gran contribucion';
  return 'Contribucion maxima';
}

export function AssignToObjectiveDialog({ task, open, onOpenChange }: AssignToObjectiveDialogProps) {
  const [objectiveId, setObjectiveId] = useState('');
  const [contribution, setContribution] = useState(5);
  const { toast } = useToast();
  const moveMutation = useMoveTask();
  const { data: objectivesData } = useObjectives({ status: 'active' });

  // Get selected objective to check tracking type
  const selectedObjective = objectivesData?.items.find((o) => o.id === objectiveId);
  const isFrequencyObjective = selectedObjective?.trackingType === 'frequency';

  // F9: Reset dialog state on close
  const handleOpenChange = useCallback(
    (newOpen: boolean) => {
      if (!newOpen) {
        setObjectiveId('');
        setContribution(5);
      }
      onOpenChange(newOpen);
    },
    [onOpenChange],
  );

  const handleSubmit = async () => {
    if (!task || !objectiveId) return;

    try {
      await moveMutation.mutateAsync({
        taskId: task.id,
        data: {
          objectiveId,
          // Frequency objectives don't have contribution, result objectives require it
          contribution: isFrequencyObjective ? null : contribution,
        },
      });

      toast({
        title: 'Tarea asignada',
        description: 'La tarea ha sido movida al objetivo',
      });

      handleOpenChange(false);
    } catch {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: 'No se pudo asignar la tarea',
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Asignar a objetivo</DialogTitle>
          <DialogDescription>
            Selecciona un objetivo y define cuanto contribuye esta tarea
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Task Preview */}
          {task && (
            <div className="p-3 bg-gray-50 rounded-md">
              <p className="text-sm font-medium text-gray-900">{task.title}</p>
            </div>
          )}

          {/* Objective Selector - F10: Accessibility */}
          <div className="space-y-2">
            <Label htmlFor="assign-objective-select">Objetivo</Label>
            <Select value={objectiveId} onValueChange={setObjectiveId}>
              <SelectTrigger id="assign-objective-select">
                <SelectValue placeholder="Selecciona un objetivo" />
              </SelectTrigger>
              <SelectContent>
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
          </div>

          {/* Contribution Slider - Only for result objectives */}
          {objectiveId && !isFrequencyObjective && (
            <div className="space-y-4 bg-blue-50 rounded-lg p-4">
              <Label className="text-base font-medium flex items-center gap-2">
                Contribucion al objetivo
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
                <span>Pequeno</span>
                <span>Moderado</span>
                <span>Grande</span>
              </div>
            </div>
          )}

          {/* Frequency objective notice */}
          {objectiveId && isFrequencyObjective && (
            <div className="p-3 bg-purple-50 rounded-lg border border-purple-200">
              <p className="text-sm text-purple-700">
                Este objetivo es de frecuencia. La tarea se asignara sin contribucion numerica.
              </p>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex gap-3">
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
            className="flex-1"
          >
            Cancelar
          </Button>
          <Button onClick={handleSubmit} disabled={!objectiveId || moveMutation.isPending} className="flex-1">
            {moveMutation.isPending ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Asignando...
              </>
            ) : (
              'Asignar tarea'
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
