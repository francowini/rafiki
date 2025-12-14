import { Facet, FACETS } from './types';

export interface FacetConfig {
  label: string;
  description: string;
  color: string;
  bgColor: string;
  borderColor: string;
  icon: string;
}

/**
 * Configuration for each life facet
 */
export const FACET_CONFIG: Record<Facet, FacetConfig> = {
  health: {
    label: 'Salud',
    description: 'Bienestar físico y mental',
    color: 'text-emerald-700',
    bgColor: 'bg-emerald-100',
    borderColor: 'border-emerald-300',
    icon: '🏃',
  },
  relationships: {
    label: 'Relaciones',
    description: 'Familia, amigos, conexiones',
    color: 'text-blue-700',
    bgColor: 'bg-blue-100',
    borderColor: 'border-blue-300',
    icon: '🤝',
  },
  career: {
    label: 'Carrera',
    description: 'Desarrollo profesional, trabajo',
    color: 'text-amber-700',
    bgColor: 'bg-amber-100',
    borderColor: 'border-amber-300',
    icon: '💼',
  },
  personal_growth: {
    label: 'Crecimiento Personal',
    description: 'Automejora, aprendizaje',
    color: 'text-purple-700',
    bgColor: 'bg-purple-100',
    borderColor: 'border-purple-300',
    icon: '🌱',
  },
  family: {
    label: 'Familia',
    description: 'Relaciones familiares',
    color: 'text-pink-700',
    bgColor: 'bg-pink-100',
    borderColor: 'border-pink-300',
    icon: '👨‍👩‍👧‍👦',
  },
  creativity: {
    label: 'Creatividad',
    description: 'Expresión creativa, artes',
    color: 'text-orange-700',
    bgColor: 'bg-orange-100',
    borderColor: 'border-orange-300',
    icon: '🎨',
  },
  community: {
    label: 'Comunidad',
    description: 'Servicio, contribución, ayudar a otros',
    color: 'text-green-700',
    bgColor: 'bg-green-100',
    borderColor: 'border-green-300',
    icon: '🤲',
  },
  spirituality: {
    label: 'Espiritualidad',
    description: 'Significado, propósito, trascendencia',
    color: 'text-indigo-700',
    bgColor: 'bg-indigo-100',
    borderColor: 'border-indigo-300',
    icon: '🙏',
  },
};

/**
 * Get facet configuration
 */
export function getFacetConfig(facet: Facet): FacetConfig {
  return FACET_CONFIG[facet];
}

/**
 * Get all facets as array
 */
export function getAllFacets(): readonly Facet[] {
  return FACETS;
}

/**
 * Format facet for display (e.g., "personal_growth" -> "Personal Growth")
 */
export function formatFacetLabel(facet: Facet): string {
  return FACET_CONFIG[facet].label;
}
