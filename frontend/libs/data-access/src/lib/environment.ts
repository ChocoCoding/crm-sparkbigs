// Cambia production a true y apunta al dominio cuando hagas el build de producción.
// Para desarrollo local mantén los valores comentados debajo.
export const environment = {
  production: true,
  apiUrl: 'https://striking-forgiveness-production.up.railway.app/api/v1',
  backendUrl: 'https://striking-forgiveness-production.up.railway.app',
};

// ── Desarrollo local ──────────────────────────────────────
// export const environment = {
//   production: false,
//   apiUrl: 'http://localhost:8080/api/v1',
//   backendUrl: 'http://localhost:8080',
// };
