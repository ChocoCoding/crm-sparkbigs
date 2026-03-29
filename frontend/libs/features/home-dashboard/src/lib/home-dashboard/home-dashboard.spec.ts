import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HomeDashboard } from './home-dashboard';

describe('HomeDashboard', () => {
  let component: HomeDashboard;
  let fixture: ComponentFixture<HomeDashboard>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [HomeDashboard],
    }).compileComponents();

    fixture = TestBed.createComponent(HomeDashboard);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
