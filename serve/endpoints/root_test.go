package endpoints

import (
	"net/http"
	"net/http/httptest"
	"ngstaticserver/test"
	"testing"
)

func TestRootRequest_notFound(t *testing.T) {
	handler := RootEndpoint{"de-CH", []string{}}
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, http.StatusNotFound)
}

func TestRootRequest_default(t *testing.T) {
	handler := RootEndpoint{"de-CH", []string{"de-CH"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "en-US")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, http.StatusTemporaryRedirect)
	test.AssertEqual(t, resp.Header.Get("Location"), "/de-CH")
}

func TestRootRequest_exactMatch(t *testing.T) {
	handler := RootEndpoint{"de", []string{"de"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "de")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, http.StatusTemporaryRedirect)
	test.AssertEqual(t, resp.Header.Get("Location"), "/de")
}

func TestRootRequest_partialMatch(t *testing.T) {
	handler := RootEndpoint{"de", []string{"de"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "de-CH")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, http.StatusTemporaryRedirect)
	test.AssertEqual(t, resp.Header.Get("Location"), "/de")
}

func TestRootRequest_partialMatchReversed(t *testing.T) {
	handler := RootEndpoint{"de-CH", []string{"de-CH"}}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "de")
	w := httptest.NewRecorder()
	handler.Handle(w, req, make(map[string]string))

	resp := w.Result()

	test.AssertEqual(t, resp.StatusCode, http.StatusTemporaryRedirect)
	test.AssertEqual(t, resp.Header.Get("Location"), "/de-CH")
}
