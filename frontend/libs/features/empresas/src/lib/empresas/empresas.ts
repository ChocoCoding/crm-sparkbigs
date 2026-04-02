import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { TranslateModule } from '@ngx-translate/core';
import {
  CompanyService,
  CompanyPayload,
  SidebarComponent,
  Company,
} from '@miapp/data-access';

@Component({
  selector: 'app-empresas',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule, SidebarComponent],
  templateUrl: './empresas.html',
  styleUrl: './empresas.css',
})
export class EmpresasComponent implements OnInit {
  private readonly svc = inject(CompanyService);

  readonly companies  = signal<Company[]>([]);
  readonly total      = signal(0);
  readonly loading    = signal(false);
  readonly errorMsg   = signal<string | null>(null);
  readonly successMsg = signal<string | null>(null);
  readonly showModal  = signal(false);
  readonly saving     = signal(false);
  readonly searchQuery = signal('');
  readonly isSearching = computed(() => this.searchQuery().length > 1);
  readonly editingId  = signal<number | null>(null);
  readonly isEditing  = computed(() => this.editingId() !== null);

  readonly form = signal<CompanyPayload>({
    name: '', sector: '', status: 'prospect',
    website: '', phone: '', address: '', relation_start_date: null,
  });

  readonly sectors = [
    'Software', 'Retail', 'Finanzas', 'Logística', 'Salud',
    'Biotecnología', 'Educación', 'Manufactura', 'Consultoría', 'Otro',
  ];

  ngOnInit(): void { this.load(); }

  load(): void {
    this.loading.set(true);
    this.errorMsg.set(null);
    const obs$ = this.isSearching()
      ? this.svc.search(this.searchQuery())
      : this.svc.getAll();

    obs$.subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.companies.set(res.data.list ?? []);
          this.total.set(res.data.total);
        }
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error cargando empresas');
        this.loading.set(false);
      },
    });
  }

  onSearch(event: Event): void {
    this.searchQuery.set((event.target as HTMLInputElement).value);
    this.load();
  }

  openModal(): void {
    this.editingId.set(null);
    this.form.set({ name: '', sector: '', status: 'prospect', website: '', phone: '', address: '', relation_start_date: null });
    this.showModal.set(true);
  }

  openEdit(company: Company): void {
    this.editingId.set(company.id);
    this.form.set({
      name: company.name, sector: company.sector, status: company.status,
      website: company.website, phone: company.phone, address: company.address,
      relation_start_date: company.relation_start_date,
    });
    this.showModal.set(true);
  }

  closeModal(): void { this.showModal.set(false); this.editingId.set(null); }

  patchForm(field: keyof CompanyPayload, value: string): void {
    this.form.update(f => ({ ...f, [field]: value }));
  }

  submit(): void {
    const f = this.form();
    if (!f.name?.trim()) { this.errorMsg.set('El nombre es obligatorio'); return; }
    this.saving.set(true);
    const id = this.editingId();
    const req$ = id ? this.svc.update(id, f) : this.svc.create(f);
    req$.subscribe({
      next: (res) => {
        if (res.success) {
          this.successMsg.set(id ? 'Empresa actualizada' : 'Empresa creada');
          setTimeout(() => this.successMsg.set(null), 3000);
          this.closeModal();
          this.load();
        }
        this.saving.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error guardando empresa');
        this.saving.set(false);
      },
    });
  }

  deleteCompany(id: number): void {
    if (!confirm('¿Eliminar esta empresa?')) return;
    this.svc.delete(id).subscribe({
      next: () => { this.successMsg.set('Empresa eliminada'); setTimeout(() => this.successMsg.set(null), 3000); this.load(); },
      error: (err: HttpErrorResponse) => { this.errorMsg.set(err.error?.error?.message ?? 'Error'); },
    });
  }

  statusStyle(status: Company['status']): string {
    const map: Record<Company['status'], string> = {
      active:   'background:#e8f5e9;color:#006e2a;',
      prospect: 'background:#e8eaf6;color:#4c56af;',
      inactive: 'background:var(--surface-low,#edf0f8);color:var(--on-surface-subtle,#9a9db8);',
    };
    return map[status] ?? map['inactive'];
  }

  statusLabel(status: Company['status']): string {
    return { active: 'Activo', prospect: 'Prospecto', inactive: 'Inactivo' }[status] ?? status;
  }
}
