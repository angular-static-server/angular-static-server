package serve

import (
	"fmt"
	"net/http"
	"ngstaticserver/constants"
	"ngstaticserver/serve/config"
	"ngstaticserver/serve/endpoints"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dimfeld/httptreemux/v5"
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
		Value:   constants.CspTemplate,
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

	slog.Debug("HTTP server setup start")
	app := createApp(params)
	defer app.Close()

	router := app.createRouter()
	slog.Debug("HTTP server setup complete")
	return http.ListenAndServe(fmt.Sprintf(":%v", params.Port), router)
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
	return App{params, appVariables, dotEnv, fileWatcher}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (app App) createRouter() *httptreemux.TreeMux {
	router := httptreemux.New()
	router.PanicHandler = httptreemux.SimplePanicHandler
	router.Use(func(next httptreemux.HandlerFunc) httptreemux.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request, m map[string]string) {
			w := &loggingResponseWriter{rw, http.StatusOK}
			requestIdentity := fmt.Sprintf("%v %v %v", r.Method, r.URL.Path, r.Proto)
			slog.Debug(requestIdentity, "state", "request start")
			next(w, r, m)
			slog.Info(requestIdentity, "status", w.statusCode)
			slog.Debug(requestIdentity, "state", "request complete")
		}
	})
	versionEndpoint := endpoints.VersionEndpoint(filepath.Join(app.params.WorkingDirectory, "version.json"))
	heartbeatEndpoint := endpoints.HeartbeatEndpoint()
	router.GET("/__version__", versionEndpoint.Handle)
	router.GET("/__heartbeat__", heartbeatEndpoint.Handle)
	router.GET("/__lbheartbeat__", heartbeatEndpoint.Handle)

	indexPaths := make([]string, 0)
	err := filepath.Walk(app.params.WorkingDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "/index.html") {
			indexPaths = append(indexPaths, path)
		} else if info.IsDir() || strings.HasSuffix(path, ".br") || strings.HasSuffix(path, ".gz") {
			return nil
		}

		requestPath, _ := filepath.Rel(app.params.WorkingDirectory, path)
		handler, err := endpoints.ResolveFileEndpoint(path, app.params.CacheControlMaxAge)
		if err != nil {
			return err
		}

		router.GET(fmt.Sprintf("/%v", requestPath), handler.Handle)
		return nil
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to walk files in %v", app.params.WorkingDirectory), "error", err)
	}

	sort.Slice(indexPaths, func(i, j int) bool {
		return len(indexPaths[i]) > len(indexPaths[j])
	})
	cspValue := app.params.CspTemplate
	if len(cspValue) > 0 {
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_DEFAULT_SRC}", app.params.CspDefaultSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_CONNECT_SRC}", app.params.CspConnectSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_FONT_SRC}", app.params.CspFontSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_IMG_SRC}", app.params.CspImgSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_SCRIPT_SRC}", app.params.CspScriptSrc)
		cspValue = strings.ReplaceAll(cspValue, "${_CSP_STYLE_SRC}", app.params.CspStyleSrc)
	}
	hasRootIndex := false
	for _, path := range indexPaths {
		dir := filepath.Dir(path)
		requestPath, _ := filepath.Rel(app.params.WorkingDirectory, dir)
		if requestPath == "." {
			requestPath = ""
			hasRootIndex = true
		} else if !strings.HasSuffix(requestPath, "/") {
			requestPath += "/"
		}
		handler := endpoints.ResolveIndexEndpoint(path, int(app.params.CompressionThreshold), cspValue, app.appVariables)
		router.GET(fmt.Sprintf("/%v", requestPath), handler.Handle)
		router.GET(fmt.Sprintf("/%v*filepath", requestPath), handler.Handle)
	}

	if len(indexPaths) > 0 && !hasRootIndex {
		handler := endpoints.ResolveRootEndpoint(app.params.WorkingDirectory, app.params.I18nDefault)
		router.GET("/", handler.Handle)
		router.GET("/*filepath", handler.Handle)
	}

	return router
}

func (app *App) Close() {
	app.fileWatcher.Close()
}
