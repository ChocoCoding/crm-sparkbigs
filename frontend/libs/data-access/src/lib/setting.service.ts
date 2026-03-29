import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiResponse, Setting } from './models';
import { environment } from './environment';

export interface SettingPayload {
  category: string;
  key: string;
  value: string;
  label?: string;
  input_type?: string;
}

@Injectable({ providedIn: 'root' })
export class SettingService {
  private http = inject(HttpClient);
  private base = `${environment.apiUrl}/settings`;

  getAll(): Observable<ApiResponse<{ list: Setting[] }>> {
    return this.http.get<ApiResponse<{ list: Setting[] }>>(this.base);
  }

  getByCategory(category: string): Observable<ApiResponse<{ list: Setting[] }>> {
    return this.http.get<ApiResponse<{ list: Setting[] }>>(`${this.base}/${category}`);
  }

  upsert(payload: SettingPayload): Observable<ApiResponse<{ setting: Setting }>> {
    return this.http.put<ApiResponse<{ setting: Setting }>>(this.base, payload);
  }

  delete(id: number): Observable<ApiResponse<{ message: string }>> {
    return this.http.delete<ApiResponse<{ message: string }>>(`${this.base}/${id}`);
  }

  seedDefaults(): Observable<ApiResponse<{ message: string }>> {
    return this.http.post<ApiResponse<{ message: string }>>(`${this.base}/seed`, {});
  }
}
