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
