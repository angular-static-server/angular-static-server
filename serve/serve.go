package serve

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"net/http"
	"ngstaticserver/compress"
	"ngstaticserver/constants"
	"ngstaticserver/serve/config"
	"ngstaticserver/serve/headers"
	"ngstaticserver/serve/response"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

const CspNonceName = "NGSS_CSP_NONCE"

var CspNoncePlaceholder = fmt.Sprintf("${%v}", CspNonceName)

var Flags = []cli.Flag{
	&cli.IntFlag{
		EnvVars: []string{"_PORT"},
		Name:    "port",
		Aliases: []string{"p"},
		Value:   8080,
	},
	&cli.Int64Flag{
		EnvVars: []string{"_CACHE_CONTROL_MAX_AGE"},
		Name:    "cache-control-max-age",
		Value:   60 * 60 * 24 * 365,
	},
	&cli.IntFlag{
		EnvVars: []string{"_CACHE_SIZE"},
		Name:    "cache-size",
		Value:   constants.DefaultCacheSize,
	},
	&cli.Int64Flag{
		EnvVars: []string{"_COMPRESSION_THRESHOLD"},
		Name:    "compression-threshold",
		Value:   constants.DefaultCompressionThreshold,
	},
	&cli.StringFlag{
		EnvVars: []string{"_LOG_LEVEL"},
		Name:    "log-level",
		Aliases: []string{"l"},
		Value:   "INFO",
	},
	&cli.StringFlag{
		EnvVars: []string{"_LOG_FORMAT"},
		Name:    "log-format",
		Value:   "text",
	},
	&cli.StringFlag{
		EnvVars: []string{"_I18N_DEFAULT"},
		Name:    "i18n-default",
		Value:   "",
	},
	&cli.PathFlag{
		EnvVars: []string{"_DOTENV_PATH"},
		Name:    "dotenv-path",
		Value:   "/config/.env",
	},
	&cli.StringFlag{
		EnvVars: []string{"_CSP_TEMPLATE"},
		Name:    "csp-template",
		Value: strings.Join([]string{
			"default-src 'self' ${_CSP_STYLE_SRC};",
			"connect-src 'self' ${_CSP_CONNECT_SRC};",
			"font-src 'self' ${_CSP_FONT_SRC};",
			"img-src 'self' ${_CSP_IMG_SRC};",
			"script-src 'self' ${NGSS_CSP_NONCE} ${_CSP_SCRIPT_SRC};",
			"style-src 'self' ${NGSS_CSP_NONCE} ${_CSP_STYLE_SRC};",
		}, " "),
	},
	&cli.StringFlag{
		EnvVars: []string{"_CSP_DEFAULT_SRC"},
		Name:    "csp-default-src",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"_CSP_CONNECT_SRC"},
		Name:    "csp-connect-src",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"_CSP_FONT_SRC"},
		Name:    "csp-font-src",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"_CSP_IMG_SRC"},
		Name:    "csp-img-src",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"_CSP_SCRIPT_SRC"},
		Name:    "csp-script-src",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"_CSP_STYLE_SRC"},
		Name:    "csp-style-src",
		Value:   "",
	},
	&cli.StringFlag{
		EnvVars: []string{"_X_FRAME_OPTIONS"},
		Name:    "x-frame-options",
		Value:   "DENY",
	},
}

type ServerParams struct {
	WorkingDirectory     string
	Port                 int
	DotEnvPath           string
	CacheControlMaxAge   int64
	CacheSize            int
	CompressionThreshold int64
	I18nDefault          string
	LogLevel             string
	LogFormat            string
	CspTemplate          string
	CspDefaultSrc        string
	CspConnectSrc        string
	CspFontSrc           string
	CspImgSrc            string
	CspScriptSrc         string
	CspStyleSrc          string
	XFrameOptions        string
}

type App struct {
	params       *ServerParams
	resolver     response.EntityResolver
	appVariables *config.AppVariables
	env          *config.DotEnv
	fileWatcher  *config.FileWatcher
}

