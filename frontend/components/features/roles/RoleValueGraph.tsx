'use client';

import { useCallback, useState, useEffect, useRef } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  Node,
  Edge,
  useNodesState,
  useEdgesState,
  Handle,
  Position,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { Role, RoleWithValues, Value, STATUS_ACTIVE } from '@/lib/types';
import { api } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { useRoleValueConnection } from '@/hooks/use-role-value-connection';
import { getRoleFacetConfig, getRoleTypeConfig } from '@/lib/role-utils';
import { FACET_CONFIG } from '@/lib/value-utils';
import { useToast } from '@/hooks/use-toast';

interface RoleValueGraphProps {
  refresh: number;
  onRoleEdit?: (role: Role) => void;
}

// Custom Role Node Component
function RoleNode({ data }: { data: { role: Role; isConnectionSource: boolean } }) {
  const facetConfig = getRoleFacetConfig(data.role.facet);
  const typeConfig = getRoleTypeConfig(data.role.roleType);

  return (
    <div
      className={`px-4 py-3 rounded-lg border-2 min-w-[180px] transition-all ${
        data.isConnectionSource
          ? 'border-purple-500 bg-purple-50 shadow-lg ring-2 ring-purple-300'
          : 'border-gray-200 bg-white hover:border-purple-300 hover:shadow-md'
      }`}
    >
      <Handle
        type="source"
        position={Position.Bottom}
        id="source"
        className="!bg-purple-500 !w-3 !h-3"
      />
      <div className="flex items-center gap-2 mb-2">
        <span className="text-lg">{typeConfig.icon}</span>
        <span className="font-semibold text-gray-900 text-sm">{data.role.name}</span>
      </div>
      <div className="flex items-center gap-1">
        <Badge variant="outline" className={`text-xs ${facetConfig.color} ${facetConfig.bgColor}`}>
          {facetConfig.icon} {facetConfig.label}
        </Badge>
      </div>
      {data.isConnectionSource && (
        <p className="text-xs text-purple-600 mt-2">Haz clic en un valor para conectar</p>
      )}
    </div>
  );
}

// Custom Value Node Component
function ValueNode({
  data,
}: {
  data: { value: Value; isConnected: boolean; isConnectionMode: boolean };
}) {
  const facetConfig = FACET_CONFIG[data.value.facet];

  return (
    <div
      className={`px-4 py-3 rounded-lg border-2 min-w-[160px] transition-all cursor-pointer ${
        data.isConnected
          ? 'border-green-400 bg-green-50'
          : data.isConnectionMode
            ? 'border-gray-300 bg-gray-50 hover:border-purple-300 hover:bg-purple-50'
            : 'border-gray-200 bg-white'
      }`}
    >
      <Handle
        type="target"
        position={Position.Top}
        id="target"
        className="!bg-gray-400 !w-3 !h-3"
      />
      <div className="flex items-center gap-2 mb-1">
        <span>{facetConfig.icon}</span>
        <Badge variant="outline" className={`text-xs ${facetConfig.color} ${facetConfig.bgColor}`}>
          {facetConfig.label}
        </Badge>
      </div>
      <p className="text-sm text-gray-700 line-clamp-2">{data.value.content}</p>
      {data.isConnected && (
        <p className="text-xs text-green-600 mt-1">Conectado (clic para desconectar)</p>
      )}
    </div>
  );
}

const nodeTypes = {
  roleNode: RoleNode,
  valueNode: ValueNode,
};

