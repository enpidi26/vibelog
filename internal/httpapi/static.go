package httpapi

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	vibelogweb "github.com/enpidi26/vibelog/internal/web"
)

// spaHandler menyajikan frontend yang di-embed ke binary. Kalau build tidak
// menyertakan frontend (mis. `go run` saat development), hasilnya nil dan
// server hanya melayani API.
func spaHandler() http.Handler {
	dist, ok := vibelogweb.FS()
	if !ok {
		return nil
	}
	files := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "." {
			name = "index.html"
		}

		if st, err := fs.Stat(dist, name); err == nil && !st.IsDir() {
			if strings.HasPrefix(name, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			files.ServeHTTP(w, r)
			return
		}

		// File statis yang hilang jangan dibalas index.html: browser akan
		// menyimpannya sebagai HTML dan error-nya menyesatkan.
		if strings.HasPrefix(name, "assets/") || path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}

		index, err := fs.ReadFile(dist, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(index)
	})
}