func Action(c *cli.Context) error {
	params, err := parseServerParams(c)
	if err != nil {
		return err
	}

	if params.CacheSize < 1024 {
		slog.Warn(fmt.Sprintf("Minimum cache size is 1024 (configured %v). Resetting to 1024.", params.CacheSize))
	}

	// Configure slog logger
	var handler slog.Handler
	level := slog.LevelInfo
	err = level.UnmarshalText([]byte(params.LogLevel))
	handlerOptions := &slog.HandlerOptions{Level: level}
	if params.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	}
	slog.SetDefault(slog.New(handler))

	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to set log level %v. Resetting to INFO.\n", level))
	}

	app := createApp(params)
	defer app.Close()

	heartbeat := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("UP"))
	}
	http.HandleFunc("/__heartbeat__", heartbeat)
	http.HandleFunc("/__lbheartbeat__", heartbeat)
	http.HandleFunc("/", app.handleRequest)
	return http.ListenAndServe(fmt.Sprintf(":%v", params.Port), nil)
}

func parseServerParams(c *cli.Context) (*ServerParams, error) {
	var workingDirectory string
	var err error
	if c.NArg() > 0 {
		workingDirectory, err = filepath.Abs(c.Args().Get(0))
		if err != nil {
			return nil, fmt.Errorf("unable to resolve the absolute path of %v\n%v", c.Args().Get(0), err)
		}
	} else {
		workingDirectory, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve current working directory: %v", err)
		}
	}

	return &ServerParams{
		WorkingDirectory:     workingDirectory,
		Port:                 c.Int("port"),
		DotEnvPath:           c.Path("dotenv-path"),
		CacheControlMaxAge:   c.Int64("cache-control-max-age"),
		CacheSize:            c.Int("cache-size"),
		CompressionThreshold: c.Int64("compression-threshold"),
		I18nDefault:          c.String("i18n-default"),
		LogLevel:             c.String("log-level"),
		LogFormat:            c.String("log-format"),
		CspTemplate:          c.String("csp-template"),
		CspDefaultSrc:        c.String("csp-default-src"),
		CspConnectSrc:        c.String("csp-connect-src"),
		CspFontSrc:           c.String("csp-font-src"),
		CspImgSrc:            c.String("csp-img-src"),
		CspScriptSrc:         c.String("csp-script-src"),
		CspStyleSrc:          c.String("csp-style-src"),
		XFrameOptions:        c.String("x-frame-options"),
	}, nil
}

func createApp(params *ServerParams) App {
	fileWatcher := config.CreateFileWatcher()
	appVariables := config.InitializeAppVariables(params.WorkingDirectory)
	dotEnv := config.CreateDotEnv(params.DotEnvPath, appVariables.MergeVariables)
	fileWatcher.Watch(dotEnv)
	entityResolver := response.CreateEntityResolver(params.WorkingDirectory, params.CacheSize)
	return App{params, entityResolver, appVariables, dotEnv, fileWatcher}
}

