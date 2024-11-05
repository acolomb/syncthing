import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSortModule } from '@angular/material/sort';
import { MatTableModule } from '@angular/material/table';

import { DeviceListComponent } from './device-list.component';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';
import { ChangeDetectorRef } from '@angular/core';

describe('DeviceListComponent', () => {
  let component: DeviceListComponent;
  let fixture: ComponentFixture<DeviceListComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
    declarations: [DeviceListComponent],
    imports: [],
    providers: [DeviceListComponent, ChangeDetectorRef, provideHttpClient(withInterceptorsFromDi())]
}).compileComponents();

    component = TestBed.inject(DeviceListComponent);
  }));

  it('should compile', () => {
    expect(component).toBeTruthy();
  });
});
