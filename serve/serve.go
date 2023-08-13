package serve

import (
	"bytes"
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"ngstaticserver/compress"
	"ngstaticserver/constants"
	"ngstaticserver/ngsscjson"
	"ngstaticserver/serve/acceptencoding"
	"ngstaticserver/serve/dotenv"
	"ngstaticserver/serve/response"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

const CspNonceName = "NGSSC_CSP_NONCE"
const CspHashName = "NGSSC_CSP_HASH"

var CspNoncePlaceholder = fmt.Sprintf("${%v}", CspNonceName)
var CspHashPlaceholder = fmt.Sprintf("${%v}", CspHashName)

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
		Value: fmt.Sprintf(
			"default-src 'self'; style-src 'self' ${%v}; script-src 'self' ${%v} ${%v};",
			CspNonceName,
			CspHashName,
			CspNonceName,
		),
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
	XFrameOptions        string
}

type App struct {
	params      *ServerParams
	resolver    response.EntityResolver
	ngsscConfig ngsscjson.NgsscConfig
	env         dotenv.DotEnv
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
		XFrameOptions:        c.String("x-frame-options"),
	}, nil
}

func createApp(params *ServerParams) App {
	ngsscConfig := tryReadNgsscJson(params)
	return App{
		params:      params,
		resolver:    response.CreateEntityResolver(params.WorkingDirectory, params.CacheSize),
		ngsscConfig: ngsscConfig,
		env: dotenv.Create(params.DotEnvPath, func(envVariables map[string]*string) {
			if len(ngsscConfig.EnvironmentVariables) > 0 {
				ngsscConfig.MergeVariables(envVariables)
			} else {
				ngsscConfig.PopulatedEnvironmentVariables = envVariables
			}
		}),
	}
}

func tryReadNgsscJson(params *ServerParams) ngsscjson.NgsscConfig {
	ngsscjsonPath := filepath.Join(params.WorkingDirectory, "ngssc.json")
	info, err := os.Stat(ngsscjsonPath)
	if err == nil && !info.IsDir() {
		slog.Info(fmt.Sprintf("Detected ngssc.json file at %v. Reading configuration.", ngsscjsonPath))
		ngsscConfig, err := ngsscjson.NgsscJsonConfigFromPath(ngsscjsonPath)
		if err == nil {
			return ngsscConfig
		}
		slog.Warn(fmt.Sprintf("%v, creating default configuration", err))
	}

	return ngsscjson.NgsscConfig{
		FilePath:                      params.WorkingDirectory,
		Variant:                       "global",
		EnvironmentVariables:          make([]string, 0),
		PopulatedEnvironmentVariables: make(map[string]*string),
		FilePattern:                   "**/index.html",
	}
}

func (app *App) Close() {
	app.env.Close()
}

func (app *App) handleRequest(w http.ResponseWriter, r *http.Request) {
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

	if entity.IsFingerprinted() && app.params.CacheControlMaxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", app.params.CacheControlMaxAge))
	} else {
		// https://web.dev/http-cache/?hl=en#flowchart
		w.Header().Set("Cache-Control", "no-cache")
	}

	if entity.ContentType != "" {
		w.Header().Set("Content-Type", entity.ContentType)
	}

	var encoding acceptencoding.Encoding = acceptencoding.NO_COMPRESSION
	if app.params.CompressionThreshold <= entity.Size && entity.Compressed {
		slog.Debug(
			requestIdentity,
			"state",
			fmt.Sprintf(
				"compression possible (threshold %v <= size %v && compressible=%v)",
				app.params.CompressionThreshold,
				entity.Size,
				entity.Compressed))
		acceptedEncoding := acceptencoding.ResolveAcceptEncoding(r)
		if acceptedEncoding.AllowsBrotli() {
			slog.Debug(requestIdentity, "state", "compression with brotli")
			w.Header().Set("Content-Encoding", "br")
			encoding = acceptencoding.BROTLI
		} else if acceptedEncoding.AllowsGzip() {
			slog.Debug(requestIdentity, "state", "compression with gzip")
			w.Header().Set("Content-Encoding", "gzip")
			encoding = acceptencoding.GZIP
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
	slog.Info(requestIdentity, "status", http.StatusOK)
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
	encoding acceptencoding.Encoding,
) []byte {
	if len(app.params.XFrameOptions) > 0 {
		w.Header().Add("X-Frame-Options", app.params.XFrameOptions)
	}

	cspNonce := ""
	if app.isCspApplicable(entity.Content) && app.params.CspTemplate != "" {
		cspNonce = generateNonce()
		_, ok := app.ngsscConfig.PopulatedEnvironmentVariables[CspNonceName]
		if ok {
			app.ngsscConfig.PopulatedEnvironmentVariables[CspNonceName] = &cspNonce
		}
		cspValue := app.params.CspTemplate
		cspValue = strings.ReplaceAll(cspValue, CspHashPlaceholder, app.ngsscConfig.GenerateIifeScriptHash(""))
		cspValue = strings.ReplaceAll(cspValue, CspNoncePlaceholder, fmt.Sprintf("'nonce-%v'", cspNonce))
		w.Header().Set("Content-Security-Policy", cspValue)
	} else if len(app.ngsscConfig.PopulatedEnvironmentVariables) == 0 {
		if encoding.ContainsBrotli() && entity.ContentBrotli != nil {
			return entity.ContentBrotli
		} else if encoding.ContainsGzip() && entity.ContentGzip != nil {
			return entity.ContentGzip
		} else {
			return entity.Content
		}
	}

	content := app.ngsscConfig.Apply(entity.Content)
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
	_, ok := app.ngsscConfig.PopulatedEnvironmentVariables[CspNonceName]
	return ok || strings.Contains(string(indexContent), CspNoncePlaceholder)
}

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var runeCharts = []rune(chars)

func generateNonce() string {
	bytes := make([]byte, 10)

	if _, err := rand.Read(bytes); err != nil {
		slog.Warn("Failed to use secure random to generate CSP nonce. Falling back to insecure variant.")
		localRand := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
		pick := make([]rune, 10)
		for i := range pick {
			pick[i] = runeCharts[localRand.Intn(len(chars))]
		}

		return string(pick)
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}

	return string(bytes)
}
