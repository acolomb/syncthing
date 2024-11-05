import { async, TestBed } from '@angular/core/testing';

import { ChartComponent } from './chart.component';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

class MockService {
  getEach() {
    // unimplemented
  }
};

describe('ChartComponent', () => {
  let component: ChartComponent;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [ChartComponent, provideHttpClient(withInterceptorsFromDi())]
}).compileComponents();
    component = TestBed.inject(ChartComponent);
  }));

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
