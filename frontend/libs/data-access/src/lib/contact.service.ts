import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from './environment';
import { ApiResponse, Contact, PaginatedData } from './models';

export interface CreateContactPayload {
  name: string;
  email?: string;
  phone?: string;
  company?: string;
}

export interface UpdateContactPayload extends CreateContactPayload {
  status?: Contact['status'];
}

@Injectable({ providedIn: 'root' })
export class ContactService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/contacts`;

  getAll(
    offset = 0,
    limit = 20
  ): Observable<ApiResponse<PaginatedData<Contact>>> {
    const params = new HttpParams()
      .set('offset', offset)
      .set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<Contact>>>(this.apiUrl, {
      params,
    });
  }

  getById(id: number): Observable<ApiResponse<{ contact: Contact }>> {
    return this.http.get<ApiResponse<{ contact: Contact }>>(
      `${this.apiUrl}/${id}`
    );
  }

  create(
    payload: CreateContactPayload
  ): Observable<ApiResponse<{ contact: Contact }>> {
    return this.http.post<ApiResponse<{ contact: Contact }>>(
      this.apiUrl,
      payload
    );
  }

  update(
    id: number,
    payload: UpdateContactPayload
  ): Observable<ApiResponse<{ contact: Contact }>> {
    return this.http.put<ApiResponse<{ contact: Contact }>>(
      `${this.apiUrl}/${id}`,
      payload
    );
  }

  delete(id: number): Observable<ApiResponse<unknown>> {
    return this.http.delete<ApiResponse<unknown>>(`${this.apiUrl}/${id}`);
  }
}
