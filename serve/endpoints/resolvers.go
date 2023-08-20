package endpoints

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"log/slog"
	"net/http"
	"ngstaticserver/serve/config"
	"ngstaticserver/serve/headers"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var fingerprintRegex = regexp.MustCompile("\\.[a-zA-Z0-9]{16,}\\.(js|mjs|css)$")

/**
 * We use the following Cache-Control headers:
 *
 * max-age=...: When a file has a fingerprint (hash) in the file name
 * no-cache: When a file is not fingerprinted
 *
 * https://web.dev/http-cache/?hl=en#flowchart
 */

type Endpoint interface {
	Handle(w http.ResponseWriter, r *http.Request, p map[string]string)
}

func VersionEndpoint(filePath string) Endpoint {
	handler, err := ResolveFileEndpoint(filePath, 0)
	if err != nil {
		handler = InlineStringEndpoint{filePath, []byte("{\n  \"undefined\": \"app does not have a version.json file\"\n}")}
	}

	return handler
}

func HeartbeatEndpoint() Endpoint {
	return InlineStringEndpoint{"heartbeat.txt", []byte("UP")}
}

func ResolveFileEndpoint(filePath string, cacheControlMaxAge int64) (Endpoint, error) {
	hasBrotli := fileExists(filePath + ".br")
	hasGzip := fileExists(filePath + ".gz")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	cacheControl := "no-cache"
	if fingerprintRegex.MatchString(filePath) {
		cacheControl = fmt.Sprintf("max-age=%d", cacheControlMaxAge)
	}
	if hasBrotli && hasGzip {
		return BrotliGzipFileEndpoint{filePath, s.ModTime(), cacheControl}, nil
	} else if hasBrotli {
		return BrotliFileEndpoint{filePath, s.ModTime(), cacheControl}, nil
	} else if hasGzip {
		return GzipFileEndpoint{filePath, s.ModTime(), cacheControl}, nil
	} else {
		return UncompressedFileEndpoint{filePath, s.ModTime(), cacheControl}, nil
	}
}

func ResolveIndexEndpoint(filePath string, compressionThreshold int, csp string, appVariables *config.AppVariables) Endpoint {
	var encoding headers.Encoding = headers.NO_COMPRESSION
	if fileExists(filePath + ".br") {
		encoding ^= headers.BROTLI
	}
	if fileExists(filePath + ".gz") {
		encoding ^= headers.GZIP
	}
	content, _ := os.ReadFile(filePath)
	contentAsString := string(content)
	s, _ := os.Stat(filePath)

	if len(csp) > 0 && (strings.Contains(contentAsString, "${NGSS_CSP_NONCE}") || appVariables.Has("NGSS_CSP_NONCE")) {
		csp, err := detectCspTokens(contentAsString, csp)
		if err != nil {
			slog.Warn(fmt.Sprintf("Failed to parse HTML in %v", filePath), "error", err)
		}
		return CspIndexEndpoint{filePath, compressionThreshold, appVariables, csp}
	} else {
		return IndexEndpoint{filePath, encoding, compressionThreshold, s.ModTime(), appVariables}
	}
}

func ResolveRootEndpoint(workingDirectory, i18nDefault string) Endpoint {
	entries, _ := os.ReadDir(workingDirectory)
	paths := make([]string, 0)
	hasDefault := false
	for _, e := range entries {
		if e.IsDir() {
			paths = append(paths, e.Name())
			if !hasDefault && e.Name() == i18nDefault {
				hasDefault = true
			}
		}
	}
	if len(i18nDefault) == 0 {
		i18nDefault = paths[0]
	} else if !hasDefault {
		slog.Warn(fmt.Sprintf("i18n default %v does not exist (%v)", i18nDefault, strings.Join(paths, ", ")))
		i18nDefault = paths[0]
	}

	return RootEndpoint{i18nDefault, paths}
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

func detectCspTokens(content, csp string) (string, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		csp = strings.Replace(csp, "${NGSS_CSP_STYLE_HASH}", "", -1)
		return csp, err
	}
	scriptHashes := []string{}
	styleHashes := []string{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" && requiresCspHash(n) {
			scriptHashes = append(scriptHashes, generateCspHash(n))
		} else if n.Type == html.ElementNode && n.Data == "style" && requiresCspHash(n) {
			styleHashes = append(styleHashes, generateCspHash(n))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	csp = strings.Replace(
		csp,
		"${NGSS_CSP_SCRIPT_HASH}",
		strings.TrimSpace("${NGSS_CSP_SCRIPT_HASH} "+strings.Join(scriptHashes, " ")),
		-1)
	csp = strings.Replace(csp, "${NGSS_CSP_STYLE_HASH}", strings.Join(styleHashes, " "), -1)

	return csp, nil
}

func requiresCspHash(node *html.Node) bool {
	hasNonce := false
	for _, a := range node.Attr {
		if a.Key == "nonce" {
			hasNonce = true
			break
		}
	}
	return !hasNonce && node.FirstChild != nil
}

func generateCspHash(node *html.Node) string {
	buffer := &bytes.Buffer{}
	textContent(node, buffer)
	return fmt.Sprintf("'sha512-%x'", sha512.Sum512(buffer.Bytes()))
}

func textContent(n *html.Node, buffer *bytes.Buffer) {
	if n.Type == html.TextNode {
		buffer.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		textContent(c, buffer)
	}
}
