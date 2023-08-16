import { ApplicationConfig, CSP_NONCE } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideHttpClient } from '@angular/common/http';

import { routes } from './app.routes';
import { environment } from '../environments/environment';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    {
        provide: CSP_NONCE,
        useValue: environment.cspNonce,
    },
    provideHttpClient(),
    provideAnimations()
]
};
