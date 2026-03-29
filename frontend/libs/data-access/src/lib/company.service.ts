import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { ApiResponse, Company, PaginatedData } from './models';

export interface CompanyPayload {
  name: string;
  sector?: string;
  status?: Company['status'];
  website?: string;
  phone?: string;
  address?: string;
  relation_start_date?: string | null; // "YYYY-MM-DD"
}

@Injectable({ providedIn: 'root' })
export class CompanyService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/companies`;

  getAll(offset = 0, limit = 20): Observable<ApiResponse<PaginatedData<Company>>> {
    const params = new HttpParams().set('offset', offset).set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<Company>>>(this.apiUrl, { params });
  }

  search(query: string, offset = 0, limit = 20): Observable<ApiResponse<PaginatedData<Company>>> {
    const params = new HttpParams().set('q', query).set('offset', offset).set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<Company>>>(`${this.apiUrl}/search`, { params });
  }

  getById(id: number): Observable<ApiResponse<{ company: Company }>> {
    return this.http.get<ApiResponse<{ company: Company }>>(`${this.apiUrl}/${id}`);
  }

  create(payload: CompanyPayload): Observable<ApiResponse<{ company: Company }>> {
    return this.http.post<ApiResponse<{ company: Company }>>(this.apiUrl, payload);
  }

  update(id: number, payload: CompanyPayload): Observable<ApiResponse<{ company: Company }>> {
    return this.http.put<ApiResponse<{ company: Company }>>(`${this.apiUrl}/${id}`, payload);
  }

  delete(id: number): Observable<ApiResponse<unknown>> {
    return this.http.delete<ApiResponse<unknown>>(`${this.apiUrl}/${id}`);
  }
}
