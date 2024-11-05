import { TestBed } from '@angular/core/testing';

import { FolderService } from './folder.service';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('FolderService', () => {
  let service: FolderService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [FolderService, provideHttpClient(withInterceptorsFromDi())]
});
    service = TestBed.inject(FolderService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
