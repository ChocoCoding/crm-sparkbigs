import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from './auth.service';

/** Protege rutas que requieren usuario autenticado */
export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.authenticated()) {
    return true;
  }

  return router.createUrlTree(['/login']);
};

/** Protege rutas que requieren rol admin */
export const adminGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.isAdmin()) {
    return true;
  }

  // Usuario logueado pero sin permisos → redirigir al home
  if (auth.authenticated()) {
    return router.createUrlTree(['/']);
  }

  return router.createUrlTree(['/login']);
};
