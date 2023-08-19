package serve

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"ngstaticserver/constants"
	"ngstaticserver/test"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v2"
)

var Licenses = "3rdpartylicenses.txt"
var IndexHtml = "index.html"

func TestAction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Flags:  Flags,
				Action: Action,
			},
		},
	}
	go app.RunContext(ctx, []string{"path-to-binary", "serve"})

	select {
	case <-ctx.Done():
		test.AssertEqual(t, ctx.Err(), context.DeadlineExceeded)
	}
}

func TestStartingServer(t *testing.T) {
	app, _ := createTestApp(t)
	ts := httptest.NewServer(app.createRouter())
	defer ts.Close()
}

func TestFileRequest(t *testing.T) {
	app, context := createTestApp(t)
	content := context.ReadFile(Licenses)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%v", Licenses), nil)
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	test.AssertEqual(t, string(body), content)
}

func TestFileRequestBrotli(t *testing.T) {
	for _, e := range []string{"*", "br", "gz, deflate, br"} {
		app, context := createTestApp(t)
		polyfill := context.FindFile("polyfills.")
		context.CompressFile(polyfill)
		content := context.ReadFile(polyfill)

		req := httptest.NewRequest("GET", fmt.Sprintf("/%v", polyfill), nil)
		req.Header.Add("Accept-Encoding", e)
		w := httptest.NewRecorder()
		app.createRouter().ServeHTTP(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		test.AssertEqual(t, resp.StatusCode, 200)
		test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/javascript; charset=utf-8")
		test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "br")
		responseContent := string(test.DecompressBrotli(body))
		test.AssertEqual(t, responseContent, content)
	}
}

func TestIndexRequestBrotli(t *testing.T) {
	for _, s := range []bool{false, true} {
		for _, e := range []string{"*", "br"} {
			app, context := createTestAppWithInit(t, func(context test.TestDir, params *ServerParams) {
				context.ImportTestApp("ngssc")
				params.CompressionThreshold = 10
				context.CompressFile(IndexHtml)
				if s {
					context.RemoveFile("ngssc.json")
				}
			})
			content := context.ReadFile(IndexHtml)
			parts := regexp.MustCompile("(<!--CONFIG-->|\\${NGSS_CSP_NONCE})").Split(content, -1)

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Add("Accept-Encoding", e)
			w := httptest.NewRecorder()
			app.createRouter().ServeHTTP(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			test.AssertEqual(t, resp.StatusCode, 200)
			test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
			test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "br")
			responseContent := string(test.DecompressBrotli(body))
			for k, v := range parts {
				if k == 0 {
					test.AssertTrue(t, strings.HasPrefix(responseContent, v))
				} else if k < len(parts)-1 {
					test.AssertTrue(t, strings.Contains(responseContent, v))
				} else {
					test.AssertTrue(t, strings.HasSuffix(responseContent, v))
				}
			}
		}
	}
}

func TestFileRequestGzip(t *testing.T) {
	app, context := createTestApp(t)
	polyfill := context.FindFile("polyfills.")
	context.CompressFile(polyfill)
	content := context.ReadFile(polyfill)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%v", polyfill), nil)
	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/javascript; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "gzip")
	responseContent := string(test.DecompressGzip(body))
	test.AssertEqual(t, responseContent, content)
}

func TestIndexRequestGzip(t *testing.T) {
	for _, s := range []bool{false, true} {
		app, context := createTestAppWithInit(t, func(context test.TestDir, params *ServerParams) {
			context.ImportTestApp("ngssc")
			params.CompressionThreshold = 10
			context.CompressFile(IndexHtml)
			if s {
				context.RemoveFile("ngssc.json")
			}
		})
		content := context.ReadFile(IndexHtml)
		parts := regexp.MustCompile("(<!--CONFIG-->|\\${NGSS_CSP_NONCE})").Split(content, -1)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Add("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		app.createRouter().ServeHTTP(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		test.AssertEqual(t, resp.StatusCode, 200)
		test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
		test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "gzip")
		responseContent := string(test.DecompressGzip(body))
		for k, v := range parts {
			if k == 0 {
				test.AssertTrue(t, strings.HasPrefix(responseContent, v))
			} else if k < len(parts)-1 {
				test.AssertTrue(t, strings.Contains(responseContent, v))
			} else {
				test.AssertTrue(t, strings.HasSuffix(responseContent, v))
			}
		}
	}
}

func TestMultipleIndex(t *testing.T) {
	var expectedIndexContent string
	app, _ := createTestAppWithInit(t, func(context test.TestDir, params *ServerParams) {
		context.ImportTestApp("i18n")
		expectedIndexContent = context.ReadFile("de-CH/index.html")
	})
	parts := regexp.MustCompile("(</title>|\\${NGSS_CSP_NONCE})").Split(expectedIndexContent, -1)

	req := httptest.NewRequest("GET", "/de-CH/example/path/to/request", nil)
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	bodyText := string(body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	for k, v := range parts {
		if k == 0 {
			test.AssertTrue(t, strings.HasPrefix(bodyText, v))
		} else if k < len(parts)-1 {
			test.AssertTrue(t, strings.Contains(bodyText, v))
		} else {
			test.AssertTrue(t, strings.HasSuffix(bodyText, v))
		}
	}
}

func TestNoNgsscJson(t *testing.T) {
	app, _ := createTestAppWithInit(t, func(context test.TestDir, _ *ServerParams) {
		context.ImportTestApp("minimal")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertTrue(t, !strings.Contains(string(body), "ngssc"))
}

func TestNotFound(t *testing.T) {
	app, _ := createTestAppWithInit(t, func(context test.TestDir, params *ServerParams) {})

	req := httptest.NewRequest("GET", "/example.txt", nil)
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, 404)
}

func TestNonGetRequest(t *testing.T) {
	app, _ := createTestApp(t)

	req := httptest.NewRequest("PUT", "/example.txt", nil)
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, 405)
}

func TestHeadRequest(t *testing.T) {
	app, _ := createTestApp(t)

	req := httptest.NewRequest("HEAD", "/example.txt", nil)
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, string(body), "")
}

func TestLanguageRedirect(t *testing.T) {
	app, _ := createTestAppWithInit(t, func(context test.TestDir, params *ServerParams) {
		context.ImportTestApp("i18n")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.createRouter().ServeHTTP(w, req)

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, http.StatusTemporaryRedirect)
	test.AssertEqual(t, resp.Header.Get("Location"), "/de-CH")
}

func createTestApp(t *testing.T) (App, test.TestDir) {
	return createTestAppWithInit(t, func(context test.TestDir, params *ServerParams) {
		context.ImportTestApp("ngssc")
	})
}

func createTestAppWithInit(t *testing.T, init func(context test.TestDir, params *ServerParams)) (App, test.TestDir) {
	context := test.NewTestDir(t)
	cspTemplate := regexp.MustCompile("\\$\\{_CSP_[^_]+_SRC\\}").ReplaceAllString(constants.CspTemplate, "")
	params := &ServerParams{
		WorkingDirectory:     context.Path,
		Port:                 0,
		CacheControlMaxAge:   31536000,
		CompressionThreshold: constants.DefaultCompressionThreshold,
		LogLevel:             "ERROR",
		CspTemplate:          cspTemplate,
		XFrameOptions:        "DENY",
	}
	init(context, params)
	app := createApp(params)
	t.Cleanup(func() {
		app.Close()
	})
	return app, context
}
