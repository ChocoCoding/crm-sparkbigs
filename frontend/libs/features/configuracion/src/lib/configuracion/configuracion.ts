import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { SidebarComponent } from '@miapp/data-access';

@Component({
  selector: 'app-configuracion',
  standalone: true,
  imports: [CommonModule, TranslateModule, SidebarComponent],
  templateUrl: './configuracion.html',
  styleUrl: './configuracion.css',
})
export class ConfiguracionComponent {}
