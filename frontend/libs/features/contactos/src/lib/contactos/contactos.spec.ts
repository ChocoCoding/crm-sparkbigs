import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Contactos } from './contactos';

describe('Contactos', () => {
  let component: Contactos;
  let fixture: ComponentFixture<Contactos>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Contactos],
    }).compileComponents();

    fixture = TestBed.createComponent(Contactos);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
