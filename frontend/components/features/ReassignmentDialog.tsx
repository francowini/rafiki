'use client';

import { useState } from 'react';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, Info, AlertTriangle } from 'lucide-react';
import { Value, LifeVision } from '@/lib/types';
import { api } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';
import { getFacetConfig } from '@/lib/value-utils';

interface ReassignmentDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  valueToArchive: Value;
  activeLifeVisions: LifeVision[];
  availableValues: Value[];
  onComplete: () => void;
}

export function ReassignmentDialog({
  open,
  onOpenChange,
  valueToArchive,
  activeLifeVisions,
  availableValues,
  onComplete,
}: ReassignmentDialogProps) {
  const [selectedValueId, setSelectedValueId] = useState<string>('');
  const [archiveVisions, setArchiveVisions] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const { toast } = useToast();

  const facetConfig = getFacetConfig(valueToArchive.facet);
  const canProceed = archiveVisions || selectedValueId;

  const handleProceed = async () => {
    setIsProcessing(true);

    try {
      if (archiveVisions) {
        // Archive all life visions
        await Promise.all(activeLifeVisions.map((vision) => api.lifeVisions.archive(vision.id)));
        toast({
          title: 'Visiones retiradas',
          description: `${activeLifeVisions.length} visiones retiradas exitosamente.`,
        });
      } else {
        // Reassign all life visions to selected value
        await Promise.all(
          activeLifeVisions.map((vision) => api.lifeVisions.reassign(vision.id, selectedValueId)),
        );
        toast({
          title: 'Visiones reasignadas',
          description: `${activeLifeVisions.length} visiones reasignadas exitosamente.`,
        });
      }

      // Now archive the value
      await api.values.archive(valueToArchive.id);
      toast({
        title: 'Valor retirado',
        description: 'Tu valor ha sido retirado exitosamente.',
      });

      onComplete();
      onOpenChange(false);
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : 'No se pudo completar la operacion. Por favor intenta de nuevo.';
      toast({
        variant: 'destructive',
        title: 'Error',
        description: message,
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const handleCancel = () => {
    setSelectedValueId('');
    setArchiveVisions(false);
    onOpenChange(false);
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="max-w-lg">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-amber-500" />
            Reasignar visiones de vida
          </AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <span className="block">
                Este valor tiene{' '}
                <strong className="text-amber-600">
                  {activeLifeVisions.length} vision(es) de vida
                </strong>{' '}
                activa(s).
              </span>
              <span className="block text-muted-foreground">
                Elige un valor activo para reasignarlas, o retiralas junto con el valor.
              </span>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>

        {/* Value being archived */}
        <div className="p-3 bg-gray-50 rounded-md border">
          <div className="flex items-center gap-2 mb-1">
            <Badge
              variant="outline"
              className={`${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor}`}
            >
              <span className="mr-1">{facetConfig.icon}</span>
              {facetConfig.label}
            </Badge>
          </div>
          <p className="text-sm font-medium line-clamp-2">{valueToArchive.content}</p>
        </div>

        <Alert className="bg-blue-50 border-blue-200">
          <Info className="h-4 w-4 text-blue-600" />
          <AlertDescription className="text-sm text-blue-700">
            Las visiones reasignadas permaneceran activas en el nuevo valor.
          </AlertDescription>
        </Alert>

        {/* Reassignment Options */}
        <div className="space-y-4">
          {/* Option 1: Reassign to another value */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Reasignar a:</label>
            <Select
              value={selectedValueId}
              onValueChange={(value) => {
                setSelectedValueId(value);
                setArchiveVisions(false);
              }}
              disabled={archiveVisions}
            >
              <SelectTrigger>
                <SelectValue placeholder="Selecciona un valor..." />
              </SelectTrigger>
              <SelectContent>
                {availableValues.map((value) => {
                  const config = getFacetConfig(value.facet);
                  return (
                    <SelectItem key={value.id} value={value.id}>
                      <div className="flex items-center gap-2">
                        <span>{config.icon}</span>
                        <span className="line-clamp-1">{value.content}</span>
                      </div>
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          </div>

          {/* Option 2: Archive all visions */}
          <div className="flex items-center gap-2 p-3 bg-gray-50 rounded-md border">
            <input
              type="checkbox"
              id="archive-visions"
              checked={archiveVisions}
              onChange={(e) => {
                setArchiveVisions(e.target.checked);
                if (e.target.checked) setSelectedValueId('');
              }}
              className="h-4 w-4 rounded border-gray-300 text-rose-600 focus:ring-rose-500"
            />
            <label htmlFor="archive-visions" className="text-sm cursor-pointer">
              Retirar todas las visiones (no reasignar)
            </label>
          </div>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel onClick={handleCancel} disabled={isProcessing}>
            Cancelar
          </AlertDialogCancel>
          <Button
            onClick={handleProceed}
            disabled={!canProceed || isProcessing}
            className="bg-amber-600 hover:bg-amber-700"
          >
            {isProcessing ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Procesando...
              </>
            ) : archiveVisions ? (
              'Retirar todo'
            ) : (
              'Reasignar y retirar'
            )}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
