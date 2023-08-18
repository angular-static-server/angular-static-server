package endpoints

import (
	"fmt"
	"io"
	"net/http/httptest"
	"ngstaticserver/test"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const File = "file.txt"

func TestFileRequest_brotligzip(t *testing.T) {
	context, handler := createTestContext_brotligzip(t)
	content := context.ReadFile(File)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%v", File), nil)
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	test.AssertEqual(t, string(body), content)
}

func TestFileRequestBrotli_brotligzip(t *testing.T) {
	context, handler := createTestContext_brotligzip(t)
	context.CompressFile(File)
	content := context.ReadFile(File)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%v", File), nil)
	req.Header.Add("Accept-Encoding", "br")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "br")
	responseContent := string(test.DecompressBrotli(body))
	test.AssertEqual(t, responseContent, content)
}

func TestFileRequestGzip_brotligzip(t *testing.T) {
	context, handler := createTestContext_brotligzip(t)
	context.CompressFile(File)
	context.RemoveFile(File + ".br")
	content := context.ReadFile(File)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%v", File), nil)
	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	test.AssertEqual(t, resp.Header.Get("Content-Encoding"), "gzip")
	responseContent := string(test.DecompressGzip(body))
	test.AssertEqual(t, responseContent, content)
}

func createTestContext_brotligzip(t *testing.T) (test.TestDir, Endpoint) {
	context := test.NewTestDir(t)
	context.WriteFile(File, strings.Repeat("example", 10))
	return context, BrotliGzipFileEndpoint{filepath.Join(context.Path, File), time.Now(), "no-store"}
}
