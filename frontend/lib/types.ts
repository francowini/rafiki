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

export const THINK_CATEGORY_LABELS: Record<ThinkCategory, string> = {
  personal: 'Personal',
  work: 'Trabajo',
  ideas: 'Ideas',
  learning: 'Aprendizaje',
  reflection: 'Reflexión',
};

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

/**
 * Pagination parameters for list endpoints.
 * @param orderBy - Combined field and direction, e.g., 'date_created,DESC'
 *                  Format: "field,DIRECTION" where DIRECTION is ASC or DESC
 *                  Available fields vary by endpoint (e.g., think_id, category, date_created, date_updated)
 */
export interface PaginationParams {
  page?: number;
  rows?: number;
  orderBy?: string;
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

export interface MomentStats {
  thisWeek: number;
  thisMonth: number;
  last30Days: number;
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

// Entity status for soft delete functionality
export type EntityStatus = 'active' | 'archived';

// Entity status constants
export const STATUS_ACTIVE: EntityStatus = 'active';
export const STATUS_ARCHIVED: EntityStatus = 'archived';

export interface Value {
  id: string;
  content: string;
  facet: Facet;
  displayOrder: number;
  status: EntityStatus;
  archivedAt: string | null; // ISO 8601
  dateCreated: string; // ISO 8601
  dateUpdated: string; // ISO 8601
}

// Extended Value with life vision count (for UI display)
export interface ValueWithLifeVisionCount extends Value {
  activeLifeVisionCount?: number;
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
  status: EntityStatus;
  archivedAt: string | null; // ISO 8601
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

// ============================================================================
// Objective Types
// ============================================================================

export type TrackingType = 'result' | 'frequency';
export type ObjectiveStatus = 'active' | 'completed' | 'abandoned' | 'paused';
export type RecordStatus = 'completed' | 'intentionally_skipped' | 'skipped';
export type FrequencyType = 'daily' | 'n_per_week' | 'n_per_month';

export const RECORD_STATUS_LABELS: Record<RecordStatus, string> = {
  completed: 'Completado',
  intentionally_skipped: 'Omitido intencionalmente',
  skipped: 'Omitido',
};

export const OBJECTIVE_STATUS_LABELS: Record<ObjectiveStatus, string> = {
  active: 'Activo',
  completed: 'Completado',
  abandoned: 'Abandonado',
  paused: 'Pausado',
};

export interface Objective {
  id: string;
  lifeVisionId: string;
  title: string;
  trackingType: TrackingType;
  status: ObjectiveStatus;

  // Result fields (null if frequency)
  targetMetric?: number;
  currentMetric?: number;
  beginMetric?: number;

  // Frequency fields (null if result)
  frequencyType?: FrequencyType;
  frequencyCount?: number;
  complianceTargetPct?: number;

  archivedAt?: string | null;
  dateCreated: string;
  dateUpdated: string;
}

export interface ObjectiveWithRelations extends Objective {
  lifeVision?: LifeVision;
  value?: Value;
  currentWeekCompliance?: number;
  currentMonthCompliance?: number;
}

export interface NewObjective {
  lifeVisionId: string;
  title: string;
  trackingType: TrackingType;
  targetMetric?: number;
  beginMetric?: number;
  frequencyType?: FrequencyType;
  frequencyCount?: number;
  complianceTargetPct?: number;
}

export interface UpdateObjective {
  title?: string;
  lifeVisionId?: string;
  status?: ObjectiveStatus;
  targetMetric?: number;
  beginMetric?: number;
  frequencyType?: FrequencyType;
  frequencyCount?: number;
  complianceTargetPct?: number;
}

export interface ObjectiveListResponse {
  items: Objective[];
  total: number;
}

// ============================================================================
// Objective Record Types
// ============================================================================

export interface ObjectiveRecord {
  id: string;
  objectiveId: string;
  recordDate: string; // YYYY-MM-DD
  status: RecordStatus;
  notes?: string;
  dateCreated: string;
}

export interface NewObjectiveRecord {
  recordDate: string; // YYYY-MM-DD (today or yesterday only)
  status: RecordStatus;
  notes?: string;
}

export interface ObjectiveRecordListResponse {
  items: ObjectiveRecord[];
  total: number;
}

// ============================================================================
// Activity Heatmap Types
// ============================================================================

export interface ActivityItem {
  type: 'task' | 'record';
  id: string;
  title?: string;
  contribution?: number;
  status?: string;
  notes?: string;
}

export interface ObjectiveActivityDay {
  date: string; // YYYY-MM-DD
  hasActivity: boolean;
  items: ActivityItem[];
}

export interface ObjectiveActivityData {
  year: number;
  days: ObjectiveActivityDay[];
  totalCompletions: number;
  currentStreak: number;
  longestStreak: number;
  streakUnit: 'days' | 'weeks' | 'months';
}

// ============================================================================
// Task Types
// ============================================================================

export type TaskStatus = 'pending' | 'completed' | 'cancelled';

export const TASK_STATUS_LABELS: Record<TaskStatus, string> = {
  pending: 'Pendiente',
  completed: 'Completada',
  cancelled: 'Cancelada',
};

export interface Task {
  id: string;
  objectiveId: string | null;
  title: string;
  description: string | null;
  contribution: number | null;
  status: TaskStatus;
  completedAt: string | null;
  dateCreated: string;
  dateUpdated: string;
}

export interface NewTask {
  objectiveId: string | null;
  title: string;
  description?: string | null;
  contribution?: number | null;
}

export interface UpdateTask {
  title?: string;
  description?: string | null;
  clearDescription?: boolean;
  contribution?: number | null;
}

export interface CompleteTaskResponse {
  task: Task;
  objectiveProgress: number | null;
  objectiveTracking: 'result' | 'frequency' | null; // Tracking type for smart prompt
}

export interface MoveTaskRequest {
  objectiveId: string;
  contribution: number | null;
}

export interface TaskListResponse {
  items: Task[];
  total: number;
  page: number;
  rowsPerPage: number;
}

export interface TaskFilters {
  objectiveId?: string;
  status?: TaskStatus;
  inboxOnly?: boolean;
}