func (app *App) Close() {
	app.fileWatcher.Close()
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (app *App) handleRequest(rw http.ResponseWriter, r *http.Request) {
	w := &loggingResponseWriter{rw, http.StatusOK}
	requestIdentity := fmt.Sprintf("%v %v %v", r.Method, r.URL.Path, r.Proto)
	slog.Debug(requestIdentity, "state", "request start")
	if r.Method != "GET" && r.Method != "HEAD" {
		errorResponse(w, requestIdentity, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	entity := app.resolver.Resolve(r.URL.Path)
	if entity.IsNotFound() {
		acceptLanguage := strings.Split(r.Header.Get("Accept-Language"), ",")
		acceptLanguage = append(acceptLanguage, app.params.I18nDefault)
		if match := app.resolver.MatchLanguage(acceptLanguage); match != "" {
			w.Header().Set("Location", fmt.Sprintf("/%v", match))
			w.WriteHeader(http.StatusTemporaryRedirect)
			slog.Info(requestIdentity, "status", http.StatusTemporaryRedirect)
			return
		}
		errorResponse(w, requestIdentity, http.StatusNotFound, "Not Found")
		return
	}

	if entity.IsIndex() {
		// Due to CSP limitations (stale nonce), we cannot allow browsers to cache index responses
		w.Header().Set("Cache-Control", "no-store")
	} else if entity.IsFingerprinted() && app.params.CacheControlMaxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", app.params.CacheControlMaxAge))
	} else {
		// https://web.dev/http-cache/?hl=en#flowchart
		w.Header().Set("Cache-Control", "no-cache")
	}

	if entity.ContentType != "" {
		w.Header().Set("Content-Type", entity.ContentType)
	}

	var encoding headers.Encoding = headers.NO_COMPRESSION
	if app.params.CompressionThreshold <= entity.Size && entity.Compressed {
		slog.Debug(
			requestIdentity,
			"state",
			fmt.Sprintf(
				"compression possible (threshold %v <= size %v && compressible=%v)",
				app.params.CompressionThreshold,
				entity.Size,
				entity.Compressed))
		acceptedEncoding := headers.ResolveAcceptEncoding(r)
		if acceptedEncoding.AllowsBrotli() {
			slog.Debug(requestIdentity, "state", "compression with brotli")
			w.Header().Set("Content-Encoding", "br")
			encoding = headers.BROTLI
		} else if acceptedEncoding.AllowsGzip() {
			slog.Debug(requestIdentity, "state", "compression with gzip")
			w.Header().Set("Content-Encoding", "gzip")
			encoding = headers.GZIP
		}
	} else {
		slog.Debug(
			requestIdentity,
			"state",
			fmt.Sprintf(
				"compression not applicable (threshold %v > size %v || compressible=%v)",
				app.params.CompressionThreshold,
				entity.Size,
				entity.Compressed))
	}

	var content []byte
	if entity.IsIndex() {
		content = app.renderIndex(w, entity, encoding)
	} else if encoding.ContainsBrotli() {
		content = entity.ContentBrotli
	} else if encoding.ContainsGzip() {
		content = entity.ContentGzip
	} else {
		content = entity.Content
	}

	http.ServeContent(w, r, entity.Path, entity.ModTime, bytes.NewReader(content))
	slog.Info(requestIdentity, "status", w.statusCode)
}

func errorResponse(w http.ResponseWriter, requestIdentity string, statusCode int, statusMessage string) {
	slog.Info(requestIdentity, "status", http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(`{"code": %v, "status": "%v"}`, statusCode, statusMessage)))
}

func (app *App) renderIndex(
	w http.ResponseWriter,
	entity response.ResponseEntity,
	encoding headers.Encoding,
) []byte {
	if len(app.params.XFrameOptions) > 0 {
		w.Header().Add("X-Frame-Options", app.params.XFrameOptions)
	}

	cspNonce := ""
	if app.isCspApplicable(entity.Content) && app.params.CspTemplate != "" {
		cspNonce = generateNonce()
		app.appVariables.Update(CspNonceName, cspNonce)
		cspValue := app.params.CspTemplate
		cspValue = strings.ReplaceAll(cspValue, CspNoncePlaceholder, fmt.Sprintf("'nonce-%v'", cspNonce))
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_DEFAULT_SRC}", app.params.CspDefaultSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_CONNECT_SRC}", app.params.CspConnectSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_FONT_SRC}", app.params.CspFontSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_IMG_SRC}", app.params.CspImgSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_SCRIPT_SRC}", app.params.CspScriptSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_STYLE_SRC}", app.params.CspStyleSrc)
		w.Header().Set("Content-Security-Policy", cspValue)
	} else if app.appVariables.IsEmpty() {
		if encoding.ContainsBrotli() && entity.ContentBrotli != nil {
			w.Header().Set("Content-Encoding", "br")
			return entity.ContentBrotli
		} else if encoding.ContainsGzip() && entity.ContentGzip != nil {
			w.Header().Set("Content-Encoding", "gzip")
			return entity.ContentGzip
		} else {
			return entity.Content
		}
	}

	content := app.appVariables.Insert(entity.Content, cspNonce)
	if cspNonce != "" {
		content = strings.ReplaceAll(content, CspNoncePlaceholder, cspNonce)
	}
	byteContent := []byte(content)
	if int64(len(content)) < app.params.CompressionThreshold {
		return byteContent
	} else if encoding.ContainsBrotli() {
		return compress.CompressWithBrotliFast(byteContent)
	} else if encoding.ContainsGzip() {
		return compress.CompressWithGzipFast(byteContent)
	} else {
		return byteContent
	}
}

func (app *App) isCspApplicable(indexContent []byte) bool {
	return app.appVariables.Has(CspNonceName) || strings.Contains(string(indexContent), CspNoncePlaceholder)
}

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var runeCharts = []rune(chars)

func generateNonce() string {
	result := make([]byte, 16)
	for i := 0; i < 16; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			slog.Warn("Failed to use secure random to generate CSP nonce. Falling back to less secure variant.")
			localRand := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
			pick := make([]rune, 16)
			for i := range pick {
				pick[i] = runeCharts[localRand.Intn(len(chars))]
			}

			return string(pick)
		}
		result[i] = chars[value.Int64()]
	}

	return string(result)
}
