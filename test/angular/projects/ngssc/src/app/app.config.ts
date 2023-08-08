import { ApplicationConfig, CSP_NONCE } from '@angular/core';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';
import { environment } from '../environments/environment';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    {
      provide: CSP_NONCE,
      useValue: environment.cspNonce,
    },
  ]
};
