package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"nutricionist/auth"
	"nutricionist/db"
)

type platilloRow struct {
	ID            string  `json:"id"`
	Nombre        string  `json:"nombre"`
	Categoria     string  `json:"categoria"`
	Region        *string `json:"region"`
	PorcionDesc   *string `json:"porcionDesc"`
	PorcionG      float64 `json:"porcionG"`
	Calorias      float64 `json:"calorias"`
	Proteinas     float64 `json:"proteinas"`
	Carbohidratos float64 `json:"carbohidratos"`
	Grasas        float64 `json:"grasas"`
	Fibra         float64 `json:"fibra"`
	ImagenURL     *string `json:"imagen_url"`
	AptoDiabetes  bool    `json:"aptoDiabetes"`
	AptoHiper     bool    `json:"aptoHipertension"`
	AptoSobrepeso bool    `json:"aptoSobrepeso"`
	AltoProteina  bool    `json:"altoProteina"`
	BajoGrasa     bool    `json:"bajoGrasa"`
	Vegetariano   bool    `json:"vegetariano"`
}

const platilloSelect = `SELECT id,nombre,categoria,region,porcion_desc,
	COALESCE(porcion_g,0),COALESCE(calorias,0),COALESCE(proteinas,0),
	COALESCE(carbohidratos,0),COALESCE(grasas,0),COALESCE(fibra,0),imagen_url,
	apto_diabetes,apto_hipertension,apto_sobrepeso,alto_proteina,bajo_grasa,vegetariano
FROM platillos`

func scanPlatillos(rows interface {
	Next() bool
	Scan(...any) error
}) []platilloRow {
	var lista []platilloRow
	for rows.Next() {
		var p platilloRow
		rows.Scan(&p.ID, &p.Nombre, &p.Categoria, &p.Region, &p.PorcionDesc, &p.PorcionG,
			&p.Calorias, &p.Proteinas, &p.Carbohidratos, &p.Grasas, &p.Fibra, &p.ImagenURL,
			&p.AptoDiabetes, &p.AptoHiper, &p.AptoSobrepeso, &p.AltoProteina, &p.BajoGrasa, &p.Vegetariano)
		lista = append(lista, p)
	}
	if lista == nil {
		lista = []platilloRow{}
	}
	return lista
}

func casoFilter(caso string) string {
	switch caso {
	case "diabetes":
		return " AND apto_diabetes=1"
	case "hipertension":
		return " AND apto_hipertension=1"
	case "sobrepeso":
		return " AND apto_sobrepeso=1"
	case "alto_proteina":
		return " AND alto_proteina=1"
	case "bajo_grasa":
		return " AND bajo_grasa=1"
	case "vegetariano":
		return " AND vegetariano=1"
	}
	return ""
}

func (h *Handler) ListPlatillos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	categoria := r.URL.Query().Get("categoria")
	region := r.URL.Query().Get("region")
	caso := r.URL.Query().Get("caso")

	query := platilloSelect + " WHERE 1=1"
	args := []any{}

	if q != "" {
		query += " AND nombre LIKE ?"
		args = append(args, "%"+q+"%")
	}
	if categoria != "" {
		query += " AND categoria=?"
		args = append(args, categoria)
	}
	if region != "" {
		query += " AND region=?"
		args = append(args, region)
	}
	query += casoFilter(caso) + " ORDER BY nombre LIMIT 100"

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error buscando platillos")
		return
	}
	defer rows.Close()
	writeJSON(w, http.StatusOK, scanPlatillos(rows))
}

