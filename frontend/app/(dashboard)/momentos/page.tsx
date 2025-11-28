'use client';

import { useState } from 'react';
import { MomentForm } from '@/components/features/MomentForm';
import { MomentList } from '@/components/features/MomentList';
import { MomentDetail } from '@/components/features/MomentDetail';
import { Moment } from '@/lib/types';
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
import { Plus, Download } from 'lucide-react';
import { api } from '@/lib/api';
import { ExportDialog } from '@/components/features/export';

export default function MomentosPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [selectedMoment, setSelectedMoment] = useState<Moment | null>(null);
  const [momentToEdit, setMomentToEdit] = useState<Moment | null>(null);
  const [momentToDelete, setMomentToDelete] = useState<Moment | null>(null);
  const [refresh, setRefresh] = useState(0);

  const handleCreateSuccess = () => {
    setIsFormOpen(false);
    setRefresh((prev) => prev + 1);
  };

  const handleEditSuccess = () => {
    setIsEditFormOpen(false);
    setMomentToEdit(null);
    setRefresh((prev) => prev + 1);
  };

  const handleEdit = (moment: Moment) => {
    setMomentToEdit(moment);
    setIsEditFormOpen(true);
  };

  const handleDelete = async () => {
    if (!momentToDelete) return;

    try {
      await api.moments.delete(momentToDelete.id);
      setMomentToDelete(null);
      setRefresh((prev) => prev + 1);
    } catch (err) {
      console.error('Error deleting moment:', err);
    }
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Moments</h1>
          <p className="text-muted-foreground mt-1">Record and reflect on your difficult moments</p>
        </div>

        <div className="flex gap-3">
          {/* Export Button */}
          <ExportDialog>
            <Button variant="outline" size="lg">
              <Download className="h-5 w-5 mr-2" />
              Export
            </Button>
          </ExportDialog>

          {/* New Moment Button */}
          <Sheet open={isFormOpen} onOpenChange={setIsFormOpen}>
            <SheetTrigger asChild>
              <Button size="lg" className="bg-purple-600 hover:bg-purple-700">
                <Plus className="h-5 w-5 mr-2" />
                New moment
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
              <SheetHeader>
                <SheetTitle className="text-xl">Record a moment</SheetTitle>
                <SheetDescription>
                  Take your time to describe what happened. There are no right or wrong answers.
                </SheetDescription>
              </SheetHeader>
              <div className="mt-6">
                <MomentForm onSuccess={handleCreateSuccess} onCancel={() => setIsFormOpen(false)} />
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>

      {/* Moment List */}
      <MomentList
        refresh={refresh}
        onMomentClick={setSelectedMoment}
        onMomentEdit={handleEdit}
        onMomentDelete={setMomentToDelete}
      />

      {/* Moment Detail Modal */}
      <MomentDetail
        moment={selectedMoment}
        open={!!selectedMoment}
        onClose={() => setSelectedMoment(null)}
      />

      {/* Edit Form Sheet */}
      <Sheet open={isEditFormOpen} onOpenChange={setIsEditFormOpen}>
        <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="text-xl">Edit moment</SheetTitle>
            <SheetDescription>Update your moment information.</SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <MomentForm
              moment={momentToEdit}
              onSuccess={handleEditSuccess}
              onCancel={() => {
                setIsEditFormOpen(false);
                setMomentToEdit(null);
              }}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        open={!!momentToDelete}
        onOpenChange={(open) => !open && setMomentToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. The moment will be permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
