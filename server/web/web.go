// Package web serves the embedded Alpha Network web application.
//
// The frontend is a small vanilla ES-module SPA (no build step). Serving it
// from the Go binary keeps the launcher self-contained and lets the same UI
// run in Qt WebEngine without a separate static host.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed static
var staticFiles embed.FS

// appRoutes are the SPA entry points. They all return index.html and the
// client-side router decides which view to mount.
var appRoutes = map[string]struct{}{
	"/login":     {},
	"/register":  {},
	"/dashboard": {},
	"/players":   {},
	"/admin":     {},
}

// legacyRedirects keeps old bookmarks and launcher links working after the
// single-page app replaced the separate HTML files.
var legacyRedirects = map[string]string{
	"/skin-manager.html": "/dashboard",
	"/dashboard.html":    "/dashboard",
	"/admin.html":        "/admin",
	"/login.html":        "/login",
	"/register.html":     "/register",
	"/players.html":      "/players",
}

// Handler returns an http.Handler serving the embedded website under "/".
func Handler() http.Handler {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		// embed guarantees the directory exists; treat as fatal at build level.
		panic(err)
	}

	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		panic("web: embedded index.html is missing: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(sub))
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		requestPath := r.URL.Path
		if target, ok := legacyRedirects[requestPath]; ok {
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		if requestPath == "/favicon.ico" {
			// The SVG favicon is linked from index.html; answer the implicit
			// legacy request quietly instead of producing a 404 in Qt logs.
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Never expose directory listings from the embedded bundle.
		if requestPath != "/" && strings.HasSuffix(requestPath, "/") {
			http.NotFound(w, r)
			return
		}

		if requestPath == "/" || isAppRoute(requestPath) || !hasFileExtension(requestPath) {
			serveIndex(w, r, index)
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux
}

func isAppRoute(requestPath string) bool {
	_, ok := appRoutes[requestPath]
	return ok
}

func hasFileExtension(requestPath string) bool {
	return path.Ext(requestPath) != ""
}

func serveIndex(w http.ResponseWriter, r *http.Request, index []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(index)
}
