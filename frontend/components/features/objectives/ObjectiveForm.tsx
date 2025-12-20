'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { HelpTooltip } from '@/components/ui/HelpTooltip';
import { Loader2, Target, BarChart3 } from 'lucide-react';
import { useCreateObjective, useUpdateObjective } from '@/lib/hooks/use-objectives';
import { objectiveSchema, ObjectiveFormData } from '@/lib/schemas/objective-schema';
import { api } from '@/lib/api';
import { cn } from '@/lib/utils';
import type { Objective, TrackingType, FrequencyType, LifeVision } from '@/lib/types';

interface ObjectiveFormProps {
  objective?: Objective | null;
  onSuccess: () => void;
  onCancel: () => void;
}

export function ObjectiveForm({ objective, onSuccess, onCancel }: ObjectiveFormProps) {
  const [lifeVisions, setLifeVisions] = useState<LifeVision[]>([]);
  const [loadingLifeVisions, setLoadingLifeVisions] = useState(true);
  const [trackingType, setTrackingType] = useState<TrackingType>(
    objective?.trackingType || 'frequency',
  );

  const isEditing = !!objective;

  const createMutation = useCreateObjective();
  const updateMutation = useUpdateObjective();

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
    reset,
  } = useForm<ObjectiveFormData>({
    resolver: zodResolver(objectiveSchema),
    defaultValues: {
      trackingType: objective?.trackingType || 'frequency',
      lifeVisionId: objective?.lifeVisionId || '',
      title: objective?.title || '',
      targetMetric: objective?.targetMetric || 100,
      frequencyType: objective?.frequencyType || 'daily',
      frequencyCount: objective?.frequencyCount || 3,
      complianceTargetPct: objective?.complianceTargetPct || 80,
    },
  });

  const watchedFrequencyType = watch('frequencyType');

  useEffect(() => {
    async function loadLifeVisions() {
      try {
        const response = await api.lifeVisions.getAll();
        setLifeVisions(response.items);
      } catch (error) {
        console.error('Failed to load life visions:', error);
      } finally {
        setLoadingLifeVisions(false);
      }
    }
    loadLifeVisions();
  }, []);

  const onSubmit = async (data: ObjectiveFormData) => {
    try {
      if (isEditing && objective) {
        await updateMutation.mutateAsync({
          id: objective.id,
          data: {
            title: data.title,
            lifeVisionId: data.lifeVisionId,
            ...(data.trackingType === 'result' && {
              targetMetric: data.targetMetric,
            }),
            ...(data.trackingType === 'frequency' && {
              frequencyType: data.frequencyType,
              frequencyCount: data.frequencyCount,
              complianceTargetPct: data.complianceTargetPct,
            }),
          },
        });
      } else {
        await createMutation.mutateAsync({
          title: data.title,
          lifeVisionId: data.lifeVisionId,
          trackingType: data.trackingType,
          ...(data.trackingType === 'result' && {
            targetMetric: data.targetMetric,
          }),
          ...(data.trackingType === 'frequency' && {
            frequencyType: data.frequencyType,
            frequencyCount: data.frequencyCount,
            complianceTargetPct: data.complianceTargetPct,
          }),
        });
      }
      reset();
      onSuccess();
    } catch (error) {
      console.error('Failed to save objective:', error);
    }
  };

  const handleTrackingTypeChange = (type: TrackingType) => {
    setTrackingType(type);
    setValue('trackingType', type);
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Tracking Type Selector - only show when creating */}
      {!isEditing && (
        <div className="space-y-3">
          <Label className="flex items-center gap-2">
            Tipo de objetivo
            <HelpTooltip content="Resultado: progreso hacia una meta numérica. Frecuencia: hábito que se repite regularmente." />
          </Label>
          <div className="grid grid-cols-2 gap-3">
            <button
              type="button"
              onClick={() => handleTrackingTypeChange('frequency')}
              className={cn(
                'flex flex-col items-center p-4 border-2 rounded-lg transition-all',
                trackingType === 'frequency'
                  ? 'border-purple-500 bg-purple-50'
                  : 'border-gray-200 hover:border-purple-300',
              )}
            >
              <BarChart3
                className={cn(
                  'h-8 w-8 mb-2',
                  trackingType === 'frequency' ? 'text-purple-600' : 'text-gray-400',
                )}
              />
              <span className="font-medium">Frecuencia</span>
              <span className="text-xs text-muted-foreground text-center mt-1">
                Hábito diario/semanal
              </span>
            </button>

            <button
              type="button"
              onClick={() => handleTrackingTypeChange('result')}
              className={cn(
                'flex flex-col items-center p-4 border-2 rounded-lg transition-all',
                trackingType === 'result'
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200 hover:border-blue-300',
              )}
            >
              <Target
                className={cn(
                  'h-8 w-8 mb-2',
                  trackingType === 'result' ? 'text-blue-600' : 'text-gray-400',
                )}
              />
              <span className="font-medium">Resultado</span>
              <span className="text-xs text-muted-foreground text-center mt-1">
                Meta numérica medible
              </span>
            </button>
          </div>
        </div>
      )}

      {/* Life Vision Selector */}
      <div className="space-y-2">
        <Label htmlFor="lifeVisionId" className="flex items-center gap-2">
          Visión de vida
          <HelpTooltip content="Selecciona la visión de vida a la que este objetivo contribuye." />
        </Label>
        {loadingLifeVisions ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Cargando visiones...
          </div>
        ) : (
          <Select
            value={watch('lifeVisionId')}
            onValueChange={(value) => setValue('lifeVisionId', value)}
          >
            <SelectTrigger id="lifeVisionId">
              <SelectValue placeholder="Selecciona una visión de vida" />
            </SelectTrigger>
            <SelectContent>
              {lifeVisions.map((lv) => (
                <SelectItem key={lv.id} value={lv.id}>
                  {lv.content.length > 60 ? `${lv.content.substring(0, 60)}...` : lv.content}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
        {errors.lifeVisionId && (
          <p className="text-sm text-red-500">{errors.lifeVisionId.message}</p>
        )}
      </div>

      {/* Title */}
      <div className="space-y-2">
        <Label htmlFor="title">Título del objetivo</Label>
        <Input
          id="title"
          {...register('title')}
          placeholder={
            trackingType === 'frequency'
              ? 'Ej: Meditar 10 minutos cada día'
              : 'Ej: Leer 12 libros este año'
          }
        />
        {errors.title && <p className="text-sm text-red-500">{errors.title.message}</p>}
      </div>

      {/* Result-specific fields */}
      {trackingType === 'result' && (
        <div className="space-y-2">
          <Label htmlFor="targetMetric" className="flex items-center gap-2">
            Meta numérica
            <HelpTooltip content="El valor que quieres alcanzar. Por ejemplo: 12 libros, 100 km, 10000 pasos." />
          </Label>
          <Input
            id="targetMetric"
            type="number"
            min="1"
            {...register('targetMetric', { valueAsNumber: true })}
            placeholder="Ej: 12"
          />
          {'targetMetric' in errors && errors.targetMetric && (
            <p className="text-sm text-red-500">{errors.targetMetric.message}</p>
          )}
        </div>
      )}

      {/* Frequency-specific fields */}
      {trackingType === 'frequency' && (
        <>
          <div className="space-y-2">
            <Label htmlFor="frequencyType" className="flex items-center gap-2">
              Frecuencia
              <HelpTooltip content="Con qué frecuencia planeas realizar esta actividad." />
            </Label>
            <Select
              value={watch('frequencyType')}
              onValueChange={(value) => setValue('frequencyType', value as FrequencyType)}
            >
              <SelectTrigger id="frequencyType">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="daily">Diario</SelectItem>
                <SelectItem value="n_per_week">Veces por semana</SelectItem>
                <SelectItem value="n_per_month">Veces por mes</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {watchedFrequencyType !== 'daily' && (
            <div className="space-y-2">
              <Label htmlFor="frequencyCount">
                {watchedFrequencyType === 'n_per_week'
                  ? 'Veces por semana'
                  : 'Veces por mes'}
              </Label>
              <Input
                id="frequencyCount"
                type="number"
                min="1"
                max={watchedFrequencyType === 'n_per_week' ? 7 : 31}
                {...register('frequencyCount', { valueAsNumber: true })}
              />
              {'frequencyCount' in errors && errors.frequencyCount && (
                <p className="text-sm text-red-500">{errors.frequencyCount.message}</p>
              )}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="complianceTargetPct" className="flex items-center gap-2">
              Meta de cumplimiento (%)
              <HelpTooltip content="Porcentaje de días/semanas que consideras exitoso. 80% es un buen objetivo para empezar." />
            </Label>
            <Input
              id="complianceTargetPct"
              type="number"
              min="1"
              max="100"
              {...register('complianceTargetPct', { valueAsNumber: true })}
            />
            {'complianceTargetPct' in errors && errors.complianceTargetPct && (
              <p className="text-sm text-red-500">{errors.complianceTargetPct.message}</p>
            )}
          </div>
        </>
      )}

      {/* Actions */}
      <div className="flex gap-3 pt-4">
        <Button type="button" variant="outline" onClick={onCancel} className="flex-1">
          Cancelar
        </Button>
        <Button type="submit" disabled={isPending} className="flex-1">
          {isPending ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              {isEditing ? 'Guardando...' : 'Creando...'}
            </>
          ) : isEditing ? (
            'Guardar cambios'
          ) : (
            'Crear objetivo'
          )}
        </Button>
      </div>
    </form>
  );
}
