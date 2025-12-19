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
import type { Objective, TrackingType, FrecuenciaType, LifeVision } from '@/lib/types';

interface ObjectiveFormProps {
  objective?: Objective | null;
  onSuccess: () => void;
  onCancel: () => void;
}

export function ObjectiveForm({ objective, onSuccess, onCancel }: ObjectiveFormProps) {
  const [lifeVisions, setLifeVisions] = useState<LifeVision[]>([]);
  const [loadingLifeVisions, setLoadingLifeVisions] = useState(true);
  const [trackingType, setTrackingType] = useState<TrackingType>(
    objective?.tipoTracking || 'frecuencia',
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
      tipoTracking: objective?.tipoTracking || 'frecuencia',
      lifeVisionId: objective?.lifeVisionId || '',
      titulo: objective?.titulo || '',
      metricaObjetivo: objective?.metricaObjetivo || 100,
      frecuenciaTipo: objective?.frecuenciaTipo || 'daily',
      frecuenciaN: objective?.frecuenciaN || 3,
      cumplimientoTargetPct: objective?.cumplimientoTargetPct || 80,
    },
  });

  const watchedFrecuenciaTipo = watch('frecuenciaTipo');

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
            titulo: data.titulo,
            lifeVisionId: data.lifeVisionId,
            ...(data.tipoTracking === 'resultado' && {
              metricaObjetivo: data.metricaObjetivo,
            }),
            ...(data.tipoTracking === 'frecuencia' && {
              frecuenciaTipo: data.frecuenciaTipo,
              frecuenciaN: data.frecuenciaN,
              cumplimientoTargetPct: data.cumplimientoTargetPct,
            }),
          },
        });
      } else {
        await createMutation.mutateAsync({
          titulo: data.titulo,
          lifeVisionId: data.lifeVisionId,
          tipoTracking: data.tipoTracking,
          ...(data.tipoTracking === 'resultado' && {
            metricaObjetivo: data.metricaObjetivo,
          }),
          ...(data.tipoTracking === 'frecuencia' && {
            frecuenciaTipo: data.frecuenciaTipo,
            frecuenciaN: data.frecuenciaN,
            cumplimientoTargetPct: data.cumplimientoTargetPct,
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
    setValue('tipoTracking', type);
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
              onClick={() => handleTrackingTypeChange('frecuencia')}
              className={cn(
                'flex flex-col items-center p-4 border-2 rounded-lg transition-all',
                trackingType === 'frecuencia'
                  ? 'border-purple-500 bg-purple-50'
                  : 'border-gray-200 hover:border-purple-300',
              )}
            >
              <BarChart3
                className={cn(
                  'h-8 w-8 mb-2',
                  trackingType === 'frecuencia' ? 'text-purple-600' : 'text-gray-400',
                )}
              />
              <span className="font-medium">Frecuencia</span>
              <span className="text-xs text-muted-foreground text-center mt-1">
                Hábito diario/semanal
              </span>
            </button>

            <button
              type="button"
              onClick={() => handleTrackingTypeChange('resultado')}
              className={cn(
                'flex flex-col items-center p-4 border-2 rounded-lg transition-all',
                trackingType === 'resultado'
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200 hover:border-blue-300',
              )}
            >
              <Target
                className={cn(
                  'h-8 w-8 mb-2',
                  trackingType === 'resultado' ? 'text-blue-600' : 'text-gray-400',
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
        <Label htmlFor="titulo">Título del objetivo</Label>
        <Input
          id="titulo"
          {...register('titulo')}
          placeholder={
            trackingType === 'frecuencia'
              ? 'Ej: Meditar 10 minutos cada día'
              : 'Ej: Leer 12 libros este año'
          }
        />
        {errors.titulo && <p className="text-sm text-red-500">{errors.titulo.message}</p>}
      </div>

      {/* Resultado-specific fields */}
      {trackingType === 'resultado' && (
        <div className="space-y-2">
          <Label htmlFor="metricaObjetivo" className="flex items-center gap-2">
            Meta numérica
            <HelpTooltip content="El valor que quieres alcanzar. Por ejemplo: 12 libros, 100 km, 10000 pasos." />
          </Label>
          <Input
            id="metricaObjetivo"
            type="number"
            min="1"
            {...register('metricaObjetivo', { valueAsNumber: true })}
            placeholder="Ej: 12"
          />
          {'metricaObjetivo' in errors && errors.metricaObjetivo && (
            <p className="text-sm text-red-500">{errors.metricaObjetivo.message}</p>
          )}
        </div>
      )}

      {/* Frecuencia-specific fields */}
      {trackingType === 'frecuencia' && (
        <>
          <div className="space-y-2">
            <Label htmlFor="frecuenciaTipo" className="flex items-center gap-2">
              Frecuencia
              <HelpTooltip content="Con qué frecuencia planeas realizar esta actividad." />
            </Label>
            <Select
              value={watch('frecuenciaTipo')}
              onValueChange={(value) => setValue('frecuenciaTipo', value as FrecuenciaType)}
            >
              <SelectTrigger id="frecuenciaTipo">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="daily">Diario</SelectItem>
                <SelectItem value="n_por_semana">Veces por semana</SelectItem>
                <SelectItem value="n_por_mes">Veces por mes</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {watchedFrecuenciaTipo !== 'daily' && (
            <div className="space-y-2">
              <Label htmlFor="frecuenciaN">
                {watchedFrecuenciaTipo === 'n_por_semana'
                  ? 'Veces por semana'
                  : 'Veces por mes'}
              </Label>
              <Input
                id="frecuenciaN"
                type="number"
                min="1"
                max={watchedFrecuenciaTipo === 'n_por_semana' ? 7 : 31}
                {...register('frecuenciaN', { valueAsNumber: true })}
              />
              {'frecuenciaN' in errors && errors.frecuenciaN && (
                <p className="text-sm text-red-500">{errors.frecuenciaN.message}</p>
              )}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="cumplimientoTargetPct" className="flex items-center gap-2">
              Meta de cumplimiento (%)
              <HelpTooltip content="Porcentaje de días/semanas que consideras exitoso. 80% es un buen objetivo para empezar." />
            </Label>
            <Input
              id="cumplimientoTargetPct"
              type="number"
              min="1"
              max="100"
              {...register('cumplimientoTargetPct', { valueAsNumber: true })}
            />
            {'cumplimientoTargetPct' in errors && errors.cumplimientoTargetPct && (
              <p className="text-sm text-red-500">{errors.cumplimientoTargetPct.message}</p>
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
