import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { ApiResponse, Contact, PaginatedData } from './models';

export interface ContactPayload {
  name: string;
  email?: string;
  phone?: string;
  position?: string;
  status?: Contact['status'];
  company_id?: number | null;
}

@Injectable({ providedIn: 'root' })
export class ContactService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/contacts`;

  getAll(offset = 0, limit = 20): Observable<ApiResponse<PaginatedData<Contact>>> {
    const params = new HttpParams().set('offset', offset).set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<Contact>>>(this.apiUrl, { params });
  }

  search(query: string, offset = 0, limit = 20): Observable<ApiResponse<PaginatedData<Contact>>> {
    const params = new HttpParams().set('q', query).set('offset', offset).set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<Contact>>>(`${this.apiUrl}/search`, { params });
  }

  getById(id: number): Observable<ApiResponse<{ contact: Contact }>> {
    return this.http.get<ApiResponse<{ contact: Contact }>>(`${this.apiUrl}/${id}`);
  }

  create(payload: ContactPayload): Observable<ApiResponse<{ contact: Contact }>> {
    return this.http.post<ApiResponse<{ contact: Contact }>>(this.apiUrl, payload);
  }

  update(id: number, payload: ContactPayload): Observable<ApiResponse<{ contact: Contact }>> {
    return this.http.put<ApiResponse<{ contact: Contact }>>(`${this.apiUrl}/${id}`, payload);
  }

  delete(id: number): Observable<ApiResponse<unknown>> {
    return this.http.delete<ApiResponse<unknown>>(`${this.apiUrl}/${id}`);
  }
}
