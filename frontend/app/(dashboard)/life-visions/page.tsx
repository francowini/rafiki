'use client';

import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';
import { LifeVision, Value } from '@/lib/types';
import { LifeVisionForm } from '@/components/features/life-visions/LifeVisionForm';
import { LifeVisionsByValue } from '@/components/features/life-visions/LifeVisionsByValue';
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
import { Skeleton } from '@/components/ui/skeleton';
import { Plus, Info, Sparkles, Loader2 } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import Link from 'next/link';

export default function LifeVisionsPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [visionToEdit, setVisionToEdit] = useState<LifeVision | null>(null);
  const [visionToDelete, setVisionToDelete] = useState<LifeVision | null>(null);
  const [values, setValues] = useState<Value[]>([]);
  const [visions, setVisions] = useState<LifeVision[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isDeleting, setIsDeleting] = useState(false);
  const { toast } = useToast();

  const fetchData = useCallback(async () => {
    try {
      const [valuesRes, visionsRes] = await Promise.all([
        api.values.getAll(),
        api.lifeVisions.getAll(),
      ]);
      setValues(valuesRes.items);
      setVisions(visionsRes.items);
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : 'No se pudieron cargar los datos. Por favor intenta de nuevo.';
      toast({
        variant: 'destructive',
        title: 'Error al cargar datos',
        description: message,
      });
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreateSuccess = () => {
    setIsFormOpen(false);
    fetchData();
  };

  const handleEditSuccess = () => {
    setIsEditFormOpen(false);
    setVisionToEdit(null);
    fetchData();
  };

  const handleEdit = (vision: LifeVision) => {
    setVisionToEdit(vision);
    setIsEditFormOpen(true);
  };

  const handleDelete = async () => {
    if (!visionToDelete) return;

    setIsDeleting(true);
    try {
      await api.lifeVisions.delete(visionToDelete.id);
      toast({
        title: 'Visión eliminada',
        description: 'Tu visión de vida ha sido eliminada exitosamente.',
      });
      setVisionToDelete(null);
      fetchData();
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : 'No se pudo eliminar la visión. Por favor intenta de nuevo.';
      toast({
        variant: 'destructive',
        title: 'Error al eliminar visión',
        description: message,
      });
    } finally {
      setIsDeleting(false);
    }
  };

  const getVisionsForValue = (valueId: string) => {
    return visions.filter((v) => v.valueId === valueId);
  };

  if (isLoading) {
    return (
      <div className="container mx-auto py-8 px-4 max-w-7xl">
        <div className="mb-6">
          <Skeleton className="h-10 w-48 mb-2" />
          <Skeleton className="h-5 w-96" />
        </div>
        <Skeleton className="h-24 w-full mb-6" />
        <div className="space-y-6">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-32 w-full" />
          ))}
        </div>
      </div>
    );
  }

  if (values.length === 0) {
    return (
      <div className="container mx-auto py-8 px-4 max-w-7xl">
        <div className="mb-6">
          <div className="flex items-center gap-2 mb-2">
            <Sparkles className="h-8 w-8 text-gray-600" />
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Visiones de Vida</h1>
          </div>
          <p className="text-muted-foreground">Define cómo quieres vivir cada uno de tus valores</p>
        </div>

        <div className="border rounded-lg p-8 text-center">
          <Sparkles className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium mb-2">Aún no hay valores definidos</h3>
          <p className="text-muted-foreground mb-6">
            Necesitas crear valores primero antes de agregar visiones de vida.
          </p>
          <Link href="/values">
            <Button>Define tus valores</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <Sparkles className="h-8 w-8 text-gray-600" />
              <h1 className="text-3xl font-bold tracking-tight text-gray-900">Visiones de Vida</h1>
            </div>
            <p className="text-muted-foreground">
              Define cómo quieres vivir cada uno de tus valores
            </p>
          </div>

          <Sheet open={isFormOpen} onOpenChange={setIsFormOpen}>
            <SheetTrigger asChild>
              <Button size="lg">
                <Plus className="h-5 w-5 mr-2" />
                Nueva visión
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
              <SheetHeader>
                <SheetTitle className="text-xl">Crear una visión de vida</SheetTitle>
                <SheetDescription>
                  Describe cómo quieres vivir un valor específico. Pinta una imagen de tu futuro
                  ideal.
                </SheetDescription>
              </SheetHeader>
              <div className="mt-6">
                <LifeVisionForm
                  values={values}
                  onSuccess={handleCreateSuccess}
                  onCancel={() => setIsFormOpen(false)}
                />
              </div>
            </SheetContent>
          </Sheet>
        </div>

        <Alert className="bg-gray-50 border-gray-200">
          <Info className="h-4 w-4 text-gray-600" />
          <AlertDescription className="text-sm text-gray-700">
            Cada valor puede tener múltiples visiones, pero recomendamos máximo 2 por valor para
            mayor claridad. Las visiones te ayudan a traducir valores abstractos en metas de vida
            concretas.
          </AlertDescription>
        </Alert>
      </div>

      <div className="space-y-8">
        {values.map((value) => (
          <LifeVisionsByValue
            key={value.id}
            value={value}
            visions={getVisionsForValue(value.id)}
            onEdit={handleEdit}
            onDelete={setVisionToDelete}
          />
        ))}
      </div>

      <Sheet open={isEditFormOpen} onOpenChange={setIsEditFormOpen}>
        <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="text-xl">Editar visión de vida</SheetTitle>
            <SheetDescription>
              Actualiza tu declaración de visión o muévela a un valor diferente.
            </SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <LifeVisionForm
              lifeVision={visionToEdit}
              values={values}
              onSuccess={handleEditSuccess}
              onCancel={() => {
                setIsEditFormOpen(false);
                setVisionToEdit(null);
              }}
            />
          </div>
        </SheetContent>
      </Sheet>

      <AlertDialog
        open={!!visionToDelete}
        onOpenChange={(open) => !open && setVisionToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>¿Estás seguro?</AlertDialogTitle>
            <AlertDialogDescription>
              Esto eliminará permanentemente esta visión de vida. Esta acción no se puede deshacer.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive hover:bg-destructive/90"
            >
              {isDeleting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Eliminando...
                </>
              ) : (
                'Eliminar'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
