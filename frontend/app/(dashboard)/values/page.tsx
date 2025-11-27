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
        title: 'Value deleted',
        description: 'Your value has been successfully deleted.',
      });
      setValueToDelete(null);
      setRefresh((prev) => prev + 1);
    } catch (err: any) {
      toast({
        variant: 'destructive',
        title: 'Error deleting value',
        description: err.message || 'Failed to delete value. Please try again.',
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
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Values</h1>
            <p className="text-muted-foreground mt-1">Define what matters most in your life</p>
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
                  New value
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
                <SheetHeader>
                  <SheetTitle className="text-xl">Define a value</SheetTitle>
                  <SheetDescription>
                    What do you care about? What guides your decisions and actions?
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
              <p className="text-sm text-muted-foreground italic">Maximum of 10 values reached</p>
            )}
          </div>
        </div>

        {/* Info Alert */}
        <Alert className="bg-gray-50 border-gray-200">
          <Info className="h-4 w-4 text-gray-600" />
          <AlertDescription className="text-sm text-gray-700">
            You can define up to <strong>10 core values</strong>. Drag to reorder and set your
            priorities. Gaps are allowed when you delete values.
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
            <SheetTitle className="text-xl">Edit value</SheetTitle>
            <SheetDescription>Update your value statement or priority.</SheetDescription>
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
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete this value. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive hover:bg-destructive/90"
            >
              {isDeleting ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
