package endpoints

import (
	"net/http"
	"ngstaticserver/serve/headers"
	"os"
	"time"
)

type GzipFileEndpoint struct {
	Path         string
	ModTime      time.Time
	CacheControl string
}

func (endpoint GzipFileEndpoint) Handle(w http.ResponseWriter, r *http.Request, p map[string]string) {
	acceptedEncoding := headers.ResolveAcceptEncoding(r)
	path := endpoint.Path
	if acceptedEncoding.AllowsGzip() {
		path += ".gz"
		w.Header().Set("Content-Encoding", "gzip")
	}
	f, _ := os.Open(path)
	defer f.Close()
	w.Header().Set("Cache-Control", endpoint.CacheControl)
	http.ServeContent(w, r, endpoint.Path, endpoint.ModTime, f)
}
