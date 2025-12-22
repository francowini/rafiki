import { z } from 'zod';

export const taskSchema = z
  .object({
    objectiveId: z.string().uuid().optional().nullable(),
    title: z
      .string()
      .min(3, 'El título debe tener al menos 3 caracteres')
      .max(200, 'El título debe tener máximo 200 caracteres')
      .trim(),
    description: z
      .string()
      .max(2000, 'La descripción debe tener máximo 2000 caracteres')
      .trim()
      .optional()
      .nullable(),
    contribution: z
      .number()
      .int('La contribución debe ser un número entero')
      .min(1, 'La contribución debe ser al menos 1')
      .max(10, 'La contribución debe ser máximo 10')
      .optional()
      .nullable(),
  })
  .refine(
    (data) => {
      if (data.objectiveId && !data.contribution) {
        return false;
      }
      return true;
    },
    {
      message: 'La contribución es requerida para tareas vinculadas a objetivos',
      path: ['contribution'],
    },
  );

export type TaskFormData = z.infer<typeof taskSchema>;
