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
    .min(10, 'Vision must be at least 10 characters')
    .max(500, 'Vision must be less than 500 characters'),
  valueId: z.string().uuid('Please select a value'),
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
          title: 'Vision updated',
          description: 'Your life vision has been successfully updated.',
        });
      } else {
        const newVision: NewLifeVision = {
          content: data.content,
          valueId: data.valueId,
        };
        await api.lifeVisions.create(newVision);
        toast({
          title: 'Vision created',
          description: 'Your life vision has been successfully created.',
        });
      }

      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || `Error ${isEditMode ? 'updating' : 'creating'} vision`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <Alert className="bg-rose-50 border-rose-200">
        <AlertDescription className="text-sm text-rose-900">
          Your life visions are private and only visible to you.
        </AlertDescription>
      </Alert>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-2">
        <Label htmlFor="valueId" className="text-base font-medium">
          Value
        </Label>
        <Select value={selectedValueId} onValueChange={(value) => setValue('valueId', value)}>
          <SelectTrigger id="valueId">
            <SelectValue placeholder="Select a value" />
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
          Your vision
        </Label>
        <Textarea
          id="content"
          {...register('content')}
          placeholder="Example: In 5 years, I see myself leading a team that builds products that help millions of people live healthier lives..."
          className="min-h-32 text-base"
          maxLength={500}
        />
        <p className="text-sm text-muted-foreground">
          Describe how you want to live this value. Be specific and paint a picture (10-500
          characters). We recommend max 2 visions per value for clarity.
        </p>
        {errors.content && <p className="text-sm text-destructive">{errors.content.message}</p>}
      </div>

      <div className="flex gap-3 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} className="flex-1">
            Cancel
          </Button>
        )}
        <Button
          type="submit"
          disabled={isSubmitting}
          className="flex-1 bg-rose-600 hover:bg-rose-700"
        >
          {isSubmitting
            ? isEditMode
              ? 'Updating...'
              : 'Saving...'
            : isEditMode
              ? 'Update vision'
              : 'Save vision'}
        </Button>
      </div>
    </form>
  );
}
