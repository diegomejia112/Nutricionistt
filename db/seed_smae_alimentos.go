package db

import (
	_ "embed"
	"database/sql"
	"strconv"
	"strings"
)

// Catálogo de 2159 alimentos con datos reales de la tabla SMAE de la práctica
// del nutriólogo (planilla "Nutriphasev"), clasificados en las categorías que
// ya usa la app. Los alimentos poco comunes en México (mariscos exóticos,
// quesos importados, carnes de caza, productos empacados de EUA, etc.) quedan
// con activo=0 por default — el nutriólogo puede reactivarlos individualmente
// desde /alimentos si los necesita.
//
//go:embed data/smae_alimentos.tsv
var smaeAlimentosTSV string

// SeedSMAEAlimentos inserta el catálogo SMAE completo (idempotente vía
// INSERT OR IGNORE — no duplica si ya corrió antes o si el nombre ya existe).
func SeedSMAEAlimentos(database *sql.DB) error {
	lines := strings.Split(strings.TrimRight(smaeAlimentosTSV, "\n"), "\n")

	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO alimentos
			(id, nombre, categoria, calorias, proteinas, carbohidratos, grasas,
			 fibra, sodio, potasio, porcion_desc, activo)
		SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM alimentos WHERE nombre = ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, line := range lines {
		cols := strings.Split(line, "\t")
		if len(cols) != 11 {
			continue
		}
		nombre, categoria := cols[0], cols[1]
		calorias, _ := strconv.ParseFloat(cols[2], 64)
		proteinas, _ := strconv.ParseFloat(cols[3], 64)
		carbohidratos, _ := strconv.ParseFloat(cols[4], 64)
		grasas, _ := strconv.ParseFloat(cols[5], 64)
		fibra, _ := strconv.ParseFloat(cols[6], 64)
		sodio, _ := strconv.ParseFloat(cols[7], 64)
		potasio, _ := strconv.ParseFloat(cols[8], 64)
		porcionDesc := cols[9]
		activo := cols[10] == "1"

		id := GenerateID()
		if _, err := stmt.Exec(
			id, nombre, categoria, calorias, proteinas, carbohidratos, grasas,
			fibra, sodio, potasio, porcionDesc, activo, nombre,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
