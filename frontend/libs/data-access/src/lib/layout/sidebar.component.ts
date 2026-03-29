import { Component, inject, computed } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule, TranslateModule],
  template: `
    <aside class="flex flex-col w-64 min-h-screen bg-indigo-900 text-white px-4 py-6">
      <!-- Logo -->
      <div class="mb-8 px-2">
        <span class="text-xl font-bold tracking-tight">
          {{ 'APP.NAME' | translate }}
        </span>
      </div>

      <!-- Nav -->
      <nav class="flex-1 space-y-1">
        <a routerLink="/"
           routerLinkActive="bg-indigo-700"
           [routerLinkActiveOptions]="{ exact: true }"
           class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
          </svg>
          {{ 'APP.SIDEBAR.DASHBOARD' | translate }}
        </a>

        @if (isAdmin()) {
          <a routerLink="/admin"
             routerLinkActive="bg-indigo-700"
             class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"/>
            </svg>
            {{ 'APP.SIDEBAR.ADMIN' | translate }}
          </a>
        }
      </nav>

      <!-- User + Logout -->
      <div class="border-t border-indigo-700 pt-4 mt-4">
        <p class="text-xs text-indigo-300 px-2 mb-2 truncate">{{ userEmail() }}</p>
        <button (click)="onLogout()"
          class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700 transition-colors text-left">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
          </svg>
          {{ 'APP.SIDEBAR.LOGOUT' | translate }}
        </button>
      </div>
    </aside>
  `,
})
export class SidebarComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly isAdmin = this.auth.isAdmin;
  readonly userEmail = computed(() => this.auth.user()?.email ?? '');

  onLogout(): void {
    this.auth.logout().subscribe({
      next: () => this.router.navigate(['/login']),
      error: () => {
        this.auth.clearSession();
        this.router.navigate(['/login']);
      },
    });
  }
}
