'use client';

import { useState } from 'react';
import { NewThink, ThinkCategory } from '@/lib/types';
import { api, APIError } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';

const categories: ThinkCategory[] = ['personal', 'work', 'ideas', 'learning', 'reflection'];

interface ThinkFormProps {
  onSuccess?: () => void;
}

export function ThinkForm({ onSuccess }: ThinkFormProps) {
  const [category, setCategory] = useState<ThinkCategory>('personal');
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!content.trim()) {
      setError('El contenido no puede estar vacío');
      return;
    }

    setIsSubmitting(true);

    try {
      const newThink: NewThink = { category, content };
      await api.thinks.create(newThink);

      // Reset form
      setContent('');
      setCategory('personal');

      // Call success callback
      if (onSuccess) {
        onSuccess();
      }
    } catch (err) {
      if (err instanceof APIError) {
        setError(`Error al crear el pensamiento: ${err.message}`);
      } else {
        setError('Ocurrió un error inesperado');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Crear Nuevo Pensamiento</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <label htmlFor="category" className="text-sm font-medium">
              Categoría
            </label>
            <Select value={category} onValueChange={(v) => setCategory(v as ThinkCategory)}>
              <SelectTrigger>
                <SelectValue placeholder="Selecciona una categoría" />
              </SelectTrigger>
              <SelectContent>
                {categories.map((cat) => {
                  const categoryLabels: Record<ThinkCategory, string> = {
                    personal: 'Personal',
                    work: 'Trabajo',
                    ideas: 'Ideas',
                    learning: 'Aprendizaje',
                    reflection: 'Reflexión',
                  };
                  return (
                    <SelectItem key={cat} value={cat}>
                      {categoryLabels[cat]}
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <label htmlFor="content" className="text-sm font-medium">
              Contenido
            </label>
            <Textarea
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="¿Qué estás pensando?"
              rows={5}
              className="resize-none"
            />
          </div>

          <Button type="submit" disabled={isSubmitting} className="w-full">
            {isSubmitting ? 'Creando...' : 'Crear Pensamiento'}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
