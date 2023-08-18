package endpoints

import (
	"net/http"
	"ngstaticserver/serve/headers"
	"os"
	"time"
)

type BrotliGzipFileEndpoint struct {
	Path         string
	ModTime      time.Time
	CacheControl string
}

func (endpoint BrotliGzipFileEndpoint) Handle(w http.ResponseWriter, r *http.Request, p map[string]string) {
	acceptedEncoding := headers.ResolveAcceptEncoding(r)
	path := endpoint.Path
	if acceptedEncoding.AllowsBrotli() {
		path += ".br"
		w.Header().Set("Content-Encoding", "br")
	} else if acceptedEncoding.AllowsGzip() {
		path += ".gz"
		w.Header().Set("Content-Encoding", "gzip")
	}
	f, _ := os.Open(path)
	defer f.Close()
	w.Header().Set("Cache-Control", endpoint.CacheControl)
	http.ServeContent(w, r, endpoint.Path, endpoint.ModTime, f)
}
