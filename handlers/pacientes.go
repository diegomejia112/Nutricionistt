package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"nutricionist/auth"
	"nutricionist/db"
)

func (h *Handler) ListPacientes(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 200 {
		limit = l
	}

	type Row struct {
		ID              string   `json:"id"`
		Nombre          string   `json:"nombre"`
		Apellidos       string   `json:"apellidos"`
		Email           *string  `json:"email"`
		Telefono        *string  `json:"telefono"`
		FechaNacimiento *string  `json:"fecha_nacimiento"`
		Sexo            *string  `json:"sexo"`
		Altura          *float64 `json:"altura"`
		Peso            *float64 `json:"peso"`
		Objetivo        *string  `json:"objetivo"`
		Patologias      *string  `json:"patologias"`
		Edad            *int     `json:"edad"`
	}

	var query string
	var args []any
	cols := `id,nombre,apellidos,email,telefono,fecha_nacimiento,sexo,altura,peso_inicial,objetivo,patologias,
		CAST((julianday('now')-julianday(fecha_nacimiento))/365.25 AS INTEGER) as edad`

	if q != "" {
		like := "%" + q + "%"
		query = `SELECT ` + cols + ` FROM pacientes WHERE nutricionista_id=? AND activo=1
			AND (nombre LIKE ? OR apellidos LIKE ? OR patologias LIKE ?) ORDER BY apellidos,nombre LIMIT ?`
		args = []any{userID, like, like, like, limit}
	} else {
		query = `SELECT ` + cols + ` FROM pacientes WHERE nutricionista_id=? AND activo=1
			ORDER BY apellidos,nombre LIMIT ?`
		args = []any{userID, limit}
	}

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error consultando pacientes")
		return
	}
	defer rows.Close()

	var list []Row
	for rows.Next() {
		var p Row
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Apellidos, &p.Email, &p.Telefono,
			&p.FechaNacimiento, &p.Sexo, &p.Altura, &p.Peso, &p.Objetivo, &p.Patologias, &p.Edad); err != nil {
			continue
		}
		list = append(list, p)
	}
	if list == nil {
		list = []Row{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) CreatePaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var body struct {
		Nombre          string   `json:"nombre"`
		Apellidos       string   `json:"apellidos"`
		Email           string   `json:"email"`
		Telefono        string   `json:"telefono"`
		FechaNacimiento string   `json:"fechaNacimiento"`
		Sexo            string   `json:"sexo"`
		Peso            *float64 `json:"peso"`
		Altura          *float64 `json:"altura"`
		Objetivo        string   `json:"objetivo"`
		Enfermedades    string   `json:"enfermedades"`
		Alergias        string   `json:"alergias"`
		Preferencias    string   `json:"preferencias"`
		AlimentosEvitar string   `json:"alimentosEvitar"`
		HorarioComidas  string   `json:"horarioComidas"`
		NivelActividad  string   `json:"nivelActividad"`
		Notas           string   `json:"notas"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if strings.TrimSpace(body.Nombre) == "" {
		writeError(w, http.StatusBadRequest, "nombre es obligatorio")
		return
	}

	id := db.GenerateID()
	_, err := h.db.ExecContext(r.Context(), `
		INSERT INTO pacientes
			(id,nutricionista_id,nombre,apellidos,email,telefono,fecha_nacimiento,
			 sexo,peso_inicial,altura,objetivo,patologias,alergias,
			 preferencias,alimentos_evitar,horario_comidas,nivel_actividad,notas,activo)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)`,
		id, userID, body.Nombre, body.Apellidos, body.Email, body.Telefono,
		body.FechaNacimiento, body.Sexo, body.Peso, body.Altura,
		body.Objetivo, body.Enfermedades, body.Alergias,
		body.Preferencias, body.AlimentosEvitar, body.HorarioComidas, body.NivelActividad, body.Notas)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error creando paciente")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *Handler) GetPaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	type PacienteRow struct {
		ID              string   `json:"id"`
		Nombre          string   `json:"nombre"`
		Apellidos       string   `json:"apellidos"`
		Email           *string  `json:"email"`
		Telefono        *string  `json:"telefono"`
		FechaNacimiento *string  `json:"fecha_nacimiento"`
		Sexo            *string  `json:"sexo"`
		Altura          *float64 `json:"altura"`
		Peso            *float64 `json:"peso"`
		IMC             *float64 `json:"imc"`
		Objetivo        *string  `json:"objetivo"`
		Enfermedades    *string  `json:"enfermedades"`
		Alergias        *string  `json:"alergias"`
		Preferencias    *string  `json:"preferencias"`
		AlimentosEvitar *string  `json:"alimentosEvitar"`
		HorarioComidas  *string  `json:"horarioComidas"`
		NivelActividad  *string  `json:"nivelActividad"`
		Notas           *string  `json:"notas"`
		Edad            *int     `json:"edad"`
	}
	type ConsultaRow struct {
		ID              string   `json:"id"`
		Fecha           string   `json:"fecha"`
		Peso            *float64 `json:"peso"`
		IMC             *float64 `json:"imc"`
		Glucosa         *float64 `json:"glucosa"`
		Motivo          *string  `json:"motivo"`
		Observaciones   *string  `json:"observaciones"`
		Recomendaciones *string  `json:"recomendaciones"`
		Notas           *string  `json:"notas"`
	}
	type PlanRow struct {
		ID          string `json:"id"`
		Nombre      string `json:"nombre"`
		FechaInicio string `json:"fecha_inicio"`
		GeneradoIA  bool   `json:"generado_ia"`
	}
	type Restriccion struct {
		ID          string  `json:"id"`
		Ingrediente string  `json:"ingrediente"`
		Motivo      *string `json:"motivo"`
	}

	var p PacienteRow
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id,nombre,apellidos,email,telefono,fecha_nacimiento,sexo,
		       altura,peso_inicial,
		       CASE WHEN altura>0 THEN ROUND(peso_inicial/((altura/100)*(altura/100)),1) END,
		       objetivo,patologias,alergias,
		       preferencias,alimentos_evitar,horario_comidas,nivel_actividad,notas,
		       CAST((julianday('now')-julianday(fecha_nacimiento))/365.25 AS INTEGER)
		FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1`, pacienteID, userID).
		Scan(&p.ID, &p.Nombre, &p.Apellidos, &p.Email, &p.Telefono, &p.FechaNacimiento,
			&p.Sexo, &p.Altura, &p.Peso, &p.IMC, &p.Objetivo, &p.Enfermedades,
			&p.Alergias, &p.Preferencias, &p.AlimentosEvitar, &p.HorarioComidas,
			&p.NivelActividad, &p.Notas, &p.Edad)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error cargando paciente")
		return
	}

	cRows, _ := h.db.QueryContext(r.Context(), `
		SELECT id,fecha,peso,imc,glucosa,motivo,observaciones,recomendaciones,notas
		FROM consultas WHERE paciente_id=? ORDER BY fecha DESC LIMIT 20`, pacienteID)
	var consultas []ConsultaRow
	if cRows != nil {
		defer cRows.Close()
		for cRows.Next() {
			var c ConsultaRow
			cRows.Scan(&c.ID, &c.Fecha, &c.Peso, &c.IMC, &c.Glucosa,
				&c.Motivo, &c.Observaciones, &c.Recomendaciones, &c.Notas)
			consultas = append(consultas, c)
		}
	}
	if consultas == nil {
		consultas = []ConsultaRow{}
	}

	pRows, _ := h.db.QueryContext(r.Context(), `
		SELECT id,nombre,fecha_inicio,generado_ia FROM planes
		WHERE paciente_id=? AND activo=1 ORDER BY created_at DESC`, pacienteID)
	var planes []PlanRow
	if pRows != nil {
		defer pRows.Close()
		for pRows.Next() {
			var pl PlanRow
			pRows.Scan(&pl.ID, &pl.Nombre, &pl.FechaInicio, &pl.GeneradoIA)
			planes = append(planes, pl)
		}
	}
	if planes == nil {
		planes = []PlanRow{}
	}

	rRows, _ := h.db.QueryContext(r.Context(), `
		SELECT id,ingrediente,motivo FROM paciente_restricciones
		WHERE paciente_id=? ORDER BY ingrediente`, pacienteID)
	var restricciones []Restriccion
	if rRows != nil {
		defer rRows.Close()
		for rRows.Next() {
			var rr Restriccion
			rRows.Scan(&rr.ID, &rr.Ingrediente, &rr.Motivo)
			restricciones = append(restricciones, rr)
		}
	}
	if restricciones == nil {
		restricciones = []Restriccion{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"paciente":      p,
		"consultas":     consultas,
		"planes":        planes,
		"restricciones": restricciones,
	})
}

