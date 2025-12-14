'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/lib/api';
import { LifeVision, NewLifeVision, Value } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getFacetConfig } from '@/lib/value-utils';
import { useToast } from '@/hooks/use-toast';

const lifeVisionSchema = z.object({
  content: z
    .string()
    .min(10, 'La visión debe tener al menos 10 caracteres')
    .max(500, 'La visión debe tener menos de 500 caracteres'),
  valueId: z.string().uuid('Por favor selecciona un valor'),
});

type LifeVisionFormData = z.infer<typeof lifeVisionSchema>;

interface LifeVisionFormProps {
  lifeVision?: LifeVision | null;
  preselectedValueId?: string;
  values: Value[];
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function LifeVisionForm({
  lifeVision,
  preselectedValueId,
  values,
  onSuccess,
  onCancel,
}: LifeVisionFormProps) {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isEditMode = !!lifeVision;
  const { toast } = useToast();

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
    reset,
  } = useForm<LifeVisionFormData>({
    resolver: zodResolver(lifeVisionSchema),
    defaultValues: {
      content: '',
      valueId: preselectedValueId || '',
    },
  });

  useEffect(() => {
    setError(null);
    if (lifeVision) {
      reset({
        content: lifeVision.content,
        valueId: lifeVision.valueId,
      });
    } else {
      reset({
        content: '',
        valueId: preselectedValueId || '',
      });
    }
  }, [lifeVision, preselectedValueId, reset]);

  const selectedValueId = watch('valueId');
  const selectedValue = values.find((v) => v.id === selectedValueId);

  const onSubmit = async (data: LifeVisionFormData) => {
    setError(null);
    setIsSubmitting(true);

    try {
      if (isEditMode && lifeVision) {
        await api.lifeVisions.update(lifeVision.id, {
          content: data.content,
          valueId: data.valueId,
        });
        toast({
          title: 'Visión actualizada',
          description: 'Tu visión de vida ha sido actualizada exitosamente.',
        });
      } else {
        const newVision: NewLifeVision = {
          content: data.content,
          valueId: data.valueId,
        };
        await api.lifeVisions.create(newVision);
        toast({
          title: 'Visión creada',
          description: 'Tu visión de vida ha sido creada exitosamente.',
        });
      }

      if (onSuccess) {
        onSuccess();
      }
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : `Error al ${isEditMode ? 'actualizar' : 'crear'} la visión`;
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <Alert className="bg-gray-50 border-gray-200">
        <AlertDescription className="text-sm text-gray-700">
          Tus visiones de vida son privadas y solo visibles para ti.
        </AlertDescription>
      </Alert>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-2">
        <Label htmlFor="valueId" className="text-base font-medium">
          Valor
        </Label>
        <Select value={selectedValueId} onValueChange={(value) => setValue('valueId', value)}>
          <SelectTrigger id="valueId">
            <SelectValue placeholder="Selecciona un valor" />
          </SelectTrigger>
          <SelectContent>
            {values.map((value) => {
              const facetConfig = getFacetConfig(value.facet);
              return (
                <SelectItem key={value.id} value={value.id}>
                  <div className="flex items-center gap-2">
                    <span className="text-gray-500">#{value.displayOrder}</span>
                    <span>{facetConfig.icon}</span>
                    <span className="truncate max-w-[300px]">{value.content}</span>
                  </div>
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
        {selectedValue && (
          <p className="text-sm text-muted-foreground">
            {getFacetConfig(selectedValue.facet).label}
          </p>
        )}
        {errors.valueId && <p className="text-sm text-destructive">{errors.valueId.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="content" className="text-base font-medium">
          Tu visión
        </Label>
        <Textarea
          id="content"
          {...register('content')}
          placeholder="Ejemplo: En 5 años, me veo liderando un equipo que construye productos que ayudan a millones de personas a vivir vidas más saludables..."
          className="min-h-32 text-base"
          maxLength={500}
        />
        <p className="text-sm text-muted-foreground">
          Describe cómo quieres vivir este valor. Sé específico y pinta una imagen (10-500
          caracteres). Recomendamos máximo 2 visiones por valor para mayor claridad.
        </p>
        {errors.content && <p className="text-sm text-destructive">{errors.content.message}</p>}
      </div>

      <div className="flex gap-3 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} className="flex-1">
            Cancelar
          </Button>
        )}
        <Button
          type="submit"
          disabled={isSubmitting}
          className="flex-1 bg-rose-600 hover:bg-rose-700"
        >
          {isSubmitting
            ? isEditMode
              ? 'Actualizando...'
              : 'Guardando...'
            : isEditMode
              ? 'Actualizar visión'
              : 'Guardar visión'}
        </Button>
      </div>
    </form>
  );
}
