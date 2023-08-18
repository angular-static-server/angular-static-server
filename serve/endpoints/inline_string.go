package endpoints

import (
	"bytes"
	"net/http"
	"time"
)

type InlineStringEndpoint struct {
	Path    string
	Content []byte
}

func (endpoint InlineStringEndpoint) Handle(w http.ResponseWriter, r *http.Request, p map[string]string) {
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, endpoint.Path, time.Now(), bytes.NewReader(endpoint.Content))
}
