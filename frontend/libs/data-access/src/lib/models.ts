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

export interface Company {
  id: number;
  user_id: number;
  name: string;
  sector: string;
  status: 'prospect' | 'active' | 'inactive';
  website: string;
  phone: string;
  address: string;
  relation_start_date: string | null;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Contact {
  id: number;
  user_id: number;
  company_id: number | null;
  name: string;
  email: string;
  phone: string;
  position: string;
  status: 'active' | 'inactive' | 'lead';
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  // Relación cargada con Preload en el backend
  company?: Pick<Company, 'id' | 'name' | 'sector'>;
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

export interface Meeting {
  id: number;
  user_id: number;
  company_id: number;
  contact_id: number | null;
  title: string;
  start_at: string;        // ISO 8601
  duration_min: number;
  status: 'scheduled' | 'completed' | 'cancelled';
  notes: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  // Relaciones Preload
  company?: Pick<Company, 'id' | 'name'>;
  contact?: Pick<Contact, 'id' | 'name' | 'position'>;
}

export interface Subscription {
  id: number;
  user_id: number;
  company_id: number;
  name: string;
  plan_type: string;
  status: 'active' | 'cancelled' | 'paused' | 'expired';
  amount: number;
  currency: string;
  billing_cycle: 'monthly' | 'annual' | 'one_time';
  start_date: string;          // YYYY-MM-DD
  renewal_date: string | null; // YYYY-MM-DD
  notes: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  company?: Pick<Company, 'id' | 'name'>;
}

export interface Setting {
  id: number;
  user_id: number;
  category: string;
  key: string;
  value: string;
  label: string;
  input_type: 'text' | 'number' | 'boolean' | 'select' | 'email';
}

// ─── Payloads de Auth ────────────────────────────────────────────────────────

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

export interface AuthLoginData extends AuthTokens {
  user: User;
}
