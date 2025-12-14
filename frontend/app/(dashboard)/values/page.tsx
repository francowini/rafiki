'use client';

import { useState } from 'react';
import { ValueForm } from '@/components/features/ValueForm';
import { ValueList } from '@/components/features/ValueList';
import { Value } from '@/lib/types';
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
import { Plus, Info } from 'lucide-react';
import { api } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';

export default function ValuesPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [valueToEdit, setValueToEdit] = useState<Value | null>(null);
  const [valueToDelete, setValueToDelete] = useState<Value | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [valuesCount, setValuesCount] = useState(0);
  const [isDeleting, setIsDeleting] = useState(false);
  const { toast } = useToast();

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

  const handleDelete = async () => {
    if (!valueToDelete) return;

    setIsDeleting(true);
    try {
      await api.values.delete(valueToDelete.id);
      toast({
        title: 'Valor eliminado',
        description: 'Tu valor ha sido eliminado exitosamente.',
      });
      setValueToDelete(null);
      setRefresh((prev) => prev + 1);
    } catch (err: any) {
      toast({
        variant: 'destructive',
        title: 'Error al eliminar valor',
        description: err.message || 'No se pudo eliminar el valor. Por favor intenta de nuevo.',
      });
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Valores</h1>
            <p className="text-muted-foreground mt-1">Define lo que más importa en tu vida</p>
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
                    ¿Qué te importa? ¿Qué guía tus decisiones y acciones?
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
              <p className="text-sm text-muted-foreground italic">Máximo de 10 valores alcanzado</p>
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
      </div>

      {/* Value List */}
      <ValueList
        refresh={refresh}
        onValueEdit={handleEdit}
        onValueDelete={setValueToDelete}
        onValuesCountChange={setValuesCount}
      />

      {/* Edit Form Sheet */}
      <Sheet open={isEditFormOpen} onOpenChange={setIsEditFormOpen}>
        <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="text-xl">Editar valor</SheetTitle>
            <SheetDescription>Actualiza tu declaración de valor o prioridad.</SheetDescription>
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

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={!!valueToDelete} onOpenChange={(open) => !open && setValueToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>¿Estás seguro?</AlertDialogTitle>
            <AlertDialogDescription>
              Esto eliminará permanentemente este valor. Esta acción no se puede deshacer.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive hover:bg-destructive/90"
            >
              {isDeleting ? 'Eliminando...' : 'Eliminar'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
