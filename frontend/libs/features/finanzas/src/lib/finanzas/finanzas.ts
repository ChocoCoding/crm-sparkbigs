import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { SidebarComponent, SubscriptionService, CompanyService, SubscriptionPayload } from '@miapp/data-access';
import type { Subscription, Company } from '@miapp/data-access';

@Component({
  selector: 'app-finanzas',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule, SidebarComponent],
  templateUrl: './finanzas.html',
  styleUrl: './finanzas.css',
})
export class FinanzasComponent implements OnInit {
  private svc = inject(SubscriptionService);
  private companySvc = inject(CompanyService);

  // ── Estado ──────────────────────────────────────────────
  subscriptions = signal<Subscription[]>([]);
  expiring       = signal<Subscription[]>([]);
  companies      = signal<Company[]>([]);
  loading        = signal(false);
  errorMsg       = signal<string | null>(null);
  showModal      = signal(false);
  editTarget     = signal<Subscription | null>(null);

  // ── Formulario ──────────────────────────────────────────
  form = signal<SubscriptionPayload>({
    company_id: 0,
    name: '',
    plan_type: '',
    status: 'active',
    amount: 0,
    currency: 'EUR',
    billing_cycle: 'monthly',
    start_date: new Date().toISOString().substring(0, 10),
    renewal_date: null,
    notes: '',
  });

  // ── Computed ─────────────────────────────────────────────
  readonly totalMRR = computed(() =>
    this.subscriptions()
      .filter(s => s.status === 'active' && s.billing_cycle === 'monthly')
      .reduce((acc, s) => acc + s.amount, 0)
  );
  readonly totalARR = computed(() =>
    this.subscriptions()
      .filter(s => s.status === 'active' && s.billing_cycle === 'annual')
      .reduce((acc, s) => acc + s.amount, 0)
  );
  readonly activeCount = computed(() =>
    this.subscriptions().filter(s => s.status === 'active').length
  );

  ngOnInit() {
    this.load();
    this.loadExpiring();
    this.loadCompanies();
  }

  load() {
    this.loading.set(true);
    this.errorMsg.set(null);
    this.svc.getAll().subscribe({
      next: res => {
        if (res.success && res.data) this.subscriptions.set(res.data.list);
        this.loading.set(false);
      },
      error: err => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error cargando suscripciones');
        this.loading.set(false);
      },
    });
  }

  loadExpiring() {
    this.svc.getExpiringSoon(30).subscribe({
      next: res => { if (res.success && res.data) this.expiring.set(res.data.list); },
    });
  }

  loadCompanies() {
    this.companySvc.getAll().subscribe({
      next: res => { if (res.success && res.data) this.companies.set(res.data.list); },
    });
  }

  openCreate() {
    this.editTarget.set(null);
    this.form.set({
      company_id: 0,
      name: '',
      plan_type: '',
      status: 'active',
      amount: 0,
      currency: 'EUR',
      billing_cycle: 'monthly',
      start_date: new Date().toISOString().substring(0, 10),
      renewal_date: null,
      notes: '',
    });
    this.showModal.set(true);
  }

  openEdit(s: Subscription) {
    this.editTarget.set(s);
    this.form.set({
      company_id: s.company_id,
      name: s.name,
      plan_type: s.plan_type,
      status: s.status,
      amount: s.amount,
      currency: s.currency,
      billing_cycle: s.billing_cycle,
      start_date: s.start_date?.substring(0, 10) ?? '',
      renewal_date: s.renewal_date?.substring(0, 10) ?? null,
      notes: s.notes,
    });
    this.showModal.set(true);
  }

  patchForm(field: string, value: unknown) {
    this.form.update(f => ({ ...f, [field]: value }));
  }

  submit() {
    const f = this.form();
    if (!f.name || !f.company_id) return;
    const target = this.editTarget();
    const obs = target
      ? this.svc.update(target.id, f)
      : this.svc.create(f);
    obs.subscribe({
      next: () => { this.showModal.set(false); this.load(); this.loadExpiring(); },
      error: err => this.errorMsg.set(err.error?.error?.message ?? 'Error al guardar'),
    });
  }

  delete(id: number) {
    if (!confirm('¿Eliminar esta suscripción?')) return;
    this.svc.delete(id).subscribe({ next: () => { this.load(); this.loadExpiring(); } });
  }

  statusColor(status: string): string {
    const map: Record<string, string> = {
      active:    'background:#e8f5e9;color:#006e2a;',
      cancelled: 'background:#fce8e6;color:#b3261e;',
      paused:    'background:#fff8e1;color:#b45309;',
      expired:   'background:var(--surface-low,#edf0f8);color:var(--on-surface-subtle,#9a9db8);',
    };
    return map[status] ?? 'background:var(--surface-low,#edf0f8);color:var(--on-surface-subtle,#9a9db8);';
  }

  statusLabel(status: string): string {
    const map: Record<string, string> = {
      active: 'Activa', cancelled: 'Cancelada', paused: 'Pausada', expired: 'Expirada',
    };
    return map[status] ?? status;
  }

  cycleLabel(cycle: string): string {
    const map: Record<string, string> = { monthly: 'Mensual', annual: 'Anual', one_time: 'Pago único' };
    return map[cycle] ?? cycle;
  }

  daysUntil(dateStr: string | null): number {
    if (!dateStr) return 999;
    const diff = new Date(dateStr).getTime() - Date.now();
    return Math.ceil(diff / 86400000);
  }
}
