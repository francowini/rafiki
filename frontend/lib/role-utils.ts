import { RoleFacet, RoleType, ROLE_FACETS } from './types';

export interface RoleFacetConfig {
  label: string;
  icon: string;
  color: string;
  bgColor: string;
  borderColor: string;
  description: string;
}

export interface RoleTypeConfig {
  label: string;
  icon: string;
  color: string;
  bgColor: string;
  borderColor: string;
}

/**
 * Configuration for each role facet
 */
export const ROLE_FACET_CONFIG: Record<RoleFacet, RoleFacetConfig> = {
  family_relationships: {
    label: 'Familia',
    icon: '👨‍👩‍👧‍👦',
    color: 'text-pink-600',
    bgColor: 'bg-pink-100',
    borderColor: 'border-pink-300',
    description: 'Roles familiares y de crianza',
  },
  friendships_social: {
    label: 'Amistades',
    icon: '🤝',
    color: 'text-blue-600',
    bgColor: 'bg-blue-100',
    borderColor: 'border-blue-300',
    description: 'Relaciones sociales y amistades',
  },
  romantic_relationships: {
    label: 'Pareja',
    icon: '❤️',
    color: 'text-rose-600',
    bgColor: 'bg-rose-100',
    borderColor: 'border-rose-300',
    description: 'Relaciones románticas',
  },
  work_career: {
    label: 'Carrera',
    icon: '💼',
    color: 'text-amber-600',
    bgColor: 'bg-amber-100',
    borderColor: 'border-amber-300',
    description: 'Trabajo y carrera profesional',
  },
  education_growth: {
    label: 'Aprendizaje',
    icon: '📚',
    color: 'text-purple-600',
    bgColor: 'bg-purple-100',
    borderColor: 'border-purple-300',
    description: 'Educación y crecimiento',
  },
  physical_health: {
    label: 'Salud Física',
    icon: '🏃',
    color: 'text-emerald-600',
    bgColor: 'bg-emerald-100',
    borderColor: 'border-emerald-300',
    description: 'Ejercicio y bienestar físico',
  },
  mental_wellbeing: {
    label: 'Bienestar Mental',
    icon: '🧠',
    color: 'text-teal-600',
    bgColor: 'bg-teal-100',
    borderColor: 'border-teal-300',
    description: 'Salud mental y emocional',
  },
  creative_expression: {
    label: 'Creatividad',
    icon: '🎨',
    color: 'text-orange-600',
    bgColor: 'bg-orange-100',
    borderColor: 'border-orange-300',
    description: 'Expresión artística y creativa',
  },
  leadership_mentorship: {
    label: 'Liderazgo',
    icon: '🌟',
    color: 'text-violet-600',
    bgColor: 'bg-violet-100',
    borderColor: 'border-violet-300',
    description: 'Liderazgo y mentoría',
  },
  spirituality_meaning: {
    label: 'Espiritualidad',
    icon: '🙏',
    color: 'text-indigo-600',
    bgColor: 'bg-indigo-100',
    borderColor: 'border-indigo-300',
    description: 'Vida espiritual y significado',
  },
  community_citizenship: {
    label: 'Comunidad',
    icon: '🤲',
    color: 'text-green-600',
    bgColor: 'bg-green-100',
    borderColor: 'border-green-300',
    description: 'Servicio comunitario',
  },
  financial_stewardship: {
    label: 'Finanzas',
    icon: '💰',
    color: 'text-yellow-600',
    bgColor: 'bg-yellow-100',
    borderColor: 'border-yellow-300',
    description: 'Gestión financiera',
  },
  environmental_nature: {
    label: 'Naturaleza',
    icon: '🌿',
    color: 'text-lime-600',
    bgColor: 'bg-lime-100',
    borderColor: 'border-lime-300',
    description: 'Conexión con la naturaleza',
  },
  recreation_leisure: {
    label: 'Recreación',
    icon: '🎮',
    color: 'text-cyan-600',
    bgColor: 'bg-cyan-100',
    borderColor: 'border-cyan-300',
    description: 'Ocio y entretenimiento',
  },
  health_wellbeing: {
    label: 'Salud General',
    icon: '❤️‍🩹',
    color: 'text-red-600',
    bgColor: 'bg-red-100',
    borderColor: 'border-red-300',
    description: 'Bienestar general',
  },
};

