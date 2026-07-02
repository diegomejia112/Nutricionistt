package handlers

import (
	"net/http"
)

func (h *Handler) ListAlimentos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	categoria := r.URL.Query().Get("categoria")

	limit := 80
	if q != "" {
		limit = 40 // when searching, fewer results is fine
	}

	query := `SELECT id,nombre,categoria,
		COALESCE(calorias,0),COALESCE(proteinas,0),COALESCE(carbohidratos,0),
		COALESCE(grasas,0),COALESCE(fibra,0),COALESCE(sodio,0),
		COALESCE(potasio,0),COALESCE(calcio,0),COALESCE(hierro,0),porcion_desc
	  FROM alimentos WHERE 1=1`
	args := []any{}

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
		ID            string   `json:"id"`
		Nombre        string   `json:"nombre"`
		Categoria     string   `json:"categoria"`
		Calorias      float64  `json:"calorias"`
		Proteinas     float64  `json:"proteinas"`
		Carbohidratos float64  `json:"carbohidratos"`
		Grasas        float64  `json:"grasas"`
		Fibra         float64  `json:"fibra"`
		Sodio         float64  `json:"sodio"`
		Potasio       float64  `json:"potasio"`
		Calcio        float64  `json:"calcio"`
		Hierro        float64  `json:"hierro"`
		PorcionDesc   *string  `json:"porcionDesc"`
	}

	var alimentos []Alimento
	for rows.Next() {
		var a Alimento
		if err := rows.Scan(&a.ID, &a.Nombre, &a.Categoria, &a.Calorias,
			&a.Proteinas, &a.Carbohidratos, &a.Grasas, &a.Fibra,
			&a.Sodio, &a.Potasio, &a.Calcio, &a.Hierro, &a.PorcionDesc); err != nil {
			continue
		}
		alimentos = append(alimentos, a)
	}
	if alimentos == nil {
		alimentos = []Alimento{}
	}
	writeJSON(w, http.StatusOK, alimentos)
}
