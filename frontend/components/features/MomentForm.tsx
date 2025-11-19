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

const momentSchema = z.object({
  momentDate: z.string().min(1, 'Date is required'),
  situation: z.string().min(1, 'This field is required'),
  thoughts: z.string().min(1, 'This field is required'),
  physicalSymptoms: z.string().min(1, 'This field is required'),
  behavior: z.string().min(1, 'This field is required'),
  consequences: z.string().min(1, 'This field is required'),
  valuesReflection: z.string().min(1, 'This field is required'),
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
      setError(err.message || `Error ${isEditMode ? 'updating' : 'saving'} the moment`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Privacy Message */}
      <Alert className="bg-purple-50 border-purple-200">
        <AlertDescription className="text-sm text-purple-900">
          Only you can see this. Your data is safe and private.
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
          Date and time
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
        <Label htmlFor="situation" className="text-base font-medium">
          Situation (Where were you? What happened before?)
        </Label>
        <Textarea
          id="situation"
          {...register('situation')}
          placeholder="Example: I was at home, I had just eaten lunch alone..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Briefly describe where you were and what was happening before you felt bad.
        </p>
        {errors.situation && <p className="text-sm text-destructive">{errors.situation.message}</p>}
      </div>

      {/* Thoughts Field */}
      <div className="space-y-2">
        <Label htmlFor="thoughts" className="text-base font-medium">
          Thoughts that appeared (What did I think?)
        </Label>
        <Textarea
          id="thoughts"
          {...register('thoughts')}
          placeholder="Example: I'm wasting my time, I don't know what I'm going to do with my life..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          What thoughts came to your mind? What were you telling yourself?
        </p>
        {errors.thoughts && <p className="text-sm text-destructive">{errors.thoughts.message}</p>}
      </div>

      {/* Physical Symptoms Field */}
      <div className="space-y-2">
        <Label htmlFor="physicalSymptoms" className="text-base font-medium">
          Physical symptoms or emotions (What did I feel?)
        </Label>
        <Textarea
          id="physicalSymptoms"
          {...register('physicalSymptoms')}
          placeholder="Example: Heart palpitations, sweaty palms, anxiety, sadness..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          How did your body feel? What emotions did you experience?
        </p>
        {errors.physicalSymptoms && (
          <p className="text-sm text-destructive">{errors.physicalSymptoms.message}</p>
        )}
      </div>

      {/* Behavior Field */}
      <div className="space-y-2">
        <Label htmlFor="behavior" className="text-base font-medium">
          What you did (What did I do?)
        </Label>
        <Textarea
          id="behavior"
          {...register('behavior')}
          placeholder="Example: I started cleaning, went for a walk, called a friend..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          What did you do in response? Describe your actions.
        </p>
        {errors.behavior && <p className="text-sm text-destructive">{errors.behavior.message}</p>}
      </div>

      {/* Consequences Field */}
      <div className="space-y-2">
        <Label htmlFor="consequences" className="text-base font-medium">
          Immediate consequences
        </Label>
        <Textarea
          id="consequences"
          {...register('consequences')}
          placeholder="Example: I felt a bit better, but then I felt sadder..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          How did you feel right after doing that? What changed?
        </p>
        {errors.consequences && (
          <p className="text-sm text-destructive">{errors.consequences.message}</p>
        )}
      </div>

      {/* Values Reflection Field */}
      <div className="space-y-2">
        <Label htmlFor="valuesReflection" className="text-base font-medium">
          Did I avoid or approach something important to me?
        </Label>
        <Textarea
          id="valuesReflection"
          {...register('valuesReflection')}
          placeholder="Example: I avoided being with myself, I didn't progress on anything I want for myself..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Think about whether what you did brought you closer or further from something you value.
        </p>
        {errors.valuesReflection && (
          <p className="text-sm text-destructive">{errors.valuesReflection.message}</p>
        )}
      </div>

      {/* Intensity Slider */}
      <div className="space-y-4">
        <Label htmlFor="intensity" className="text-base font-medium">
          Distress intensity: {intensity}/10
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
          <span>Mild</span>
          <span>Moderate</span>
          <span>Intense</span>
        </div>
        {errors.intensity && <p className="text-sm text-destructive">{errors.intensity.message}</p>}
      </div>

      {/* Action Buttons */}
      <div className="flex gap-3 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} className="flex-1">
            Cancel
          </Button>
        )}
        <Button
          type="submit"
          disabled={isSubmitting}
          className="flex-1 bg-purple-600 hover:bg-purple-700"
        >
          {isSubmitting
            ? isEditMode
              ? 'Updating...'
              : 'Saving...'
            : isEditMode
              ? 'Update moment'
              : 'Save moment'}
        </Button>
      </div>
    </form>
  );
}
