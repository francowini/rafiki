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
      .min(0, 'La contribución debe ser 0 o mayor')
      .max(10, 'La contribución debe ser máximo 10')
      .optional()
      .nullable(),
  })
  .refine(
    (data) => {
      // Contribution is optional for linked tasks (can be 0 or null)
      // Only require it to be defined if objectiveId is set
      if (data.objectiveId && data.contribution === undefined) {
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
