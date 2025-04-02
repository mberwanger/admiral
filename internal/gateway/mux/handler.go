package mux

import (
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	apiPattern         = regexp.MustCompile(`^/api/v\d+/`)
	staticRoutePattern = regexp.MustCompile(`^/static*`)
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
		// Serve from the embedded API handler.
		a.next.ServeHTTP(w, r)
		return
	}

	// Check if assets are okay to serve by calling the Fetch endpoint and verifying it returns a 200.
	//rec := httptest.NewRecorder()
	origPath := r.URL.Path
	//r.URL.Path = "/v1/assets/fetch"
	//a.next.ServeHTTP(rec, r)
	//
	//if rec.Code != http.StatusOK {
	//	copyHTTPResponse(rec.Result(), w)
	//	return
	//}

	// Set the original path.
	r.URL.Path = origPath

	// if enableStaticBaseRoute is set to true, we wont attempt to serve assets if there is no extension in the path.
	// This is to prevent serving the SPA when the user is trying to access a nested route.
	//if a.isStaticPathRoutable(r.URL.Path) {
	//	r.URL.Path = "/"
	//	a.fileServer.ServeHTTP(w, r)
	//}

	// Serve!
	if f, err := a.fileSystem.Open(r.URL.Path); err != nil {
		//// If not a known static asset and an asset provider is configured, try streaming from the configured provider.
		//if a.assetCfg != nil && a.assetCfg.Provider != nil && strings.HasPrefix(r.URL.Path, staticAssetPath) {
		//	// We attach this header simply for observability purposes.
		//	// Otherwise its difficult to know if the assets are being served from the configured provider.
		//	w.Header().Set("x-clutch-asset-passthrough", "true")
		//
		//	asset, err := a.assetProviderHandler(r.Context(), r.URL.Path)
		//	if err != nil {
		//		w.WriteHeader(http.StatusInternalServerError)
		//		_, _ = w.Write([]byte(fmt.Sprintf("Error getting assets from the configured asset provider: %v", err)))
		//		return
		//	}
		//	defer asset.Close()
		//
		//	_, err = io.Copy(w, asset)
		//	if err != nil {
		//		w.WriteHeader(http.StatusInternalServerError)
		//		_, _ = w.Write([]byte(fmt.Sprintf("Error getting assets from the configured asset provider: %v", err)))
		//		return
		//	}
		//	return
		//}

		// If not a known static asset serve the SPA.
		r.URL.Path = "/"
	} else {
		_ = f.Close()
	}

	a.fileServer.ServeHTTP(w, r)
}
