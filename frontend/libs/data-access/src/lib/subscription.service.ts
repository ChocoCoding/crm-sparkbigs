import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiResponse, PaginatedData, Subscription } from './models';
import { environment } from './environment';

export interface SubscriptionPayload {
  company_id: number;
  name: string;
  plan_type: string;
  status?: string;
  amount: number;
  currency?: string;
  billing_cycle?: string;
  start_date: string;       // YYYY-MM-DD
  renewal_date?: string | null;
  notes?: string;
}

@Injectable({ providedIn: 'root' })
export class SubscriptionService {
  private http = inject(HttpClient);
  private base = `${environment.apiUrl}/subscriptions`;

  getAll(offset = 0, limit = 50): Observable<ApiResponse<PaginatedData<Subscription>>> {
    return this.http.get<ApiResponse<PaginatedData<Subscription>>>(
      `${this.base}?offset=${offset}&limit=${limit}`
    );
  }

  getExpiringSoon(days = 30): Observable<ApiResponse<{ list: Subscription[] }>> {
    return this.http.get<ApiResponse<{ list: Subscription[] }>>(
      `${this.base}/expiring?days=${days}`
    );
  }

  getById(id: number): Observable<ApiResponse<{ subscription: Subscription }>> {
    return this.http.get<ApiResponse<{ subscription: Subscription }>>(`${this.base}/${id}`);
  }

  create(payload: SubscriptionPayload): Observable<ApiResponse<{ subscription: Subscription }>> {
    return this.http.post<ApiResponse<{ subscription: Subscription }>>(this.base, payload);
  }

  update(id: number, payload: SubscriptionPayload): Observable<ApiResponse<{ subscription: Subscription }>> {
    return this.http.put<ApiResponse<{ subscription: Subscription }>>(`${this.base}/${id}`, payload);
  }

  delete(id: number): Observable<ApiResponse<{ message: string }>> {
    return this.http.delete<ApiResponse<{ message: string }>>(`${this.base}/${id}`);
  }
}
