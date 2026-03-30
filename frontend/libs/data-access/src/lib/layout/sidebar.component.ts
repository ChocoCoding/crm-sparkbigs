import { Component, inject, computed, signal } from '@angular/core';
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
  styles: [`
    :host { display: contents; }

    aside {
      background: linear-gradient(180deg, #131b52 0%, #0d1340 100%);
      width: 220px;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      flex-shrink: 0;
      padding: 24px 12px;
    }

    /* Logo */
    .logo-wrap {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 0 8px;
      margin-bottom: 36px;
    }
    .logo-icon {
      width: 32px; height: 32px;
      background: linear-gradient(135deg, #00ced1 0%, #00969a 100%);
      border-radius: 8px;
      display: flex; align-items: center; justify-content: center;
      flex-shrink: 0;
    }
    .logo-text {
      font-family: 'Manrope', sans-serif;
      font-size: 16px;
      font-weight: 800;
      color: #ffffff;
      letter-spacing: -0.02em;
    }
    .logo-sub {
      font-family: 'Inter', sans-serif;
      font-size: 10px;
      color: rgba(255,255,255,0.4);
      font-weight: 400;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }

    /* Nav items */
    nav { flex: 1; display: flex; flex-direction: column; gap: 2px; }

    a.nav-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px 12px;
      border-radius: 10px;
      font-family: 'Inter', sans-serif;
      font-size: 13.5px;
      font-weight: 500;
      color: rgba(255,255,255,0.6);
      text-decoration: none;
      cursor: pointer;
    }
    a.nav-item:hover {
      background: rgba(255,255,255,0.07);
      color: rgba(255,255,255,0.9);
    }
    a.nav-item.active {
      background: rgba(0,206,209,0.18);
      color: #00ced1;
    }
    a.nav-item.active .nav-icon { color: #00ced1; }

    .nav-icon {
      width: 18px; height: 18px;
      flex-shrink: 0;
      color: rgba(255,255,255,0.45);
    }
    a.nav-item:hover .nav-icon { color: rgba(255,255,255,0.8); }

    /* CTA crear nuevo */
    .btn-create {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      width: 100%;
      padding: 11px 16px;
      border-radius: 10px;
      background: linear-gradient(135deg, #00ced1 0%, #00969a 100%);
      color: #fff;
      font-family: 'Inter', sans-serif;
      font-size: 13px;
      font-weight: 600;
      border: none;
      cursor: pointer;
      margin-bottom: 20px;
      box-shadow: 0 4px 16px rgba(0, 206, 209, 0.3);
    }
    .btn-create:hover {
      box-shadow: 0 6px 20px rgba(0, 206, 209, 0.45);
      transform: translateY(-1px);
    }

    /* User profile */
    .user-profile {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px 8px;
      border-radius: 10px;
      cursor: pointer;
    }
    .user-profile:hover { background: rgba(255,255,255,0.06); }
    .user-avatar {
      width: 34px; height: 34px;
      border-radius: 50%;
      background: linear-gradient(135deg, #4c56af, #00ced1);
      display: flex; align-items: center; justify-content: center;
      font-family: 'Inter', sans-serif;
      font-size: 12px;
      font-weight: 700;
      color: #fff;
      flex-shrink: 0;
    }
    .user-name {
      font-family: 'Inter', sans-serif;
      font-size: 12.5px;
      font-weight: 600;
      color: #fff;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .user-role {
      font-family: 'Inter', sans-serif;
      font-size: 10.5px;
      color: rgba(255,255,255,0.4);
      font-weight: 400;
    }

    .logout-btn {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 9px 12px;
      border-radius: 8px;
      font-family: 'Inter', sans-serif;
      font-size: 12.5px;
      font-weight: 500;
      color: rgba(255,255,255,0.4);
      background: none;
      border: none;
      cursor: pointer;
      width: 100%;
      text-align: left;
    }
    .logout-btn:hover {
      background: rgba(255,255,255,0.06);
      color: rgba(255,255,255,0.7);
    }

    .sidebar-divider {
      height: 1px;
      background: rgba(255,255,255,0.07);
      margin: 16px 0 12px;
    }

    /* ── Hamburguesa (solo móvil) ────────────────────────── */
    .mob-toggle {
      display: none;
      position: fixed;
      top: 12px;
      left: 12px;
      z-index: 201;
      width: 40px;
      height: 40px;
      border-radius: 10px;
      background: #131b52;
      border: none;
      cursor: pointer;
      align-items: center;
      justify-content: center;
      color: #fff;
      box-shadow: 0 2px 12px rgba(0,0,0,0.25);
      flex-shrink: 0;
    }

    /* Backdrop */
    .mob-backdrop {
      display: none;
      position: fixed;
      inset: 0;
      background: rgba(0,0,0,0.5);
      backdrop-filter: blur(4px);
      z-index: 199;
    }
    .mob-backdrop.open { display: block; }

    @media (max-width: 768px) {
      .mob-toggle { display: flex; }

      aside {
        position: fixed !important;
        top: 0;
        left: 0;
        height: 100vh;
        z-index: 200;
        transform: translateX(-260px);
        transition: transform 0.25s ease;
        overflow-y: auto;
      }
      aside.mob-open {
        transform: translateX(0);
      }
      /* Botón cerrar visible en móvil */
      .mob-close { display: flex !important; }
    }

    .mob-close {
      display: none;
      margin-left: auto;
      width: 28px;
      height: 28px;
      border-radius: 7px;
      border: none;
      background: rgba(255,255,255,0.1);
      color: rgba(255,255,255,0.7);
      cursor: pointer;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }
  `],
  template: `
    <!-- Botón hamburguesa (solo visible en móvil) -->
    <button class="mob-toggle" (click)="mobileOpen.set(true)" aria-label="Abrir menú">
      <svg width="18" height="18" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16"/>
      </svg>
    </button>

    <!-- Backdrop -->
    <div class="mob-backdrop" [class.open]="mobileOpen()" (click)="mobileOpen.set(false)"></div>

    <aside [class.mob-open]="mobileOpen()">

      <!-- Logo -->
      <div class="logo-wrap">
        <div class="logo-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
            <path d="M13 2L4.5 13.5H11L10 22L20.5 10H14L13 2Z"
              fill="#fff" stroke="#fff" stroke-width="1.5"
              stroke-linejoin="round" stroke-linecap="round"/>
          </svg>
        </div>
        <div>
          <div class="logo-text">SparkBIGS</div>
          <div class="logo-sub">CRM de Datos</div>
        </div>
        <button class="mob-close" (click)="mobileOpen.set(false)" aria-label="Cerrar menú">
          <svg width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <!-- Navegación principal -->
      <nav>
        @for (item of visibleNavItems(); track item.route) {
          <a [routerLink]="item.route"
             routerLinkActive="active"
             [routerLinkActiveOptions]="{ exact: item.exact ?? false }"
             class="nav-item">
            <span class="nav-icon" [innerHTML]="item.icon"></span>
            {{ item.label | translate }}
          </a>
        }
      </nav>

      <!-- CTA + Perfil -->
      <div>
        <button class="btn-create">
          <svg width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
          </svg>
          Crear Nuevo
        </button>

        <div class="sidebar-divider"></div>

        <div class="user-profile">
          <div class="user-avatar">{{ userInitials() }}</div>
          <div style="flex:1;min-width:0;">
            <div class="user-name">{{ userName() }}</div>
            <div class="user-role">{{ userEmail() }}</div>
          </div>
        </div>

        <button class="logout-btn" (click)="onLogout()">
          <svg width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round"
              d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
          </svg>
          {{ 'NAV.LOGOUT' | translate }}
        </button>
      </div>

    </aside>
  `,
})
export class SidebarComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly mobileOpen = signal(false);

  readonly isAdmin = this.auth.isAdmin;
  readonly userName = computed(() => this.auth.user()?.name ?? '');
  readonly userEmail = computed(() => this.auth.user()?.email ?? '');
  readonly userInitials = computed(() => {
    const name = this.auth.user()?.name ?? '';
    return name.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase() || 'U';
  });

  readonly navItems: NavItem[] = [
    {
      label: 'NAV.INICIO', route: '/', exact: true,
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
             </svg>`,
    },
    {
      label: 'NAV.EMPRESAS', route: '/empresas',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
             </svg>`,
    },
    {
      label: 'NAV.CONTACTOS', route: '/contactos',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/>
             </svg>`,
    },
    {
      label: 'NAV.REUNIONES', route: '/reuniones',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
             </svg>`,
    },
    {
      label: 'NAV.FINANZAS', route: '/finanzas',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
             </svg>`,
    },
    {
      label: 'NAV.CONFIGURACION', route: '/configuracion',
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
               <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
             </svg>`,
    },
    {
      label: 'NAV.ADMIN', route: '/admin', adminOnly: true,
      icon: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
               <path stroke-linecap="round" stroke-linejoin="round"
                 d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
             </svg>`,
    },
  ];

  readonly visibleNavItems = computed(() =>
    this.navItems.filter(item => !item.adminOnly || this.isAdmin())
  );

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
