// Auth types
export interface User {
  sub: string;        // User ID from JWT
  email: string;      // Email from JWT
  name: string;       // Display name from JWT
  roles: string[];    // User roles from JWT
}

export interface AuthTokenResponse {
  token: string;
}

export interface DecodedToken {
  sub: string;        // user_id
  email: string;      // user email
  name: string;       // user name
  iss: string;        // issuer
  exp: number;        // expiration timestamp
  iat: number;        // issued at timestamp
  roles: string[];    // user roles
}

// Think types
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