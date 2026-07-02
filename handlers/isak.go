package handlers

import "math"

// Coeficientes Durnin-Womersley (1974) para densidad corporal a partir de la
// suma de 4 pliegues cutáneos (tríceps + bíceps + subescapular + suprailiaco,
// en mm), por sexo y banda de edad. D = a - b*log10(sumaPliegues).
type dwCoef struct{ a, b float64 }

var dwMasculino = []struct {
	edadMax int
	c       dwCoef
}{
	{19, dwCoef{1.1620, 0.0630}},
	{29, dwCoef{1.1631, 0.0632}},
	{39, dwCoef{1.1422, 0.0544}},
	{49, dwCoef{1.1620, 0.0700}},
	{999, dwCoef{1.1715, 0.0779}},
}

var dwFemenino = []struct {
	edadMax int
	c       dwCoef
}{
	{19, dwCoef{1.1549, 0.0678}},
	{29, dwCoef{1.1599, 0.0717}},
	{39, dwCoef{1.1423, 0.0632}},
	{49, dwCoef{1.1333, 0.0612}},
	{999, dwCoef{1.1339, 0.0645}},
}

// CalcularGrasaISAK estima el % de grasa corporal con el método
// Durnin-Womersley (densidad corporal por suma de 4 pliegues) + ecuación de
// Siri. Devuelve (porcentaje, true) si hay datos suficientes, o (0, false)
// si falta algún pliegue, la edad o el sexo.
//
// Esta es una aproximación clínica estándar, no un diagnóstico — depende de
// que los pliegues se tomen con técnica ISAK correcta.
func CalcularGrasaISAK(tricep, biceps, subescapular, suprailiaco *float64, edad int, sexo string) (float64, bool) {
	if tricep == nil || biceps == nil || subescapular == nil || suprailiaco == nil {
		return 0, false
	}
	if edad <= 0 || (sexo != "M" && sexo != "F") {
		return 0, false
	}

	suma := *tricep + *biceps + *subescapular + *suprailiaco
	if suma <= 0 {
		return 0, false
	}

	tabla := dwFemenino
	if sexo == "M" {
		tabla = dwMasculino
	}
	var c dwCoef
	for _, banda := range tabla {
		c = banda.c
		if edad <= banda.edadMax {
			break
		}
	}

	densidad := c.a - c.b*math.Log10(suma)
	if densidad <= 0 {
		return 0, false
	}

	porcentaje := (495 / densidad) - 450
	if porcentaje < 0 {
		porcentaje = 0
	}
	return math.Round(porcentaje*10) / 10, true
}
