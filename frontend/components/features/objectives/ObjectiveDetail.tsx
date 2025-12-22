'use client';

import { useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ObjectiveHeatmap } from './ObjectiveHeatmap';
import { FrequencyLogDialog } from './FrequencyLogDialog';
import { TasksSection } from '@/components/features/tasks';
import {
  useObjective,
  useObjectiveRecords,
  useUpdateObjectiveStatus,
  useIncrementProgress,
} from '@/lib/hooks/use-objectives';
import {
  Target,
  BarChart3,
  Calendar,
  TrendingUp,
  CheckCircle2,
  MinusCircle,
  XCircle,
  Plus,
  Minus,
} from 'lucide-react';
import { OBJECTIVE_STATUS_LABELS, RECORD_STATUS_LABELS } from '@/lib/types';
import type { ObjectiveStatus } from '@/lib/types';

interface ObjectiveDetailProps {
  objectiveId: string;
}

export function ObjectiveDetail({ objectiveId }: ObjectiveDetailProps) {
  const currentYear = new Date().getFullYear();
  const [logDialogOpen, setLogDialogOpen] = useState(false);

  const { data: objective, isLoading } = useObjective(objectiveId);
  const { data: records } = useObjectiveRecords(objectiveId);
  const updateStatusMutation = useUpdateObjectiveStatus();
  const incrementProgressMutation = useIncrementProgress();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (!objective) {
    return (
      <div className="text-center py-12">
        <Target className="h-12 w-12 text-gray-400 mx-auto mb-4" />
        <p className="text-muted-foreground">Objetivo no encontrado</p>
      </div>
    );
  }

  const isResult = objective.trackingType === 'result';
  const isTerminal = objective.status === 'completed' || objective.status === 'abandoned';
  const progress =
    isResult && objective.currentMetric && objective.targetMetric
      ? Math.round((objective.currentMetric / objective.targetMetric) * 100)
      : 0;

  const handleStatusChange = (newStatus: ObjectiveStatus) => {
    // Don't allow changes for terminal states
    if (isTerminal) return;
    updateStatusMutation.mutate({ id: objectiveId, status: newStatus });
  };

  const handleIncrementProgress = (increment: number) => {
    incrementProgressMutation.mutate({ id: objectiveId, increment });
  };

  const getRecordIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle2 className="h-4 w-4 text-green-600" />;
      case 'intentionally_skipped':
        return <MinusCircle className="h-4 w-4 text-blue-600" />;
      case 'skipped':
        return <XCircle className="h-4 w-4 text-gray-400" />;
      default:
        return null;
    }
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="pr-8">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <h1 className="text-2xl font-bold mb-2">{objective.title}</h1>
            <div className="flex items-center gap-2">
              <Badge
                variant="outline"
                className={
                  isResult
                    ? 'bg-blue-100 text-blue-700'
                    : 'bg-purple-100 text-purple-700'
                }
              >
                {isResult ? (
                  <>
                    <Target className="h-3 w-3 mr-1" />
                    Resultado
                  </>
                ) : (
                  <>
                    <BarChart3 className="h-3 w-3 mr-1" />
                    Frecuencia
                  </>
                )}
              </Badge>
            </div>
          </div>

          {/* Status selector */}
          <Select
            value={objective.status}
            onValueChange={handleStatusChange}
            disabled={isTerminal}
          >
            <SelectTrigger
              className="w-36"
              title={isTerminal ? 'El estado no puede cambiarse (estado final)' : undefined}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(OBJECTIVE_STATUS_LABELS).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Progress section for Result */}
      {isResult && (
        <div className="bg-blue-50 rounded-lg p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold text-lg flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-blue-600" />
              Progreso
            </h2>
            <div className="flex items-center gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleIncrementProgress(-1)}
                disabled={incrementProgressMutation.isPending || (objective.currentMetric || 0) <= 0}
              >
                <Minus className="h-4 w-4" />
              </Button>
              <span className="font-bold text-xl min-w-[80px] text-center">
                {objective.currentMetric || 0} / {objective.targetMetric}
              </span>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleIncrementProgress(1)}
                disabled={incrementProgressMutation.isPending}
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <Progress value={progress} className="h-3" />
          <p className="text-sm text-muted-foreground text-center">{progress}% completado</p>
        </div>
      )}

      {/* Tasks section - available for both Result and Frequency objectives */}
      <TasksSection
        objectiveId={objectiveId}
        objectiveName={objective.title}
        targetMetric={objective.targetMetric}
        trackingType={objective.trackingType}
      />

      {/* Log record button for Frequency */}
      {!isResult && (
        <div className="bg-purple-50 rounded-lg p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-lg flex items-center gap-2">
              <Calendar className="h-5 w-5 text-purple-600" />
              Registro diario
            </h2>
            <Button onClick={() => setLogDialogOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Registrar hoy
            </Button>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold text-purple-600">
              {objective.complianceTargetPct}%
            </p>
            <p className="text-sm text-muted-foreground">Meta de cumplimiento</p>
          </div>
        </div>
      )}

      {/* Activity Heatmap for Frequency */}
      {!isResult && (
        <div className="space-y-4">
          <h2 className="font-semibold text-lg flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            Actividad {currentYear}
          </h2>
          <ObjectiveHeatmap objectiveId={objectiveId} year={currentYear} />
        </div>
      )}

      {/* Recent records for Frequency */}
      {!isResult && records && records.items.length > 0 && (
        <div className="space-y-4">
          <h2 className="font-semibold text-lg">Registros recientes</h2>
          <div className="space-y-2">
            {records.items.slice(0, 10).map((record) => (
              <div
                key={record.id}
                className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
              >
                <div className="flex items-center gap-3">
                  {getRecordIcon(record.status)}
                  <span className="text-sm font-medium">
                    {RECORD_STATUS_LABELS[record.status]}
                  </span>
                </div>
                <span className="text-sm text-muted-foreground">{record.recordDate}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Metadata */}
      <div className="text-sm text-muted-foreground border-t pt-4">
        <p>Creado: {new Date(objective.dateCreated).toLocaleDateString('es-ES')}</p>
        <p>Actualizado: {new Date(objective.dateUpdated).toLocaleDateString('es-ES')}</p>
      </div>

      {/* Log Dialog */}
      {!isResult && (
        <FrequencyLogDialog
          objectiveId={objectiveId}
          objectiveName={objective.title}
          open={logDialogOpen}
          onOpenChange={setLogDialogOpen}
        />
      )}
    </div>
  );
}
