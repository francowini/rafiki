'use client';

import { useState } from 'react';
import { RoleForm } from '@/components/features/roles/RoleForm';
import { RoleList } from '@/components/features/roles/RoleList';
import { RoleValueGraph } from '@/components/features/roles/RoleValueGraph';
import { ValueSelector } from '@/components/features/roles/ValueSelector';
import { Role, Value, RoleWithValues, STATUS_ACTIVE } from '@/lib/types';
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
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Plus, Info, Archive, RotateCcw, LayoutGrid, GitBranch, Link2 } from 'lucide-react';
import { api } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';
import { ROLE_MESSAGES } from '@/lib/role-utils';

export default function RolesPage() {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [roleToEdit, setRoleToEdit] = useState<Role | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [rolesCount, setRolesCount] = useState(0);
  const [activeTab, setActiveTab] = useState('cards');
  const { toast } = useToast();

  // Archive/Restore state
  const [showArchived, setShowArchived] = useState(false);
  const [roleToArchive, setRoleToArchive] = useState<Role | null>(null);
  const [roleToRestore, setRoleToRestore] = useState<Role | null>(null);
  const [isArchiving, setIsArchiving] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);

  // Value connection sheet state
  const [roleToConnect, setRoleToConnect] = useState<Role | null>(null);
  const [isConnectSheetOpen, setIsConnectSheetOpen] = useState(false);
  const [availableValues, setAvailableValues] = useState<Value[]>([]);
  const [selectedValueIds, setSelectedValueIds] = useState<string[]>([]);
  const [isSavingConnections, setIsSavingConnections] = useState(false);

  const handleCreateSuccess = () => {
    setIsFormOpen(false);
    setRefresh((prev) => prev + 1);
  };

  const handleEditSuccess = () => {
    setIsEditFormOpen(false);
    setRoleToEdit(null);
    setRefresh((prev) => prev + 1);
  };

  const handleEdit = (role: Role) => {
    setRoleToEdit(role);
    setIsEditFormOpen(true);
  };

  // Archive handler
  const handleArchive = (role: Role) => {
    setRoleToArchive(role);
  };

  const confirmArchive = async () => {
    if (!roleToArchive) return;
    setIsArchiving(true);

    try {
      await api.roles.archive(roleToArchive.id);
      toast({
        title: 'Rol retirado',
        description: 'Tu rol ha sido retirado exitosamente.',
      });
      setRoleToArchive(null);
      setRefresh((prev) => prev + 1);
    } catch (err: unknown) {
      toast({
        variant: 'destructive',
        title: 'Error al retirar rol',
        description: err instanceof Error ? err.message : 'Error desconocido',
      });
    } finally {
      setIsArchiving(false);
    }
  };

  // Restore handler
  const handleRestore = async () => {
    if (!roleToRestore) return;
    setIsRestoring(true);

    try {
      await api.roles.restore(roleToRestore.id);
      toast({
        title: 'Rol restaurado',
        description: 'Tu rol está activo nuevamente.',
      });
      setRoleToRestore(null);
      setRefresh((prev) => prev + 1);
    } catch (err: unknown) {
      toast({
        variant: 'destructive',
        title: 'Error al restaurar rol',
        description: err instanceof Error ? err.message : 'Error desconocido',
      });
    } finally {
      setIsRestoring(false);
    }
  };

  // Connect values handler
  const handleConnectValues = async (role: Role) => {
    setRoleToConnect(role);
    setIsConnectSheetOpen(true);

    try {
      // Fetch available values
      const valuesRes = await api.values.getAll();
      const activeValues = valuesRes.items.filter((v) => v.status === STATUS_ACTIVE);
      setAvailableValues(activeValues);

      // Fetch current connections
      const roleWithValues: RoleWithValues = await api.roles.getById(role.id);
      setSelectedValueIds(roleWithValues.valueIds || []);
    } catch (err) {
      console.error('Failed to load connection data:', err);
      toast({
        variant: 'destructive',
        title: 'Error al cargar datos',
        description: 'No se pudieron cargar los valores disponibles.',
      });
      setIsConnectSheetOpen(false);
      setRoleToConnect(null);
    }
  };

  const saveValueConnections = async () => {
    if (!roleToConnect) return;
    setIsSavingConnections(true);

    try {
      await api.roles.bulkSetValues(roleToConnect.id, selectedValueIds);
      toast({
        title: 'Conexiones guardadas',
        description: 'Los valores se conectaron correctamente.',
      });
      setIsConnectSheetOpen(false);
      setRoleToConnect(null);
      setSelectedValueIds([]);
      setRefresh((prev) => prev + 1);
    } catch (err) {
      console.error('Failed to save connections:', err);
      toast({
        variant: 'destructive',
        title: 'Error al guardar conexiones',
        description: err instanceof Error ? err.message : 'Error desconocido',
      });
    } finally {
      setIsSavingConnections(false);
    }
  };

  const closeConnectSheet = () => {
    // Reset state on close (F9 compliance)
    setIsConnectSheetOpen(false);
    setRoleToConnect(null);
    setSelectedValueIds([]);
    setAvailableValues([]);
  };

  return (
    <div className="container mx-auto py-8 px-4 max-w-7xl">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Roles</h1>
            <p className="text-muted-foreground mt-1">
              Define los roles que desempeñas en tu vida y conéctalos con tus valores
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Sheet
              open={isFormOpen}
              onOpenChange={(open) => {
                // Only allow opening if under the limit
                if (open && rolesCount >= 15) {
                  return;
                }
                setIsFormOpen(open);
              }}
            >
              <SheetTrigger asChild>
                <Button
                  size="lg"
                  disabled={rolesCount >= 15}
                  aria-disabled={rolesCount >= 15}
                  className="bg-purple-600 hover:bg-purple-700"
                >
                  <Plus className="h-5 w-5 mr-2" />
                  Nuevo rol
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
                <SheetHeader>
                  <SheetTitle className="text-xl">Crear un rol</SheetTitle>
                  <SheetDescription>
                    ¿Qué rol desempeñas en tu vida? Define cómo quieres actuar en ese rol.
                  </SheetDescription>
                </SheetHeader>
                <div className="mt-6">
                  <RoleForm
                    existingRolesCount={rolesCount}
                    onSuccess={handleCreateSuccess}
                    onCancel={() => setIsFormOpen(false)}
                  />
                </div>
              </SheetContent>
            </Sheet>
            {rolesCount >= 15 && (
              <p className="text-sm text-muted-foreground italic">
                {ROLE_MESSAGES.maxRolesReached.title}
              </p>
            )}
          </div>
        </div>

        {/* Info Alert */}
        <Alert className="bg-gray-50 border-gray-200">
          <Info className="h-4 w-4 text-gray-600" />
          <AlertDescription className="text-sm text-gray-700">
            Puedes definir hasta <strong>15 roles</strong>. Cada rol puede conectarse con múltiples
            valores. Usa la vista de grafo para visualizar las conexiones.
          </AlertDescription>
        </Alert>

        {/* Show Archived Toggle */}
        <div className="flex items-center gap-3 mt-4">
          <Switch
            id="show-archived"
            checked={showArchived}
            onCheckedChange={setShowArchived}
            aria-label="Mostrar roles retirados"
          />
          <Label htmlFor="show-archived" className="cursor-pointer flex items-center gap-1.5">
            <Archive className="h-4 w-4" />
            Mostrar retirados
          </Label>
        </div>
      </div>

      {/* Tabs for Cards / Graph views */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="mb-6">
          <TabsTrigger value="cards" className="flex items-center gap-2">
            <LayoutGrid className="h-4 w-4" />
            Tarjetas
          </TabsTrigger>
          <TabsTrigger value="graph" className="flex items-center gap-2">
            <GitBranch className="h-4 w-4" />
            Grafo
          </TabsTrigger>
        </TabsList>

        <TabsContent value="cards">
          <RoleList
            refresh={refresh}
            showArchived={showArchived}
            onRoleEdit={handleEdit}
            onRoleArchive={handleArchive}
            onRoleRestore={setRoleToRestore}
            onRoleConnectValues={handleConnectValues}
            onRolesCountChange={setRolesCount}
          />
        </TabsContent>

        <TabsContent value="graph">
          <RoleValueGraph refresh={refresh} onRoleEdit={handleEdit} />
        </TabsContent>
      </Tabs>

      {/* Edit Form Sheet */}
      <Sheet open={isEditFormOpen} onOpenChange={setIsEditFormOpen}>
        <SheetContent side="right" className="w-full sm:max-w-2xl overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="text-xl">Editar rol</SheetTitle>
            <SheetDescription>Actualiza la información de tu rol.</SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <RoleForm
              role={roleToEdit}
              existingRolesCount={rolesCount}
              onSuccess={handleEditSuccess}
              onCancel={() => {
                setIsEditFormOpen(false);
                setRoleToEdit(null);
              }}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* Connect Values Sheet */}
      <Sheet open={isConnectSheetOpen} onOpenChange={(open) => !open && closeConnectSheet()}>
        <SheetContent side="right" className="w-full sm:max-w-lg overflow-y-auto">
          <SheetHeader>
            <SheetTitle className="text-xl flex items-center gap-2">
              <Link2 className="h-5 w-5 text-purple-600" />
              Conectar valores
            </SheetTitle>
            <SheetDescription>
              {roleToConnect && (
                <>
                  Selecciona los valores que guían cómo actúas en tu rol de{' '}
                  <strong>{roleToConnect.name}</strong>.
                </>
              )}
            </SheetDescription>
          </SheetHeader>
          <div className="mt-6">
            <ValueSelector
              values={availableValues}
              selectedValueIds={selectedValueIds}
              onChange={setSelectedValueIds}
              disabled={isSavingConnections}
            />
            <div className="flex justify-end gap-3 pt-6 border-t mt-6">
              <Button
                variant="outline"
                onClick={closeConnectSheet}
                disabled={isSavingConnections}
              >
                Cancelar
              </Button>
              <Button
                onClick={saveValueConnections}
                disabled={isSavingConnections}
                className="bg-purple-600 hover:bg-purple-700"
              >
                {isSavingConnections ? 'Guardando...' : 'Guardar conexiones'}
              </Button>
            </div>
          </div>
        </SheetContent>
      </Sheet>

      {/* Archive Confirmation Dialog */}
      <AlertDialog open={!!roleToArchive} onOpenChange={(open) => !open && setRoleToArchive(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <Archive className="h-5 w-5 text-amber-600" />
              Retirar rol
            </AlertDialogTitle>
            <AlertDialogDescription>
              Este rol será retirado pero no eliminado. Podrás restaurarlo más tarde si lo deseas.
              Las conexiones con valores se mantendrán.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isArchiving}>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmArchive}
              disabled={isArchiving}
              className="bg-amber-600 hover:bg-amber-700"
            >
              {isArchiving ? 'Retirando...' : 'Retirar'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Restore Confirmation Dialog */}
      <AlertDialog
        open={!!roleToRestore}
        onOpenChange={(open) => !open && setRoleToRestore(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <RotateCcw className="h-5 w-5 text-green-600" />
              Restaurar rol
            </AlertDialogTitle>
            <AlertDialogDescription>
              Este rol volverá a estar activo y podrás seguir conectándolo con tus valores.
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
    </div>
  );
}
