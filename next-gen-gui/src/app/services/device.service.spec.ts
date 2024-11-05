import { TestBed } from '@angular/core/testing';

import { DeviceService } from './device.service';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('DeviceService', () => {
  let service: DeviceService;

  beforeEach(() => {
    TestBed.configureTestingModule({
    imports: [],
    providers: [DeviceService, provideHttpClient(withInterceptorsFromDi())]
});
    service = TestBed.inject(DeviceService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
