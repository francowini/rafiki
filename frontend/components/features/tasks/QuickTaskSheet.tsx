'use client';

import { useCallback } from 'react';
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

  const form = useForm<QuickTaskFormData>({
    resolver: zodResolver(quickTaskSchema),
    defaultValues: {
      title: '',
      objectiveId: null,
    },
  });

  const selectedObjectiveId = form.watch('objectiveId');

  // Convert null to sentinel value for Select display
  const selectValue = selectedObjectiveId ?? INBOX_VALUE;

  // Reset form when dialog closes (F9: State Management)
  const handleOpenChange = useCallback(
    (newOpen: boolean) => {
      if (!newOpen) {
        form.reset();
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
      await createMutation.mutateAsync({
        title: data.title,
        objectiveId: data.objectiveId ?? null,
        contribution: null, // Inbox tasks don't have contribution
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
                    {obj.title}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Si no seleccionas un objetivo, la tarea ira a tu Inbox.
            </p>
          </div>

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
