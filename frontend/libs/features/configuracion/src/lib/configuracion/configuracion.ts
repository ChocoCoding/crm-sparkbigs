import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { SidebarComponent, SettingService } from '@miapp/data-access';
import type { Setting } from '@miapp/data-access';

@Component({
  selector: 'app-configuracion',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule, SidebarComponent],
  templateUrl: './configuracion.html',
  styleUrl: './configuracion.css',
})
export class ConfiguracionComponent implements OnInit {
  private svc = inject(SettingService);

  settings   = signal<Setting[]>([]);
  loading    = signal(false);
  saving     = signal<Record<string, boolean>>({});
  savedKeys  = signal<Record<string, boolean>>({});
  errorMsg   = signal<string | null>(null);
  seeded     = signal(false);

  // Valores editables en memoria (key → value)
  editValues = signal<Record<string, string>>({});

  // Categorías únicas para agrupar
  readonly categories = computed(() => {
    const cats = this.settings().map(s => s.category);
    return [...new Set(cats)];
  });

  settingsByCategory(cat: string): Setting[] {
    return this.settings().filter(s => s.category === cat);
  }

  categoryLabel(cat: string): string {
    const map: Record<string, string> = {
      general:        'General',
      notifications:  'Notificaciones',
      integrations:   'Integraciones',
    };
    return map[cat] ?? cat.charAt(0).toUpperCase() + cat.slice(1);
  }

  categoryIcon(cat: string): string {
    const map: Record<string, string> = {
      general:       'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
      notifications: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9',
      integrations:  'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1',
    };
    return map[cat] ?? 'M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4';
  }

  ngOnInit() {
    this.load();
  }

  load() {
    this.loading.set(true);
    this.svc.getAll().subscribe({
      next: res => {
        if (res.success && res.data) {
          this.settings.set(res.data.list);
          // Inicializar valores editables
          const vals: Record<string, string> = {};
          res.data.list.forEach(s => { vals[s.key] = s.value; });
          this.editValues.set(vals);
        }
        this.loading.set(false);
      },
      error: err => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error cargando ajustes');
        this.loading.set(false);
      },
    });
  }

  getValue(key: string): string {
    return this.editValues()[key] ?? '';
  }

  setValue(key: string, value: string) {
    this.editValues.update(v => ({ ...v, [key]: value }));
  }

  save(s: Setting) {
    this.saving.update(m => ({ ...m, [s.key]: true }));
    this.svc.upsert({
      category:   s.category,
      key:        s.key,
      value:      this.getValue(s.key),
      label:      s.label,
      input_type: s.input_type,
    }).subscribe({
      next: () => {
        this.saving.update(m => ({ ...m, [s.key]: false }));
        this.savedKeys.update(m => ({ ...m, [s.key]: true }));
        setTimeout(() => this.savedKeys.update(m => ({ ...m, [s.key]: false })), 2000);
      },
      error: err => {
        this.saving.update(m => ({ ...m, [s.key]: false }));
        this.errorMsg.set(err.error?.error?.message ?? 'Error al guardar');
      },
    });
  }

  seedDefaults() {
    this.svc.seedDefaults().subscribe({
      next: () => { this.seeded.set(true); this.load(); },
      error: err => this.errorMsg.set(err.error?.error?.message ?? 'Error al inicializar'),
    });
  }

  isSaving(key: string): boolean { return !!this.saving()[key]; }
  isSaved(key: string): boolean  { return !!this.savedKeys()[key]; }
}