func (h *Handler) UpdatePaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")

	var body struct {
		Nombre          *string  `json:"nombre"`
		Apellidos       *string  `json:"apellidos"`
		Email           *string  `json:"email"`
		Telefono        *string  `json:"telefono"`
		FechaNacimiento *string  `json:"fechaNacimiento"`
		Sexo            *string  `json:"sexo"`
		Peso            *float64 `json:"peso"`
		Altura          *float64 `json:"altura"`
		Objetivo        *string  `json:"objetivo"`
		Enfermedades    *string  `json:"enfermedades"`
		Alergias        *string  `json:"alergias"`
		Preferencias    *string  `json:"preferencias"`
		AlimentosEvitar *string  `json:"alimentosEvitar"`
		HorarioComidas  *string  `json:"horarioComidas"`
		NivelActividad  *string  `json:"nivelActividad"`
		Notas           *string  `json:"notas"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	h.db.ExecContext(r.Context(), `
		UPDATE pacientes SET
			nombre=COALESCE(?,nombre), apellidos=COALESCE(?,apellidos),
			email=COALESCE(?,email), telefono=COALESCE(?,telefono),
			fecha_nacimiento=COALESCE(?,fecha_nacimiento), sexo=COALESCE(?,sexo),
			peso_inicial=COALESCE(?,peso_inicial), altura=COALESCE(?,altura),
			objetivo=COALESCE(?,objetivo), patologias=COALESCE(?,patologias),
			alergias=COALESCE(?,alergias), preferencias=COALESCE(?,preferencias),
			alimentos_evitar=COALESCE(?,alimentos_evitar),
			horario_comidas=COALESCE(?,horario_comidas),
			nivel_actividad=COALESCE(?,nivel_actividad),
			notas=COALESCE(?,notas), updated_at=?
		WHERE id=? AND nutricionista_id=? AND activo=1`,
		body.Nombre, body.Apellidos, body.Email, body.Telefono,
		body.FechaNacimiento, body.Sexo, body.Peso, body.Altura,
		body.Objetivo, body.Enfermedades, body.Alergias,
		body.Preferencias, body.AlimentosEvitar, body.HorarioComidas, body.NivelActividad,
		body.Notas, time.Now(), pacienteID, userID)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) DeletePaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")
	h.db.ExecContext(r.Context(), `
		UPDATE pacientes SET activo=0,updated_at=? WHERE id=? AND nutricionista_id=? AND activo=1`,
		time.Now(), pacienteID, userID)
	w.WriteHeader(http.StatusNoContent)
}
