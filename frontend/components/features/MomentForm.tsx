"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { api } from "@/lib/api";
import { NewMoment } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Slider } from "@/components/ui/slider";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";

const momentSchema = z.object({
  momentDate: z.string().min(1, "La fecha es requerida"),
  situation: z.string().min(1, "Este campo es requerido"),
  thoughts: z.string().min(1, "Este campo es requerido"),
  physicalSymptoms: z.string().min(1, "Este campo es requerido"),
  behavior: z.string().min(1, "Este campo es requerido"),
  consequences: z.string().min(1, "Este campo es requerido"),
  valuesReflection: z.string().min(1, "Este campo es requerido"),
  intensity: z.number().min(0).max(10),
});

type MomentFormData = z.infer<typeof momentSchema>;

interface MomentFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function MomentForm({ onSuccess, onCancel }: MomentFormProps) {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
  } = useForm<MomentFormData>({
    resolver: zodResolver(momentSchema),
    defaultValues: {
      momentDate: new Date().toISOString().slice(0, 16), // Current datetime
      intensity: 5,
    },
  });

  const intensity = watch("intensity");

  const onSubmit = async (data: MomentFormData) => {
    setError(null);
    setIsSubmitting(true);

    try {
      // Convert datetime-local to ISO 8601
      const momentDate = new Date(data.momentDate).toISOString();

      const newMoment: NewMoment = {
        ...data,
        momentDate,
      };

      await api.moments.create(newMoment);

      // Clear localStorage draft if exists
      localStorage.removeItem("moment-draft");

      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || "Error al guardar el momento");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Privacy Message */}
      <Alert className="bg-purple-50 border-purple-200">
        <AlertDescription className="text-sm text-purple-900">
          Solo tu puedes ver esto. Tus datos estan seguros y privados.
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
          {...register("momentDate")}
          className="text-base"
        />
        {errors.momentDate && (
          <p className="text-sm text-destructive">{errors.momentDate.message}</p>
        )}
      </div>

      {/* Situation Field */}
      <div className="space-y-2">
        <Label htmlFor="situation" className="text-base font-medium">
          Situacion (Donde estabas? Que paso antes?)
        </Label>
        <Textarea
          id="situation"
          {...register("situation")}
          placeholder="Ejemplo: Estaba en casa, recien habia almorzado solo..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Describe brevemente donde estabas y que estaba pasando antes de sentirte mal.
        </p>
        {errors.situation && (
          <p className="text-sm text-destructive">{errors.situation.message}</p>
        )}
      </div>

      {/* Thoughts Field */}
      <div className="space-y-2">
        <Label htmlFor="thoughts" className="text-base font-medium">
          Pensamientos que aparecieron (Que pense?)
        </Label>
        <Textarea
          id="thoughts"
          {...register("thoughts")}
          placeholder="Ejemplo: Estoy perdiendo el tiempo, no se que voy a hacer con mi vida..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Que pensamientos vinieron a tu mente? Que te estabas diciendo a ti mismo/a?
        </p>
        {errors.thoughts && (
          <p className="text-sm text-destructive">{errors.thoughts.message}</p>
        )}
      </div>

      {/* Physical Symptoms Field */}
      <div className="space-y-2">
        <Label htmlFor="physicalSymptoms" className="text-base font-medium">
          Sintomas fisicos o emociones (Que senti?)
        </Label>
        <Textarea
          id="physicalSymptoms"
          {...register("physicalSymptoms")}
          placeholder="Ejemplo: Palpitaciones, sudor en las manos, ansiedad, tristeza..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Como se sintio tu cuerpo? Que emociones experimentaste?
        </p>
        {errors.physicalSymptoms && (
          <p className="text-sm text-destructive">{errors.physicalSymptoms.message}</p>
        )}
      </div>

      {/* Behavior Field */}
      <div className="space-y-2">
        <Label htmlFor="behavior" className="text-base font-medium">
          Lo que hiciste (Que hago?)
        </Label>
        <Textarea
          id="behavior"
          {...register("behavior")}
          placeholder="Ejemplo: Me puse a limpiar, fui a caminar, llame a un amigo..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Que hiciste en respuesta? Describe tus acciones.
        </p>
        {errors.behavior && (
          <p className="text-sm text-destructive">{errors.behavior.message}</p>
        )}
      </div>

      {/* Consequences Field */}
      <div className="space-y-2">
        <Label htmlFor="consequences" className="text-base font-medium">
          Consecuencias inmediatas
        </Label>
        <Textarea
          id="consequences"
          {...register("consequences")}
          placeholder="Ejemplo: Me senti un poco mejor, pero despues me senti mas triste..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Como te sentiste justo despues de hacer eso? Que cambio?
        </p>
        {errors.consequences && (
          <p className="text-sm text-destructive">{errors.consequences.message}</p>
        )}
      </div>

      {/* Values Reflection Field */}
      <div className="space-y-2">
        <Label htmlFor="valuesReflection" className="text-base font-medium">
          Evite o me acerque a algo importante para mi?
        </Label>
        <Textarea
          id="valuesReflection"
          {...register("valuesReflection")}
          placeholder="Ejemplo: Evite estar conmigo mismo, no avance con nada que quiero para mi..."
          className="min-h-24 text-base"
        />
        <p className="text-sm text-muted-foreground">
          Piensa si lo que hiciste te acerco o te alejo de algo que valoras.
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
          onValueChange={(value) => setValue("intensity", value[0])}
          className="py-4"
        />
        <div className="flex justify-between text-sm text-muted-foreground">
          <span>Leve</span>
          <span>Moderado</span>
          <span>Intenso</span>
        </div>
        {errors.intensity && (
          <p className="text-sm text-destructive">{errors.intensity.message}</p>
        )}
      </div>

      {/* Action Buttons */}
      <div className="flex gap-3 pt-4">
        {onCancel && (
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            className="flex-1"
          >
            Cancelar
          </Button>
        )}
        <Button
          type="submit"
          disabled={isSubmitting}
          className="flex-1 bg-purple-600 hover:bg-purple-700"
        >
          {isSubmitting ? "Guardando..." : "Guardar momento"}
        </Button>
      </div>
    </form>
  );
}
