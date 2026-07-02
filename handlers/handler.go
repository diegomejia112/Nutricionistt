package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	db      *sql.DB
	dataDir string
}

// New creates a Handler with the given database and data directory.
func New(db *sql.DB, dataDir string) *Handler {
	return &Handler{db: db, dataDir: dataDir}
}

func (h *Handler) imageDir() string { return h.dataDir + "/imagenes" }

// json helpers

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
