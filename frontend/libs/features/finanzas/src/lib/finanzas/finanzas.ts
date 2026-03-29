import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { SidebarComponent } from '@miapp/data-access';

@Component({
  selector: 'app-finanzas',
  standalone: true,
  imports: [CommonModule, TranslateModule, SidebarComponent],
  templateUrl: './finanzas.html',
  styleUrl: './finanzas.css',
})
export class FinanzasComponent {}
