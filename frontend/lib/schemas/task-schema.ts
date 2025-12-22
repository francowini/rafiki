import { z } from 'zod';

export const taskSchema = z
  .object({
    objectiveId: z.string().uuid().optional().nullable(),
    title: z
      .string()
      .min(3, 'El titulo debe tener al menos 3 caracteres')
      .max(200, 'El titulo debe tener maximo 200 caracteres')
      .trim(),
    description: z
      .string()
      .max(2000, 'La descripcion debe tener maximo 2000 caracteres')
      .trim()
      .optional()
      .nullable(),
    contribution: z
      .number()
      .int('La contribucion debe ser un numero entero')
      .min(1, 'La contribucion debe ser al menos 1')
      .max(10, 'La contribucion debe ser maximo 10')
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
      message: 'La contribucion es requerida para tareas vinculadas a objetivos',
      path: ['contribution'],
    },
  );

export type TaskFormData = z.infer<typeof taskSchema>;
