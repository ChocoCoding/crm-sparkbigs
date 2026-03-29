import { Component, inject, computed } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../auth.service';

interface NavItem {
  label: string;
  route: string;
  icon: string;
  exact?: boolean;
  adminOnly?: boolean;
}

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule, TranslateModule],
  template: `
    <aside class="flex flex-col w-64 min-h-screen bg-slate-900 text-white px-3 py-5 shrink-0">

      <!-- Logo -->
      <div class="px-3 mb-8">
        <p class="text-xs font-semibold text-slate-400 uppercase tracking-widest mb-0.5">SparkBIGS</p>
        <p class="text-base font-bold text-white leading-tight">CRM de Datos</p>
      </div>

      <!-- Navegación principal -->
      <nav class="flex-1 space-y-0.5">
        @for (item of navItems; track item.route) {
          @if (!item.adminOnly || isAdmin()) {
            <a [routerLink]="item.route"
               routerLinkActive="bg-slate-700 text-white"
               [routerLinkActiveOptions]="{ exact: item.exact ?? false }"
               class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-slate-300 hover:bg-slate-700 hover:text-white transition-colors">
              <span class="w-5 h-5 shrink-0" [innerHTML]="item.icon"></span>
              {{ item.label | translate }}
            </a>
          }
        }
      </nav>

      <!-- Footer: usuario + logout -->
      <div class="border-t border-slate-700 pt-4 mt-4 space-y-1">
        <div class="px-3 mb-2">
          <p class="text-xs font-medium text-white truncate">{{ userName() }}</p>
          <p class="text-xs text-slate-400 truncate">{{ userEmail() }}</p>
        </div>
        <button (click)="onLogout()"
          class="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-slate-300 hover:bg-slate-700 hover:text-white transition-colors text-left">
          <span class="w-5 h-5 shrink-0" [innerHTML]="iconLogout"></span>
          {{ 'NAV.LOGOUT' | translate }}
        </button>
      </div>
    </aside>
  `,
})
export class SidebarComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly isAdmin = this.auth.isAdmin;
  readonly userName = computed(() => this.auth.user()?.name ?? '');
  readonly userEmail = computed(() => this.auth.user()?.email ?? '');

  readonly navItems: NavItem[] = [
    {
      label: 'NAV.INICIO',
      route: '/',
      exact: true,
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
             </svg>`,
    },
    {
      label: 'NAV.EMPRESAS',
      route: '/empresas',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
             </svg>`,
    },
    {
      label: 'NAV.CONTACTOS',
      route: '/contactos',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
             </svg>`,
    },
    {
      label: 'NAV.REUNIONES',
      route: '/reuniones',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
             </svg>`,
    },
    {
      label: 'NAV.FINANZAS',
      route: '/finanzas',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
             </svg>`,
    },
    {
      label: 'NAV.CONFIGURACION',
      route: '/configuracion',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                 <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
             </svg>`,
    },
  ];

  readonly iconLogout = `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
    <path stroke-linecap="round" stroke-linejoin="round"
      d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
  </svg>`;

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
