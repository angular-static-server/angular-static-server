package endpoints

import (
	"io"
	"net/http/httptest"
	"ngstaticserver/constants"
	"ngstaticserver/serve/config"
	"ngstaticserver/serve/headers"
	"ngstaticserver/test"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestIndexRequest_uncompressed(t *testing.T) {
	context, handler := createTestContext_index(t, headers.NO_COMPRESSION)
	content := context.ReadFile("de-CH/index.html")

	req := httptest.NewRequest("GET", "/de-CH", nil)
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertEqual(t, string(body), content)
}

func TestIndexRequest_gzip(t *testing.T) {
	context, handler := createTestContext_index(t, headers.GZIP)
	content := context.ReadFile("de-CH/index.html")

	req := httptest.NewRequest("GET", "/de-CH", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "gzip")
	responseContent := string(test.DecompressGzip(body))
	test.AssertEqual(t, responseContent, content)
}

func TestIndexRequest_brotli(t *testing.T) {
	context, handler := createTestContext_index(t, headers.BROTLI)
	content := context.ReadFile("de-CH/index.html")

	req := httptest.NewRequest("GET", "/de-CH", nil)
	req.Header.Add("Accept-Encoding", "br")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "br")
	responseContent := string(test.DecompressBrotli(body))
	test.AssertEqual(t, responseContent, content)
}

func TestIndexRequest_uncompressed_withVariables(t *testing.T) {
	context, handler := createTestContext_index(t, headers.NO_COMPRESSION)
	insertVariables(handler.AppVariables)
	parts := strings.Split(context.ReadFile("de-CH/index.html"), "</title>")

	req := httptest.NewRequest("GET", "/de-CH", nil)
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertTrue(t, strings.HasPrefix(content, parts[0]))
	test.AssertTrue(t, strings.HasSuffix(content, parts[1]))
}

func TestIndexRequest_gzip_withVariables(t *testing.T) {
	context, handler := createTestContext_index(t, headers.GZIP)
	insertVariables(handler.AppVariables)
	parts := strings.Split(context.ReadFile("de-CH/index.html"), "</title>")

	req := httptest.NewRequest("GET", "/de-CH", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(test.DecompressGzip(body))

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "gzip")
	test.AssertTrue(t, strings.HasPrefix(content, parts[0]))
	test.AssertTrue(t, strings.HasSuffix(content, parts[1]))
}

func TestIndexRequest_brotli_withVariables(t *testing.T) {
	context, handler := createTestContext_index(t, headers.BROTLI)
	insertVariables(handler.AppVariables)
	parts := strings.Split(context.ReadFile("de-CH/index.html"), "</title>")

	req := httptest.NewRequest("GET", "/de-CH", nil)
	req.Header.Add("Accept-Encoding", "br")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(test.DecompressBrotli(body))

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "br")
	test.AssertTrue(t, strings.HasPrefix(content, parts[0]))
	test.AssertTrue(t, strings.HasSuffix(content, parts[1]))
}

func TestCspIndexRequest_uncompressed(t *testing.T) {
	context, handler := createTestContext_cspIndex(t, headers.NO_COMPRESSION)
	parts := regexp.MustCompile("(</title>|\\${NGSS_CSP_NONCE})").Split(context.ReadFile("de-CH/index.html"), -1)

	req := httptest.NewRequest("GET", "/de-CH", nil)
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	for k, v := range parts {
		if k == 0 {
			test.AssertTrue(t, strings.HasPrefix(content, v))
		} else if k < len(parts)-1 {
			test.AssertTrue(t, strings.Contains(content, v))
		} else {
			test.AssertTrue(t, strings.HasSuffix(content, v))
		}
	}
}

func TestCspIndexRequest_uncompressed_withVariables(t *testing.T) {
	context, handler := createTestContext_cspIndex(t, headers.NO_COMPRESSION)
	insertVariables(handler.AppVariables)
	parts := regexp.MustCompile("(</title>|\\${NGSS_CSP_NONCE})").Split(context.ReadFile("de-CH/index.html"), -1)

	req := httptest.NewRequest("GET", "/de-CH", nil)
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	for k, v := range parts {
		if k == 0 {
			test.AssertTrue(t, strings.HasPrefix(content, v))
		} else if k < len(parts)-1 {
			test.AssertTrue(t, strings.Contains(content, v))
		} else {
			test.AssertTrue(t, strings.HasSuffix(content, v))
		}
	}
}

func TestCspIndexRequest_gzip_withVariables(t *testing.T) {
	context, handler := createTestContext_cspIndex(t, headers.GZIP)
	insertVariables(handler.AppVariables)
	parts := regexp.MustCompile("(</title>|\\${NGSS_CSP_NONCE})").Split(context.ReadFile("de-CH/index.html"), -1)

	req := httptest.NewRequest("GET", "/de-CH", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(test.DecompressGzip(body))

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "gzip")
	for k, v := range parts {
		if k == 0 {
			test.AssertTrue(t, strings.HasPrefix(content, v))
		} else if k < len(parts)-1 {
			test.AssertTrue(t, strings.Contains(content, v))
		} else {
			test.AssertTrue(t, strings.HasSuffix(content, v))
		}
	}
}

func TestCspIndexRequest_brotli_withVariables(t *testing.T) {
	context, handler := createTestContext_cspIndex(t, headers.BROTLI)
	insertVariables(handler.AppVariables)
	parts := regexp.MustCompile("(</title>|\\${NGSS_CSP_NONCE})").Split(context.ReadFile("de-CH/index.html"), -1)

	req := httptest.NewRequest("GET", "/de-CH", nil)
	req.Header.Add("Accept-Encoding", "br")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	content := string(test.DecompressBrotli(body))

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/html; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Cache-Control"), "no-cache")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "br")
	for k, v := range parts {
		if k == 0 {
			test.AssertTrue(t, strings.HasPrefix(content, v))
		} else if k < len(parts)-1 {
			test.AssertTrue(t, strings.Contains(content, v))
		} else {
			test.AssertTrue(t, strings.HasSuffix(content, v))
		}
	}
}

func createTestContext_index(t *testing.T, encoding headers.Encoding) (test.TestDir, IndexEndpoint) {
	context := test.NewTestDir(t)
	context.ImportTestApp("i18n")
	context.CompressFile("de-CH/index.html")
	return context, IndexEndpoint{
		filepath.Join(context.Path, "de-CH/index.html"),
		encoding,
		int(constants.DefaultCompressionThreshold),
		time.Now(),
		config.DefaultAppVariables(),
	}
}

func createTestContext_cspIndex(t *testing.T, encoding headers.Encoding) (test.TestDir, CspIndexEndpoint) {
	context := test.NewTestDir(t)
	context.ImportTestApp("i18n")
	context.CompressFile("de-CH/index.html")
	return context, CspIndexEndpoint{
		filepath.Join(context.Path, "de-CH/index.html"),
		int(constants.DefaultCompressionThreshold),
		config.DefaultAppVariables(),
		constants.CspTemplate,
	}
}

func insertVariables(appVariables *config.AppVariables) {
	value := "value"
	appVariables.MergeVariables(map[string]*string{
		"TEST": &value,
	})
}
