package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	// Generar el plan DIA POR DIA (7 llamadas cortas en paralelo) en vez de una
	// sola llamada para toda la semana: una respuesta de 7 dias completos con
	// macros por alimento fácilmente pasa de los 8192 tokens máximos de
	// DeepSeek y llega truncada (JSON incompleto). Un solo día es una
	// fracción del tamaño y nunca se acerca a ese límite.
	type diaResultado struct {
		tiempos       []any
		totalCalorias float64
		err           error
	}
	resultados := make([]diaResultado, 7)
	var wg sync.WaitGroup
	for i := 0; i < 7; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			dia := idx + 1
			tiempos, total, err := h.generarDiaIA(r.Context(), apiKey, pac,
				ingredientesRestringidos, body.PreferenciaRegion, body.RestriccionesExtra,
				body.CaloriasObj, dia, diasNombre[dia])
			resultados[idx] = diaResultado{tiempos, total, err}
		}(i)
	}
	wg.Wait()

	var diasFallidos []string
	sumaCalorias := 0.0
	for i, res := range resultados {
		if res.err != nil {
			diasFallidos = append(diasFallidos, diasNombre[i+1])
		} else {
			sumaCalorias += res.totalCalorias
		}
	}
	if len(diasFallidos) > 0 {
		writeError(w, http.StatusBadGateway,
			"no se pudo generar el plan para: "+strings.Join(diasFallidos, ", ")+" (intenta de nuevo)")
		return
	}

	// Save the plan to DB
	planID := db.GenerateID()

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error creando plan")
		return
	}
	defer tx.Rollback()

	descripcion := fmt.Sprintf("Plan semanal para %s %s, generado por IA", pac.Nombre, pac.Apellidos)
	if pac.Objetivo.Valid && pac.Objetivo.String != "" {
		descripcion += " — " + pac.Objetivo.String
	}
	var calTarget *float64
	if body.CaloriasObj != nil {
		calTarget = body.CaloriasObj
	} else if sumaCalorias > 0 {
		promedio := sumaCalorias / 7
		calTarget = &promedio
	}

	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO planes (id,paciente_id,nombre,descripcion,fecha_inicio,
			calorias_obj,generado_ia,prompt_ia,activo)
		VALUES (?,?,?,?,?,?,1,?,1)`,
		planID, body.PacienteID, body.NombrePlan, descripcion,
		time.Now().Format("2006-01-02"), calTarget, "generado dia por dia"); err != nil {
		writeError(w, http.StatusInternalServerError, "error guardando plan")
		return
	}

	// Guardar dias, tiempos y alimentos del plan generado por IA
	for dia := 1; dia <= 7; dia++ {
		diaID := db.GenerateID()
		if _, err := tx.ExecContext(r.Context(),
			"INSERT INTO dias_plan (id,plan_id,dia_semana) VALUES (?,?,?)",
			diaID, planID, dia); err != nil {
			writeError(w, http.StatusInternalServerError, "error guardando dias")
			return
		}

		tiemposIA := resultados[dia-1].tiempos

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

func buildPromptDia(pac pacientePrompt, ingredientesRestringidos []string, region, restriccionesExtra string, caloriasObjDia *float64, diaNum int, diaNombre string) string {
	sexoStr := "Femenino"
	if pac.Sexo.String == "M" {
		sexoStr = "Masculino"
	}

	calStr := "calcular segun IMC y objetivo"
	if caloriasObjDia != nil {
		calStr = fmt.Sprintf("%.0f kcal para este dia", *caloriasObjDia)
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
			"Genera SOLO el dia %d de 7 (%s) de un plan de alimentacion semanal para el "+
			"siguiente paciente. NO generes los otros dias, solo este uno:\n"+
			"- Nombre: %s %s\n"+
			"- Sexo: %s, Edad: %s\n"+
			"- Peso: %s, Altura: %s",
		diaNum, diaNombre, pac.Nombre, pac.Apellidos, sexoStr, edadStr, pesoStr, altStr))
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
1. Usar platillos mexicanos reales (tacos, enchiladas, pozole, chilaquiles, sopas, guisados, tortas, quesadillas, caldos) — distintos a los de otros dias de la semana si es posible
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
12. BALANCE OBLIGATORIO: este dia debe incluir alimentos de TODOS los grupos del
    Sistema Mexicano de Alimentos Equivalentes (SMAE) — cereales, verduras,
    frutas, alimentos de origen animal (AOA: carnes/huevo/pescado), leguminosas,
    lacteos, aceites/grasas. NUNCA generes un dia sin fruta o sin verdura, salvo
    que una restriccion medica del paciente lo prohiba explicitamente (ej.
    restriccion de potasio en enfermedad renal puede limitar ciertas frutas,
    pero no eliminar el grupo completo sin justificacion clinica)
13. Responde UNICAMENTE los datos de este UN dia — no incluyas los otros 6 dias
    de la semana bajo ninguna circunstancia

Responde SOLO JSON valido con esta estructura exacta (un solo dia):
{
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
}`)

	return sb.String()
}

// generarDiaIA genera UN día del plan semanal con DeepSeek. Reintenta una vez
// si la respuesta falla o llega con JSON incompleto (truncado).
func (h *Handler) generarDiaIA(ctx context.Context, apiKey string, pac pacientePrompt,
	ingredientesRestringidos []string, region, restriccionesExtra string,
	caloriasObjSemana *float64, diaNum int, diaNombre string) ([]any, float64, error) {

	var caloriasObjDia *float64
	if caloriasObjSemana != nil {
		v := *caloriasObjSemana / 7
		caloriasObjDia = &v
	}
	prompt := buildPromptDia(pac, ingredientesRestringidos, region, restriccionesExtra, caloriasObjDia, diaNum, diaNombre)

	var lastErr error
	for intento := 0; intento < 2; intento++ {
		tiempos, total, err := llamarDeepSeekDia(ctx, apiKey, prompt)
		if err == nil {
			return tiempos, total, nil
		}
		lastErr = err
		fmt.Printf("[IA] dia %d (%s) intento %d fallo: %v\n", diaNum, diaNombre, intento+1, err)
	}
	return nil, 0, lastErr
}

func llamarDeepSeekDia(ctx context.Context, apiKey, prompt string) ([]any, float64, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":      8192,
		"response_format": map[string]string{"type": "json_object"},
	})

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deepseekURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error conectando con DeepSeek: %w", err)
	}
	defer resp.Body.Close()

	rawData, _ := io.ReadAll(resp.Body)

	var dsResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rawData, &dsResp); err != nil || len(dsResp.Choices) == 0 {
		if dsResp.Error != nil {
			return nil, 0, fmt.Errorf("DeepSeek: %s", dsResp.Error.Message)
		}
		return nil, 0, fmt.Errorf("respuesta invalida de DeepSeek")
	}

	content := cleanJSONResponse(dsResp.Choices[0].Message.Content)
	var diaJSON map[string]any
	if err := json.Unmarshal([]byte(content), &diaJSON); err != nil {
		preview := content
		if len(preview) > 300 {
			preview = preview[:300]
		}
		fmt.Printf("[IA] JSON parse error: %v\nContent preview: %s\n", err, preview)
		return nil, 0, fmt.Errorf("JSON invalido: %w", err)
	}

	tiempos, _ := diaJSON["tiempos"].([]any)
	if len(tiempos) == 0 {
		return nil, 0, fmt.Errorf("respuesta sin tiempos de comida")
	}
	total, _ := diaJSON["totalCalorias"].(float64)
	return tiempos, total, nil
}
