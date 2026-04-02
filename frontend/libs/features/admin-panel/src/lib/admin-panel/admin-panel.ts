import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { TranslateModule } from '@ngx-translate/core';
import {
  ApiResponse,
  PaginatedData,
  SidebarComponent,
  User,
  License,
  environment,
} from '@miapp/data-access';

interface CreateUserPayload {
  email: string;
  name: string;
  password: string;
  role: 'admin' | 'user';
}

interface UpdateUserPayload {
  name: string;
  role: 'admin' | 'user';
  is_active: boolean;
}

interface LicensePayload {
  plan: 'free' | 'pro' | 'enterprise';
}

@Component({
  selector: 'app-admin-panel',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule, SidebarComponent],
  templateUrl: './admin-panel.html',
  styleUrl: './admin-panel.css',
})
export class AdminPanelComponent implements OnInit {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/admin/users`;

  // ── Lista ──────────────────────────────────────────────────────────────────
  readonly users      = signal<User[]>([]);
  readonly total      = signal(0);
  readonly loading    = signal(false);
  readonly error      = signal<string | null>(null);
  readonly successMsg = signal<string | null>(null);
  readonly search     = signal('');

  readonly filteredUsers = computed(() => {
    const q = this.search().toLowerCase();
    if (!q) return this.users();
    return this.users().filter(
      u => u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q)
    );
  });

  // ── Modal Crear ────────────────────────────────────────────────────────────
  readonly showCreateModal = signal(false);
  readonly createForm      = signal<CreateUserPayload>({ email: '', name: '', password: '', role: 'user' });
  readonly createLoading   = signal(false);

  // ── Modal Editar ───────────────────────────────────────────────────────────
  readonly showEditModal = signal(false);
  readonly editTarget    = signal<User | null>(null);
  readonly editForm      = signal<UpdateUserPayload>({ name: '', role: 'user', is_active: true });
  readonly editLoading   = signal(false);

  // ── Modal Licencia ─────────────────────────────────────────────────────────
  readonly showLicenseModal = signal(false);
  readonly licenseTarget    = signal<User | null>(null);
  readonly licenseForm      = signal<LicensePayload>({ plan: 'free' });
  readonly licenseLoading   = signal(false);

  // ── Confirmación Eliminar ──────────────────────────────────────────────────
  readonly confirmDeleteId = signal<number | null>(null);

  ngOnInit(): void { this.loadUsers(); }

  loadUsers(): void {
    this.loading.set(true);
    this.error.set(null);
    this.http.get<ApiResponse<PaginatedData<User>>>(this.apiUrl).subscribe({
      next: res => {
        if (res.success && res.data) {
          this.users.set(res.data.list);
          this.total.set(res.data.total);
        }
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.error?.message ?? 'Error cargando usuarios');
        this.loading.set(false);
      },
    });
  }

  // ── Crear ──────────────────────────────────────────────────────────────────
  openCreate(): void {
    this.createForm.set({ email: '', name: '', password: '', role: 'user' });
    this.showCreateModal.set(true);
  }

  submitCreate(): void {
    const f = this.createForm();
    if (!f.email || !f.name || !f.password) return;
    this.createLoading.set(true);
    this.http.post<ApiResponse<{ user: User }>>(this.apiUrl, f).subscribe({
      next: res => {
        if (res.success) { this.showCreateModal.set(false); this.showSuccess('Usuario creado correctamente'); this.loadUsers(); }
        this.createLoading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.error?.message ?? 'Error creando usuario');
        this.createLoading.set(false);
      },
    });
  }

  // ── Editar ─────────────────────────────────────────────────────────────────
  openEdit(user: User): void {
    this.editTarget.set(user);
    this.editForm.set({ name: user.name, role: user.role, is_active: user.is_active });
    this.showEditModal.set(true);
  }

  submitEdit(): void {
    const target = this.editTarget();
    if (!target) return;
    this.editLoading.set(true);
    this.http.put<ApiResponse<{ user: User }>>(`${this.apiUrl}/${target.id}`, this.editForm()).subscribe({
      next: res => {
        if (res.success) { this.showEditModal.set(false); this.showSuccess('Usuario actualizado correctamente'); this.loadUsers(); }
        this.editLoading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.error?.message ?? 'Error actualizando usuario');
        this.editLoading.set(false);
      },
    });
  }

  // ── Licencia ───────────────────────────────────────────────────────────────
  openLicense(user: User): void {
    this.licenseTarget.set(user);
    this.licenseForm.set({ plan: user.license?.plan ?? 'free' });
    this.showLicenseModal.set(true);
  }

  submitLicense(): void {
    const target = this.licenseTarget();
    if (!target) return;
    this.licenseLoading.set(true);
    this.http.put<ApiResponse<{ license: License }>>(`${this.apiUrl}/${target.id}/license`, this.licenseForm()).subscribe({
      next: res => {
        if (res.success) { this.showLicenseModal.set(false); this.showSuccess('Licencia actualizada correctamente'); this.loadUsers(); }
        this.licenseLoading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.error?.message ?? 'Error actualizando licencia');
        this.licenseLoading.set(false);
      },
    });
  }

  // ── Eliminar / Desactivar ──────────────────────────────────────────────────
  confirmDelete(id: number): void { this.confirmDeleteId.set(id); }
  cancelDelete(): void            { this.confirmDeleteId.set(null); }

  executeDelete(): void {
    const id = this.confirmDeleteId();
    if (id === null) return;
    this.http.delete<ApiResponse<{ message: string }>>(`${this.apiUrl}/${id}`).subscribe({
      next: res => {
        if (res.success) { this.confirmDeleteId.set(null); this.showSuccess(res.data?.message ?? 'Usuario desactivado'); this.loadUsers(); }
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error?.error?.message ?? 'Error desactivando usuario');
        this.confirmDeleteId.set(null);
      },
    });
  }

  // ── Badges / Estilos ───────────────────────────────────────────────────────
  roleBadgeStyle(role: User['role']): string {
    return role === 'admin'
      ? 'background:#ede7f6;color:#5e35b1;'
      : 'background:var(--surface-low,#edf0f8);color:var(--on-surface-subtle,#9a9db8);';
  }

  planBadgeStyle(plan?: string): string {
    const map: Record<string, string> = {
      free:       'background:var(--surface-low,#edf0f8);color:var(--on-surface-subtle,#9a9db8);',
      pro:        'background:#e3f2fd;color:#1565c0;',
      enterprise: 'background:#e8f5e9;color:#2e7d32;',
    };
    return map[plan ?? 'free'] ?? map['free'];
  }

  activeStyle(active: boolean): string {
    return active ? 'background:#e8f5e9;color:#2e7d32;' : 'background:#fce8e6;color:#b3261e;';
  }

  // ── Helpers ────────────────────────────────────────────────────────────────
  private showSuccess(msg: string): void {
    this.successMsg.set(msg);
    setTimeout(() => this.successMsg.set(null), 3500);
  }

  updateCreateField(field: keyof CreateUserPayload, value: string): void {
    this.createForm.update(f => ({ ...f, [field]: value }));
  }

  updateEditField<K extends keyof UpdateUserPayload>(field: K, value: UpdateUserPayload[K]): void {
    this.editForm.update(f => ({ ...f, [field]: value }));
  }

  updateLicensePlan(plan: 'free' | 'pro' | 'enterprise'): void {
    this.licenseForm.update(f => ({ ...f, plan }));
  }
}
