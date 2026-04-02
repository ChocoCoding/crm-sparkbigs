import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import {
  MeetingService, MeetingPayload,
  CompanyService, ContactService,
  SidebarComponent,
  Meeting, Company, Contact,
} from '@miapp/data-access';

interface CalendarDay {
  date: Date;
  isCurrentMonth: boolean;
  isToday: boolean;
  meetings: Meeting[];
}

@Component({
  selector: 'app-reuniones',
  standalone: true,
  imports: [CommonModule, FormsModule, SidebarComponent],
  templateUrl: './reuniones.html',
  styleUrl: './reuniones.css',
})
export class ReunionesComponent implements OnInit {
  private readonly svc        = inject(MeetingService);
  private readonly companySvc = inject(CompanyService);
  private readonly contactSvc = inject(ContactService);

  // ─── Estado principal ────────────────────────────────────────
  readonly meetings   = signal<Meeting[]>([]);
  readonly upcoming   = signal<Meeting[]>([]);
  readonly companies  = signal<Company[]>([]);
  readonly contacts   = signal<Contact[]>([]);
  readonly total      = signal(0);
  readonly loading    = signal(false);
  readonly errorMsg   = signal<string | null>(null);
  readonly successMsg = signal<string | null>(null);
  readonly showModal  = signal(false);
  readonly saving     = signal(false);
  readonly activeTab  = signal<'calendar' | 'list'>('calendar');
  readonly editingId  = signal<number | null>(null);
  readonly isEditing  = computed(() => this.editingId() !== null);

  // ─── Calendario ──────────────────────────────────────────────
  readonly calendarDate = signal(new Date());

  readonly calendarTitle = computed(() => {
    return this.calendarDate().toLocaleDateString('es-ES', { month: 'long', year: 'numeric' });
  });

  readonly scheduledCount = computed(() => this.meetings().filter(m => m.status === 'scheduled').length);
  readonly completedCount = computed(() => this.meetings().filter(m => m.status === 'completed').length);
  readonly cancelledCount = computed(() => this.meetings().filter(m => m.status === 'cancelled').length);

  readonly calendarDays = computed((): CalendarDay[] => {
    const today = new Date();
    const ref   = this.calendarDate();
    const year  = ref.getFullYear();
    const month = ref.getMonth();

    const firstDay = new Date(year, month, 1);
    const lastDay  = new Date(year, month + 1, 0);

    // Día de la semana del primero (lunes = 0)
    const startPad = (firstDay.getDay() + 6) % 7;

    const days: CalendarDay[] = [];

    // Días del mes anterior
    for (let i = startPad - 1; i >= 0; i--) {
      const d = new Date(year, month, -i);
      days.push({ date: d, isCurrentMonth: false, isToday: false, meetings: [] });
    }

    // Días del mes actual
    for (let d = 1; d <= lastDay.getDate(); d++) {
      const date = new Date(year, month, d);
      const isToday = date.toDateString() === today.toDateString();
      const dayMeetings = this.meetings().filter(m => {
        const mDate = new Date(m.start_at);
        return mDate.getFullYear() === year &&
               mDate.getMonth() === month &&
               mDate.getDate() === d;
      });
      days.push({ date, isCurrentMonth: true, isToday, meetings: dayMeetings });
    }

    // Completar hasta múltiplo de 7
    const remaining = (7 - (days.length % 7)) % 7;
    for (let d = 1; d <= remaining; d++) {
      const date = new Date(year, month + 1, d);
      days.push({ date, isCurrentMonth: false, isToday: false, meetings: [] });
    }

    return days;
  });

  // ─── Formulario ──────────────────────────────────────────────
  readonly form = signal<MeetingPayload & { start_date: string; start_time: string }>({
    title: '', company_id: 0, contact_id: null,
    start_at: '', start_date: '', start_time: '10:00',
    duration_min: 60, status: 'scheduled', notes: '',
  });

  ngOnInit(): void {
    this.load();
    this.loadUpcoming();
    this.loadCompanies();
    this.loadContacts();
  }

