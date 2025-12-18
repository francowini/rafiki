'use client';

import { Role, STATUS_ARCHIVED } from '@/lib/types';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Edit, MoreVertical, Archive, RotateCcw, Link2 } from 'lucide-react';
import { getRoleFacetConfig, getRoleTypeConfig, ROLE_GRADIENT_MAP } from '@/lib/role-utils';
import { formatDateShort } from '@/lib/date-utils';

interface RoleCardProps {
  role: Role;
  connectedValuesCount?: number;
  onEdit?: () => void;
  onArchive?: () => void;
  onRestore?: () => void;
  onConnectValues?: () => void;
}

export function RoleCard({
  role,
  connectedValuesCount = 0,
  onEdit,
  onArchive,
  onRestore,
  onConnectValues,
}: RoleCardProps) {
  const facetConfig = getRoleFacetConfig(role.facet);
  const typeConfig = getRoleTypeConfig(role.roleType);
  const gradientClass = ROLE_GRADIENT_MAP[role.facet] || 'bg-white';
  const isArchived = role.status === STATUS_ARCHIVED;

  return (
    <Card
      className={`transition-all ${gradientClass} ${
        isArchived
          ? 'opacity-60 border-dashed border-2 border-gray-300 bg-gray-50/50'
          : 'hover:border-purple-200'
      } hover:shadow-md`}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 flex-wrap flex-1">
            {/* Role Type Badge */}
            <Badge
              variant="outline"
              className={`${typeConfig.bgColor} ${typeConfig.color} ${typeConfig.borderColor}`}
            >
              <span className="mr-1">{typeConfig.icon}</span>
              {typeConfig.label}
            </Badge>

            {/* Facet Badge */}
            <Badge
              variant="outline"
              className={`${facetConfig.bgColor} ${facetConfig.color} ${facetConfig.borderColor}`}
            >
              <span className="mr-1">{facetConfig.icon}</span>
              {facetConfig.label}
            </Badge>

            {/* Archived Badge */}
            {isArchived && (
              <Badge variant="outline" className="bg-gray-100 text-gray-600 border-gray-400">
                <Archive className="h-3 w-3 mr-1" />
                Retirado
              </Badge>
            )}

            {/* Connected Values Badge (only for active roles) */}
            {!isArchived && connectedValuesCount > 0 && (
              <Badge variant="outline" className="bg-purple-50 text-purple-700 border-purple-300">
                <Link2 className="h-3 w-3 mr-1" />
                {connectedValuesCount} valor{connectedValuesCount !== 1 ? 'es' : ''}
              </Badge>
            )}
          </div>

          {/* Dropdown menu for actions */}
          {(onEdit || onArchive || onRestore || onConnectValues) && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-8 w-8 p-0" aria-label="Acciones">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {!isArchived && onEdit && (
                  <DropdownMenuItem onClick={onEdit}>
                    <Edit className="h-4 w-4 mr-2" />
                    Editar
                  </DropdownMenuItem>
                )}
                {!isArchived && onConnectValues && (
                  <DropdownMenuItem onClick={onConnectValues}>
                    <Link2 className="h-4 w-4 mr-2" />
                    Conectar valores
                  </DropdownMenuItem>
                )}
                {!isArchived && onArchive && (
                  <DropdownMenuItem
                    onClick={onArchive}
                    className="text-amber-600 focus:text-amber-600"
                  >
                    <Archive className="h-4 w-4 mr-2" />
                    Retirar
                  </DropdownMenuItem>
                )}
                {isArchived && onRestore && (
                  <DropdownMenuItem
                    onClick={onRestore}
                    className="text-green-600 focus:text-green-600"
                  >
                    <RotateCcw className="h-4 w-4 mr-2" />
                    Restaurar
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        <div>
          <h3 className="font-semibold text-lg text-gray-900">{role.name}</h3>
          {role.description && (
            <p className="text-sm text-muted-foreground mt-1 leading-relaxed">{role.description}</p>
          )}
        </div>

        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>Actualizado {formatDateShort(role.dateUpdated)}</span>
          {isArchived && role.archivedAt && (
            <span className="flex items-center gap-1 text-gray-500">
              <Archive className="h-3 w-3" />
              Retirado el {formatDateShort(role.archivedAt)}
            </span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
