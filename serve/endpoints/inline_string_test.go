package endpoints

import (
	"fmt"
	"io"
	"net/http/httptest"
	"ngstaticserver/test"
	"testing"
)

func TestFileRequest_inline_string(t *testing.T) {
	handler := createTestContext_inline_string(t)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%v", File), nil)
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	test.AssertEqual(t, resp.StatusCode, 200)
	test.AssertEqual(t, resp.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	test.AssertEqual(t, string(body), string(handler.Content))
}

func createTestContext_inline_string(t *testing.T) InlineStringEndpoint {
	return InlineStringEndpoint{File, []byte("test")}
}
