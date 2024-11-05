import { TestBed } from '@angular/core/testing';

import { DbCompletionService } from './db-completion.service';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('DbCompletionService', () => {
  let service: DbCompletionService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [DbCompletionService, provideHttpClient(withInterceptorsFromDi())]
});
    service = TestBed.inject(DbCompletionService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
