import { Routes } from '@angular/router';
import { authGuard } from '@miapp/data-access';

export const appRoutes: Routes = [
  // ── Pública ──────────────────────────────────────────────────
  {
    path: 'login',
    loadComponent: () =>
      import('@miapp/features/login').then((m) => m.LoginComponent),
  },

  // ── Autenticadas ─────────────────────────────────────────────
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@miapp/features/home-dashboard').then(
        (m) => m.HomeDashboardComponent
      ),
  },
  {
    path: 'empresas',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@miapp/features/empresas').then((m) => m.EmpresasComponent),
  },
  {
    path: 'contactos',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@miapp/features/contactos').then((m) => m.ContactosComponent),
  },
  {
    path: 'reuniones',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@miapp/features/reuniones').then((m) => m.ReunionesComponent),
  },
  {
    path: 'finanzas',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@miapp/features/finanzas').then((m) => m.FinanzasComponent),
  },
  {
    path: 'configuracion',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@miapp/features/configuracion').then(
        (m) => m.ConfiguracionComponent
      ),
  },

  // ── Fallback ─────────────────────────────────────────────────
  { path: '**', redirectTo: '' },
];
