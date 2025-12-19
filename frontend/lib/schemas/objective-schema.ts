import { z } from 'zod';

export const objectiveSchema = z.discriminatedUnion('tipoTracking', [
  // Resultado tracking
  z.object({
    tipoTracking: z.literal('resultado'),
    lifeVisionId: z.string().min(1, 'Visión de vida es requerida'),
    titulo: z.string().min(5, 'Mínimo 5 caracteres').max(200, 'Máximo 200 caracteres'),
    metricaObjetivo: z.number().min(1, 'La meta debe ser mayor a 0'),
  }),

  // Frecuencia tracking
  z
    .object({
      tipoTracking: z.literal('frecuencia'),
      lifeVisionId: z.string().min(1, 'Visión de vida es requerida'),
      titulo: z.string().min(5, 'Mínimo 5 caracteres').max(200, 'Máximo 200 caracteres'),
      frecuenciaTipo: z.enum(['daily', 'n_por_semana', 'n_por_mes']),
      frecuenciaN: z.number().min(1).max(31).optional(),
      cumplimientoTargetPct: z.number().min(1, 'Mínimo 1%').max(100, 'Máximo 100%'),
    })
    .refine(
      (data) => {
        // For n_por_semana and n_por_mes, frecuenciaN is required
        if (data.frecuenciaTipo === 'n_por_semana' || data.frecuenciaTipo === 'n_por_mes') {
          return data.frecuenciaN !== undefined && data.frecuenciaN >= 1;
        }
        // For daily, frecuenciaN is optional (can be omitted)
        return true;
      },
      {
        message: 'La frecuencia (N) es requerida para tipo semanal o mensual',
        path: ['frecuenciaN'],
      },
    ),
]);

export type ObjectiveFormData = z.infer<typeof objectiveSchema>;

export const recordSchema = z.object({
  fechaRegistro: z.string().min(1, 'La fecha es requerida'),
  status: z.enum(['completado', 'omitido_intencional', 'omitido']),
  notes: z.string().optional(),
});

export type RecordFormData = z.infer<typeof recordSchema>;
