import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { TranslateModule } from '@ngx-translate/core';
import {
  ContactService,
  ContactPayload,
  CompanyService,
  SidebarComponent,
  Contact,
  Company,
} from '@miapp/data-access';

@Component({
  selector: 'app-contactos',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule, SidebarComponent],
  templateUrl: './contactos.html',
  styleUrl: './contactos.css',
})
export class ContactosComponent implements OnInit {
  private readonly svc     = inject(ContactService);
  private readonly companySvc = inject(CompanyService);

  readonly contacts    = signal<Contact[]>([]);
  readonly companies   = signal<Company[]>([]);
  readonly total       = signal(0);
  readonly loading     = signal(false);
  readonly errorMsg    = signal<string | null>(null);
  readonly successMsg  = signal<string | null>(null);
  readonly showModal   = signal(false);
  readonly saving      = signal(false);
  readonly searchQuery = signal('');

  readonly form = signal<ContactPayload>({
    name: '', email: '', phone: '', position: '',
    status: 'active', company_id: null,
  });

  ngOnInit(): void {
    this.load();
    this.loadCompanies();
  }

  load(): void {
    this.loading.set(true);
    this.errorMsg.set(null);

    const q = this.searchQuery();
    const obs$ = q.length > 1
      ? this.svc.search(q)
      : this.svc.getAll();

    obs$.subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.contacts.set(res.data.list ?? []);
          this.total.set(res.data.total);
        }
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error cargando contactos');
        this.loading.set(false);
      },
    });
  }

  loadCompanies(): void {
    this.companySvc.getAll(0, 100).subscribe({
      next: (res) => {
        if (res.success && res.data) this.companies.set(res.data.list);
      },
    });
  }

  onSearch(event: Event): void {
    this.searchQuery.set((event.target as HTMLInputElement).value);
    this.load();
  }

  openModal(): void {
    this.form.set({ name: '', email: '', phone: '', position: '', status: 'active', company_id: null });
    this.showModal.set(true);
  }

  closeModal(): void { this.showModal.set(false); }

  patchForm(field: keyof ContactPayload, value: string | number | null): void {
    this.form.update(f => ({ ...f, [field]: value }));
  }

  submitCreate(): void {
    const f = this.form();
    if (!f.name?.trim()) { this.errorMsg.set('El nombre es obligatorio'); return; }

    this.saving.set(true);
    this.svc.create(f).subscribe({
      next: (res) => {
        if (res.success) {
          this.successMsg.set('Contacto creado correctamente');
          setTimeout(() => this.successMsg.set(null), 3000);
          this.closeModal();
          this.load();
        }
        this.saving.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error creando contacto');
        this.saving.set(false);
      },
    });
  }

  deleteContact(id: number): void {
    if (!confirm('¿Eliminar este contacto?')) return;
    this.svc.delete(id).subscribe({
      next: () => {
        this.successMsg.set('Contacto eliminado');
        setTimeout(() => this.successMsg.set(null), 3000);
        this.load();
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error eliminando contacto');
      },
    });
  }

  companyName(contact: Contact): string {
    return contact.company?.name ?? '—';
  }
}
