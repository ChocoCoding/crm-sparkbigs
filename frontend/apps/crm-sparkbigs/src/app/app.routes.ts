import { Routes } from '@angular/router';
import { authGuard, adminGuard } from '@miapp/data-access';

export const appRoutes: Routes = [
  {
    path: 'login',
    loadComponent: () =>
      import('@miapp/features/login').then((m) => m.LoginComponent),
  },
  {
    path: 'admin',
    canActivate: [adminGuard],
    loadComponent: () =>
      import('@miapp/features/admin-panel').then((m) => m.AdminPanelComponent),
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@miapp/features/home-dashboard').then(
        (m) => m.HomeDashboardComponent
      ),
  },
  { path: '**', redirectTo: '' },
];