  load(): void {
    this.loading.set(true);
    this.svc.getAll(0, 100).subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.meetings.set(res.data.list ?? []);
          this.total.set(res.data.total);
        }
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error cargando reuniones');
        this.loading.set(false);
      },
    });
  }

  loadUpcoming(): void {
    this.svc.getUpcoming(5).subscribe({
      next: (res) => { if (res.success && res.data) this.upcoming.set(res.data.list ?? []); },
    });
  }

  loadCompanies(): void {
    this.companySvc.getAll(0, 100).subscribe({
      next: (res) => { if (res.success && res.data) this.companies.set(res.data.list); },
    });
  }

  loadContacts(): void {
    this.contactSvc.getAll(0, 100).subscribe({
      next: (res) => { if (res.success && res.data) this.contacts.set(res.data.list); },
    });
  }

  // ─── Calendario ──────────────────────────────────────────────
  prevMonth(): void {
    const d = this.calendarDate();
    this.calendarDate.set(new Date(d.getFullYear(), d.getMonth() - 1, 1));
  }

  nextMonth(): void {
    const d = this.calendarDate();
    this.calendarDate.set(new Date(d.getFullYear(), d.getMonth() + 1, 1));
  }

  goToday(): void {
    this.calendarDate.set(new Date());
  }

  // ─── Modal ───────────────────────────────────────────────────
  openModal(): void {
    this.editingId.set(null);
    const today = new Date().toISOString().split('T')[0];
    this.form.set({
      title: '', company_id: 0, contact_id: null,
      start_at: '', start_date: today, start_time: '10:00',
      duration_min: 60, status: 'scheduled', notes: '',
    });
    this.showModal.set(true);
  }

  openEdit(meeting: Meeting): void {
    this.editingId.set(meeting.id);
    const d = new Date(meeting.start_at);
    const start_date = d.toISOString().split('T')[0];
    const start_time = d.toTimeString().slice(0, 5);
    this.form.set({
      title: meeting.title, company_id: meeting.company_id,
      contact_id: meeting.contact_id, start_at: meeting.start_at,
      start_date, start_time, duration_min: meeting.duration_min,
      status: meeting.status, notes: meeting.notes,
    });
    this.showModal.set(true);
  }

  closeModal(): void { this.showModal.set(false); this.editingId.set(null); }

  patchForm(field: string, value: string | number | null): void {
    this.form.update(f => ({ ...f, [field]: value }));
  }

  submit(): void {
    const f = this.form();
    if (!f.title?.trim()) { this.errorMsg.set('El título es obligatorio'); return; }
    if (!f.company_id)    { this.errorMsg.set('La empresa es obligatoria'); return; }
    if (!f.start_date)    { this.errorMsg.set('La fecha es obligatoria'); return; }

    const startAt = new Date(`${f.start_date}T${f.start_time}:00`).toISOString();
    const payload: MeetingPayload = {
      title: f.title, company_id: f.company_id, contact_id: f.contact_id || null,
      start_at: startAt, duration_min: f.duration_min, status: f.status, notes: f.notes,
    };

    this.saving.set(true);
    const id = this.editingId();
    const req$ = id ? this.svc.update(id, payload) : this.svc.create(payload);
    req$.subscribe({
      next: (res) => {
        if (res.success) {
          this.successMsg.set(id ? 'Reunión actualizada' : 'Reunión creada');
          setTimeout(() => this.successMsg.set(null), 3000);
          this.closeModal();
          this.load();
          this.loadUpcoming();
        }
        this.saving.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error guardando reunión');
        this.saving.set(false);
      },
    });
  }

  deleteMeeting(id: number): void {
    if (!confirm('¿Eliminar esta reunión?')) return;
    this.svc.delete(id).subscribe({
      next: () => {
        this.successMsg.set('Reunión eliminada');
        setTimeout(() => this.successMsg.set(null), 3000);
        this.load();
        this.loadUpcoming();
      },
      error: (err: HttpErrorResponse) => {
        this.errorMsg.set(err.error?.error?.message ?? 'Error');
      },
    });
  }

  // ─── Helpers de presentación ─────────────────────────────────
  formatTime(isoString: string): string {
    return new Date(isoString).toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' });
  }

  formatDateTime(isoString: string): string {
    return new Date(isoString).toLocaleString('es-ES', {
      day: '2-digit', month: 'short', year: 'numeric',
      hour: '2-digit', minute: '2-digit',
    });
  }

  statusLabel(status: Meeting['status']): string {
    return { scheduled: 'Programada', completed: 'Completada', cancelled: 'Cancelada' }[status] ?? status;
  }

  statusColor(status: Meeting['status']): string {
    return {
      scheduled: 'background:#dbeafe;color:#1d4ed8',
      completed: 'background:#dcfce7;color:#15803d',
      cancelled: 'background:#f3f4f6;color:#6b7280',
    }[status] ?? '';
  }

  meetingDotColor(status: Meeting['status']): string {
    return { scheduled: '#1A237E', completed: '#00C853', cancelled: '#9ca3af' }[status] ?? '#9ca3af';
  }
}
