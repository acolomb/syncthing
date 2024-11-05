import { TestBed } from '@angular/core/testing';

import { SystemStatusService } from './system-status.service';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('SystemStatusService', () => {
  let service: SystemStatusService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [SystemStatusService, provideHttpClient(withInterceptorsFromDi())]
});
    service = TestBed.inject(SystemStatusService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