export function RoleValueGraph({ refresh, onRoleEdit: _onRoleEdit }: RoleValueGraphProps) {
  const [roles, setRoles] = useState<Role[]>([]);
  const [rolesWithValues, setRolesWithValues] = useState<Record<string, string[]>>({});
  const [values, setValues] = useState<Value[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  const connectionMode = useRoleValueConnection();
  const { toast } = useToast();

  // Use requestId pattern for stale response handling (F2 compliance)
  const requestIdRef = useRef(0);

  // Fetch data
  useEffect(() => {
    const fetchData = async () => {
      const currentRequestId = ++requestIdRef.current;
      setIsLoading(true);

      try {
        // Fetch roles and values in parallel
        const [rolesRes, valuesRes] = await Promise.all([
          api.roles.getAll(),
          api.values.getAll(),
        ]);

        if (currentRequestId !== requestIdRef.current) return;

        const activeRoles = rolesRes.items.filter((r) => r.status === STATUS_ACTIVE);
        const activeValues = valuesRes.items.filter((v) => v.status === STATUS_ACTIVE);

        setRoles(activeRoles);
        setValues(activeValues);

        // Fetch value connections for each role
        const connectionsMap: Record<string, string[]> = {};
        const roleDetailsPromises = activeRoles.map(async (role) => {
          try {
            const roleWithValues: RoleWithValues = await api.roles.getById(role.id);
            connectionsMap[role.id] = roleWithValues.valueIds || [];
          } catch {
            connectionsMap[role.id] = [];
          }
        });

        await Promise.allSettled(roleDetailsPromises);

        if (currentRequestId === requestIdRef.current) {
          setRolesWithValues(connectionsMap);
        }
      } catch (err) {
        if (currentRequestId !== requestIdRef.current) return;
        console.error('Failed to fetch graph data:', err);
        toast({
          variant: 'destructive',
          title: 'Error al cargar el grafo',
          description: 'No se pudieron cargar los datos del grafo.',
        });
      } finally {
        if (currentRequestId === requestIdRef.current) {
          setIsLoading(false);
        }
      }
    };

    fetchData();
  }, [refresh, toast]);

  // Create nodes and edges when data changes
  useEffect(() => {
    if (isLoading) return;

    // Create role nodes (top row)
    const roleNodes: Node[] = roles.map((role, idx) => ({
      id: role.id,
      type: 'roleNode',
      data: {
        role,
        isConnectionSource: connectionMode.sourceRoleId === role.id,
      },
      position: { x: 100 + idx * 250, y: 50 },
      draggable: true,
    }));

    // Create value nodes (bottom row)
    const valueNodes: Node[] = values.map((value, idx) => ({
      id: value.id,
      type: 'valueNode',
      data: {
        value,
        isConnected:
          connectionMode.sourceRoleId !== null &&
          rolesWithValues[connectionMode.sourceRoleId]?.includes(value.id),
        isConnectionMode: connectionMode.isActive,
      },
      position: { x: 50 + idx * 200, y: 300 },
      draggable: true,
    }));

    setNodes([...roleNodes, ...valueNodes]);

    // Create edges from role-value connections
    const newEdges: Edge[] = Object.entries(rolesWithValues).flatMap(([roleId, valueIds]) =>
      valueIds.map((valueId) => ({
        id: `${roleId}-${valueId}`,
        source: roleId,
        target: valueId,
        sourceHandle: 'source',
        targetHandle: 'target',
        type: 'default',
        style: { stroke: '#9333ea', strokeWidth: 2 },
        animated: connectionMode.sourceRoleId === roleId,
      })),
    );

    setEdges(newEdges);
  }, [
    roles,
    values,
    rolesWithValues,
    connectionMode.sourceRoleId,
    connectionMode.isActive,
    isLoading,
    setNodes,
    setEdges,
  ]);

  // Handle node click
  const onNodeClick = useCallback(
    async (event: React.MouseEvent, node: Node) => {
      if (connectionMode.isLoading) return;

      if (connectionMode.isActive) {
        // In connection mode: toggle connection on value nodes
        if (node.type === 'valueNode' && connectionMode.sourceRoleId) {
          const isConnected = rolesWithValues[connectionMode.sourceRoleId]?.includes(node.id);
          const success = await connectionMode.toggleValueConnection(node.id, isConnected);

          if (success) {
            // Update local state
            setRolesWithValues((prev) => {
              const currentConnections = prev[connectionMode.sourceRoleId!] || [];
              const newConnections = isConnected
                ? currentConnections.filter((id) => id !== node.id)
                : [...currentConnections, node.id];
              return { ...prev, [connectionMode.sourceRoleId!]: newConnections };
            });
          }
        }
      } else {
        // Not in connection mode: enter mode on role click
        if (node.type === 'roleNode') {
          connectionMode.enterConnectionMode(node.id);
        }
      }
    },
    [connectionMode, rolesWithValues],
  );

  if (isLoading) {
    return (
      <div className="h-[600px] border rounded-lg p-8 flex items-center justify-center">
        <div className="text-center space-y-4">
          <Skeleton className="h-8 w-48 mx-auto" />
          <Skeleton className="h-4 w-32 mx-auto" />
        </div>
      </div>
    );
  }

  if (roles.length === 0 && values.length === 0) {
    return (
      <div className="h-[400px] border-2 border-dashed border-gray-200 rounded-lg flex items-center justify-center">
        <div className="text-center">
          <p className="text-muted-foreground mb-2">No hay roles ni valores para visualizar.</p>
          <p className="text-sm text-gray-500">
            Crea algunos roles y valores para ver el grafo de conexiones.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-[600px] border rounded-lg relative">
      {connectionMode.isActive && (
        <div className="absolute top-4 left-4 z-10 bg-purple-100 text-purple-900 px-4 py-2 rounded-lg flex items-center gap-2 shadow-md">
          <span className="font-medium">Modo conexión activo</span>
          <Button
            size="sm"
            variant="outline"
            onClick={connectionMode.exitConnectionMode}
            disabled={connectionMode.isLoading}
          >
            Listo (ESC)
          </Button>
        </div>
      )}

      {!connectionMode.isActive && roles.length > 0 && (
        <div className="absolute top-4 left-4 z-10 bg-gray-100 text-gray-700 px-4 py-2 rounded-lg text-sm">
          Haz clic en un rol para entrar en modo conexión
        </div>
      )}

      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.5}
        maxZoom={1.5}
      >
        <Background color="#f1f5f9" gap={16} />
        <Controls />
        <MiniMap
          nodeColor={(node) => (node.type === 'roleNode' ? '#9333ea' : '#6b7280')}
          maskColor="rgba(0, 0, 0, 0.1)"
        />
      </ReactFlow>
    </div>
  );
}
