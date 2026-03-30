import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { DashboardService, DashboardStats, SidebarComponent } from '@miapp/data-access';

interface MrrBar { pct: number; color: string; label: string; }

@Component({
  selector: 'app-home-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, TranslateModule, SidebarComponent],
  templateUrl: './home-dashboard.html',
  styleUrl: './home-dashboard.css',
})
export class HomeDashboardComponent implements OnInit {
  private svc = inject(DashboardService);

  stats    = signal<DashboardStats | null>(null);
  loading  = signal(true);
  errorMsg = signal<string | null>(null);

  readonly greeting = computed(() => {
    const h = new Date().getHours();
    if (h < 12) return 'Buenos días';
    if (h < 20) return 'Buenas tardes';
    return 'Buenas noches';
  });

  // ── Barras del gráfico MRR (4 meses decorativos con el valor real como base) ──
  readonly mrrBars = computed((): MrrBar[] => {
    const mrr = this.stats()?.mrr ?? 0;
    // Tendencia ascendente: 65% → 78% → 90% → 100% del MRR actual
    const proportions = [0.65, 0.78, 0.90, 1.0];
    const labels = ['ENE', 'FEB', 'MAR', 'ABR'];
    const colors = ['#00ced180', '#00ced199', '#00ced1b3', '#00ced1'];
    return proportions.map((p, i) => ({
      pct: mrr === 0 ? (40 + i * 15) : Math.round(p * 100), // si no hay datos, barras demo
      color: colors[i],
      label: labels[i],
    }));
  });

  // Porcentaje de suscripciones activas para la barra de desglose
  readonly activeSubPct = computed(() => {
    const active = this.stats()?.active_subscriptions ?? 0;
    const expiring = this.stats()?.expiring_soon?.length ?? 0;
    const total = active + expiring;
    if (total === 0) return 70; // demo
    return Math.round((active / total) * 100);
  });

  readonly expiringSubPct = computed(() => 100 - this.activeSubPct());

  ngOnInit() {
    this.svc.getStats().subscribe({
      next: res => {
        if (res.success && res.data) this.stats.set(res.data.stats);
        this.loading.set(false);
      },
      error: err => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error cargando el dashboard');
        this.loading.set(false);
      },
    });
  }

  daysUntil(dateStr: string | null): number {
    if (!dateStr) return 999;
    return Math.ceil((new Date(dateStr).getTime() - Date.now()) / 86400000);
  }

  companyStatusStyle(status: string): string {
    const map: Record<string, string> = {
      active:   'background:#e8f5e9;color:#006e2a;',
      prospect: 'background:#e8eaf6;color:#4c56af;',
      inactive: 'background:var(--surface-low,#edf0f8);color:var(--on-surface-subtle,#9a9db8);',
    };
    return map[status] ?? map['inactive'];
  }
}
