import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { provideHttpClient, withInterceptorsFromDi, withXsrfConfiguration } from '@angular/common/http';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { MatLegacyTableModule as MatTableModule } from '@angular/material/legacy-table';
import { MatLegacyPaginatorModule as MatPaginatorModule } from '@angular/material/legacy-paginator';
import { MatSortModule } from '@angular/material/sort';
import { MatLegacyInputModule as MatInputModule } from '@angular/material/legacy-input';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatLegacyCardModule as MatCardModule } from '@angular/material/legacy-card';
import { MatLegacyProgressBarModule as MatProgressBarModule } from '@angular/material/legacy-progress-bar';
import { MatLegacyDialogModule as MatDialogModule } from '@angular/material/legacy-dialog';
import { MatLegacyListModule as MatListModule } from '@angular/material/legacy-list'
import { MatLegacyButtonModule as MatButtonModule } from '@angular/material/legacy-button';
import { FlexLayoutModule } from '@angular/flex-layout';

import { httpInterceptorProviders } from './http-interceptors';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';

import { StatusListComponent } from './lists/status-list/status-list.component';
import { DeviceListComponent } from './lists/device-list/device-list.component';
import { DonutChartComponent } from './charts/donut-chart/donut-chart.component';
import { DashboardComponent } from './dashboard/dashboard.component';
import { ListToggleComponent } from './list-toggle/list-toggle.component';

import { HttpClientInMemoryWebApiModule } from 'angular-in-memory-web-api';
import { InMemoryConfigDataService } from './services/in-memory-config-data.service';

import { deviceID } from './api-utils';
import { environment } from '../environments/environment';
import { ChartItemComponent } from './charts/chart-item/chart-item.component';
import { ChartComponent } from './charts/chart/chart.component';
import { FolderListComponent } from './lists/folder-list/folder-list.component';
import { DialogComponent } from './dialog/dialog.component';
import { CardComponent, CardTitleComponent, CardContentComponent } from './card/card.component';
import { TrimPipe } from './trim.pipe';

@NgModule({ declarations: [
        AppComponent,
        StatusListComponent,
        DeviceListComponent,
        ListToggleComponent,
        DashboardComponent,
        DonutChartComponent,
        ChartComponent,
        ChartItemComponent,
        FolderListComponent,
        DialogComponent,
        CardComponent,
        CardTitleComponent,
        CardContentComponent,
        TrimPipe,
    ],
    bootstrap: [AppComponent], imports: [BrowserModule,
        AppRoutingModule,
        BrowserAnimationsModule,
        MatInputModule,
        MatTableModule,
        MatPaginatorModule,
        MatSortModule,
        MatButtonToggleModule,
        MatCardModule,
        MatProgressBarModule,
        MatDialogModule,
        MatListModule,
        MatButtonModule,
        FlexLayoutModule,
        environment.production ?
            [] : HttpClientInMemoryWebApiModule.forRoot(InMemoryConfigDataService, { dataEncapsulation: false, delay: 10 })], providers: [httpInterceptorProviders, provideHttpClient(withInterceptorsFromDi(), withXsrfConfiguration({
            headerName: 'X-CSRF-Token-' + deviceID(),
            cookieName: 'CSRF-Token-' + deviceID(),
        }))] })

export class AppModule { }


