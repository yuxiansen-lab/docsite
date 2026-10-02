// Package handler wires HTTP routes to the service and static assets.
package handler

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"docsite/internal/model"
	"docsite/internal/service"
)

// Handler is the concrete HTTP handler set.
type Handler struct {
	svc *service.Service
}

// New builds a *Service and returns an http.Handler with all routes
// registered.
func New() (http.Handler, error) {
	svc, err := service.New()
	if err != nil {
		return nil, err
	}
	h := &Handler{svc: svc}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/config", h.handleConfig)
	mux.HandleFunc("/api/menu", h.handleMenu)
	mux.HandleFunc("/api/doc", h.handleDoc)
	mux.HandleFunc("/api/search", h.handleSearch)
	mux.Handle("/docs/assets/", h.assetsHandler())
	mux.Handle("/", h.staticHandler())

	return &security{next: mux}, nil
}

type security struct{ next http.Handler }

func (s *security) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "same-origin")
	s.next.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		// Headers already sent; just log-free abort.
		_ = err
	}
}

func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.svc.Site())
}

func (h *Handler) handleMenu(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.svc.Menu())
}

func (h *Handler) handleDoc(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		writeError(w, http.StatusBadRequest, "missing ?path=")
		return
	}
	doc, err := h.svc.Doc(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to render document")
		return
	}
	writeJSON(w, doc)
}

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	hits := h.svc.Search(q)
	if hits == nil {
		hits = []model.SearchHit{}
	}
	writeJSON(w, map[string]any{"results": hits})
}

// staticHandler serves the embedded web/ tree. Unknown paths without a
// file extension fall back to index.html (SPA routing); paths that look
// like real static files must 404 instead of silently returning HTML.
func (h *Handler) staticHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := h.svc.Web().Open(p); err == nil {
			http.ServeFileFS(w, r, h.svc.Web(), p)
			return
		}
		if path.Ext(p) != "" {
			http.NotFound(w, r)
			return
		}
		http.ServeFileFS(w, r, h.svc.Web(), "index.html")
	})
}

// assetsHandler serves /docs/assets/* from the embedded docs tree.
func (h *Handler) assetsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/docs/assets/")
		if rel == "" || strings.Contains(rel, "..") || strings.ContainsAny(rel, "\x00/") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.ServeFileFS(w, r, h.svc.Assets(), rel)
	})
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
