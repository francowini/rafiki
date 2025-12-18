'use client';

import { useState, useEffect, useRef } from 'react';
import { Role, NewRole, Value, RoleType, RoleFacet, STATUS_ACTIVE } from '@/lib/types';
import { api } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ValueSelector } from './ValueSelector';
import { useToast } from '@/hooks/use-toast';
import { getAllRoleTypes, getAllRoleFacets, getRoleTypeConfig, getRoleFacetConfig } from '@/lib/role-utils';

interface RoleFormProps {
  role?: Role | null;
  existingRolesCount: number;
  onSuccess: () => void;
  onCancel: () => void;
}

export function RoleForm({ role, existingRolesCount, onSuccess, onCancel }: RoleFormProps) {
  const [name, setName] = useState(role?.name || '');
  const [roleType, setRoleType] = useState<RoleType>(role?.roleType || 'personal');
  const [facet, setFacet] = useState<RoleFacet>(role?.facet || 'family_relationships');
  const [description, setDescription] = useState(role?.description || '');
  const [selectedValueIds, setSelectedValueIds] = useState<string[]>([]);
  const [availableValues, setAvailableValues] = useState<Value[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isFetchingValues, setIsFetchingValues] = useState(true);
  const { toast } = useToast();

  // Use requestId pattern for stale response handling (F2 compliance)
  const requestIdRef = useRef(0);

  // Fetch available values on mount
  useEffect(() => {
    const fetchValues = async () => {
      const currentRequestId = ++requestIdRef.current;
      setIsFetchingValues(true);

      try {
        const response = await api.values.getAll();
        if (currentRequestId !== requestIdRef.current) return;

        // Only show active values
        const activeValues = response.items.filter((v) => v.status === STATUS_ACTIVE);
        setAvailableValues(activeValues);

        // If editing, fetch role's connected values
        if (role) {
          try {
            const roleWithValues = await api.roles.getById(role.id);
            if (currentRequestId !== requestIdRef.current) return;
            setSelectedValueIds(roleWithValues.valueIds || []);
          } catch (err) {
            console.error('Failed to fetch role values:', err);
          }
        }
      } catch (err) {
        if (currentRequestId !== requestIdRef.current) return;
        console.error('Failed to fetch values:', err);
        toast({
          variant: 'destructive',
          title: 'Error al cargar valores',
          description: 'No se pudieron cargar los valores disponibles.',
        });
      } finally {
        if (currentRequestId === requestIdRef.current) {
          setIsFetchingValues(false);
        }
      }
    };

    fetchValues();
  }, [role, toast]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast({
        variant: 'destructive',
        title: 'Error de validación',
        description: 'El nombre del rol es requerido.',
      });
      return;
    }

    setIsLoading(true);

    try {
      if (role) {
        // Update existing role
        await api.roles.update(role.id, {
          name: name.trim(),
          roleType,
          facet,
          description: description.trim(),
        });

        // Update value connections
        await api.roles.bulkSetValues(role.id, selectedValueIds);

        toast({
          title: 'Rol actualizado',
          description: 'El rol se actualizó correctamente.',
        });
      } else {
        // Create new role
        const newRole: NewRole = {
          name: name.trim(),
          roleType,
          facet,
          description: description.trim(),
          displayOrder: existingRolesCount + 1,
          valueIds: selectedValueIds,
        };

        await api.roles.create(newRole);

        toast({
          title: 'Rol creado',
          description: 'El rol se creó correctamente. Ahora puedes conectarlo con tus valores.',
        });
      }

      // Reset form state before closing (F9 compliance)
      setName('');
      setRoleType('personal');
      setFacet('family_relationships');
      setDescription('');
      setSelectedValueIds([]);

      onSuccess();
    } catch (err) {
      console.error('Failed to save role:', err);
      toast({
        variant: 'destructive',
        title: role ? 'Error al actualizar rol' : 'Error al crear rol',
        description: err instanceof Error ? err.message : 'Error desconocido',
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    // Reset form state on cancel (F9 compliance)
    setName('');
    setRoleType('personal');
    setFacet('family_relationships');
    setDescription('');
    setSelectedValueIds([]);
    onCancel();
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Name */}
      <div className="space-y-2">
        <Label htmlFor="role-name">Nombre del rol *</Label>
        <Input
          id="role-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Ej: Padre, Líder de equipo, Amigo..."
          disabled={isLoading}
          maxLength={100}
        />
        <p className="text-xs text-muted-foreground">
          Un nombre corto que describa este rol en tu vida.
        </p>
      </div>

      {/* Role Type */}
      <div className="space-y-2">
        <Label htmlFor="role-type">Tipo de rol</Label>
        <Select
          value={roleType}
          onValueChange={(value) => setRoleType(value as RoleType)}
          disabled={isLoading}
        >
          <SelectTrigger id="role-type">
            <SelectValue placeholder="Selecciona un tipo" />
          </SelectTrigger>
          <SelectContent>
            {getAllRoleTypes().map((type) => {
              const config = getRoleTypeConfig(type);
              return (
                <SelectItem key={type} value={type}>
                  <span className="flex items-center gap-2">
                    <span>{config.icon}</span>
                    <span>{config.label}</span>
                  </span>
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
      </div>

      {/* Facet */}
      <div className="space-y-2">
        <Label htmlFor="role-facet">Área de vida</Label>
        <Select
          value={facet}
          onValueChange={(value) => setFacet(value as RoleFacet)}
          disabled={isLoading}
        >
          <SelectTrigger id="role-facet">
            <SelectValue placeholder="Selecciona un área" />
          </SelectTrigger>
          <SelectContent>
            {getAllRoleFacets().map((f) => {
              const config = getRoleFacetConfig(f);
              return (
                <SelectItem key={f} value={f}>
                  <span className="flex items-center gap-2">
                    <span>{config.icon}</span>
                    <span>{config.label}</span>
                  </span>
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
      </div>

      {/* Description */}
      <div className="space-y-2">
        <Label htmlFor="role-description">Descripción (opcional)</Label>
        <Textarea
          id="role-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Describe qué significa este rol para ti..."
          disabled={isLoading}
          rows={3}
          maxLength={500}
        />
      </div>

      {/* Value Selector */}
      <div className="space-y-2">
        <Label>Valores conectados</Label>
        <p className="text-xs text-muted-foreground mb-2">
          Selecciona los valores que guían cómo actúas en este rol.
        </p>
        {isFetchingValues ? (
          <div className="border rounded-md p-4 text-center text-muted-foreground">
            Cargando valores...
          </div>
        ) : (
          <ValueSelector
            values={availableValues}
            selectedValueIds={selectedValueIds}
            onChange={setSelectedValueIds}
            disabled={isLoading}
          />
        )}
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-4 border-t">
        <Button type="button" variant="outline" onClick={handleCancel} disabled={isLoading}>
          Cancelar
        </Button>
        <Button type="submit" disabled={isLoading} className="bg-purple-600 hover:bg-purple-700">
          {isLoading ? 'Guardando...' : role ? 'Actualizar' : 'Crear rol'}
        </Button>
      </div>
    </form>
  );
}
