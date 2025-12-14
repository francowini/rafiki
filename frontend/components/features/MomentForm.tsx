'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/lib/api';
import { NewMoment, Moment } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Slider } from '@/components/ui/slider';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { HelpTooltip } from '@/components/ui/HelpTooltip';

const momentSchema = z.object({
  momentDate: z.string().min(1, 'La fecha es requerida'),
  situation: z.string().min(1, 'Este campo es requerido'),
  thoughts: z.string().min(1, 'Este campo es requerido'),
  physicalSymptoms: z.string().min(1, 'Este campo es requerido'),
  behavior: z.string().min(1, 'Este campo es requerido'),
  consequences: z.string().min(1, 'Este campo es requerido'),
  valuesReflection: z.string().min(1, 'Este campo es requerido'),
  intensity: z.number().min(0).max(10),
});

type MomentFormData = z.infer<typeof momentSchema>;

interface MomentFormProps {
  moment?: Moment | null; // If provided, form is in edit mode
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function MomentForm({ moment, onSuccess, onCancel }: MomentFormProps) {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isEditMode = !!moment;

  // Format current local time for datetime-local input
  const getLocalDateTimeString = (date?: Date | string) => {
    const d = date ? new Date(date) : new Date();
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${minutes}`;
  };

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
    reset,
  } = useForm<MomentFormData>({
    resolver: zodResolver(momentSchema),
    defaultValues: {
      momentDate: getLocalDateTimeString(),
      intensity: 5,
    },
  });

  // Reset form when moment changes (edit mode)
  useEffect(() => {
    if (moment) {
      reset({
        momentDate: getLocalDateTimeString(moment.momentDate),
        situation: moment.situation,
        thoughts: moment.thoughts,
        physicalSymptoms: moment.physicalSymptoms,
        behavior: moment.behavior,
        consequences: moment.consequences,
        valuesReflection: moment.valuesReflection,
        intensity: moment.intensity,
      });
    } else {
      reset({
        momentDate: getLocalDateTimeString(),
        intensity: 5,
        situation: '',
        thoughts: '',
        physicalSymptoms: '',
        behavior: '',
        consequences: '',
        valuesReflection: '',
      });
    }
  }, [moment, reset]);

  const intensity = watch('intensity');

  const onSubmit = async (data: MomentFormData) => {
    setError(null);
    setIsSubmitting(true);

    try {
      // Convert datetime-local to ISO 8601
      const momentDate = new Date(data.momentDate).toISOString();

      if (isEditMode && moment) {
        // Update existing moment
        await api.moments.update(moment.id, {
          ...data,
          momentDate,
        });
      } else {
        // Create new moment
        const newMoment: NewMoment = {
          ...data,
          momentDate,
        };
        await api.moments.create(newMoment);

        // Clear localStorage draft if exists
        localStorage.removeItem('moment-draft');
      }

      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || `Error al ${isEditMode ? 'actualizar' : 'guardar'} el momento`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Privacy Message */}
      <Alert className="bg-purple-50 border-purple-200">
        <AlertDescription className="text-sm text-purple-900">
          Solo tú puedes ver esto. Tus datos están seguros y son privados.
        </AlertDescription>
      </Alert>

      {/* Error Message */}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Date/Time Field */}
      <div className="space-y-2">
        <Label htmlFor="momentDate" className="text-base font-medium">
          Fecha y hora
        </Label>
        <Input
          id="momentDate"
          type="datetime-local"
          {...register('momentDate')}
          className="text-base"
        />
        {errors.momentDate && (
          <p className="text-sm text-destructive">{errors.momentDate.message}</p>
        )}
      </div>

      {/* Situation Field */}
      <div className="space-y-2">
        <Label htmlFor="situation" className="text-base font-medium flex items-center gap-2">
          Situación (¿Dónde estabas? ¿Qué pasó antes?)
          <HelpTooltip content="Describe el contexto: lugar, personas presentes, momento del día" />
        </Label>
        <Textarea
          id="situation"
          {...register('situation')}
          placeholder="Ejemplo: Estaba en casa, había almorzado solo..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Describe brevemente dónde estabas y qué estaba pasando antes de sentirte mal.
        </p>
        {errors.situation && <p className="text-sm text-destructive">{errors.situation.message}</p>}
      </div>

      {/* Thoughts Field */}
      <div className="space-y-2">
        <Label htmlFor="thoughts" className="text-base font-medium flex items-center gap-2">
          Pensamientos que aparecieron (¿Qué pensé?)
          <HelpTooltip content="Registra los pensamientos automáticos que surgieron, aunque parezcan irracionales" />
        </Label>
        <Textarea
          id="thoughts"
          {...register('thoughts')}
          placeholder="Ejemplo: Estoy perdiendo el tiempo, no sé qué voy a hacer con mi vida..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          ¿Qué pensamientos vinieron a tu mente? ¿Qué te estabas diciendo?
        </p>
        {errors.thoughts && <p className="text-sm text-destructive">{errors.thoughts.message}</p>}
      </div>

      {/* Physical Symptoms Field */}
      <div className="space-y-2">
        <Label htmlFor="physicalSymptoms" className="text-base font-medium flex items-center gap-2">
          Síntomas físicos o emociones (¿Qué sentí?)
          <HelpTooltip content="Las sensaciones físicas son señales importantes de tu estado emocional" />
        </Label>
        <Textarea
          id="physicalSymptoms"
          {...register('physicalSymptoms')}
          placeholder="Ejemplo: Palpitaciones, manos sudorosas, ansiedad, tristeza..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          ¿Cómo se sentía tu cuerpo? ¿Qué emociones experimentaste?
        </p>
        {errors.physicalSymptoms && (
          <p className="text-sm text-destructive">{errors.physicalSymptoms.message}</p>
        )}
      </div>

      {/* Behavior Field */}
      <div className="space-y-2">
        <Label htmlFor="behavior" className="text-base font-medium flex items-center gap-2">
          Lo que hiciste (¿Qué hice?)
          <HelpTooltip content="No juzgues tus acciones, solo obsérvalas y descríbelas" />
        </Label>
        <Textarea
          id="behavior"
          {...register('behavior')}
          placeholder="Ejemplo: Empecé a limpiar, salí a caminar, llamé a un amigo..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          ¿Qué hiciste en respuesta? Describe tus acciones.
        </p>
        {errors.behavior && <p className="text-sm text-destructive">{errors.behavior.message}</p>}
      </div>

      {/* Consequences Field */}
      <div className="space-y-2">
        <Label htmlFor="consequences" className="text-base font-medium flex items-center gap-2">
          Consecuencias inmediatas
          <HelpTooltip content="Las consecuencias a corto plazo pueden ser diferentes de las de largo plazo" />
        </Label>
        <Textarea
          id="consequences"
          {...register('consequences')}
          placeholder="Ejemplo: Me sentí un poco mejor, pero luego me sentí más triste..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          ¿Cómo te sentiste justo después de hacer eso? ¿Qué cambió?
        </p>
        {errors.consequences && (
          <p className="text-sm text-destructive">{errors.consequences.message}</p>
        )}
      </div>

      {/* Values Reflection Field */}
      <div className="space-y-2">
        <Label htmlFor="valuesReflection" className="text-base font-medium flex items-center gap-2">
          ¿Evité o me acerqué a algo importante para mí?
          <HelpTooltip content="Reflexiona si tu comportamiento te acercó o alejó de tus valores personales" />
        </Label>
        <Textarea
          id="valuesReflection"
          {...register('valuesReflection')}
          placeholder="Ejemplo: Evité estar conmigo mismo, no avancé en nada que quiero para mí..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Piensa si lo que hiciste te acercó o alejó de algo que valoras.
        </p>
        {errors.valuesReflection && (
          <p className="text-sm text-destructive">{errors.valuesReflection.message}</p>
        )}
      </div>

      {/* Intensity Slider */}
      <div className="space-y-4">
        <Label htmlFor="intensity" className="text-base font-medium">
          Intensidad del malestar: {intensity}/10
        </Label>
        <Slider
          id="intensity"
          min={0}
          max={10}
          step={1}
          value={[intensity]}
          onValueChange={(value) => setValue('intensity', value[0])}
          className="py-4"
        />
        <div className="flex justify-between text-sm text-muted-foreground">
          <span>Leve</span>
          <span>Moderado</span>
          <span>Intenso</span>
        </div>
        {errors.intensity && <p className="text-sm text-destructive">{errors.intensity.message}</p>}
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
          className="flex-1 bg-purple-600 hover:bg-purple-700"
        >
          {isSubmitting
            ? isEditMode
              ? 'Actualizando...'
              : 'Guardando...'
            : isEditMode
              ? 'Actualizar momento'
              : 'Guardar momento'}
        </Button>
      </div>
    </form>
  );
}
