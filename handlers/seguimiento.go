package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nutricionist/auth"
	"nutricionist/db"
)

func (h *Handler) ListSeguimientos(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	var n int
	h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&n)
	if n == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, fecha, peso, cintura_cm, cadera_cm, brazo_cm, muslo_cm,
		       grasa_corporal, tension_sistolica, tension_diastolica, glucosa, notas
		FROM seguimientos WHERE paciente_id=? ORDER BY fecha DESC, created_at DESC`,
		pacienteID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error cargando seguimientos")
		return
	}
	defer rows.Close()

	type S struct {
		ID                string   `json:"id"`
		Fecha             string   `json:"fecha"`
		Peso              *float64 `json:"peso"`
		CinturaCm         *float64 `json:"cinturaCm"`
		CaderaCm          *float64 `json:"caderaCm"`
		BrazoCm           *float64 `json:"brazoCm"`
		MusloCm           *float64 `json:"musloCm"`
		GrasaCorporal     *float64 `json:"grasaCorporal"`
		TensionSistolica  *int     `json:"tensionSistolica"`
		TensionDiastolica *int     `json:"tensionDiastolica"`
		Glucosa           *float64 `json:"glucosa"`
		Notas             *string  `json:"notas"`
	}

	var list []S
	for rows.Next() {
		var s S
		rows.Scan(&s.ID, &s.Fecha, &s.Peso, &s.CinturaCm, &s.CaderaCm,
			&s.BrazoCm, &s.MusloCm, &s.GrasaCorporal,
			&s.TensionSistolica, &s.TensionDiastolica, &s.Glucosa, &s.Notas)
		list = append(list, s)
	}
	if list == nil {
		list = []S{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) AddSeguimiento(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	var n int
	h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&n)
	if n == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	var body struct {
		Fecha             string   `json:"fecha"`
		Peso              *float64 `json:"peso"`
		CinturaCm         *float64 `json:"cinturaCm"`
		CaderaCm          *float64 `json:"caderaCm"`
		BrazoCm           *float64 `json:"brazoCm"`
		MusloCm           *float64 `json:"musloCm"`
		GrasaCorporal     *float64 `json:"grasaCorporal"`
		TensionSistolica  *int     `json:"tensionSistolica"`
		TensionDiastolica *int     `json:"tensionDiastolica"`
		Glucosa           *float64 `json:"glucosa"`
		Notas             *string  `json:"notas"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if body.Fecha == "" {
		body.Fecha = time.Now().Format("2006-01-02")
	}

	id := db.GenerateID()
	if _, err := h.db.ExecContext(r.Context(), `
		INSERT INTO seguimientos
			(id,paciente_id,fecha,peso,cintura_cm,cadera_cm,brazo_cm,muslo_cm,
			 grasa_corporal,tension_sistolica,tension_diastolica,glucosa,notas)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, pacienteID, body.Fecha, body.Peso, body.CinturaCm, body.CaderaCm,
		body.BrazoCm, body.MusloCm, body.GrasaCorporal,
		body.TensionSistolica, body.TensionDiastolica, body.Glucosa, body.Notas); err != nil {
		writeError(w, http.StatusInternalServerError, "error guardando seguimiento")
		return
	}

	if body.Peso != nil {
		h.db.ExecContext(r.Context(),
			"UPDATE pacientes SET peso_inicial=?,updated_at=CURRENT_TIMESTAMP WHERE id=?",
			body.Peso, pacienteID)
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

func (h *Handler) DeleteSeguimiento(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")
	sid := chi.URLParam(r, "sid")

	res, err := h.db.ExecContext(r.Context(), `
		DELETE FROM seguimientos WHERE id=? AND paciente_id=?
		AND paciente_id IN (SELECT id FROM pacientes WHERE nutricionista_id=?)`,
		sid, pacienteID, userID)
	n, _ := res.RowsAffected()
	if err != nil || n == 0 {
		writeError(w, http.StatusNotFound, "seguimiento no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
