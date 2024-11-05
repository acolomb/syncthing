import { TestBed } from '@angular/core/testing';

import { SystemConfigService } from './system-config.service';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('SystemConfigService', () => {
  let service: SystemConfigService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [SystemConfigService, provideHttpClient(withInterceptorsFromDi())]
});
    service = TestBed.inject(SystemConfigService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
