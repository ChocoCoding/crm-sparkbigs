import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiResponse, Meeting, Subscription, Company } from './models';
import { environment } from './environment';

export interface DashboardStats {
  total_companies: number;
  active_companies: number;
  total_contacts: number;
  active_subscriptions: number;
  mrr: number;
  arr: number;
  meetings_this_month: number;
  upcoming_meetings: Meeting[];
  expiring_soon: Subscription[];
  recent_companies: Company[];
}

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private http = inject(HttpClient);

  getStats(): Observable<ApiResponse<{ stats: DashboardStats }>> {
    return this.http.get<ApiResponse<{ stats: DashboardStats }>>(
      `${environment.apiUrl}/dashboard`
    );
  }
}