func (h *Handler) ListVariantes(w http.ResponseWriter, r *http.Request) {
	platilloID := chi.URLParam(r, "id")
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id,nombre,descripcion,COALESCE(calorias,0),COALESCE(proteinas,0),
		       COALESCE(carbohidratos,0),COALESCE(grasas,0),COALESCE(fibra,0),porcion_desc,caso
		FROM platillo_variantes WHERE platillo_base_id=? ORDER BY caso`, platilloID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error cargando variantes")
		return
	}
	defer rows.Close()

	type Variante struct {
		ID            string  `json:"id"`
		Nombre        string  `json:"nombre"`
		Descripcion   *string `json:"descripcion"`
		Calorias      float64 `json:"calorias"`
		Proteinas     float64 `json:"proteinas"`
		Carbohidratos float64 `json:"carbohidratos"`
		Grasas        float64 `json:"grasas"`
		Fibra         float64 `json:"fibra"`
		PorcionDesc   *string `json:"porcionDesc"`
		Caso          *string `json:"caso"`
	}

	var lista []Variante
	for rows.Next() {
		var v Variante
		rows.Scan(&v.ID, &v.Nombre, &v.Descripcion, &v.Calorias, &v.Proteinas,
			&v.Carbohidratos, &v.Grasas, &v.Fibra, &v.PorcionDesc, &v.Caso)
		lista = append(lista, v)
	}
	if lista == nil {
		lista = []Variante{}
	}
	writeJSON(w, http.StatusOK, lista)
}

func (h *Handler) GetRestriccionesPaciente(w http.ResponseWriter, r *http.Request) {
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

	rows, _ := h.db.QueryContext(r.Context(),
		"SELECT id,ingrediente,motivo FROM paciente_restricciones WHERE paciente_id=? ORDER BY ingrediente",
		pacienteID)
	if rows == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	defer rows.Close()

	type Restriccion struct {
		ID          string  `json:"id"`
		Ingrediente string  `json:"ingrediente"`
		Motivo      *string `json:"motivo"`
	}
	var lista []Restriccion
	for rows.Next() {
		var r Restriccion
		rows.Scan(&r.ID, &r.Ingrediente, &r.Motivo)
		lista = append(lista, r)
	}
	if lista == nil {
		lista = []Restriccion{}
	}
	writeJSON(w, http.StatusOK, lista)
}

func (h *Handler) AddRestriccionPaciente(w http.ResponseWriter, r *http.Request) {
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
		Ingrediente string `json:"ingrediente"`
		Motivo      string `json:"motivo"`
	}
	if err := decodeBody(r, &body); err != nil || body.Ingrediente == "" {
		writeError(w, http.StatusBadRequest, "ingrediente requerido")
		return
	}

	id := db.GenerateID()
	h.db.ExecContext(r.Context(),
		"INSERT INTO paciente_restricciones (id,paciente_id,ingrediente,motivo) VALUES (?,?,?,?)",
		id, pacienteID, body.Ingrediente, body.Motivo)

	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *Handler) DeleteRestriccionPaciente(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "id")
	restriccionID := chi.URLParam(r, "restriccionId")

	var count int
	h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&count)
	if count == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	h.db.ExecContext(r.Context(),
		"DELETE FROM paciente_restricciones WHERE id=? AND paciente_id=?",
		restriccionID, pacienteID)
	w.WriteHeader(http.StatusNoContent)
}

// ListPlatillosCompatibles filtra platillos excluyendo los ingredientes restringidos del paciente
func (h *Handler) ListPlatillosCompatibles(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	pacienteID := chi.URLParam(r, "pacienteId")

	var count int
	h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		pacienteID, userID).Scan(&count)
	if count == 0 {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	q := r.URL.Query().Get("q")
	categoria := r.URL.Query().Get("categoria")
	caso := r.URL.Query().Get("caso")

	query := `SELECT DISTINCT p.id,p.nombre,p.categoria,p.region,p.porcion_desc,
		COALESCE(p.porcion_g,0),COALESCE(p.calorias,0),COALESCE(p.proteinas,0),
		COALESCE(p.carbohidratos,0),COALESCE(p.grasas,0),COALESCE(p.fibra,0),p.imagen_url,
		p.apto_diabetes,p.apto_hipertension,p.apto_sobrepeso,p.alto_proteina,p.bajo_grasa,p.vegetariano
	FROM platillos p
	WHERE p.id NOT IN (
		SELECT DISTINCT pi2.platillo_id FROM platillo_ingredientes pi2
		WHERE pi2.grupo IN (SELECT pr.ingrediente FROM paciente_restricciones pr WHERE pr.paciente_id=?)
		   OR pi2.ingrediente IN (SELECT pr.ingrediente FROM paciente_restricciones pr WHERE pr.paciente_id=?)
	)`
	args := []any{pacienteID, pacienteID}

	if q != "" {
		query += " AND p.nombre LIKE ?"
		args = append(args, "%"+q+"%")
	}
	if categoria != "" {
		query += " AND p.categoria=?"
		args = append(args, categoria)
	}
	if caso != "" {
		query += " AND p." + casoColumn(caso) + "=1"
	}
	query += " ORDER BY p.nombre LIMIT 100"

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error buscando platillos compatibles")
		return
	}
	defer rows.Close()
	writeJSON(w, http.StatusOK, scanPlatillos(rows))
}

func casoColumn(caso string) string {
	switch caso {
	case "diabetes":
		return "apto_diabetes"
	case "hipertension":
		return "apto_hipertension"
	case "sobrepeso":
		return "apto_sobrepeso"
	case "alto_proteina":
		return "alto_proteina"
	case "bajo_grasa":
		return "bajo_grasa"
	case "vegetariano":
		return "vegetariano"
	}
	return "apto_diabetes"
}
