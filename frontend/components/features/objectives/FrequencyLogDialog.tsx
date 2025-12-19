'use client';

import { useState, useMemo } from 'react';
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
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { HelpTooltip } from '@/components/ui/HelpTooltip';
import { CheckCircle2, MinusCircle, XCircle, Loader2 } from 'lucide-react';
import { useLogRecord } from '@/lib/hooks/use-objectives';
import { cn } from '@/lib/utils';
import type { RecordStatus } from '@/lib/types';

interface FrequencyLogDialogProps {
  objetivoId: string;
  objetivoName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialStatus?: RecordStatus;
}

function getDateStrings() {
  const now = new Date();
  const today = now.toISOString().split('T')[0];
  const yesterdayDate = new Date(now);
  yesterdayDate.setDate(yesterdayDate.getDate() - 1);
  const yesterday = yesterdayDate.toISOString().split('T')[0];
  return { today, yesterday };
}

export function FrequencyLogDialog({
  objetivoId,
  objetivoName,
  open,
  onOpenChange,
  initialStatus,
}: FrequencyLogDialogProps) {
  const { today, yesterday } = useMemo(() => getDateStrings(), []);

  const [selectedDate, setSelectedDate] = useState(today);
  const [status, setStatus] = useState<RecordStatus | null>(initialStatus || null);
  const [notes, setNotes] = useState('');

  const { mutate: logRecord, isPending } = useLogRecord(objetivoId);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!status) return;

    logRecord(
      { fechaRegistro: selectedDate, status, notes: notes || undefined },
      {
        onSuccess: () => {
          // Reset all dialog state before closing
          setStatus(null);
          setNotes('');
          setSelectedDate(today);
          onOpenChange(false);
        },
      },
    );
  };

  // Reset state when dialog closes (user cancels)
  const handleCancel = () => {
    setStatus(null);
    setNotes('');
    setSelectedDate(today);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>Registrar progreso</DialogTitle>
          <DialogDescription>{objetivoName}</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Date selector with grace period */}
          <div className="space-y-2">
            <Label htmlFor="date-selector" className="flex items-center gap-2">
              Fecha
              <HelpTooltip content="Puedes registrar para hoy o ayer si olvidaste hacerlo." />
            </Label>
            <Select value={selectedDate} onValueChange={setSelectedDate}>
              <SelectTrigger id="date-selector">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={today}>Hoy ({today})</SelectItem>
                <SelectItem value={yesterday}>Ayer ({yesterday})</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Status selector */}
          <div className="space-y-2">
            <Label htmlFor="status-selector">Estado</Label>
            <div
              id="status-selector"
              className="grid grid-cols-3 gap-3"
              role="group"
              aria-label="Seleccionar estado del registro"
            >
              <button
                type="button"
                onClick={() => setStatus('completado')}
                className={cn(
                  'border-2 rounded-lg p-4 text-center transition-all',
                  status === 'completado'
                    ? 'border-green-500 bg-green-50'
                    : 'border-gray-200 hover:border-green-300',
                )}
              >
                <CheckCircle2
                  className={cn(
                    'h-8 w-8 mx-auto mb-2',
                    status === 'completado' ? 'text-green-600' : 'text-gray-400',
                  )}
                />
                <span className="text-sm font-medium">Completado</span>
              </button>

              <button
                type="button"
                onClick={() => setStatus('omitido_intencional')}
                className={cn(
                  'border-2 rounded-lg p-4 text-center transition-all',
                  status === 'omitido_intencional'
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-blue-300',
                )}
              >
                <MinusCircle
                  className={cn(
                    'h-8 w-8 mx-auto mb-2',
                    status === 'omitido_intencional' ? 'text-blue-600' : 'text-gray-400',
                  )}
                />
                <span className="text-sm font-medium">Descanso</span>
              </button>

              <button
                type="button"
                onClick={() => setStatus('omitido')}
                className={cn(
                  'border-2 rounded-lg p-4 text-center transition-all',
                  status === 'omitido'
                    ? 'border-gray-500 bg-gray-50'
                    : 'border-gray-200 hover:border-gray-300',
                )}
              >
                <XCircle
                  className={cn(
                    'h-8 w-8 mx-auto mb-2',
                    status === 'omitido' ? 'text-gray-600' : 'text-gray-400',
                  )}
                />
                <span className="text-sm font-medium">Omitido</span>
              </button>
            </div>
          </div>

          {/* Optional notes */}
          <div className="space-y-2">
            <Label htmlFor="notes-textarea">Notas (opcional)</Label>
            <Textarea
              id="notes-textarea"
              placeholder="Reflexiones, obstáculos, aprendizajes..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleCancel}>
              Cancelar
            </Button>
            <Button type="submit" disabled={!status || isPending}>
              {isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Guardando...
                </>
              ) : (
                'Guardar registro'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
