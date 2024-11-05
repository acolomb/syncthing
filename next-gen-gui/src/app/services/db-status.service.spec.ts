import { TestBed } from '@angular/core/testing';

import { DbStatusService } from './db-status.service';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('DbStatusService', () => {
  let service: DbStatusService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [DbStatusService, provideHttpClient(withInterceptorsFromDi())]
});
    TestBed.configureTestingModule({});
    service = TestBed.inject(DbStatusService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
