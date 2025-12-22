import { z } from 'zod';

export const quickTaskSchema = z.object({
  title: z
    .string()
    .min(3, 'El titulo debe tener al menos 3 caracteres')
    .max(200, 'El titulo debe tener maximo 200 caracteres')
    .trim(),
  objectiveId: z.string().uuid().optional().nullable(),
});

export type QuickTaskFormData = z.infer<typeof quickTaskSchema>;
