package handlers

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/khannover/asylum-wall-app/internal/web"
)

func Static() http.Handler {
	sub, err := fs.Sub(web.FS, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean(r.URL.Path)
		if clean == "/" || clean == "." {
			serveIndex(w, sub)
			return
		}
		name := strings.TrimPrefix(clean, "/")
		if _, err := sub.Open(name); err != nil {
			serveIndex(w, sub)
			return
		}
		setAssetCache(w, name)
		fileServer.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, fsys fs.FS) {
	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	html := string(data)
	v := web.Version
	html = strings.ReplaceAll(html, "/assets/style.css", "/assets/style.css?v="+v)
	html = strings.ReplaceAll(html, "/assets/app.js", "/assets/app.js?v="+v)
	html = strings.ReplaceAll(html, "/favicon.svg", "/favicon.svg?v="+v)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Write([]byte(html))
}

func setAssetCache(w http.ResponseWriter, name string) {
	if strings.HasSuffix(name, ".html") {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		return
	}
	if strings.HasPrefix(name, "assets/") || name == "favicon.svg" {
		// Versioned URLs (?v=...) can be cached; rebuild changes the query string.
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	}
}

