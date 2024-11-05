import { TestBed } from '@angular/core/testing';

import { SystemConnectionsService } from './system-connections.service';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('SystemConnectionsService', () => {
  let service: SystemConnectionsService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [SystemConnectionsService, provideHttpClient(withInterceptorsFromDi())]
});
    service = TestBed.inject(SystemConnectionsService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
