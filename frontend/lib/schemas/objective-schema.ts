import { z } from 'zod';

export const objectiveSchema = z.discriminatedUnion('trackingType', [
  // Result tracking
  z.object({
    trackingType: z.literal('result'),
    lifeVisionId: z.string().min(1, 'Visión de vida es requerida'),
    title: z.string().min(5, 'Mínimo 5 caracteres').max(200, 'Máximo 200 caracteres'),
    targetMetric: z.number().min(1, 'La meta debe ser mayor a 0'),
  }),

  // Frequency tracking
  z
    .object({
      trackingType: z.literal('frequency'),
      lifeVisionId: z.string().min(1, 'Visión de vida es requerida'),
      title: z.string().min(5, 'Mínimo 5 caracteres').max(200, 'Máximo 200 caracteres'),
      frequencyType: z.enum(['daily', 'n_per_week', 'n_per_month']),
      frequencyCount: z.number().min(1).max(31).optional(),
      complianceTargetPct: z.number().min(1, 'Mínimo 1%').max(100, 'Máximo 100%'),
    })
    .refine(
      (data) => {
        // For n_per_week and n_per_month, frequencyCount is required
        if (data.frequencyType === 'n_per_week' || data.frequencyType === 'n_per_month') {
          return data.frequencyCount !== undefined && data.frequencyCount >= 1;
        }
        // For daily, frequencyCount is optional (can be omitted)
        return true;
      },
      {
        message: 'La frecuencia (N) es requerida para tipo semanal o mensual',
        path: ['frequencyCount'],
      },
    ),
]);

export type ObjectiveFormData = z.infer<typeof objectiveSchema>;

export const recordSchema = z.object({
  recordDate: z.string().min(1, 'La fecha es requerida'),
  status: z.enum(['completed', 'intentionally_skipped', 'skipped']),
  notes: z.string().optional(),
});

export type RecordFormData = z.infer<typeof recordSchema>;
