import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from './environment';
import { ApiResponse, AuthLoginData, AuthTokens, User } from './models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl;

  // ─── Estado reactivo con Signals ──────────────────────────────
  private readonly _currentUser = signal<User | null>(null);
  private readonly _isLoggedIn = signal<boolean>(false);

  // ─── Selectores públicos (readonly) ───────────────────────────
  readonly user = this._currentUser.asReadonly();
  readonly authenticated = this._isLoggedIn.asReadonly();
  readonly isAdmin = computed(() => this._currentUser()?.role === 'admin');
  readonly mustChangePassword = computed(
    () => this._currentUser()?.must_change_password === true
  );

  constructor() {
    this.restoreSession();
  }

  // ─── Auth ────────────────────────────────────────────────────

  login(
    email: string,
    password: string
  ): Observable<ApiResponse<AuthLoginData>> {
    return this.http
      .post<ApiResponse<AuthLoginData>>(`${this.apiUrl}/auth/login`, {
        email,
        password,
      })
      .pipe(
        tap((res) => {
          if (res.success && res.data) {
            this.storeSession(res.data);
          }
        })
      );
  }

  refreshToken(): Observable<ApiResponse<AuthTokens>> {
    const token = localStorage.getItem('refresh_token');
    return this.http
      .post<ApiResponse<AuthTokens>>(`${this.apiUrl}/auth/refresh`, {
        refresh_token: token,
      })
      .pipe(
        tap((res) => {
          if (res.success && res.data) {
            localStorage.setItem('access_token', res.data.access_token);
            localStorage.setItem('refresh_token', res.data.refresh_token);
          }
        })
      );
  }

  logout(): Observable<ApiResponse<unknown>> {
    const token = localStorage.getItem('refresh_token');
    return this.http
      .post<ApiResponse<unknown>>(`${this.apiUrl}/auth/logout`, {
        refresh_token: token,
      })
      .pipe(tap(() => this.clearSession()));
  }

  changePassword(
    currentPassword: string,
    newPassword: string
  ): Observable<ApiResponse<unknown>> {
    return this.http.put<ApiResponse<unknown>>(
      `${this.apiUrl}/auth/change-password`,
      { current_password: currentPassword, new_password: newPassword }
    );
  }

  getAccessToken(): string | null {
    return localStorage.getItem('access_token');
  }

  // ─── Session management ────────────────────────────────────────

  private storeSession(data: AuthLoginData): void {
    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('refresh_token', data.refresh_token);
    localStorage.setItem('user', JSON.stringify(data.user));
    this._currentUser.set(data.user);
    this._isLoggedIn.set(true);
  }

  private restoreSession(): void {
    const raw = localStorage.getItem('user');
    const token = localStorage.getItem('access_token');
    if (raw && token) {
      try {
        const user = JSON.parse(raw) as User;
        this._currentUser.set(user);
        this._isLoggedIn.set(true);
      } catch {
        this.clearSession();
      }
    }
  }

  clearSession(): void {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    this._currentUser.set(null);
    this._isLoggedIn.set(false);
  }
}
