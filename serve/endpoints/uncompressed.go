package endpoints

import (
	"net/http"
	"os"
	"time"
)

type UncompressedFileEndpoint struct {
	Path         string
	ModTime      time.Time
	CacheControl string
}

func (endpoint UncompressedFileEndpoint) Handle(w http.ResponseWriter, r *http.Request, p map[string]string) {
	f, _ := os.Open(endpoint.Path)
	defer f.Close()
	w.Header().Set("Cache-Control", endpoint.CacheControl)
	http.ServeContent(w, r, endpoint.Path, endpoint.ModTime, f)
}
