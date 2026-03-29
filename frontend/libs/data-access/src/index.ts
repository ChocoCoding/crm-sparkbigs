// Modelos e interfaces
export type {
  ApiResponse,
  PaginatedData,
  User,
  License,
  Company,
  Contact,
  Deal,
  Meeting,
  Subscription,
  Setting,
  AuthTokens,
  AuthLoginData,
} from './lib/models';

// Servicios
export { AuthService } from './lib/auth.service';
export { CompanyService } from './lib/company.service';
export type { CompanyPayload } from './lib/company.service';
export { ContactService } from './lib/contact.service';
export type { ContactPayload } from './lib/contact.service';
export { DealService } from './lib/deal.service';
export type { CreateDealPayload, UpdateDealPayload } from './lib/deal.service';
export { MeetingService } from './lib/meeting.service';
export type { MeetingPayload } from './lib/meeting.service';
export { SubscriptionService } from './lib/subscription.service';
export type { SubscriptionPayload } from './lib/subscription.service';
export { SettingService } from './lib/setting.service';
export type { SettingPayload } from './lib/setting.service';

// Infraestructura
export { authInterceptor } from './lib/auth.interceptor';
export { authGuard, adminGuard } from './lib/auth.guard';
export { environment } from './lib/environment';

// UI compartida
export { SidebarComponent } from './lib/layout/sidebar.component';
