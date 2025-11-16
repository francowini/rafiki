"use client";

import { useState } from "react";
import { MomentForm } from "@/components/features/MomentForm";
import { MomentList } from "@/components/features/MomentList";
import { MomentDetail } from "@/components/features/MomentDetail";
import { Moment } from "@/lib/types";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Plus } from "lucide-react";
import { api } from "@/lib/api";

export default function MomentosPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [selectedMoment, setSelectedMoment] = useState<Moment | null>(null);
  const [momentToDelete, setMomentToDelete] = useState<Moment | null>(null);
  const [refresh, setRefresh] = useState(0);

  const handleCreateSuccess = () => {
    setIsFormOpen(false);
    setRefresh((prev) => prev + 1);
  };

  const handleDelete = async () => {
    if (!momentToDelete) return;

    try {
      await api.moments.delete(momentToDelete.id);
      setMomentToDelete(null);
      setRefresh((prev) => prev + 1);
    } catch (err) {
      console.error("Error deleting moment:", err);
    }
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Momentos</h1>
          <p className="text-muted-foreground mt-1">
            Registra y reflexiona sobre tus momentos dificiles
          </p>
        </div>

        <Sheet open={isFormOpen} onOpenChange={setIsFormOpen}>
          <SheetTrigger asChild>
            <Button
              size="lg"
              className="bg-purple-600 hover:bg-purple-700"
            >
              <Plus className="h-5 w-5 mr-2" />
              Nuevo momento
            </Button>
          </SheetTrigger>
          <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
            <SheetHeader>
              <SheetTitle className="text-xl">Registrar un momento</SheetTitle>
              <SheetDescription>
                Tomate tu tiempo para describir lo que sucedio. No hay respuestas correctas o incorrectas.
              </SheetDescription>
            </SheetHeader>
            <div className="mt-6">
              <MomentForm
                onSuccess={handleCreateSuccess}
                onCancel={() => setIsFormOpen(false)}
              />
            </div>
          </SheetContent>
        </Sheet>
      </div>

      {/* Moment List */}
      <MomentList
        refresh={refresh}
        onMomentClick={setSelectedMoment}
        onMomentDelete={setMomentToDelete}
      />

      {/* Moment Detail Modal */}
      <MomentDetail
        moment={selectedMoment}
        open={!!selectedMoment}
        onClose={() => setSelectedMoment(null)}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        open={!!momentToDelete}
        onOpenChange={(open) => !open && setMomentToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Estas seguro?</AlertDialogTitle>
            <AlertDialogDescription>
              Esta accion no se puede deshacer. El momento sera eliminado permanentemente.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive hover:bg-destructive/90"
            >
              Eliminar
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
