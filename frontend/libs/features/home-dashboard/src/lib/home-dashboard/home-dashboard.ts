import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { TranslateModule } from '@ngx-translate/core';
import {
  ContactService,
  DealService,
  SidebarComponent,
  Contact,
  Deal,
} from '@miapp/data-access';

@Component({
  selector: 'app-home-dashboard',
  standalone: true,
  imports: [CommonModule, TranslateModule, SidebarComponent],
  templateUrl: './home-dashboard.html',
  styleUrl: './home-dashboard.css',
})
export class HomeDashboardComponent implements OnInit {
  private readonly contactSvc = inject(ContactService);
  private readonly dealSvc = inject(DealService);

  // Estado con Signals
  readonly contacts = signal<Contact[]>([]);
  readonly deals = signal<Deal[]>([]);
  readonly totalContacts = signal(0);
  readonly totalDeals = signal(0);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);

  ngOnInit(): void {
    this.loadData();
  }

  loadData(): void {
    this.loading.set(true);
    this.error.set(null);

    this.contactSvc.getAll(0, 5).subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.contacts.set(res.data.list);
          this.totalContacts.set(res.data.total);
        }
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(
          err.error?.error?.message ?? 'Error cargando contactos'
        );
      },
    });

    this.dealSvc.getAll(0, 5).subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.deals.set(res.data.list);
          this.totalDeals.set(res.data.total);
        }
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.error?.message ?? 'Error cargando deals');
        this.loading.set(false);
      },
    });
  }

  stageBadgeClass(stage: Deal['stage']): string {
    const map: Record<Deal['stage'], string> = {
      prospect: 'bg-gray-100 text-gray-700',
      qualified: 'bg-blue-100 text-blue-700',
      proposal: 'bg-yellow-100 text-yellow-700',
      won: 'bg-green-100 text-green-700',
      lost: 'bg-red-100 text-red-700',
    };
    return map[stage] ?? 'bg-gray-100 text-gray-700';
  }

  statusBadgeClass(status: Contact['status']): string {
    const map: Record<Contact['status'], string> = {
      active: 'bg-green-100 text-green-700',
      inactive: 'bg-gray-100 text-gray-600',
      lead: 'bg-blue-100 text-blue-700',
    };
    return map[status] ?? 'bg-gray-100 text-gray-700';
  }
}