/**
 * Configuration for each role type
 */
export const ROLE_TYPE_CONFIG: Record<RoleType, RoleTypeConfig> = {
  personal: {
    label: 'Personal',
    icon: '👤',
    color: 'text-blue-600',
    bgColor: 'bg-blue-100',
    borderColor: 'border-blue-300',
  },
  relational: {
    label: 'Relacional',
    icon: '👥',
    color: 'text-pink-600',
    bgColor: 'bg-pink-100',
    borderColor: 'border-pink-300',
  },
  professional: {
    label: 'Profesional',
    icon: '💼',
    color: 'text-amber-600',
    bgColor: 'bg-amber-100',
    borderColor: 'border-amber-300',
  },
  hobby_cause: {
    label: 'Hobby/Causa',
    icon: '🎯',
    color: 'text-green-600',
    bgColor: 'bg-green-100',
    borderColor: 'border-green-300',
  },
};

/**
 * Get role facet configuration
 */
export function getRoleFacetConfig(facet: RoleFacet): RoleFacetConfig {
  return ROLE_FACET_CONFIG[facet];
}

/**
 * Get role type configuration
 */
export function getRoleTypeConfig(roleType: RoleType): RoleTypeConfig {
  return ROLE_TYPE_CONFIG[roleType];
}

/**
 * Get all role facets as array
 */
export function getAllRoleFacets(): readonly RoleFacet[] {
  return ROLE_FACETS;
}

/**
 * Get all role types as array
 */
export function getAllRoleTypes(): RoleType[] {
  return ['personal', 'relational', 'professional', 'hobby_cause'];
}

/**
 * Gradient map for role cards based on facet
 */
export const ROLE_GRADIENT_MAP: Record<RoleFacet, string> = {
  family_relationships: 'bg-gradient-to-br from-pink-50 to-white',
  friendships_social: 'bg-gradient-to-br from-blue-50 to-white',
  romantic_relationships: 'bg-gradient-to-br from-rose-50 to-white',
  work_career: 'bg-gradient-to-br from-amber-50 to-white',
  education_growth: 'bg-gradient-to-br from-purple-50 to-white',
  physical_health: 'bg-gradient-to-br from-emerald-50 to-white',
  mental_wellbeing: 'bg-gradient-to-br from-teal-50 to-white',
  creative_expression: 'bg-gradient-to-br from-orange-50 to-white',
  leadership_mentorship: 'bg-gradient-to-br from-violet-50 to-white',
  spirituality_meaning: 'bg-gradient-to-br from-indigo-50 to-white',
  community_citizenship: 'bg-gradient-to-br from-green-50 to-white',
  financial_stewardship: 'bg-gradient-to-br from-yellow-50 to-white',
  environmental_nature: 'bg-gradient-to-br from-lime-50 to-white',
  recreation_leisure: 'bg-gradient-to-br from-cyan-50 to-white',
  health_wellbeing: 'bg-gradient-to-br from-red-50 to-white',
};

/**
 * Compassionate messaging for roles (Spanish)
 */
export const ROLE_MESSAGES = {
  orphanedValues: {
    title: 'Valores sin roles conectados',
    description:
      'Estos valores aún no están conectados a ningún rol. Esto está bien—a veces los valores son aspiracionales y todavía estamos buscando formas de vivirlos.',
    action: 'Explorar roles',
  },
  emptyRole: {
    title: 'Rol sin valores',
    description:
      'Este rol aún no tiene valores conectados. ¿Qué es lo más importante para ti cuando estás en este rol?',
    action: 'Conectar valores',
  },
  maxRolesReached: {
    title: 'Límite de roles alcanzado',
    description:
      'Has alcanzado el límite de 15 roles activos. Tener demasiados roles puede dispersar tu energía. ¿Hay roles que podrías archivar?',
    action: 'Gestionar roles',
  },
  roleCreated: {
    title: 'Rol creado',
    description: 'Ahora puedes conectar este rol con los valores que vives a través de él.',
  },
};
