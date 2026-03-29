// ─── Interfaces que mapean las respuestas estándar del backend ───────────────

/** Envoltorio estándar para todas las respuestas de la API */
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

/** Respuesta paginada */
export interface PaginatedData<T> {
  list: T[];
  total: number;
}

// ─── Modelos de Dominio ──────────────────────────────────────────────────────

export interface License {
  id: number;
  user_id: number;
  plan: 'free' | 'pro' | 'enterprise';
  is_active: boolean;
  expires_at: string | null;
}

export interface User {
  id: number;
  email: string;
  name: string;
  role: 'admin' | 'user';
  is_active: boolean;
  must_change_password: boolean;
  last_login_at: string | null;
  license?: License;
}

export interface Contact {
  id: number;
  user_id: number;
  name: string;
  email: string;
  phone: string;
  company: string;
  status: 'active' | 'inactive' | 'lead';
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Deal {
  id: number;
  user_id: number;
  contact_id: number;
  title: string;
  value: number;
  currency: string;
  stage: 'prospect' | 'qualified' | 'proposal' | 'won' | 'lost';
  metadata?: Record<string, unknown>;
  closed_at: string | null;
  created_at: string;
  updated_at: string;
}

// ─── Payloads de Auth ────────────────────────────────────────────────────────

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

export interface AuthLoginData extends AuthTokens {
  user: User;
}
