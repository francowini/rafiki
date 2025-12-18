'use client';

import { useState, useCallback, useEffect } from 'react';
import { api } from '@/lib/api';
import { useToast } from '@/hooks/use-toast';

interface ConnectionModeState {
  isActive: boolean;
  sourceRoleId: string | null;
  isLoading: boolean;
}

export function useRoleValueConnection() {
  const [state, setState] = useState<ConnectionModeState>({
    isActive: false,
    sourceRoleId: null,
    isLoading: false,
  });
  const { toast } = useToast();

  // Enter connection mode
  const enterConnectionMode = useCallback((roleId: string) => {
    setState({
      isActive: true,
      sourceRoleId: roleId,
      isLoading: false,
    });
  }, []);

  // Exit connection mode
  const exitConnectionMode = useCallback(() => {
    setState({
      isActive: false,
      sourceRoleId: null,
      isLoading: false,
    });
  }, []);

  // Toggle value connection (immediate save with proper error handling - F11)
  const toggleValueConnection = useCallback(
    async (valueId: string, isCurrentlyConnected: boolean) => {
      if (!state.sourceRoleId) return;

      setState((prev) => ({ ...prev, isLoading: true }));

      try {
        if (isCurrentlyConnected) {
          await api.roles.disconnectValue(state.sourceRoleId, valueId);
        } else {
          await api.roles.connectValue(state.sourceRoleId, valueId);
        }

        toast({
          title: isCurrentlyConnected ? 'Desconectado' : 'Conectado',
          description: 'La conexión se guardó correctamente.',
        });

        setState((prev) => ({ ...prev, isLoading: false }));
        return true;
      } catch (err) {
        console.error('Failed to toggle connection:', err);
        toast({
          variant: 'destructive',
          title: 'Error al guardar conexión',
          description: 'No se pudo guardar la conexión. Intenta de nuevo.',
        });

        // Clean up state on error (F11)
        setState((prev) => ({ ...prev, isLoading: false }));
        return false;
      }
    },
    [state.sourceRoleId, toast],
  );

  // ESC key to exit
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && state.isActive) {
        exitConnectionMode();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [state.isActive, exitConnectionMode]);

  return {
    ...state,
    enterConnectionMode,
    exitConnectionMode,
    toggleValueConnection,
  };
}
