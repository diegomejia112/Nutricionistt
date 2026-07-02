package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nutricionist/auth"
	"nutricionist/db"
)

func (h *Handler) ListConsultas(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	var count int
	if err := h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&count); err != nil || count == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id,fecha,peso,imc,cintura,cadera,
		       masa_grasa,masa_muscular,glucosa,
		       motivo,observaciones,recomendaciones,notas,created_at
		FROM consultas WHERE paciente_id=? ORDER BY fecha DESC`, pacienteID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error obteniendo consultas")
		return
	}
	defer rows.Close()

	type Consulta struct {
		ID              string   `json:"id"`
		Fecha           string   `json:"fecha"`
		Peso            *float64 `json:"peso"`
		IMC             *float64 `json:"imc"`
		Cintura         *float64 `json:"cintura"`
		Cadera          *float64 `json:"cadera"`
		MasaGrasa       *float64 `json:"masaGrasa"`
		MasaMuscular    *float64 `json:"masaMuscular"`
		Glucosa         *float64 `json:"glucosa"`
		Motivo          *string  `json:"motivo"`
		Observaciones   *string  `json:"observaciones"`
		Recomendaciones *string  `json:"recomendaciones"`
		Notas           *string  `json:"notas"`
		CreatedAt       string   `json:"createdAt"`
	}

	var consultas []Consulta
	for rows.Next() {
		var c Consulta
		if err := rows.Scan(&c.ID, &c.Fecha, &c.Peso, &c.IMC,
			&c.Cintura, &c.Cadera, &c.MasaGrasa, &c.MasaMuscular,
			&c.Glucosa, &c.Motivo, &c.Observaciones, &c.Recomendaciones,
			&c.Notas, &c.CreatedAt); err != nil {
			continue
		}
		consultas = append(consultas, c)
	}
	if consultas == nil {
		consultas = []Consulta{}
	}
	writeJSON(w, http.StatusOK, consultas)
}

func (h *Handler) CreateConsulta(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	var count int
	if err := h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&count); err != nil || count == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	var body struct {
		Fecha           string   `json:"fecha"`
		Peso            *float64 `json:"peso"`
		IMC             *float64 `json:"imc"`
		Cintura         *float64 `json:"cintura"`
		Cadera          *float64 `json:"cadera"`
		MasaGrasa       *float64 `json:"masaGrasa"`
		MasaMuscular    *float64 `json:"masaMuscular"`
		AguaCorporal    *float64 `json:"aguaCorporal"`
		PresionSist     *float64 `json:"presionSist"`
		PresionDiast    *float64 `json:"presionDiast"`
		Glucosa         *float64 `json:"glucosa"`
		Motivo          string   `json:"motivo"`
		Observaciones   string   `json:"observaciones"`
		Recomendaciones string   `json:"recomendaciones"`
		Notas           string   `json:"notas"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	if body.IMC == nil && body.Peso != nil {
		var altura *float64
		h.db.QueryRowContext(r.Context(),
			"SELECT altura FROM pacientes WHERE id=?", pacienteID).Scan(&altura)
		if altura != nil && *altura > 0 {
			imc := *body.Peso / ((*altura / 100) * (*altura / 100))
			body.IMC = &imc
		}
	}

	id := db.GenerateID()
	_, err := h.db.ExecContext(r.Context(), `
		INSERT INTO consultas
			(id,paciente_id,fecha,peso,imc,cintura,cadera,
			 masa_grasa,masa_muscular,agua_corporal,
			 presion_sist,presion_diast,glucosa,
			 motivo,observaciones,recomendaciones,notas)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, pacienteID, body.Fecha, body.Peso, body.IMC,
		body.Cintura, body.Cadera, body.MasaGrasa, body.MasaMuscular, body.AguaCorporal,
		body.PresionSist, body.PresionDiast, body.Glucosa,
		body.Motivo, body.Observaciones, body.Recomendaciones, body.Notas)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error guardando consulta")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}
