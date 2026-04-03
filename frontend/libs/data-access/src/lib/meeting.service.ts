import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { ApiResponse, Meeting, PaginatedData } from './models';

export interface MeetingPayload {
  title: string;
  company_id: number;
  contact_id?: number | null;
  start_at: string;        // ISO 8601: "2024-06-15T10:00:00Z"
  duration_min?: number;
  status?: Meeting['status'];
  notes?: string;
  summary?: string;
}

@Injectable({ providedIn: 'root' })
export class MeetingService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/meetings`;

  getAll(offset = 0, limit = 20): Observable<ApiResponse<PaginatedData<Meeting>>> {
    const params = new HttpParams().set('offset', offset).set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<Meeting>>>(this.apiUrl, { params });
  }

  getUpcoming(limit = 10): Observable<ApiResponse<{ list: Meeting[] }>> {
    const params = new HttpParams().set('limit', limit);
    return this.http.get<ApiResponse<{ list: Meeting[] }>>(`${this.apiUrl}/upcoming`, { params });
  }

  getById(id: number): Observable<ApiResponse<{ meeting: Meeting }>> {
    return this.http.get<ApiResponse<{ meeting: Meeting }>>(`${this.apiUrl}/${id}`);
  }

  create(payload: MeetingPayload): Observable<ApiResponse<{ meeting: Meeting }>> {
    return this.http.post<ApiResponse<{ meeting: Meeting }>>(this.apiUrl, payload);
  }

  update(id: number, payload: MeetingPayload): Observable<ApiResponse<{ meeting: Meeting }>> {
    return this.http.put<ApiResponse<{ meeting: Meeting }>>(`${this.apiUrl}/${id}`, payload);
  }

  delete(id: number): Observable<ApiResponse<unknown>> {
    return this.http.delete<ApiResponse<unknown>>(`${this.apiUrl}/${id}`);
  }
}
