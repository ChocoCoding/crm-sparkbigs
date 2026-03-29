// Modelos e interfaces
export type {
  ApiResponse,
  PaginatedData,
  User,
  License,
  Contact,
  Deal,
  AuthTokens,
  AuthLoginData,
} from './lib/models';

// Servicios
export { AuthService } from './lib/auth.service';
export { ContactService } from './lib/contact.service';
export type {
  CreateContactPayload,
  UpdateContactPayload,
} from './lib/contact.service';
export { DealService } from './lib/deal.service';
export type { CreateDealPayload, UpdateDealPayload } from './lib/deal.service';

// Infraestructura
export { authInterceptor } from './lib/auth.interceptor';
export { authGuard, adminGuard } from './lib/auth.guard';
export { environment } from './lib/environment';

// UI compartida
export { SidebarComponent } from './lib/layout/sidebar.component';
