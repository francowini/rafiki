// Auth types
export interface User {
  sub: string; // User ID from JWT
  email: string; // Email from JWT
  name: string; // Display name from JWT
  roles: string[]; // User roles from JWT
}

export interface AuthTokenResponse {
  token: string;
}

export interface DecodedToken {
  sub: string; // user_id
  email: string; // user email
  name: string; // user name
  iss: string; // issuer
  exp: number; // expiration timestamp
  iat: number; // issued at timestamp
  roles: string[]; // user roles
}

// Think types
export type ThinkCategory = 'personal' | 'work' | 'ideas' | 'learning' | 'reflection';

export interface Think {
  id: string;
  category: ThinkCategory;
  content: string;
  dateCreated: string;
  dateUpdated: string;
}

export interface NewThink {
  category: ThinkCategory;
  content: string;
}

export interface ThinkListResponse {
  items: Think[];
  total: number;
  page: number;
  rowsPerPage: number;
}

export interface PaginationParams {
  page?: number;
  rows?: number;
  orderBy?: 'think_id' | 'category' | 'date_created' | 'date_updated';
}

// ============================================================================
// Moment (Registro Funcional Diario) Types
// ============================================================================

export interface Moment {
  id: string;
  momentDate: string; // ISO 8601 timestamp
  situation: string;
  thoughts: string;
  physicalSymptoms: string;
  behavior: string;
  consequences: string;
  valuesReflection: string;
  intensity: number; // 0-10
  dateCreated: string; // ISO 8601
  dateUpdated: string; // ISO 8601
}

export interface NewMoment {
  momentDate: string; // ISO 8601 timestamp
  situation: string;
  thoughts: string;
  physicalSymptoms: string;
  behavior: string;
  consequences: string;
  valuesReflection: string;
  intensity: number; // 0-10
}

export interface UpdateMoment {
  momentDate?: string;
  situation?: string;
  thoughts?: string;
  physicalSymptoms?: string;
  behavior?: string;
  consequences?: string;
  valuesReflection?: string;
  intensity?: number;
}

export interface MomentListResponse {
  items: Moment[];
  total: number;
  page: number;
  rowsPerPage: number;
}

// ============================================================================
// Value Types
// ============================================================================

export const FACETS = [
  'health',
  'relationships',
  'career',
  'personal_growth',
  'family',
  'creativity',
  'community',
  'spirituality',
] as const;

export type Facet = (typeof FACETS)[number];

export interface Value {
  id: string;
  content: string;
  facet: Facet;
  displayOrder: number;
  dateCreated: string; // ISO 8601
  dateUpdated: string; // ISO 8601
}

export interface NewValue {
  content: string;
  facet: Facet;
  displayOrder: number;
}

export interface UpdateValue {
  content?: string;
  facet?: Facet;
  displayOrder?: number;
}

export interface ValueListResponse {
  items: Value[];
  total: number;
}

// ============================================================================
// Life Vision Types
// ============================================================================

export interface LifeVision {
  id: string;
  valueId: string;
  content: string;
  dateCreated: string;
  dateUpdated: string;
}

export interface NewLifeVision {
  valueId: string;
  content: string;
}

export interface UpdateLifeVision {
  content?: string;
  valueId?: string;
}

export interface LifeVisionListResponse {
  items: LifeVision[];
  total: number;
}

// ============================================================================
// Export Types
// ============================================================================

export interface ExportItem {
  id: string;
  itemType: 'moment' | 'think';
  itemDate: string; // ISO 8601 timestamp

  // Moment-specific fields (optional)
  situation?: string;
  thoughts?: string;
  physicalSymptoms?: string;
  behavior?: string;
  consequences?: string;
  valuesReflection?: string;
  intensity?: number;

  // Think-specific fields (optional)
  category?: ThinkCategory;
  content?: string;

  dateCreated: string;
}

export interface ExportParams {
  startDate: string; // ISO 8601 date string
  endDate: string; // ISO 8601 date string
  page?: number; // Page number (1-based, default: 1)
  rows?: number; // Items per page (max: 100 per backend limit)
  orderBy?: string; // Sort order (default: "item_date,DESC")
}

// Standard Query Domain response format
export interface ExportResponse {
  items: ExportItem[];
  total: number;
  page: number; // Current page number
  rowsPerPage: number; // Items per page
}

export type DateRangePreset = '7d' | '14d' | '30d' | '90d' | 'custom';

export interface DateRange {
  startDate: Date;
  endDate: Date;
}
