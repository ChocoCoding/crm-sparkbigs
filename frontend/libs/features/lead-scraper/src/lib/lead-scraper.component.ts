import { Component, OnInit, OnDestroy, inject, signal } from '@angular/core';
import { CommonModule, CurrencyPipe, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import {
  LeadScraperService,
  ScrapeJob,
  ScrapedLead,
  ScrapeAPILog,
  SidebarComponent,
} from '@miapp/data-access';

@Component({
  selector: 'miapp-lead-scraper',
  standalone: true,
  imports: [CommonModule, FormsModule, CurrencyPipe, DatePipe, SidebarComponent],
  templateUrl: './lead-scraper.component.html',
  styleUrls: ['./lead-scraper.component.css']
})
export class LeadScraperComponent implements OnInit, OnDestroy {
  private readonly scraperService = inject(LeadScraperService);
  private readonly router = inject(Router);

  // States
  activeTab = signal<'dashboard' | 'history' | 'leads'>('dashboard');

  // Dashboard Form
  searchQuery = signal('');
  searchLocation = signal('');
  maxResults = signal<number>(5);
  isScraping = signal(false);

  // Active Job State
  currentJobId = signal<string | null>(null);
  currentJobStatus = signal<string>('');
  currentJobTotal = signal(0);
  currentJobProcessed = signal(0);
  eventStream: EventSource | null = null;

  // Logs & Results
  pipelineLogs = signal<any[]>([]);
  scrapedLeads = signal<ScrapedLead[]>([]);

  // History & Globals
  jobsHistory = signal<ScrapeJob[]>([]);
  allLeads = signal<ScrapedLead[]>([]);
  
  ngOnInit(): void {
    this.loadJobsHistory();
  }

  ngOnDestroy(): void {
    if (this.eventStream) {
      this.eventStream.close();
    }
  }

  // ─── Tab Management ────────────────────────────────────────────────────────
  
  setTab(tab: 'dashboard' | 'history' | 'leads') {
    this.activeTab.set(tab);
    if (tab === 'history') {
      this.loadJobsHistory();
    } else if (tab === 'leads') {
      this.loadAllLeads();
    }
  }

  // ─── Actions ───────────────────────────────────────────────────────────────

  startScrape() {
    if (!this.searchQuery() || !this.searchLocation()) return;

    this.isScraping.set(true);
    this.pipelineLogs.set([]);
    this.scrapedLeads.set([]);
    this.currentJobTotal.set(0);
    this.currentJobProcessed.set(0);
    this.currentJobStatus.set('Iniciando...');

    this.scraperService
      .startScrape({
        query: this.searchQuery(),
        location: this.searchLocation(),
        max_results: this.maxResults(),
      })
      .subscribe({
        next: (res) => {
          if (res.success && res.data) {
            this.currentJobId.set(res.data.job_id);
            this.listenToJobStream(res.data.job_id);
          }
        },
        error: (err) => {
          this.isScraping.set(false);
          this.currentJobStatus.set('Error al iniciar: ' + (err.error?.message || err.message));
        },
      });
  }

  listenToJobStream(jobId: string) {
    if (this.eventStream) {
      this.eventStream.close();
    }

    this.eventStream = this.scraperService.streamJobEvents(jobId);

    this.eventStream.onmessage = (event) => {
      // Fiber stream sends multiple event types natively via standard SSE formatting, 
      // check the actual EventSource usage. 
      // If we use regular message, we only catch simple 'message' events.
    };
    
    // Add custom event listeners mapping to the backend event names
    const events = [
      'connected', 'pipeline_start', 'step_start', 'api_request', 'api_response', 'step_error', 
      'step_done', 'companies_found', 'company_start', 'company_error', 'lead_duplicate',
      'lead_skipped', 'lead_saved', 'company_done', 'pipeline_error', 'pipeline_done'
    ];

    events.forEach(eventName => {
      this.eventStream!.addEventListener(eventName, (event: any) => {
        this.handleEvent(eventName, event.data);
      });
    });
  }

  private handleEvent(eventName: string, dataStr: string) {
    const data = JSON.parse(dataStr);
    
    // Add to logs
    this.pipelineLogs.update(logs => [{ event: eventName, ...data }, ...logs]);

    switch (eventName) {
      case 'companies_found':
        this.currentJobTotal.set(data.total);
        this.currentJobStatus.set(`Procesando ${data.total} empresas...`);
        break;
      case 'company_done':
      case 'company_error':
        this.currentJobProcessed.update(v => v + 1);
        this.loadJobLeads(this.currentJobId()!); // Quick refresh of saved leads
        break;
      case 'pipeline_done':
        this.isScraping.set(false);
        this.currentJobStatus.set(`Completado. Procesadas: ${data.totalProcessed}`);
        if (this.eventStream) this.eventStream.close();
        if (this.currentJobId()) {
          this.loadJobLeads(this.currentJobId()!);
        }
        this.loadJobsHistory();
        break;
      case 'pipeline_error':
        this.isScraping.set(false);
        this.currentJobStatus.set(`Error: ${data.error}`);
        if (this.eventStream) this.eventStream.close();
        break;
    }
  }

  loadJobLeads(jobId: string) {
    this.scraperService.getLeadsByJob(jobId).subscribe(res => {
      if (res.success && res.data) {
        this.scrapedLeads.set(res.data.list);
      }
    });
  }

  loadJobsHistory() {
    this.scraperService.getJobs(0, 50).subscribe(res => {
      if (res.success && res.data) {
        this.jobsHistory.set(res.data.list);
      }
    });
  }

  loadAllLeads() {
    this.scraperService.getLeads(0, 200).subscribe(res => {
      if (res.success && res.data) {
        this.allLeads.set(res.data.list);
      }
    });
  }

  importLead(leadId: number) {
    if (!confirm('¿Deseas importar este lead al CRM?')) return;
    
    this.scraperService.importToCRM(leadId).subscribe({
      next: (res) => {
        if (res.success) {
          alert('Lead importado con éxito!');
          if (this.currentJobId()) {
            this.loadJobLeads(this.currentJobId()!);
          }
        }
      },
      error: (err) => {
        alert('Error importando: ' + (err.error?.message || err.message));
      }
    });
  }

  updateLeadEstado(leadId: number, newState: ScrapedLead['estado']) {
    this.scraperService.updateLeadEstado(leadId, newState).subscribe({
      next: (res) => {
        // Updated correctly in DB, visual state already linked via ngModel
      },
      error: (err) => {
        console.error('Error updating state', err);
        alert('Error al actualizar el estado: ' + (err.error?.message || err.message));
      }
    });
  }

  getEstadoBadgeStyle(estado: string): any {
    switch (estado) {
      case 'nuevo':
        return { background: '#e3f2fd', color: '#1565c0' }; // azul
      case 'descartado':
        return { background: '#ffebee', color: '#c62828' }; // rojo
      case 'potencial':
        return { background: '#fff8e1', color: '#f57f17' }; // amarillo/naranja
      case 'contactado':
        return { background: '#e8f5e9', color: '#2e7d32' }; // verde
      case 'convertido':
        return { background: '#e8f5e9', color: '#006e2a' }; // verde oscuro
      default:
        return { background: '#f4f6fb', color: '#5c5f7a' };
    }
  }
}
