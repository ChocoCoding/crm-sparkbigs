import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from './environment';
import {
  ApiResponse,
  PaginatedData,
  ScrapeJob,
  ScrapedLead,
  ScrapeAPILog,
  Company,
  Contact,
} from './models';

export interface ScrapePayload {
  query: string;
  location: string;
  max_results: number;
}

@Injectable({ providedIn: 'root' })
export class LeadScraperService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/lead-scraper`;

  // ─── Scrape Pipeline ─────────────────────────────────────────────────────────

  startScrape(payload: ScrapePayload): Observable<ApiResponse<{ job_id: string; message: string }>> {
    return this.http.post<ApiResponse<{ job_id: string; message: string }>>(
      `${this.apiUrl}/scrape`,
      payload
    );
  }

  streamJobEvents(jobId: string): EventSource {
    const url = `${this.apiUrl}/stream/${jobId}`;
    // Se asume que JWT no va en EventSource fácilmente (es un GET nativo del browser).
    // Si el backend lo protege y requiere header, habría que usar Polyfill como EventSourcePolyfill
    // o pasar un token por query param. Por ahora el backend tiene /stream/:jobId con JWT
    // En el futuro, si the auth middleware requires it, you might need to append ?token=... 
    // Para resolverlo de forma standard, si authInterceptor no atrapa EventSource,
    // puedes pasar el token:
    const token = localStorage.getItem('access_token');
    // Para simplificar, usamos EventSource nativo. Asegúrate de configurar el endpoint para
    // soportar Auth en query params si Fiber lo bloquea (o ignorar JWT para SSE).
    return new EventSource(url);
  }

  // ─── Jobs ────────────────────────────────────────────────────────────────────

  getJobs(offset = 0, limit = 20): Observable<ApiResponse<PaginatedData<ScrapeJob>>> {
    const params = new HttpParams().set('offset', offset).set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<ScrapeJob>>>(`${this.apiUrl}/jobs`, { params });
  }

  getJob(id: string): Observable<ApiResponse<{ job: ScrapeJob }>> {
    return this.http.get<ApiResponse<{ job: ScrapeJob }>>(`${this.apiUrl}/jobs/${id}`);
  }

  // ─── Leads ───────────────────────────────────────────────────────────────────

  getLeads(offset = 0, limit = 20): Observable<ApiResponse<PaginatedData<ScrapedLead>>> {
    const params = new HttpParams().set('offset', offset).set('limit', limit);
    return this.http.get<ApiResponse<PaginatedData<ScrapedLead>>>(`${this.apiUrl}/leads`, { params });
  }

  getLeadsByJob(jobId: string): Observable<ApiResponse<PaginatedData<ScrapedLead>>> {
    const params = new HttpParams().set('jobId', jobId);
    return this.http.get<ApiResponse<PaginatedData<ScrapedLead>>>(`${this.apiUrl}/leads`, { params });
  }

  getLead(id: number): Observable<ApiResponse<{ lead: ScrapedLead }>> {
    return this.http.get<ApiResponse<{ lead: ScrapedLead }>>(`${this.apiUrl}/leads/${id}`);
  }

  updateLeadEstado(id: number, estado: ScrapedLead['estado']): Observable<ApiResponse<{ message: string }>> {
    return this.http.patch<ApiResponse<{ message: string }>>(`${this.apiUrl}/leads/${id}`, { estado });
  }

  deleteLead(id: number): Observable<ApiResponse<{ message: string }>> {
    return this.http.delete<ApiResponse<{ message: string }>>(`${this.apiUrl}/leads/${id}`);
  }

  // ─── API Logs ────────────────────────────────────────────────────────────────

  getApiLogs(jobId: string): Observable<ApiResponse<{ list: ScrapeAPILog[] }>> {
    return this.http.get<ApiResponse<{ list: ScrapeAPILog[] }>>(`${this.apiUrl}/logs/${jobId}`);
  }

  // ─── Import to CRM ───────────────────────────────────────────────────────────

  importToCRM(leadId: number): Observable<ApiResponse<{ company: Company; contact: Contact; message: string }>> {
    return this.http.post<ApiResponse<{ company: Company; contact: Contact; message: string }>>(
      `${this.apiUrl}/leads/${leadId}/import`,
      {}
    );
  }
}
