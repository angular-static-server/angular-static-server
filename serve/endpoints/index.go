package endpoints

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"net/http"
	"ngstaticserver/compress"
	"ngstaticserver/serve/config"
	"ngstaticserver/serve/headers"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

type IndexEndpoint struct {
	Path                 string
	PreCompression       headers.Encoding
	CompressionThreshold int
	ModTime              time.Time
	AppVariables         *config.AppVariables
}

func (endpoint IndexEndpoint) Handle(w http.ResponseWriter, r *http.Request, p map[string]string) {
	if endpoint.AppVariables.IsEmpty() {
		endpoint.handleEmptyAppConfig(w, r, p)
	} else {
		endpoint.handleAppConfig(w, r, p)
	}
}

func (endpoint IndexEndpoint) handleEmptyAppConfig(w http.ResponseWriter, r *http.Request, p map[string]string) {
	acceptedEncoding := headers.ResolveAcceptEncoding(r)
	path := endpoint.Path
	if acceptedEncoding.AllowsBrotli() && endpoint.PreCompression.ContainsBrotli() {
		path += ".br"
		w.Header().Set("Content-Encoding", "br")
	} else if acceptedEncoding.AllowsGzip() && endpoint.PreCompression.ContainsGzip() {
		path += ".gz"
		w.Header().Set("Content-Encoding", "gzip")
	}
	f, _ := os.Open(path)
	defer f.Close()

	// https://web.dev/http-cache/?hl=en#flowchart
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, endpoint.Path, endpoint.ModTime, f)
}

func (endpoint IndexEndpoint) handleAppConfig(w http.ResponseWriter, r *http.Request, p map[string]string) {
	acceptedEncoding := headers.ResolveAcceptEncoding(r)
	content, _ := os.ReadFile(endpoint.Path)
	content, _ = endpoint.AppVariables.Insert(content, false)

	isAboveThreshold := len(content) >= endpoint.CompressionThreshold
	if isAboveThreshold && acceptedEncoding.AllowsBrotli() {
		content = compress.CompressWithBrotliFast(content)
		w.Header().Set("Content-Encoding", "br")
	} else if isAboveThreshold && acceptedEncoding.AllowsGzip() {
		content = compress.CompressWithGzipFast(content)
		w.Header().Set("Content-Encoding", "gzip")
	}

	// https://web.dev/http-cache/?hl=en#flowchart
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, endpoint.Path, endpoint.AppVariables.LastChangedAt, bytes.NewReader(content))
}

type CspIndexEndpoint struct {
	Path                 string
	CompressionThreshold int
	AppVariables         *config.AppVariables
	Csp                  string
}

func (endpoint CspIndexEndpoint) Handle(w http.ResponseWriter, r *http.Request, p map[string]string) {
	acceptedEncoding := headers.ResolveAcceptEncoding(r)
	content, _ := os.ReadFile(endpoint.Path)
	cspNonce := generateNonce()
	csp := strings.ReplaceAll(endpoint.Csp, "${NGSS_CSP_NONCE}", fmt.Sprintf("'nonce-%v'", cspNonce))
	if !endpoint.AppVariables.IsEmpty() {
		endpoint.AppVariables.Update("NGSS_CSP_NONCE", cspNonce)
		var cspHash string
		content, cspHash = endpoint.AppVariables.Insert(content, true)
		csp = strings.Replace(csp, "${NGSS_CSP_SCRIPT_HASH}", cspHash, -1)
	} else {
		csp = strings.Replace(csp, "${NGSS_CSP_SCRIPT_HASH}", "", -1)
	}

	contentAsString := string(content)
	contentAsString = strings.ReplaceAll(contentAsString, "${NGSS_CSP_NONCE}", cspNonce)
	w.Header().Set("Content-Security-Policy", csp)

	content = []byte(contentAsString)
	isAboveThreshold := len(content) >= endpoint.CompressionThreshold
	if isAboveThreshold && acceptedEncoding.AllowsBrotli() {
		content = compress.CompressWithBrotliFast(content)
		w.Header().Set("Content-Encoding", "br")
	} else if isAboveThreshold && acceptedEncoding.AllowsGzip() {
		content = compress.CompressWithGzipFast(content)
		w.Header().Set("Content-Encoding", "gzip")
	}

	// https://web.dev/http-cache/?hl=en#flowchart
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, endpoint.Path, time.Now(), bytes.NewReader(content))
}

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var runeCharts = []rune(chars)

func generateNonce() string {
	result := make([]byte, 16)
	for i := 0; i < 16; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			slog.Warn("Failed to use secure random to generate CSP nonce. Falling back to less secure variant.")
			localRand := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
			pick := make([]rune, 16)
			for i := range pick {
				pick[i] = runeCharts[localRand.Intn(len(chars))]
			}

			return string(pick)
		}
		result[i] = chars[value.Int64()]
	}

	return string(result)
}
