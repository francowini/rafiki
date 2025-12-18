'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { Role, RoleWithValues, STATUS_ACTIVE } from '@/lib/types';
import { api } from '@/lib/api';
import { RoleCard } from './RoleCard';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertCircle, Users } from 'lucide-react';
import { ROLE_MESSAGES } from '@/lib/role-utils';

interface RoleListProps {
  refresh: number;
  showArchived: boolean;
  onRoleEdit: (role: Role) => void;
  onRoleArchive: (role: Role) => void;
  onRoleRestore: (role: Role) => void;
  onRoleConnectValues: (role: Role) => void;
  onRolesCountChange?: (count: number) => void;
}

export function RoleList({
  refresh,
  showArchived,
  onRoleEdit,
  onRoleArchive,
  onRoleRestore,
  onRoleConnectValues,
  onRolesCountChange,
}: RoleListProps) {
  const [roles, setRoles] = useState<Role[]>([]);
  const [roleValuesCount, setRoleValuesCount] = useState<Record<string, number>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Use requestId pattern to handle stale responses (F2 compliance)
  const requestIdRef = useRef(0);

  const fetchRoles = useCallback(async () => {
    const currentRequestId = ++requestIdRef.current;
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.roles.getAll({ includeArchived: showArchived });

      // Check if this response is still current
      if (currentRequestId !== requestIdRef.current) return;

      setRoles(response.items);

      // Count active roles for parent component
      const activeCount = response.items.filter((r) => r.status === STATUS_ACTIVE).length;
      onRolesCountChange?.(activeCount);

      // Fetch values count for each role
      const countsMap: Record<string, number> = {};
      const roleDetailsPromises = response.items.map(async (role) => {
        try {
          const roleWithValues: RoleWithValues = await api.roles.getById(role.id);
          countsMap[role.id] = roleWithValues.valueIds?.length || 0;
        } catch {
          countsMap[role.id] = 0;
        }
      });

      // Use Promise.allSettled for batch operations (F8 compliance)
      await Promise.allSettled(roleDetailsPromises);

      if (currentRequestId === requestIdRef.current) {
        setRoleValuesCount(countsMap);
      }
    } catch (err) {
      if (currentRequestId !== requestIdRef.current) return;
      console.error('Failed to fetch roles:', err);
      setError(err instanceof Error ? err.message : 'Error al cargar los roles');
    } finally {
      if (currentRequestId === requestIdRef.current) {
        setIsLoading(false);
      }
    }
  }, [showArchived, onRolesCountChange]);

  useEffect(() => {
    fetchRoles();
  }, [fetchRoles, refresh]);

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="p-4 border rounded-lg">
            <Skeleton className="h-6 w-24 mb-2" />
            <Skeleton className="h-4 w-full mb-2" />
            <Skeleton className="h-4 w-3/4" />
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  // Filter roles based on showArchived
  const displayedRoles = showArchived
    ? roles
    : roles.filter((r) => r.status === STATUS_ACTIVE);

  if (displayedRoles.length === 0) {
    return (
      <div className="text-center py-12 border-2 border-dashed border-gray-200 rounded-lg">
        <Users className="h-12 w-12 mx-auto text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {showArchived ? 'No hay roles' : 'Comienza a definir tus roles'}
        </h3>
        <p className="text-muted-foreground max-w-sm mx-auto">
          {showArchived
            ? 'No tienes roles creados aún.'
            : 'Los roles representan las diferentes facetas de tu vida. Crea tu primer rol para empezar a conectarlo con tus valores.'}
        </p>
      </div>
    );
  }

  // Separate active and archived roles
  const activeRoles = displayedRoles.filter((r) => r.status === STATUS_ACTIVE);
  const archivedRoles = displayedRoles.filter((r) => r.status !== STATUS_ACTIVE);

  return (
    <div className="space-y-8">
      {/* Active Roles */}
      {activeRoles.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {activeRoles.map((role) => (
            <RoleCard
              key={role.id}
              role={role}
              connectedValuesCount={roleValuesCount[role.id] || 0}
              onEdit={() => onRoleEdit(role)}
              onArchive={() => onRoleArchive(role)}
              onConnectValues={() => onRoleConnectValues(role)}
            />
          ))}
        </div>
      )}

      {/* Max Roles Warning */}
      {activeRoles.length >= 15 && (
        <Alert className="bg-amber-50 border-amber-200">
          <AlertCircle className="h-4 w-4 text-amber-600" />
          <AlertDescription className="text-amber-800">
            {ROLE_MESSAGES.maxRolesReached.description}
          </AlertDescription>
        </Alert>
      )}

      {/* Archived Roles */}
      {showArchived && archivedRoles.length > 0 && (
        <div>
          <h3 className="text-lg font-medium text-gray-700 mb-4 flex items-center gap-2">
            <span className="text-gray-400">Roles retirados</span>
            <span className="text-sm text-gray-500">({archivedRoles.length})</span>
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {archivedRoles.map((role) => (
              <RoleCard
                key={role.id}
                role={role}
                connectedValuesCount={roleValuesCount[role.id] || 0}
                onRestore={() => onRoleRestore(role)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
