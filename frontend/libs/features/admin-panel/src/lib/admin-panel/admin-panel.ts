import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { TranslateModule } from '@ngx-translate/core';
import {
  ApiResponse,
  PaginatedData,
  SidebarComponent,
  User,
  environment,
} from '@miapp/data-access';

@Component({
  selector: 'app-admin-panel',
  standalone: true,
  imports: [CommonModule, TranslateModule, SidebarComponent],
  templateUrl: './admin-panel.html',
  styleUrl: './admin-panel.css',
})
export class AdminPanelComponent implements OnInit {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/admin/users`;

  // Estado con Signals
  readonly users = signal<User[]>([]);
  readonly total = signal(0);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly successMsg = signal<string | null>(null);

  ngOnInit(): void {
    this.loadUsers();
  }

  loadUsers(): void {
    this.loading.set(true);
    this.error.set(null);

    this.http
      .get<ApiResponse<PaginatedData<User>>>(this.apiUrl)
      .subscribe({
        next: (res) => {
          if (res.success && res.data) {
            this.users.set(res.data.list);
            this.total.set(res.data.total);
          }
          this.loading.set(false);
        },
        error: (err: HttpErrorResponse) => {
          this.error.set(
            err.error?.error?.message ?? 'Error cargando usuarios'
          );
          this.loading.set(false);
        },
      });
  }

  deactivateUser(id: number): void {
    this.http
      .delete<ApiResponse<{ message: string }>>(`${this.apiUrl}/${id}`)
      .subscribe({
        next: (res) => {
          if (res.success) {
            this.successMsg.set(res.data?.message ?? 'Usuario desactivado');
            setTimeout(() => this.successMsg.set(null), 3000);
            this.loadUsers();
          }
        },
        error: (err: HttpErrorResponse) => {
          this.error.set(
            err.error?.error?.message ?? 'Error desactivando usuario'
          );
        },
      });
  }

  roleBadgeClass(role: User['role']): string {
    return role === 'admin'
      ? 'bg-purple-100 text-purple-700'
      : 'bg-gray-100 text-gray-600';
  }

  planBadgeClass(plan?: string): string {
    const map: Record<string, string> = {
      free: 'bg-gray-100 text-gray-600',
      pro: 'bg-blue-100 text-blue-700',
      enterprise: 'bg-indigo-100 text-indigo-700',
    };
    return map[plan ?? 'free'] ?? 'bg-gray-100 text-gray-600';
  }
}
