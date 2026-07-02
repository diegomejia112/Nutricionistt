package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListAlimentos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	categoria := r.URL.Query().Get("categoria")
	incluirInactivos := r.URL.Query().Get("incluirInactivos") == "1"

	limit := 80
	if q != "" {
		limit = 40 // when searching, fewer results is fine
	}
	// Permite pedir el catálogo completo (ej. panel de referencia SMAE)
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 2000 {
		limit = l
	}

	query := `SELECT id,nombre,categoria,
		COALESCE(calorias,0),COALESCE(proteinas,0),COALESCE(carbohidratos,0),
		COALESCE(grasas,0),COALESCE(fibra,0),COALESCE(sodio,0),
		COALESCE(potasio,0),COALESCE(calcio,0),COALESCE(hierro,0),porcion_desc,
		COALESCE(activo,1)
	  FROM alimentos WHERE 1=1`
	args := []any{}

	if !incluirInactivos {
		query += " AND COALESCE(activo,1)=1"
	}
	if q != "" {
		query += " AND nombre LIKE ?"
		args = append(args, "%"+q+"%")
	}
	if categoria != "" {
		query += " AND categoria = ?"
		args = append(args, categoria)
	}
	query += " ORDER BY nombre LIMIT ?"
	args = append(args, limit)

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error buscando alimentos")
		return
	}
	defer rows.Close()

	type Alimento struct {
		ID            string  `json:"id"`
		Nombre        string  `json:"nombre"`
		Categoria     string  `json:"categoria"`
		Calorias      float64 `json:"calorias"`
		Proteinas     float64 `json:"proteinas"`
		Carbohidratos float64 `json:"carbohidratos"`
		Grasas        float64 `json:"grasas"`
		Fibra         float64 `json:"fibra"`
		Sodio         float64 `json:"sodio"`
		Potasio       float64 `json:"potasio"`
		Calcio        float64 `json:"calcio"`
		Hierro        float64 `json:"hierro"`
		PorcionDesc   *string `json:"porcionDesc"`
		Activo        bool    `json:"activo"`
	}

	var alimentos []Alimento
	for rows.Next() {
		var a Alimento
		if err := rows.Scan(&a.ID, &a.Nombre, &a.Categoria, &a.Calorias,
			&a.Proteinas, &a.Carbohidratos, &a.Grasas, &a.Fibra,
			&a.Sodio, &a.Potasio, &a.Calcio, &a.Hierro, &a.PorcionDesc, &a.Activo); err != nil {
			continue
		}
		alimentos = append(alimentos, a)
	}
	if alimentos == nil {
		alimentos = []Alimento{}
	}
	writeJSON(w, http.StatusOK, alimentos)
}

// UpdateAlimentoActivo activa/desactiva un alimento del catálogo global.
// PATCH /api/alimentos/{id}/activo — Body: {"activo": bool}
func (h *Handler) UpdateAlimentoActivo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		Activo bool `json:"activo"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json inválido")
		return
	}

	res, err := h.db.ExecContext(r.Context(),
		"UPDATE alimentos SET activo=? WHERE id=?", body.Activo, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error actualizando alimento")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "alimento no encontrado")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
