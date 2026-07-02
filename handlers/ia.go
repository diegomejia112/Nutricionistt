package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nutricionist/auth"
	"nutricionist/db"
)

const deepseekURL = "https://api.deepseek.com/chat/completions"

func (h *Handler) GenerarPlanIA(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var body struct {
		PacienteID        string   `json:"pacienteId"`
		NombrePlan        string   `json:"nombrePlan"`
		CaloriasObj       *float64 `json:"caloriasObj"`
		RestriccionesExtra string  `json:"restriccionesExtra"`
		PreferenciaRegion  string  `json:"preferenciaRegion"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if body.PacienteID == "" || body.NombrePlan == "" {
		writeError(w, http.StatusBadRequest, "pacienteId y nombrePlan son requeridos")
		return
	}

	apiKey := h.deepseekKey()
	if apiKey == "" {
		writeError(w, http.StatusServiceUnavailable, "API key de DeepSeek no configurada")
		return
	}

	// Cargar todos los datos del paciente
	var pac pacientePrompt
	err := h.db.QueryRowContext(r.Context(), `
		SELECT nombre, apellidos, sexo,
		       CAST((julianday('now')-julianday(fecha_nacimiento))/365.25 AS INTEGER),
		       peso_inicial, altura, objetivo, patologias, alergias,
		       preferencias, alimentos_evitar, horario_comidas, nivel_actividad, notas
		FROM pacientes WHERE id=? AND nutricionista_id=? AND activo=1`,
		body.PacienteID, userID).
		Scan(&pac.Nombre, &pac.Apellidos, &pac.Sexo, &pac.Edad, &pac.Peso, &pac.Altura,
			&pac.Objetivo, &pac.Enfermedades, &pac.Alergias, &pac.Preferencias,
			&pac.AlimentosEvitar, &pac.HorarioComidas, &pac.NivelActividad, &pac.Notas)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "paciente no encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error cargando paciente")
		return
	}

	// Cargar ingredientes restringidos del paciente
	rRows, _ := h.db.QueryContext(r.Context(),
		"SELECT ingrediente, motivo FROM paciente_restricciones WHERE paciente_id=?", body.PacienteID)
	var ingredientesRestringidos []string
	if rRows != nil {
		defer rRows.Close()
		for rRows.Next() {
			var ing, mot sql.NullString
			rRows.Scan(&ing, &mot)
			if ing.Valid {
				entry := ing.String
				if mot.Valid && mot.String != "" {
					entry += " (" + mot.String + ")"
				}
				ingredientesRestringidos = append(ingredientesRestringidos, entry)
			}
		}
	}

	prompt := buildPrompt(pac, ingredientesRestringidos, body.PreferenciaRegion, body.RestriccionesExtra, body.CaloriasObj)

	reqBody, _ := json.Marshal(map[string]any{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":      8192,
		"response_format": map[string]string{"type": "json_object"},
	})

	client := &http.Client{Timeout: 120 * time.Second}
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, deepseekURL, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "error conectando con DeepSeek")
		return
	}
	defer resp.Body.Close()

	rawData, _ := io.ReadAll(resp.Body)

	var dsResp struct {
		Choices []struct {
			Message struct{ Content string `json:"content"` } `json:"message"`
		} `json:"choices"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := json.Unmarshal(rawData, &dsResp); err != nil || len(dsResp.Choices) == 0 {
		if dsResp.Error != nil {
			writeError(w, http.StatusBadGateway, dsResp.Error.Message)
		} else {
			writeError(w, http.StatusBadGateway, "respuesta invalida de DeepSeek")
		}
		return
	}

	content := dsResp.Choices[0].Message.Content
	content = cleanJSONResponse(content)
	var planJSON map[string]any
	if err := json.Unmarshal([]byte(content), &planJSON); err != nil {
		// Log first 500 chars to help debugging
		preview := content
		if len(preview) > 500 {
			preview = preview[:500]
		}
		fmt.Printf("[IA] JSON parse error: %v\nContent preview: %s\n", err, preview)
		writeError(w, http.StatusInternalServerError, "plan generado no es JSON valido: "+err.Error())
		return
	}

	// Save the plan to DB
	planID := db.GenerateID()
	calStr := ""
	if cal, ok := planJSON["caloriasTotal"].(float64); ok {
		calStr = fmt.Sprintf("%.0f", cal)
	}
	_ = calStr

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error creando plan")
		return
	}
	defer tx.Rollback()

	descripcion, _ := planJSON["descripcion"].(string)
	var calTarget *float64
	if body.CaloriasObj != nil {
		calTarget = body.CaloriasObj
	} else if cal, ok := planJSON["caloriasTotal"].(float64); ok {
		calTarget = &cal
	}

	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO planes (id,paciente_id,nombre,descripcion,fecha_inicio,
			calorias_obj,generado_ia,prompt_ia,activo)
		VALUES (?,?,?,?,?,?,1,?,1)`,
		planID, body.PacienteID, body.NombrePlan, descripcion,
		time.Now().Format("2006-01-02"), calTarget, prompt); err != nil {
		writeError(w, http.StatusInternalServerError, "error guardando plan")
		return
	}

	// Guardar dias, tiempos y alimentos del plan generado por IA
	diasIA, _ := planJSON["dias"].([]any)
	for dia := 1; dia <= 7; dia++ {
		diaID := db.GenerateID()
		if _, err := tx.ExecContext(r.Context(),
			"INSERT INTO dias_plan (id,plan_id,dia_semana) VALUES (?,?,?)",
			diaID, planID, dia); err != nil {
			writeError(w, http.StatusInternalServerError, "error guardando dias")
			return
		}

		// Buscar el dia correspondiente en el JSON de IA
		var tiemposIA []any
		if diasIA != nil && dia-1 < len(diasIA) {
			if diaJSON, ok := diasIA[dia-1].(map[string]any); ok {
				tiemposIA, _ = diaJSON["tiempos"].([]any)
			}
		}

		for _, t := range tiemposBase {
			tiempoID := db.GenerateID()
			if _, err := tx.ExecContext(r.Context(),
				"INSERT INTO tiempos_comida (id,dia_id,nombre,orden) VALUES (?,?,?,?)",
				tiempoID, diaID, t.nombre, t.orden); err != nil {
				writeError(w, http.StatusInternalServerError, "error guardando tiempos")
				return
			}

			// Buscar alimentos de este tiempo en el JSON de IA
			if tiemposIA != nil {
				for _, ti := range tiemposIA {
					tMap, ok := ti.(map[string]any)
					if !ok {
						continue
					}
					tiNombre, _ := tMap["nombre"].(string)
					if tiNombre != t.nombre {
						continue
					}
					alimentos, _ := tMap["alimentos"].([]any)
					for _, al := range alimentos {
						aMap, ok := al.(map[string]any)
						if !ok {
							continue
						}
						nombre, _ := aMap["nombre"].(string)
						cantidad, _ := aMap["cantidad"].(float64)
						unidad, _ := aMap["unidad"].(string)
						calorias, _ := aMap["calorias"].(float64)
						proteinas, _ := aMap["proteinas"].(float64)
						carbos, _ := aMap["carbohidratos"].(float64)
						grasas, _ := aMap["grasas"].(float64)
						if nombre == "" {
							continue
						}
						if cantidad == 0 {
							cantidad = 1
						}
						if unidad == "" {
							unidad = "porcion"
						}
						tx.ExecContext(r.Context(), `
							INSERT INTO alimentos_plan
							  (id,tiempo_id,nombre_libre,cantidad,unidad,calorias,proteinas,carbohidratos,grasas)
							VALUES (?,?,?,?,?,?,?,?,?)`,
							db.GenerateID(), tiempoID, nombre, cantidad, unidad,
							calorias, proteinas, carbos, grasas)
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "error confirmando plan")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"planId": planID,
		"plan":   planJSON,
	})
}

// deepseekKey returns the DeepSeek API key from env var or from the saved file.
func (h *Handler) deepseekKey() string {
	if k := os.Getenv("DEEPSEEK_API_KEY"); k != "" {
		return k
	}
	data, err := os.ReadFile(filepath.Join(h.dataDir, "deepseek.key"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// GetDeepseekConfig returns whether the key is configured (not the key itself).
func (h *Handler) GetDeepseekConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"configurada": h.deepseekKey() != ""})
}

// SaveDeepseekConfig persists the DeepSeek API key to dataDir/deepseek.key.
func (h *Handler) SaveDeepseekConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ApiKey string `json:"apiKey"`
	}
	if err := decodeBody(r, &body); err != nil || body.ApiKey == "" {
		writeError(w, http.StatusBadRequest, "apiKey requerida")
		return
	}
	keyPath := filepath.Join(h.dataDir, "deepseek.key")
	if err := os.WriteFile(keyPath, []byte(strings.TrimSpace(body.ApiKey)), 0600); err != nil {
		writeError(w, http.StatusInternalServerError, "error guardando key")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// cleanJSONResponse strips markdown code fences and extracts the JSON object/array
// that DeepSeek sometimes wraps around its response despite json_object format.
func cleanJSONResponse(s string) string {
	s = strings.TrimSpace(s)

	// Remove ```json ... ``` or ``` ... ``` wrappers
	for _, fence := range []string{"```json", "```JSON", "```"} {
		if strings.HasPrefix(s, fence) {
			s = strings.TrimPrefix(s, fence)
			if idx := strings.LastIndex(s, "```"); idx != -1 {
				s = s[:idx]
			}
			s = strings.TrimSpace(s)
			break
		}
	}

	// Find the outermost JSON object in case there's preamble text
	start := strings.Index(s, "{")
	if start > 0 {
		s = s[start:]
	}

	// Trim trailing text after the closing brace
	if end := strings.LastIndex(s, "}"); end != -1 && end < len(s)-1 {
		s = s[:end+1]
	}

	return strings.TrimSpace(s)
}

