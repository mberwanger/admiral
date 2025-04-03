package mux

import (
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	apiPattern = regexp.MustCompile(`^/api/v\d+/`)
)

type assetHandler struct {
	next       http.Handler
	fileSystem http.FileSystem
	fileServer http.Handler
}

func (a *assetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".ico") ||
		strings.HasSuffix(r.URL.Path, ".svg") ||
		strings.HasSuffix(r.URL.Path, ".webp") {
		if !strings.Contains(r.URL.Path[1:], "/") {
			if f, err := a.fileSystem.Open(r.URL.Path); err == nil {
				defer f.Close()
				w.Header().Set("Cache-Control", "public, max-age=86400")
				http.ServeContent(w, r, r.URL.Path, time.Time{}, f)
				return
			}
		}
	}

	if apiPattern.MatchString(r.URL.Path) || r.URL.Path == "/healthcheck" {
		a.next.ServeHTTP(w, r)
		return
	}

	if f, err := a.fileSystem.Open(r.URL.Path); err != nil {
		r.URL.Path = "/"
	} else {
		_ = f.Close()
	}

	a.fileServer.ServeHTTP(w, r)
}
