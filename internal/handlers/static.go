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
			serveFile(w, r, sub, "index.html")
			return
		}
		name := strings.TrimPrefix(clean, "/")
		if _, err := sub.Open(name); err != nil {
			serveFile(w, r, sub, "index.html")
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(name, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if strings.HasSuffix(name, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(name, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}
	w.Write(data)
}