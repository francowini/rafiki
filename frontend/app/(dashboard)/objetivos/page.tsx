'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetTrigger,
} from '@/components/ui/sheet';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { ObjectiveCard } from '@/components/features/objectives/ObjectiveCard';
import { ObjectiveFrequencyCard } from '@/components/features/objectives/ObjectiveFrequencyCard';
import { ObjectiveForm } from '@/components/features/objectives/ObjectiveForm';
import { ObjectiveDetail } from '@/components/features/objectives/ObjectiveDetail';
import { FrequencyLogDialog } from '@/components/features/objectives/FrequencyLogDialog';
import { ProgressUpdateDialog } from '@/components/features/objectives/ProgressUpdateDialog';
import { useObjectives, useArchiveObjective } from '@/lib/hooks/use-objectives';
import { Target, Plus, Info, Loader2 } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import type { Objective, RecordStatus } from '@/lib/types';

export default function ObjetivosPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [selectedObjective, setSelectedObjective] = useState<Objective | null>(null);
  const [objectiveToEdit, setObjectiveToEdit] = useState<Objective | null>(null);
  const [logDialogObjective, setLogDialogObjective] = useState<{
    objective: Objective;
    initialStatus?: RecordStatus;
  } | null>(null);
  const [progressDialogObjective, setProgressDialogObjective] = useState<Objective | null>(null);

  const { toast } = useToast();
  // Show all non-archived objectives (activo + pausado)
  const { data, isLoading } = useObjectives({});
  const archiveMutation = useArchiveObjective();

  const handleQuickLog = (objective: Objective, status: RecordStatus) => {
    setLogDialogObjective({ objective, initialStatus: status });
  };

  const handleArchive = async (id: string) => {
    try {
      await archiveMutation.mutateAsync(id);
      toast({
        title: 'Objetivo archivado',
        description: 'El objetivo ha sido archivado correctamente.',
      });
    } catch {
      toast({
        title: 'Error',
        description: 'No se pudo archivar el objetivo.',
        variant: 'destructive',
      });
    }
  };

  const handleEdit = (objective: Objective) => {
    setObjectiveToEdit(objective);
    setIsFormOpen(true);
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Target className="h-8 w-8 text-gray-600" />
            <h1 className="text-3xl font-bold tracking-tight">Objetivos</h1>
          </div>
          <p className="text-muted-foreground">
            Metas específicas alineadas con tus visiones de vida
          </p>
        </div>

        <Sheet open={isFormOpen} onOpenChange={setIsFormOpen}>
          <SheetTrigger asChild>
            <Button size="lg">
              <Plus className="h-5 w-5 mr-2" />
              Nuevo objetivo
            </Button>
          </SheetTrigger>
          <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
            <SheetHeader>
              <SheetTitle>{objectiveToEdit ? 'Editar objetivo' : 'Crear objetivo'}</SheetTitle>
              <SheetDescription>
                Define un objetivo alineado con tus visiones de vida
              </SheetDescription>
            </SheetHeader>
            <div className="mt-6">
              <ObjectiveForm
                objective={objectiveToEdit}
                onSuccess={() => {
                  setIsFormOpen(false);
                  setObjectiveToEdit(null);
                }}
                onCancel={() => {
                  setIsFormOpen(false);
                  setObjectiveToEdit(null);
                }}
              />
            </div>
          </SheetContent>
        </Sheet>
      </div>

      {/* Info Alert */}
      <Alert className="mb-6 bg-gray-50 border-gray-200">
        <Info className="h-4 w-4 text-gray-600" />
        <AlertDescription className="text-sm text-gray-700">
          Los objetivos pueden ser de <strong>Resultado</strong> (progreso medible) o{' '}
          <strong>Frecuencia</strong> (hábitos diarios/semanales).
        </AlertDescription>
      </Alert>

      {/* Objectives Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
          <span className="ml-2 text-muted-foreground">Cargando objetivos...</span>
        </div>
      ) : data?.items.length === 0 ? (
        <div className="border border-dashed rounded-lg p-12 text-center">
          <Target className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium mb-2">Aún no hay objetivos</h3>
          <p className="text-muted-foreground mb-6">
            Empieza creando tu primer objetivo alineado con tus visiones de vida.
          </p>
          <Button onClick={() => setIsFormOpen(true)}>
            <Plus className="h-5 w-5 mr-2" />
            Crear primer objetivo
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {data?.items.map((objective) =>
            objective.tipoTracking === 'resultado' ? (
              <ObjectiveCard
                key={objective.id}
                objective={objective}
                onEdit={handleEdit}
                onArchive={handleArchive}
                onClick={setSelectedObjective}
                onUpdateProgress={setProgressDialogObjective}
              />
            ) : (
              <ObjectiveFrequencyCard
                key={objective.id}
                objective={objective}
                onEdit={handleEdit}
                onArchive={handleArchive}
                onLogRecord={(status) => handleQuickLog(objective, status)}
                onClick={setSelectedObjective}
              />
            ),
          )}
        </div>
      )}

      {/* Detail Sheet */}
      <Sheet open={!!selectedObjective} onOpenChange={() => setSelectedObjective(null)}>
        <SheetContent side="right" className="w-full sm:max-w-4xl overflow-y-auto">
          <SheetHeader className="sr-only">
            <SheetTitle>Detalle del objetivo</SheetTitle>
            <SheetDescription>Detalles del objetivo seleccionado</SheetDescription>
          </SheetHeader>
          {selectedObjective && <ObjectiveDetail objectiveId={selectedObjective.id} />}
        </SheetContent>
      </Sheet>

      {/* Log Dialog */}
      {logDialogObjective && (
        <FrequencyLogDialog
          objetivoId={logDialogObjective.objective.id}
          objetivoName={logDialogObjective.objective.titulo}
          open={!!logDialogObjective}
          onOpenChange={() => setLogDialogObjective(null)}
          initialStatus={logDialogObjective.initialStatus}
        />
      )}

      {/* Progress Update Dialog */}
      <ProgressUpdateDialog
        objective={progressDialogObjective}
        open={!!progressDialogObjective}
        onOpenChange={(open) => !open && setProgressDialogObjective(null)}
      />
    </div>
  );
}
