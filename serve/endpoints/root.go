package endpoints

import (
	"fmt"
	"net/http"
	"strings"
)

type RootEndpoint struct {
	DefaultPath    string
	AvailablePaths []string
}

func (endpoint RootEndpoint) Handle(w http.ResponseWriter, r *http.Request, p map[string]string) {
	acceptLanguage := strings.Split(r.Header.Get("Accept-Language"), ",")
	if match := endpoint.MatchLanguage(acceptLanguage); match != "" {
		w.Header().Set("Location", fmt.Sprintf("/%v", match))
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (endpoint RootEndpoint) MatchLanguage(languages []string) string {
	if len(endpoint.AvailablePaths) == 0 {
		return ""
	}

	for _, l := range languages {
		for _, a := range endpoint.AvailablePaths {
			if l == a {
				return l
			}
		}
	}

	for _, l := range languages {
		for _, a := range endpoint.AvailablePaths {
			if strings.HasPrefix(a, l) || strings.HasPrefix(l, a) {
				return a
			}
		}
	}

	return endpoint.DefaultPath
}
