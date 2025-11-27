'use client';

import { useState, useEffect } from 'react';
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

  const fetchData = async () => {
    try {
      const [valuesRes, visionsRes] = await Promise.all([
        api.values.getAll(),
        api.lifeVisions.getAll(),
      ]);
      setValues(valuesRes.items);
      setVisions(visionsRes.items);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load data. Please try again.';
      toast({
        variant: 'destructive',
        title: 'Error loading data',
        description: message,
      });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

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
        title: 'Vision deleted',
        description: 'Your life vision has been successfully deleted.',
      });
      setVisionToDelete(null);
      fetchData();
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Failed to delete vision. Please try again.';
      toast({
        variant: 'destructive',
        title: 'Error deleting vision',
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
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Life Visions</h1>
          </div>
          <p className="text-muted-foreground">Define how you want to live each of your values</p>
        </div>

        <div className="border rounded-lg p-8 text-center">
          <Sparkles className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium mb-2">No values defined yet</h3>
          <p className="text-muted-foreground mb-6">
            You need to create values first before adding life visions.
          </p>
          <Link href="/values">
            <Button>Define your values</Button>
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
              <h1 className="text-3xl font-bold tracking-tight text-gray-900">Life Visions</h1>
            </div>
            <p className="text-muted-foreground">Define how you want to live each of your values</p>
          </div>

          <Sheet open={isFormOpen} onOpenChange={setIsFormOpen}>
            <SheetTrigger asChild>
              <Button size="lg">
                <Plus className="h-5 w-5 mr-2" />
                New vision
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
              <SheetHeader>
                <SheetTitle className="text-xl">Create a life vision</SheetTitle>
                <SheetDescription>
                  Describe how you want to live a specific value. Paint a picture of your ideal
                  future.
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
            Each value can have multiple visions, but we recommend max 2 per value for clarity.
            Visions help you translate abstract values into concrete life goals.
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
            <SheetTitle className="text-xl">Edit life vision</SheetTitle>
            <SheetDescription>
              Update your vision statement or move it to a different value.
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
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete this life vision. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive hover:bg-destructive/90"
            >
              {isDeleting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
