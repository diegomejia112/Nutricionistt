package db

import (
	"database/sql"
	"fmt"
)

// SeedCasos inserta el catálogo de condiciones clínicas normalizadas y
// migra/etiqueta platillos existentes según reglas heurísticas.
//
// ADVERTENCIA: Las reglas de re-etiquetado son heurísticas aproximadas basadas
// en los datos nutricionales disponibles en el schema actual. NO constituyen
// una validación clínica real. Un nutriólogo debe revisar y ajustar estas
// etiquetas manualmente antes de usarlas con pacientes reales.
func SeedCasos(database *sql.DB) error {
	// 1. Catálogo de casos — slugs estables, nombres display
	casos := []struct {
		slug   string
		nombre string
	}{
		{"diabetes", "Diabetes"},
		{"hipertension", "Hipertensión"},
		{"sobrepeso", "Sobrepeso/Obesidad"},
		{"sop", "Síndrome de Ovario Poliquístico (SOP)"},
		{"hipotiroidismo", "Hipotiroidismo"},
		{"renal", "Enfermedad renal"},
		{"dislipidemia", "Dislipidemia"},
		{"gastritis", "Gastritis/Reflujo"},
		{"colon_irritable", "Colon irritable"},
		{"embarazo", "Embarazo"},
		{"lactancia", "Lactancia"},
		{"anemia", "Anemia"},
		{"higado_graso", "Hígado graso"},
		{"gota", "Gota/Ácido úrico"},
		{"osteoporosis", "Osteoporosis"},
	}

	for _, c := range casos {
		id := GenerateID()
		_, err := database.Exec(
			`INSERT OR IGNORE INTO casos (id, slug, nombre) VALUES (?, ?, ?)`,
			id, c.slug, c.nombre,
		)
		if err != nil {
			return fmt.Errorf("seed caso %s: %w", c.slug, err)
		}
	}

	// 2. Migrar columnas booleanas antiguas → platillo_casos
	migrateOldBool := func(oldColumn string, slug string) error {
		casoID, err := casoIDBySlug(database, slug)
		if err != nil {
			return fmt.Errorf("lookup slug %s: %w", slug, err)
		}
		_, err = database.Exec(`
			INSERT OR IGNORE INTO platillo_casos (platillo_id, caso_id)
			SELECT p.id, ?
			FROM platillos p
			WHERE p.`+oldColumn+` = 1
		`, casoID)
		return err
	}

	if err := migrateOldBool("apto_diabetes", "diabetes"); err != nil {
		return fmt.Errorf("migrate diabetes: %w", err)
	}
	if err := migrateOldBool("apto_hipertension", "hipertension"); err != nil {
		return fmt.Errorf("migrate hipertension: %w", err)
	}
	if err := migrateOldBool("apto_sobrepeso", "sobrepeso"); err != nil {
		return fmt.Errorf("migrate sobrepeso: %w", err)
	}

	// 3. Re-etiquetar platillos para casos nuevos con reglas heurísticas
	type heuristicRule struct {
		slug        string
		whereClause string
	}

	rules := []heuristicRule{
		{slug: "sop", whereClause: `p.apto_diabetes = 1 AND p.bajo_grasa = 1`},
		{slug: "hipotiroidismo", whereClause: `p.apto_diabetes = 1 AND p.bajo_grasa = 1`},
		{slug: "renal", whereClause: `p.sodio_mg <= 400 AND p.proteinas <= 25`},
		{slug: "dislipidemia", whereClause: `p.bajo_grasa = 1 OR p.grasas <= 10`},
		{slug: "gastritis", whereClause: `p.grasas <= 15 AND (p.fibra BETWEEN 1 AND 6)`},
		{slug: "colon_irritable", whereClause: `p.grasas <= 15 AND (p.fibra BETWEEN 1 AND 6)`},
		{slug: "embarazo", whereClause: `p.proteinas >= 15 AND p.categoria != 'Bebida'`},
		{slug: "lactancia", whereClause: `p.proteinas >= 15 AND p.categoria != 'Bebida'`},
		{slug: "anemia", whereClause: `p.proteinas >= 15`},
		{slug: "higado_graso", whereClause: `p.bajo_grasa = 1 AND p.grasas <= 10`},
		{slug: "gota", whereClause: `p.proteinas <= 20`},
		{slug: "osteoporosis", whereClause: `p.categoria IN ('Desayuno','Comida','Cena')`},
	}

	for _, rule := range rules {
		casoID, err := casoIDBySlug(database, rule.slug)
		if err != nil {
			return fmt.Errorf("lookup slug %s: %w", rule.slug, err)
		}
		query := fmt.Sprintf(`
			INSERT OR IGNORE INTO platillo_casos (platillo_id, caso_id)
			SELECT p.id, ?
			FROM platillos p
			WHERE %s
		`, rule.whereClause)
		_, err = database.Exec(query, casoID)
		if err != nil {
			return fmt.Errorf("tag %s: %w", rule.slug, err)
		}
	}

	return nil
}

// casoIDBySlug obtiene el ID del caso dado su slug.
func casoIDBySlug(database *sql.DB, slug string) (string, error) {
	var id string
	err := database.QueryRow(`SELECT id FROM casos WHERE slug = ?`, slug).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("casoIDBySlug(%s): %w", slug, err)
	}
	return id, nil
}
