# angular-static-server

A static HTTP server based on Go optimized for Angular applications
provided as a container image.

## Motivation

`angular-static-server` is a simple, opinionated HTTP server for Angular applications.

✅ File server and `index.html` lookup  
✅ Zero configuration necessary  
✅ Small binary size (container image size is ~12MB) and fast startup  
✅ Support app configuration via environment variables/.env file  
✅ Provide security via CSP header and templating
✅ Follows Angular major version cadence

Feel free to create a feature request, if a feature you need is missing.

## Usage

`Dockerfile`

```Dockerfile
FROM ghcr.io/angular-static-server/server:16

# Copy your built application into the container.
COPY --chown=10001:10001 dist/your-app .
# Optionally compress your files to gzip and brotli variants for improved performance.
RUN ["ng-server", "compress"]
```

## Container Image

The `ghcr.io/angular-static-server/server` image is a minimal image that only contains
the binary. As it extends from [scratch](https://hub.docker.com/_/scratch), it also
does not include a shell. If you need a shell or other tools, you could copy the binary
(which is located at `/usr/local/bin/ng-server`) into your own image in a multi-stage build.

See the [Dockerfile](./Dockerfile) for reference.

The image follows the guidelines defined by https://github.com/mozilla-services/Dockerflow
(except for the named user/group `app:app`, which is just `10001:10001`).
This means the following endpoints are implemented:

| Endpoint           | Functionality                                                                     |
| ------------------ | --------------------------------------------------------------------------------- |
| `/__version__`     | Returns the content of ./version.json, if available.                              |
| `/__heartbeat__`   | Returns a HTTP status 200 if healthy, 5xx if not (currently no use case for 5xx). |
| `/__lbheartbeat__` | Always returns a HTTP status 200.                                                 |

## App Configuration

No configuration is necessary, if you have a static Angular app with no external config.

If your Angular app requires configuration, an `.env` file,
[angular-server-side-configuration](https://www.npmjs.com/package/angular-server-side-configuration)
with environment variables or a combination is supported.
If both an `ngssc.json` (from `angular-server-side-configuration`) and an `.env` file is
available, the configuration from the `ngssc.json` is used and variables from `.env` override
environment variables.

Copy or mount your `.env` file to `/config/.env` in the container.

## Security

For security the [Content-Security-Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
(see [Angular documentation](https://angular.io/guide/security#content-security-policy) on CSP)
and [X-Frame-Options](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options)
headers are used (See command documentation below for details).

The CSP header is only applied, if either the `ngCspNonce="..."` attribute (recommended) or the
reserved environment variable `NGSS_CSP_NONCE` is used.

If you only want to minimally extend the allowed CSP sources, there are a list of variables
that can be used to extend a specific source: `_CSP_*_SRC`.

### ngCspNonce (recommended)

`index.html`

```html
...
<body class="mat-typography">
  <app-root ngCspNonce="${NGSS_CSP_NONCE}"></app-root>
</body>
...
```

### NGSS_CSP_NONCE

In order for this to work, you either need to use `angular-server-side-configuration`
or define `NGSS_CSP_NONCE` in `.env` with a placeholder value (which is dynamically replaced
for each request).

**Note**: At the time of writing, when not using the `ngCspNonce="..."` attribute the Angular CLI
still renders inline scripts, which breaks CSP/the app. Due to this, the `ngCspNonce` approach
is recommended.

`app.config.ts`

```ts
...
bootstrapApplication(AppComponent, {
  providers: [{
    provide: CSP_NONCE,
    useValue: globalThis.NGSS_CSP_NONCE
  }]
});
...
```

## Commands

In order not to conflict with environment variables defined by users, the configuration environment
variables for `ng-server` are prefixed with `_`.

### compress

Compresses appropriate files in the working directory to `brotli` (e.g. `main.676ae13716545088.js.br`)
and `gzip` (e.g. `main.676ae13716545088.js.gz`) variants.

In order to reduce write operations in a container, it is recommended to compress files at build
time. Limited IO operations (and especially write operations) is better for Kubernetes clusters.

Usage: `ng-server compress [options] [directory]`
Usage in `Dockerfile`: `RUN ["ng-server", "compress"]`

| Environment Variable    | Command                   | Description                                                                    | Default |
| ----------------------- | ------------------------- | ------------------------------------------------------------------------------ | ------- |
| \_COMPRESSION_THRESHOLD | `--compression-threshold` | The threshold for compression. Only files larger than this will be compressed. | `1024`  |

### serve

Start a HTTP server.

Usage: `ng-server serve [options] [directory]`
Usage in `Dockerfile`: `CMD ["ng-server", "compress"]`

| Environment Variable    | Command                   | Description                                                                                                                                                 | Default                                                                                                                                                                                    |
| ----------------------- | ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| \_PORT                  | `--port` or `-p`          | The port to listen to.                                                                                                                                      | `8080`                                                                                                                                                                                     |
| \_CACHE_CONTROL_MAX_AGE | `--cache-control-max-age` | The `Cache-Control` `max-age` value for fingerprinted files.                                                                                                | `31536000` (a year)                                                                                                                                                                        |
| \_COMPRESSION_THRESHOLD | `--compression-threshold` | The threshold for dynamic compression. This is used to check whether to use compressed versions of files or whether to compress index responses.            | `1024`                                                                                                                                                                                     |
| \_LOG_LEVEL             | `--log-level` or `-l`     | The log level. Supports `DEBUG`, `INFO`, `WARN` and `ERROR`.                                                                                                | `INFO`                                                                                                                                                                                     |
| \_LOG_FORMAT            | `--log-format`            | Supports `text` or `json`.                                                                                                                                  | `text`                                                                                                                                                                                     |
| \_I18N_DEFAULT          | `--i18n-default`          | Which i18n variant should be used, if user `Accept-Language` value matches no available variants. Defaults to alphabetically first variant, if not defined. | ``                                                                                                                                                                                         |
| \_DOTENV_PATH           | `--dotenv-path`           | Path to the optional `.env` file to be used for app configuration.                                                                                          | `/config/.env`                                                                                                                                                                             |
| \_CSP_TEMPLATE          | `--csp-template`          | The `Content-Security-Policy` template HTTP header to be used.                                                                                              | `default-src 'self'; style-src 'self' ${NGSS_CSP_NONCE} ${CSP_STYLE_PLACEHOLDER}; script-src 'self' ${NGSS_CSP_NONCE} ${CSP_SCRIPT_PLACEHOLDER}; font-src 'self' ${CSP_FONT_PLACEHOLDER};` |
| \_CSP_DEFAULT_SRC       | `--csp-default-src`       | Value to be inserted into the \_CSP_TEMPLATE in the `default-src` section.                                                                                  | ``                                                                                                                                                                                         |
| \_CSP_CONNECT_SRC       | `--csp-connect-src`       | Value to be inserted into the \_CSP_TEMPLATE in the `connect-src` section.                                                                                  | ``                                                                                                                                                                                         |
| \_CSP_FONT_SRC          | `--csp-font-src`          | Value to be inserted into the \_CSP_TEMPLATE in the `font-src` section.                                                                                     | ``                                                                                                                                                                                         |
| \_CSP_IMG_SRC           | `--csp-img-src`           | Value to be inserted into the \_CSP_TEMPLATE in the `img-src` section.                                                                                      | ``                                                                                                                                                                                         |
| \_CSP_SCRIPT_SRC        | `--csp-script-src`        | Value to be inserted into the \_CSP_TEMPLATE in the `script-src` section.                                                                                   | ``                                                                                                                                                                                         |
| \_CSP_STYLE_SRC         | `--csp-style-src`         | Value to be inserted into the \_CSP_TEMPLATE in the `style-src` section.                                                                                    | ``                                                                                                                                                                                         |
| \_X_FRAME_OPTIONS       | `--x-frame-options`       | The `X-Frame-Options` value for the HTTP header.                                                                                                            | `DENY`                                                                                                                                                                                     |
