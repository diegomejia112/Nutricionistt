package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nutricionist/auth"
	"nutricionist/db"
)

var diasNombre = []string{"", "Lunes", "Martes", "Miercoles", "Jueves", "Viernes", "Sabado", "Domingo"}
var tiemposBase = []struct{ nombre string; orden int }{
	{"Desayuno", 1}, {"Colacion AM", 2}, {"Comida", 3}, {"Colacion PM", 4}, {"Cena", 5},
}

func (h *Handler) ListPlanes(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT p.id, p.nombre, p.descripcion, p.fecha_inicio, p.calorias_obj,
		       p.generado_ia, pac.nombre, pac.apellidos
		FROM planes p
		JOIN pacientes pac ON p.paciente_id = pac.id
		WHERE pac.nutricionista_id = ? AND p.activo = 1
		ORDER BY p.created_at DESC LIMIT 50`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error consultando planes")
		return
	}
	defer rows.Close()

	type Plan struct {
		ID          string   `json:"id"`
		Nombre      string   `json:"nombre"`
		Descripcion *string  `json:"descripcion"`
		FechaInicio string   `json:"fechaInicio"`
		CaloriasObj *float64 `json:"caloriasObj"`
		GeneradoIA  bool     `json:"generadoIA"`
		Paciente    struct {
			Nombre    string `json:"nombre"`
			Apellidos string `json:"apellidos"`
		} `json:"paciente"`
	}

	var planes []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.FechaInicio,
			&p.CaloriasObj, &p.GeneradoIA, &p.Paciente.Nombre, &p.Paciente.Apellidos); err != nil {
			continue
		}
		planes = append(planes, p)
	}
	if planes == nil {
		planes = []Plan{}
	}
	writeJSON(w, http.StatusOK, planes)
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var body struct {
		PacienteID   string   `json:"pacienteId"`
		Nombre       string   `json:"nombre"`
		Descripcion  string   `json:"descripcion"`
		FechaInicio  string   `json:"fechaInicio"`
		CaloriasObj  *float64 `json:"caloriasObj"`
		ProteinasObj *float64 `json:"proteinasObj"`
		CarbsObj     *float64 `json:"carbsObj"`
		GrasasObj    *float64 `json:"grasasObj"`
		GeneradoIA   bool     `json:"generadoIA"`
		PromptIA     string   `json:"promptIA"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if body.PacienteID == "" || body.Nombre == "" {
		writeError(w, http.StatusBadRequest, "pacienteId y nombre son requeridos")
		return
	}

	var pid string
	if err := h.db.QueryRowContext(r.Context(),
		"SELECT id FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1",
		body.PacienteID, userID).Scan(&pid); err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}

	fechaInicio := body.FechaInicio
	if fechaInicio == "" {
		fechaInicio = "2006-01-02"
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error iniciando transaccion")
		return
	}
	defer tx.Rollback()

	planID := db.GenerateID()
	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO planes (id,paciente_id,nombre,descripcion,fecha_inicio,
			calorias_obj,proteinas_obj,carbos_obj,grasas_obj,generado_ia,prompt_ia)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		planID, body.PacienteID, body.Nombre, body.Descripcion, fechaInicio,
		body.CaloriasObj, body.ProteinasObj, body.CarbsObj, body.GrasasObj,
		body.GeneradoIA, body.PromptIA); err != nil {
		writeError(w, http.StatusInternalServerError, "error creando plan")
		return
	}

	for dia := 1; dia <= 7; dia++ {
		diaID := db.GenerateID()
		if _, err := tx.ExecContext(r.Context(),
			"INSERT INTO dias_plan (id,plan_id,dia_semana) VALUES (?,?,?)",
			diaID, planID, dia); err != nil {
			writeError(w, http.StatusInternalServerError, "error creando dias")
			return
		}
		for _, t := range tiemposBase {
			if _, err := tx.ExecContext(r.Context(),
				"INSERT INTO tiempos_comida (id,dia_id,nombre,orden) VALUES (?,?,?,?)",
				db.GenerateID(), diaID, t.nombre, t.orden); err != nil {
				writeError(w, http.StatusInternalServerError, "error creando tiempos")
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "error confirmando transaccion")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": planID})
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	planID := chi.URLParam(r, "id")

	type APlan struct {
		ID            string  `json:"id"`
		Cantidad      float64 `json:"cantidad"`
		Unidad        string  `json:"unidad"`
		Notas         *string `json:"notas"`
		Nombre        string  `json:"nombre"`
		Calorias      float64 `json:"calorias"`
		Proteinas     float64 `json:"proteinas"`
		Carbohidratos float64 `json:"carbohidratos"`
		Grasas        float64 `json:"grasas"`
		Tipo          string  `json:"tipo"`      // "alimento" | "platillo" | "ia"
		Categoria     *string `json:"categoria"` // SMAE category, only for alimento type
	}
	type Tiempo struct {
		ID        string  `json:"id"`
		Nombre    string  `json:"nombre"`
		Orden     int     `json:"orden"`
		Alimentos []APlan `json:"alimentos"`
	}
	type Dia struct {
		ID        string   `json:"id"`
		DiaSemana int      `json:"diaSemana"`
		Tiempos   []Tiempo `json:"tiempos"`
	}
	type Plan struct {
		ID           string   `json:"id"`
		Nombre       string   `json:"nombre"`
		Descripcion  *string  `json:"descripcion"`
		FechaInicio  string   `json:"fechaInicio"`
		CaloriasObj  *float64 `json:"caloriasObj"`
		ProteinasObj *float64 `json:"proteinasObj"`
		CarbsObj     *float64 `json:"carbsObj"`
		GrasasObj    *float64 `json:"grasasObj"`
		GeneradoIA   bool     `json:"generadoIA"`
		PacienteID   string   `json:"pacienteId"`
		Dias         []Dia    `json:"dias"`
	}

	var plan Plan
	err := h.db.QueryRowContext(r.Context(), `
		SELECT p.id,p.nombre,p.descripcion,p.fecha_inicio,p.calorias_obj,
		       p.proteinas_obj,p.carbos_obj,p.grasas_obj,p.generado_ia,p.paciente_id
		FROM planes p JOIN pacientes pac ON p.paciente_id=pac.id
		WHERE p.id=? AND pac.nutricionista_id=? AND p.activo=1`, planID, userID).
		Scan(&plan.ID, &plan.Nombre, &plan.Descripcion, &plan.FechaInicio,
			&plan.CaloriasObj, &plan.ProteinasObj, &plan.CarbsObj, &plan.GrasasObj,
			&plan.GeneradoIA, &plan.PacienteID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "plan no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error cargando plan")
		return
	}

	diasRows, err := h.db.QueryContext(r.Context(),
		"SELECT id,dia_semana FROM dias_plan WHERE plan_id=? ORDER BY dia_semana", planID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error cargando dias")
		return
	}
	defer diasRows.Close()

	for diasRows.Next() {
		var dia Dia
		diasRows.Scan(&dia.ID, &dia.DiaSemana)

		tRows, _ := h.db.QueryContext(r.Context(),
			"SELECT id,nombre,orden FROM tiempos_comida WHERE dia_id=? ORDER BY orden", dia.ID)
		if tRows != nil {
			for tRows.Next() {
				var t Tiempo
				tRows.Scan(&t.ID, &t.Nombre, &t.Orden)

				aRows, _ := h.db.QueryContext(r.Context(), `
					SELECT ap.id, ap.cantidad, ap.unidad, ap.notas,
					       COALESCE(a.nombre, pl.nombre, ap.nombre_libre, ''),
					       COALESCE(a.calorias, pl.calorias, ap.calorias, 0),
					       COALESCE(a.proteinas, pl.proteinas, ap.proteinas, 0),
					       COALESCE(a.carbohidratos, pl.carbohidratos, ap.carbohidratos, 0),
					       COALESCE(a.grasas, pl.grasas, ap.grasas, 0),
					       CASE WHEN ap.alimento_id IS NOT NULL THEN 'alimento'
					            WHEN ap.platillo_id IS NOT NULL THEN 'platillo'
					            ELSE 'ia' END,
					       a.categoria
					FROM alimentos_plan ap
					LEFT JOIN alimentos a ON ap.alimento_id = a.id
					LEFT JOIN platillos pl ON ap.platillo_id = pl.id
					WHERE ap.tiempo_id=? ORDER BY ap.rowid`, t.ID)
				if aRows != nil {
					for aRows.Next() {
						var ap APlan
						aRows.Scan(&ap.ID, &ap.Cantidad, &ap.Unidad, &ap.Notas,
							&ap.Nombre, &ap.Calorias, &ap.Proteinas, &ap.Carbohidratos, &ap.Grasas,
							&ap.Tipo, &ap.Categoria)
						t.Alimentos = append(t.Alimentos, ap)
					}
					aRows.Close()
				}
				if t.Alimentos == nil {
					t.Alimentos = []APlan{}
				}
				dia.Tiempos = append(dia.Tiempos, t)
			}
			tRows.Close()
		}
		if dia.Tiempos == nil {
			dia.Tiempos = []Tiempo{}
		}
		plan.Dias = append(plan.Dias, dia)
	}
	if plan.Dias == nil {
		plan.Dias = []Dia{}
	}
	writeJSON(w, http.StatusOK, plan)
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	planID := chi.URLParam(r, "id")

	var body struct {
		Nombre      *string  `json:"nombre"`
		Descripcion *string  `json:"descripcion"`
		CaloriasObj *float64 `json:"caloriasObj"`
		ProteinasObj *float64 `json:"proteinasObj"`
		CarbsObj    *float64 `json:"carbsObj"`
		GrasasObj   *float64 `json:"grasasObj"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	h.db.ExecContext(r.Context(), `
		UPDATE planes SET
			nombre       = COALESCE(?, nombre),
			descripcion  = COALESCE(?, descripcion),
			calorias_obj = COALESCE(?, calorias_obj),
			proteinas_obj= COALESCE(?, proteinas_obj),
			carbos_obj   = COALESCE(?, carbos_obj),
			grasas_obj   = COALESCE(?, grasas_obj)
		WHERE id=? AND paciente_id IN (SELECT id FROM pacientes WHERE nutricionista_id=?)`,
		body.Nombre, body.Descripcion, body.CaloriasObj, body.ProteinasObj,
		body.CarbsObj, body.GrasasObj, planID, userID)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	planID := chi.URLParam(r, "id")
	h.db.ExecContext(r.Context(), `
		UPDATE planes SET activo=0
		WHERE id=? AND paciente_id IN (SELECT id FROM pacientes WHERE nutricionista_id=?)`,
		planID, userID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddAlimentoPlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	planID := chi.URLParam(r, "id")

	var count int
	h.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM planes p
		JOIN pacientes pac ON p.paciente_id=pac.id
		WHERE p.id=? AND pac.nutricionista_id=? AND p.activo=1`, planID, userID).Scan(&count)
	if count == 0 {
		writeError(w, http.StatusNotFound, "plan no encontrado")
		return
	}

	var body struct {
		TiempoID      string   `json:"tiempoId"`
		AlimentoID    string   `json:"alimentoId"`
		PlatilloID    string   `json:"platilloId"`
		NombreLibre   string   `json:"nombreLibre"`
		Cantidad      float64  `json:"cantidad"`
		Unidad        string   `json:"unidad"`
		Calorias      *float64 `json:"calorias"`
		Proteinas     *float64 `json:"proteinas"`
		Carbohidratos *float64 `json:"carbohidratos"`
		Grasas        *float64 `json:"grasas"`
		Notas         string   `json:"notas"`
	}
	if err := decodeBody(r, &body); err != nil || body.TiempoID == "" {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	var aID, pID, nL *string
	if body.AlimentoID != "" {
		aID = &body.AlimentoID
	}
	if body.PlatilloID != "" {
		pID = &body.PlatilloID
	}
	if body.NombreLibre != "" {
		nL = &body.NombreLibre
	}

	id := db.GenerateID()
	if _, err := h.db.ExecContext(r.Context(), `
		INSERT INTO alimentos_plan
			(id,tiempo_id,alimento_id,platillo_id,nombre_libre,cantidad,unidad,calorias,proteinas,carbohidratos,grasas,notas)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, body.TiempoID, aID, pID, nL, body.Cantidad, body.Unidad,
		body.Calorias, body.Proteinas, body.Carbohidratos, body.Grasas, body.Notas); err != nil {
		writeError(w, http.StatusInternalServerError, "error agregando alimento")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *Handler) UpdateAlimentoPlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	planID := chi.URLParam(r, "id")
	alimentoID := chi.URLParam(r, "alimentoId")

	var body struct {
		Cantidad      *float64 `json:"cantidad"`
		Unidad        *string  `json:"unidad"`
		NombreLibre   *string  `json:"nombreLibre"`
		Calorias      *float64 `json:"calorias"`
		Proteinas     *float64 `json:"proteinas"`
		Carbohidratos *float64 `json:"carbohidratos"`
		Grasas        *float64 `json:"grasas"`
		Notas         *string  `json:"notas"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	h.db.ExecContext(r.Context(), `
		UPDATE alimentos_plan SET
			cantidad      = COALESCE(?, cantidad),
			unidad        = COALESCE(?, unidad),
			nombre_libre  = COALESCE(?, nombre_libre),
			calorias      = COALESCE(?, calorias),
			proteinas     = COALESCE(?, proteinas),
			carbohidratos = COALESCE(?, carbohidratos),
			grasas        = COALESCE(?, grasas),
			notas         = COALESCE(?, notas)
		WHERE id=? AND tiempo_id IN (
			SELECT tc.id FROM tiempos_comida tc
			JOIN dias_plan dp ON tc.dia_id=dp.id
			JOIN planes p ON dp.plan_id=p.id
			JOIN pacientes pac ON p.paciente_id=pac.id
			WHERE p.id=? AND pac.nutricionista_id=?
		)`,
		body.Cantidad, body.Unidad, body.NombreLibre,
		body.Calorias, body.Proteinas, body.Carbohidratos, body.Grasas, body.Notas,
		alimentoID, planID, userID)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) RemoveAlimentoPlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	planID := chi.URLParam(r, "id")
	alimentoID := chi.URLParam(r, "alimentoId")

	h.db.ExecContext(r.Context(), `
		DELETE FROM alimentos_plan WHERE id=? AND tiempo_id IN (
			SELECT tc.id FROM tiempos_comida tc
			JOIN dias_plan dp ON tc.dia_id=dp.id
			JOIN planes p ON dp.plan_id=p.id
			JOIN pacientes pac ON p.paciente_id=pac.id
			WHERE p.id=? AND pac.nutricionista_id=?
		)`, alimentoID, planID, userID)

	w.WriteHeader(http.StatusNoContent)
}
