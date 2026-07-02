package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"nutricionist/auth"
)

// ListCasos devuelve todo el catálogo de condiciones clínicas, ordenado por nombre.
// GET /api/casos
func (h *Handler) ListCasos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		"SELECT id, slug, nombre FROM casos ORDER BY nombre")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error al consultar casos")
		return
	}
	defer rows.Close()

	type Caso struct {
		ID     string `json:"id"`
		Slug   string `json:"slug"`
		Nombre string `json:"nombre"`
	}
	var lista []Caso
	for rows.Next() {
		var c Caso
		rows.Scan(&c.ID, &c.Slug, &c.Nombre)
		lista = append(lista, c)
	}
	if lista == nil {
		lista = []Caso{}
	}
	writeJSON(w, http.StatusOK, lista)
}

// GetCondicionesPaciente devuelve las condiciones (casos) asignadas a un paciente.
// GET /api/pacientes/{id}/condiciones
func (h *Handler) GetCondicionesPaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	var count int
	h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&count)
	if count == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT c.id, c.slug, c.nombre
		 FROM paciente_casos pc
		 JOIN casos c ON c.id = pc.caso_id
		 WHERE pc.paciente_id = ?
		 ORDER BY c.nombre`,
		pacienteID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error al consultar condiciones")
		return
	}
	defer rows.Close()

	type Condicion struct {
		ID     string `json:"id"`
		Slug   string `json:"slug"`
		Nombre string `json:"nombre"`
	}
	var lista []Condicion
	for rows.Next() {
		var c Condicion
		rows.Scan(&c.ID, &c.Slug, &c.Nombre)
		lista = append(lista, c)
	}
	if lista == nil {
		lista = []Condicion{}
	}
	writeJSON(w, http.StatusOK, lista)
}

// AddCondicionPaciente asigna una condición (caso) a un paciente.
// POST /api/pacientes/{id}/condiciones — Body: {"casoId": "..."}
func (h *Handler) AddCondicionPaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	var count int
	h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&count)
	if count == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	var body struct {
		CasoID string `json:"casoId"`
	}
	if err := decodeBody(r, &body); err != nil || body.CasoID == "" {
		writeError(w, http.StatusBadRequest, "casoId requerido")
		return
	}

	h.db.ExecContext(r.Context(),
		"INSERT OR IGNORE INTO paciente_casos (paciente_id, caso_id, created_at) VALUES (?, ?, ?)",
		pacienteID, body.CasoID, time.Now())

	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

// DeleteCondicionPaciente elimina una condición (caso) de un paciente.
// DELETE /api/pacientes/{id}/condiciones/{casoId}
func (h *Handler) DeleteCondicionPaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")
	casoID := chi.URLParam(r, "casoId")

	var count int
	h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&count)
	if count == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	h.db.ExecContext(r.Context(),
		"DELETE FROM paciente_casos WHERE paciente_id=? AND caso_id=?",
		pacienteID, casoID)

	w.WriteHeader(http.StatusNoContent)
}
