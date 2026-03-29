import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, switchMap, throwError } from 'rxjs';
import { AuthService } from './auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);

  // Las rutas de auth no necesitan token
  if (req.url.includes('/auth/login') || req.url.includes('/auth/refresh')) {
    return next(req);
  }

  const token = auth.getAccessToken();
  const authReq = token
    ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
    : req;

  return next(authReq).pipe(
    catchError((error: HttpErrorResponse) => {
      // Si 401 y había token, intentar refresh automático
      if (error.status === 401 && token) {
        return auth.refreshToken().pipe(
          switchMap((res) => {
            if (res.success && res.data) {
              const retryReq = req.clone({
                setHeaders: {
                  Authorization: `Bearer ${res.data.access_token}`,
                },
              });
              return next(retryReq);
            }
            // Refresh falló → cerrar sesión
            auth.clearSession();
            return throwError(() => error);
          }),
          catchError((refreshError) => {
            auth.clearSession();
            return throwError(() => refreshError);
          })
        );
      }
      return throwError(() => error);
    })
  );
};
