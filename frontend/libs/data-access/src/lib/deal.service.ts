import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { ApiResponse, Deal, PaginatedData } from './models';

export interface CreateDealPayload {
  contact_id?: number;
  title: string;
  value?: number;
  currency?: string;
  stage?: Deal['stage'];
}

export type UpdateDealPayload = Partial<CreateDealPayload>;

@Injectable({ providedIn: 'root' })
export class DealService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/deals`;

  getAll(
    offset = 0,
    limit = 20
  ): Observable<ApiResponse<PaginatedData<Deal>>> {
    const params = new HttpParams()
      .set('offset', offset)
      .set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<Deal>>>(this.apiUrl, {
      params,
    });
  }

  getById(id: number): Observable<ApiResponse<{ deal: Deal }>> {
    return this.http.get<ApiResponse<{ deal: Deal }>>(`${this.apiUrl}/${id}`);
  }

  create(payload: CreateDealPayload): Observable<ApiResponse<{ deal: Deal }>> {
    return this.http.post<ApiResponse<{ deal: Deal }>>(this.apiUrl, payload);
  }

  update(
    id: number,
    payload: UpdateDealPayload
  ): Observable<ApiResponse<{ deal: Deal }>> {
    return this.http.put<ApiResponse<{ deal: Deal }>>(
      `${this.apiUrl}/${id}`,
      payload
    );
  }

  delete(id: number): Observable<ApiResponse<unknown>> {
    return this.http.delete<ApiResponse<unknown>>(`${this.apiUrl}/${id}`);
  }
}
