export type ThinkCategory = "personal" | "work" | "ideas" | "learning" | "reflection";

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
  orderBy?: "think_id" | "category" | "date_created" | "date_updated";
}