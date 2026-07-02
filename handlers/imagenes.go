package handlers

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// inFlight tracks platillo IDs currently being downloaded to avoid duplicate requests.
var (
	inFlight   = map[string]bool{}
	inFlightMu sync.Mutex
)

// ServeImagen sirve la imagen de un platillo.
// Si no está en caché local, la descarga de imagen_url y la guarda.
// Devuelve 202 si está descargando (el frontend muestra skeleton y reintenta).
func (h *Handler) ServeImagen(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	// Sanitize ID — only allow alphanumeric and hyphens
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			http.NotFound(w, r)
			return
		}
	}

	cachePath := filepath.Join(h.imageDir(), id+".jpg")

	// Serve from cache if available
	if _, err := os.Stat(cachePath); err == nil {
		w.Header().Set("Cache-Control", "public, max-age=604800") // 7 days
		http.ServeFile(w, r, cachePath)
		return
	}

	// Look up imagen_url in DB
	var imagenURL sql.NullString
	err := h.db.QueryRowContext(r.Context(),
		"SELECT imagen_url FROM platillos WHERE id = ?", id,
	).Scan(&imagenURL)
	if err != nil || !imagenURL.Valid || imagenURL.String == "" {
		http.NotFound(w, r)
		return
	}

	// Trigger background download if not already in progress
	inFlightMu.Lock()
	alreadyDownloading := inFlight[id]
	if !alreadyDownloading {
		inFlight[id] = true
	}
	inFlightMu.Unlock()

	if !alreadyDownloading {
		go func() {
			defer func() {
				inFlightMu.Lock()
				delete(inFlight, id)
				inFlightMu.Unlock()
			}()
			if err := downloadImage(imagenURL.String, cachePath); err != nil {
				log.Printf("[imagenes] error descargando %s: %v", id, err)
			}
		}()
	}

	// Return 202 — frontend shows skeleton and retries in a few seconds
	w.WriteHeader(http.StatusAccepted)
}

// downloadImage descarga una imagen y la guarda en dst (JPEG).
func downloadImage(url, dst string) error {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(sanitizeURL(url))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil // silently skip bad responses
	}

	// Limit download to 5MB
	body := io.LimitReader(resp.Body, 5<<20)

	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}

	// Write to temp file first, then rename (atomic on most OSes)
	tmp := dst + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err = io.Copy(f, body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, dst)
}

// sanitizeURL ensures only http/https URLs are used.
func sanitizeURL(u string) string {
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		return u
	}
	return ""
}
