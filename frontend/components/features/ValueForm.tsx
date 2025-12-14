'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/lib/api';
import { NewValue, Value, Facet, FACETS } from '@/lib/types';
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
import { getAllFacets, getFacetConfig } from '@/lib/value-utils';

const valueSchema = z.object({
  content: z
    .string()
    .min(3, 'El valor debe tener al menos 3 caracteres')
    .max(200, 'El valor debe tener menos de 200 caracteres'),
  facet: z.enum(FACETS),
  displayOrder: z.number().int().min(1).max(10),
});

type ValueFormData = z.infer<typeof valueSchema>;

interface ValueFormProps {
  value?: Value | null; // If provided, form is in edit mode
  existingValuesCount: number; // Number of existing values (for validation)
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function ValueForm({ value, existingValuesCount, onSuccess, onCancel }: ValueFormProps) {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isEditMode = !!value;

  // Calculate next display order
  const nextDisplayOrder = isEditMode
    ? (value?.displayOrder ?? existingValuesCount + 1)
    : existingValuesCount + 1;

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
    reset,
  } = useForm<ValueFormData>({
    resolver: zodResolver(valueSchema),
    defaultValues: {
      facet: 'personal_growth',
      displayOrder: nextDisplayOrder,
    },
  });

  // Reset form when value changes (edit mode)
  useEffect(() => {
    setError(null);
    if (value) {
      reset({
        content: value.content,
        facet: value.facet,
        displayOrder: value?.displayOrder ?? nextDisplayOrder,
      });
    } else {
      reset({
        content: '',
        facet: 'personal_growth',
        displayOrder: nextDisplayOrder,
      });
    }
  }, [value, reset, nextDisplayOrder]);

  const selectedFacet = watch('facet');

  const onSubmit = async (data: ValueFormData) => {
    setError(null);
    setIsSubmitting(true);

    try {
      // Check limit (only for create mode)
      if (!isEditMode && existingValuesCount >= 10) {
        setError(
          'Solo puedes tener hasta 10 valores. Por favor elimina uno antes de agregar uno nuevo.',
        );
        setIsSubmitting(false);
        return;
      }

      if (isEditMode && value) {
        // Update existing value
        await api.values.update(value.id, {
          content: data.content,
          facet: data.facet,
          displayOrder: data.displayOrder,
        });
      } else {
        // Create new value
        const newValue: NewValue = {
          content: data.content,
          facet: data.facet,
          displayOrder: data.displayOrder,
        };
        await api.values.create(newValue);
      }

      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || `Error al ${isEditMode ? 'actualizar' : 'crear'} el valor`);
    } finally {
      setIsSubmitting(false);
    }
  };

  const allFacets = getAllFacets();
  const selectedFacetConfig = selectedFacet ? getFacetConfig(selectedFacet) : null;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Privacy Message */}
      <Alert className="bg-rose-50 border-rose-200">
        <AlertDescription className="text-sm text-rose-900">
          Tus valores son privados y solo visibles para ti.
        </AlertDescription>
      </Alert>

      {/* Error Message */}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Content Field */}
      <div className="space-y-2">
        <Label htmlFor="content" className="text-base font-medium">
          Tu declaración de valor
        </Label>
        <Textarea
          id="content"
          {...register('content')}
          placeholder="Ejemplo: Vivir con integridad y autenticidad en todas mis relaciones"
          className="min-h-24 text-base"
          maxLength={200}
        />
        <p className="text-sm text-muted-foreground">
          Describe lo que más te importa (3-200 caracteres)
        </p>
        {errors.content && <p className="text-sm text-destructive">{errors.content.message}</p>}
      </div>

      {/* Facet Selection */}
      <div className="space-y-2">
        <Label htmlFor="facet" className="text-base font-medium">
          Área de vida
        </Label>
        <Select value={selectedFacet} onValueChange={(value) => setValue('facet', value as Facet)}>
          <SelectTrigger id="facet">
            <SelectValue placeholder="Selecciona un área de vida" />
          </SelectTrigger>
          <SelectContent>
            {allFacets.map((facet) => {
              const config = getFacetConfig(facet);
              return (
                <SelectItem key={facet} value={facet}>
                  <div className="flex items-center gap-2">
                    <span>{config.icon}</span>
                    <div>
                      <div className="font-medium">{config.label}</div>
                      <div className="text-xs text-muted-foreground">{config.description}</div>
                    </div>
                  </div>
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
        {selectedFacetConfig && (
          <p className="text-sm text-muted-foreground">{selectedFacetConfig.description}</p>
        )}
        {errors.facet && <p className="text-sm text-destructive">{errors.facet.message}</p>}
      </div>

      {/* Display Order Field */}
      <div className="space-y-2">
        <Label htmlFor="displayOrder" className="text-base font-medium">
          Prioridad (1 = más importante)
        </Label>
        <Select
          value={watch('displayOrder').toString()}
          onValueChange={(value) => setValue('displayOrder', parseInt(value))}
        >
          <SelectTrigger id="displayOrder">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map((num) => (
              <SelectItem key={num} value={num.toString()}>
                #{num} {num === 1 && '(Valor Fundamental)'}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-sm text-muted-foreground">
          Ordena tus valores de más a menos importante
        </p>
        {errors.displayOrder && (
          <p className="text-sm text-destructive">{errors.displayOrder.message}</p>
        )}
      </div>

      {/* Action Buttons */}
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
              ? 'Actualizar valor'
              : 'Guardar valor'}
        </Button>
      </div>
    </form>
  );
}