type pacientePrompt struct {
	Nombre          string
	Apellidos       string
	Sexo            sql.NullString
	Edad            sql.NullInt64
	Peso            sql.NullFloat64
	Altura          sql.NullFloat64
	Objetivo        sql.NullString
	Enfermedades    sql.NullString
	Alergias        sql.NullString
	Preferencias    sql.NullString
	AlimentosEvitar sql.NullString
	HorarioComidas  sql.NullString
	NivelActividad  sql.NullString
	Notas           sql.NullString
}

func buildPrompt(pac pacientePrompt, ingredientesRestringidos []string, region, restriccionesExtra string, caloriasObj *float64) string {
	sexoStr := "Femenino"
	if pac.Sexo.String == "M" {
		sexoStr = "Masculino"
	}

	calStr := "calcular segun IMC y objetivo"
	if caloriasObj != nil {
		calStr = fmt.Sprintf("%.0f kcal/dia", *caloriasObj)
	}

	pesoStr := "no especificado"
	if pac.Peso.Valid && pac.Peso.Float64 > 0 {
		pesoStr = fmt.Sprintf("%.1f kg", pac.Peso.Float64)
	}
	altStr := "no especificada"
	if pac.Altura.Valid && pac.Altura.Float64 > 0 {
		altStr = fmt.Sprintf("%.0f cm", pac.Altura.Float64)
	}
	edadStr := "no especificada"
	if pac.Edad.Valid && pac.Edad.Int64 > 0 {
		edadStr = fmt.Sprintf("%d anos", pac.Edad.Int64)
	}

	var imcStr string
	if pac.Peso.Valid && pac.Altura.Valid && pac.Peso.Float64 > 0 && pac.Altura.Float64 > 0 {
		imc := pac.Peso.Float64 / ((pac.Altura.Float64 / 100) * (pac.Altura.Float64 / 100))
		imcStr = fmt.Sprintf("%.1f", imc)
	}

	// Construir secciones del prompt
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"Eres un nutriologo experto en el Sistema Mexicano de Alimentos Equivalentes (SMAE) y gastronomia mexicana.\n\n"+
			"Genera un plan de alimentacion semanal completo (7 dias) para el siguiente paciente:\n"+
			"- Nombre: %s %s\n"+
			"- Sexo: %s, Edad: %s\n"+
			"- Peso: %s, Altura: %s",
		pac.Nombre, pac.Apellidos, sexoStr, edadStr, pesoStr, altStr))
	if imcStr != "" {
		sb.WriteString(", IMC: " + imcStr)
	}
	sb.WriteString("\n- Calorias objetivo: " + calStr)

	if pac.Objetivo.Valid && pac.Objetivo.String != "" {
		sb.WriteString("\n- Objetivo clinico: " + pac.Objetivo.String)
	}
	if pac.NivelActividad.Valid && pac.NivelActividad.String != "" {
		sb.WriteString("\n- Nivel de actividad fisica: " + pac.NivelActividad.String)
	}
	if pac.Enfermedades.Valid && pac.Enfermedades.String != "" {
		sb.WriteString("\n- Enfermedades/condiciones medicas: " + pac.Enfermedades.String)
	}
	if pac.Alergias.Valid && pac.Alergias.String != "" {
		sb.WriteString("\n- ALERGIAS E INTOLERANCIAS (PROHIBIDO incluir): " + pac.Alergias.String)
	}
	if pac.AlimentosEvitar.Valid && pac.AlimentosEvitar.String != "" {
		sb.WriteString("\n- Alimentos que NO come / no le gustan (evitar): " + pac.AlimentosEvitar.String)
	}
	if len(ingredientesRestringidos) > 0 {
		sb.WriteString("\n- Ingredientes especificamente restringidos (NO incluir): " + strings.Join(ingredientesRestringidos, ", "))
	}
	if pac.Preferencias.Valid && pac.Preferencias.String != "" {
		sb.WriteString("\n- Alimentos preferidos / que le gustan (priorizar si es posible): " + pac.Preferencias.String)
	}
	if pac.HorarioComidas.Valid && pac.HorarioComidas.String != "" {
		sb.WriteString("\n- Horario de comidas: " + pac.HorarioComidas.String)
	}
	if region != "" {
		sb.WriteString("\n- Preferencia regional de cocina: " + region)
	}
	if restriccionesExtra != "" {
		sb.WriteString("\n- Indicaciones adicionales del nutriologo: " + restriccionesExtra)
	}
	if pac.Notas.Valid && pac.Notas.String != "" {
		sb.WriteString("\n- Notas del expediente: " + pac.Notas.String)
	}

	sb.WriteString(`

REGLAS ESTRICTAS:
1. Usar platillos mexicanos reales y variados cada dia (tacos, enchiladas, pozole, chilaquiles, sopas, guisados, tortas, quesadillas, caldos)
2. Exactamente 5 tiempos: Desayuno, Colacion AM, Comida, Colacion PM, Cena
3. Respetar ABSOLUTAMENTE todas las alergias, intolerancias y restricciones indicadas
4. Usar medidas caseras mexicanas (tazas, cucharadas, piezas, porciones, gramos)
5. Si hay diabetes tipo 2: limitar azucares simples, priorizar leguminosas, verduras, granos enteros; eliminar azucar añadida
6. Si hay hipertension: limitar sodio, evitar embutidos, carnes procesadas, quesos salados
7. Si objetivo es bajar de peso: deficit calorico moderado, alta proteina, mucha fibra
8. Si hay colesterol alto o trigliceridos altos: evitar grasas saturadas, fritos, azucares simples
9. Si hay enfermedad renal: limitar proteina, potasio y fosforo
10. Adaptar porciones al IMC y nivel de actividad del paciente
11. Incluir macronutrientes para CADA alimento (calorias, proteinas, carbohidratos, grasas)
12. BALANCE OBLIGATORIO: cada dia debe incluir alimentos de TODOS los grupos del
    Sistema Mexicano de Alimentos Equivalentes (SMAE) — cereales, verduras,
    frutas, alimentos de origen animal (AOA: carnes/huevo/pescado), leguminosas,
    lacteos, aceites/grasas. NUNCA generes un dia sin fruta o sin verdura, salvo
    que una restriccion medica del paciente lo prohiba explicitamente (ej.
    restriccion de potasio en enfermedad renal puede limitar ciertas frutas,
    pero no eliminar el grupo completo sin justificacion clinica)

Responde SOLO JSON valido con esta estructura exacta:
{
  "descripcion": "Plan semanal para [nombre], enfocado en [objetivo]",
  "caloriasTotal": 2000,
  "distribucionMacros": {"proteinas": 25, "carbos": 50, "grasas": 25},
  "dias": [
    {
      "diaSemana": 1,
      "nombreDia": "Lunes",
      "totalCalorias": 2000,
      "tiempos": [
        {
          "nombre": "Desayuno",
          "totalCalorias": 400,
          "alimentos": [
            {
              "nombre": "Avena con leche descremada y platano",
              "cantidad": 1,
              "unidad": "taza",
              "calorias": 280,
              "proteinas": 8,
              "carbohidratos": 52,
              "grasas": 5
            }
          ]
        }
      ]
    }
  ]
}`)

	return sb.String()
}
