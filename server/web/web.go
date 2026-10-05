// Package web serves the official project website: a small HTML front end
// for registration, login and skin management. Pages are embedded into the
// binary so deployment needs no separate static files.
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*.html
var staticFiles embed.FS

// Handler returns an http.Handler serving the embedded website under "/".
func Handler() http.Handler {
	mux := http.NewServeMux()

	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		// embed guarantees the directory exists; treat as fatal at build level.
		panic(err)
	}

	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/", indexRedirect{next: fileServer})
	return mux
}

// indexRedirect serves index.html for the site root and delegates otherwise.
type indexRedirect struct {
	next http.Handler
}

func (h indexRedirect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		r.URL.Path = "/skin-manager.html"
	case "/dashboard":
		r.URL.Path = "/dashboard.html"
	case "/admin":
		r.URL.Path = "/admin.html"
	}
	h.next.ServeHTTP(w, r)
}