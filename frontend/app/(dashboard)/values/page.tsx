'use client';

import { useState } from 'react';
import { ValueForm } from '@/components/features/ValueForm';
import { ValueList } from '@/components/features/ValueList';
import { ReassignmentDialog } from '@/components/features/ReassignmentDialog';
import { Value, LifeVision } from '@/lib/types';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Plus, Info, Archive, RotateCcw } from 'lucide-react';
import { api, APIError } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';

export default function ValuesPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [valueToEdit, setValueToEdit] = useState<Value | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [valuesCount, setValuesCount] = useState(0);
  const { toast } = useToast();

  // Archive/Restore state
  const [showArchived, setShowArchived] = useState(false);
  const [valueToArchive, setValueToArchive] = useState<Value | null>(null);
  const [valueToRestore, setValueToRestore] = useState<Value | null>(null);
  const [showReassignmentDialog, setShowReassignmentDialog] = useState(false);
  const [activeLifeVisions, setActiveLifeVisions] = useState<LifeVision[]>([]);
  const [availableValues, setAvailableValues] = useState<Value[]>([]);
  const [_isArchiving, setIsArchiving] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);

  const handleCreateSuccess = () => {
    setIsFormOpen(false);
    setRefresh((prev) => prev + 1);
  };

  const handleEditSuccess = () => {
    setIsEditFormOpen(false);
    setValueToEdit(null);
    setRefresh((prev) => prev + 1);
  };

  const handleEdit = (value: Value) => {
    setValueToEdit(value);
    setIsEditFormOpen(true);
  };

  // Archive handler - handles 409 (has active life visions)
  const handleArchive = async (value: Value) => {
    setValueToArchive(value);
    setIsArchiving(true);

    try {
      // Try to archive directly
      await api.values.archive(value.id);
      toast({
        title: 'Valor retirado',
        description: 'Tu valor ha sido retirado exitosamente.',
      });
      setValueToArchive(null);
      setRefresh((prev) => prev + 1);
    } catch (err: unknown) {
      // Check if has active life visions (409 Conflict)
      if (err instanceof APIError && err.status === 409) {
        // Fetch life visions for this value
        const visionsRes = await api.lifeVisions.getAll({ valueId: value.id });
        const activeVisions = visionsRes.items.filter((v) => v.status === 'active');

        // Fetch available values for reassignment
        const valuesRes = await api.values.getAll();
        const available = valuesRes.items.filter(
          (v) => v.id !== value.id && v.status === 'active',
        );

        setActiveLifeVisions(activeVisions);
        setAvailableValues(available);
        setShowReassignmentDialog(true);
      } else {
        toast({
          variant: 'destructive',
          title: 'Error al retirar valor',
          description: err instanceof Error ? err.message : 'Error desconocido',
        });
        setValueToArchive(null);
      }
    } finally {
      setIsArchiving(false);
    }
  };

  // Restore handler
  const handleRestore = async () => {
    if (!valueToRestore) return;
    setIsRestoring(true);

    try {
      await api.values.restore(valueToRestore.id);
      toast({
        title: 'Valor restaurado',
        description: 'Tu valor esta activo nuevamente.',
      });
      setValueToRestore(null);
      setRefresh((prev) => prev + 1);
    } catch (err: unknown) {
      toast({
        variant: 'destructive',
        title: 'Error al restaurar valor',
        description: err instanceof Error ? err.message : 'Error desconocido',
      });
    } finally {
      setIsRestoring(false);
    }
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Valores</h1>
            <p className="text-muted-foreground mt-1">Define lo que mas importa en tu vida</p>
          </div>

          <div className="flex items-center gap-2">
            <Sheet
              open={isFormOpen}
              onOpenChange={(open) => {
                // Only allow opening if under the limit
                if (open && valuesCount >= 10) {
                  return;
                }
                setIsFormOpen(open);
              }}
            >
              <SheetTrigger asChild>
                <Button
                  size="lg"
                  disabled={valuesCount >= 10}
                  aria-disabled={valuesCount >= 10}
                  className="bg-rose-600 hover:bg-rose-700"
                >
                  <Plus className="h-5 w-5 mr-2" />
                  Nuevo valor
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
                <SheetHeader>
                  <SheetTitle className="text-xl">Define un valor</SheetTitle>
                  <SheetDescription>
                    ¿Que te importa? ¿Que guia tus decisiones y acciones?
                  </SheetDescription>
                </SheetHeader>
                <div className="mt-6">
                  <ValueForm
                    existingValuesCount={valuesCount}
                    onSuccess={handleCreateSuccess}
                    onCancel={() => setIsFormOpen(false)}
                  />
                </div>
              </SheetContent>
            </Sheet>
            {valuesCount >= 10 && (
              <p className="text-sm text-muted-foreground italic">Maximo de 10 valores alcanzado</p>
            )}
          </div>
        </div>

        {/* Info Alert */}
        <Alert className="bg-gray-50 border-gray-200">
          <Info className="h-4 w-4 text-gray-600" />
          <AlertDescription className="text-sm text-gray-700">
            Puedes definir hasta <strong>10 valores fundamentales</strong>. Arrastra para reordenar
            y establecer tus prioridades. Se permiten espacios cuando eliminas valores.
          </AlertDescription>
        </Alert>

        {/* Show Archived Toggle */}
        <div className="flex items-center gap-2 mt-4">
          <input
            type="checkbox"
            id="show-archived"
            checked={showArchived}
            onChange={(e) => setShowArchived(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300 text-rose-600 focus:ring-rose-500"
          />
          <label htmlFor="show-archived" className="text-sm cursor-pointer flex items-center gap-1">
            <Archive className="h-4 w-4" />
            Mostrar retirados
          </label>
        </div>
      </div>

      {/* Value List */}
      <ValueList
        refresh={refresh}
        showArchived={showArchived}
        onValueEdit={handleEdit}
        onValueArchive={handleArchive}
        onValueRestore={setValueToRestore}
        onValuesCountChange={setValuesCount}
      />

      {/* Edit Form Sheet */}
      <Sheet open={isEditFormOpen} onOpenChange={setIsEditFormOpen}>
        <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="text-xl">Editar valor</SheetTitle>
            <SheetDescription>Actualiza tu declaracion de valor o prioridad.</SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <ValueForm
              value={valueToEdit}
              existingValuesCount={valuesCount}
              onSuccess={handleEditSuccess}
              onCancel={() => {
                setIsEditFormOpen(false);
                setValueToEdit(null);
              }}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* Restore Confirmation Dialog */}
      <AlertDialog
        open={!!valueToRestore}
        onOpenChange={(open) => !open && setValueToRestore(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <RotateCcw className="h-5 w-5 text-green-600" />
              Restaurar valor
            </AlertDialogTitle>
            <AlertDialogDescription>
              Este valor volvera a estar activo y podras asignarle nuevas visiones de vida.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRestoring}>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRestore}
              disabled={isRestoring}
              className="bg-green-600 hover:bg-green-700"
            >
              {isRestoring ? 'Restaurando...' : 'Restaurar'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Reassignment Dialog */}
      {valueToArchive && (
        <ReassignmentDialog
          open={showReassignmentDialog}
          onOpenChange={(open) => {
            setShowReassignmentDialog(open);
            if (!open) {
              setValueToArchive(null);
              setActiveLifeVisions([]);
              setAvailableValues([]);
            }
          }}
          valueToArchive={valueToArchive}
          activeLifeVisions={activeLifeVisions}
          availableValues={availableValues}
          onComplete={() => {
            setValueToArchive(null);
            setActiveLifeVisions([]);
            setAvailableValues([]);
            setRefresh((prev) => prev + 1);
          }}
        />
      )}
    </div>
  );
}
