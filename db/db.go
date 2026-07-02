package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"

	_ "modernc.org/sqlite"
)

// GenerateID generates a UUID v4 using crypto/rand
func GenerateID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		panic(fmt.Sprintf("failed to generate UUID: %v", err))
	}
	// Set version 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant bits
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// Init opens/creates the SQLite database and executes the schema
func Init(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode and foreign keys
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS nutricionistas (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE,
		nombre TEXT,
		apellidos TEXT,
		cedula TEXT,
		password TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS pacientes (
		id TEXT PRIMARY KEY,
		nutricionista_id TEXT,
		nombre TEXT,
		apellidos TEXT,
		fecha_nacimiento DATE,
		sexo TEXT,
		email TEXT,
		telefono TEXT,
		altura REAL,
		peso_inicial REAL,
		objetivo TEXT,
		actividad_fisica TEXT,
		patologias TEXT,
		alergias TEXT,
		notas TEXT,
		activo INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (nutricionista_id) REFERENCES nutricionistas(id)
	);

	CREATE TABLE IF NOT EXISTS consultas (
		id TEXT PRIMARY KEY,
		paciente_id TEXT,
		fecha DATE,
		peso REAL,
		imc REAL,
		cintura REAL,
		cadera REAL,
		brazo_derecho REAL,
		muslo_izq REAL,
		masa_grasa REAL,
		masa_muscular REAL,
		agua_corporal REAL,
		presion_sist REAL,
		presion_diast REAL,
		glucosa REAL,
		motivo TEXT,
		observaciones TEXT,
		recomendaciones TEXT,
		notas TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (paciente_id) REFERENCES pacientes(id)
	);

	CREATE TABLE IF NOT EXISTS planes (
		id TEXT PRIMARY KEY,
		paciente_id TEXT,
		nombre TEXT,
		descripcion TEXT,
		fecha_inicio DATE,
		fecha_fin DATE,
		calorias_obj REAL,
		proteinas_obj REAL,
		carbos_obj REAL,
		grasas_obj REAL,
		generado_ia INTEGER DEFAULT 0,
		prompt_ia TEXT,
		activo INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (paciente_id) REFERENCES pacientes(id)
	);

	CREATE TABLE IF NOT EXISTS dias_plan (
		id TEXT PRIMARY KEY,
		plan_id TEXT,
		dia_semana INTEGER,
		FOREIGN KEY (plan_id) REFERENCES planes(id)
	);

	CREATE TABLE IF NOT EXISTS tiempos_comida (
		id TEXT PRIMARY KEY,
		dia_id TEXT,
		nombre TEXT,
		orden INTEGER,
		FOREIGN KEY (dia_id) REFERENCES dias_plan(id)
	);

	CREATE TABLE IF NOT EXISTS alimentos_plan (
		id TEXT PRIMARY KEY,
		tiempo_id TEXT,
		alimento_id TEXT,
		platillo_id TEXT,
		nombre_libre TEXT,
		cantidad REAL,
		unidad TEXT,
		calorias REAL,
		proteinas REAL,
		carbohidratos REAL,
		grasas REAL,
		notas TEXT,
		FOREIGN KEY (tiempo_id) REFERENCES tiempos_comida(id),
		FOREIGN KEY (alimento_id) REFERENCES alimentos(id),
		FOREIGN KEY (platillo_id) REFERENCES platillos(id)
	);

	CREATE TABLE IF NOT EXISTS alimentos (
		id TEXT PRIMARY KEY,
		nombre TEXT,
		categoria TEXT,
		calorias REAL,
		proteinas REAL,
		carbohidratos REAL,
		grasas REAL,
		fibra REAL,
		sodio REAL,
		potasio REAL,
		calcio REAL,
		hierro REAL,
		porcion_desc TEXT,
		mexicano INTEGER DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS platillos (
		id TEXT PRIMARY KEY,
		nombre TEXT,
		categoria TEXT,
		region TEXT,
		porcion_desc TEXT,
		porcion_g REAL,
		calorias REAL,
		proteinas REAL,
		carbohidratos REAL,
		grasas REAL,
		fibra REAL DEFAULT 0,
		sodio_mg REAL DEFAULT 0,
		imagen_url TEXT,
		apto_diabetes INTEGER DEFAULT 0,
		apto_hipertension INTEGER DEFAULT 0,
		apto_sobrepeso INTEGER DEFAULT 0,
		alto_proteina INTEGER DEFAULT 0,
		bajo_grasa INTEGER DEFAULT 0,
		vegetariano INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Variantes de platillos para diferentes necesidades (ej. version light, sin lacteos, etc.)
	CREATE TABLE IF NOT EXISTS platillo_variantes (
		id TEXT PRIMARY KEY,
		platillo_base_id TEXT NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT,
		calorias REAL,
		proteinas REAL,
		carbohidratos REAL,
		grasas REAL,
		fibra REAL DEFAULT 0,
		porcion_desc TEXT,
		caso TEXT,
		FOREIGN KEY (platillo_base_id) REFERENCES platillos(id)
	);

	-- Ingredientes que componen un platillo (para filtrado por restricciones)
	CREATE TABLE IF NOT EXISTS platillo_ingredientes (
		id TEXT PRIMARY KEY,
		platillo_id TEXT NOT NULL,
		ingrediente TEXT NOT NULL,
		grupo TEXT,
		FOREIGN KEY (platillo_id) REFERENCES platillos(id)
	);

	-- Ingredientes/alimentos restringidos por paciente
	CREATE TABLE IF NOT EXISTS paciente_restricciones (
		id TEXT PRIMARY KEY,
		paciente_id TEXT NOT NULL,
		ingrediente TEXT NOT NULL,
		motivo TEXT,
		FOREIGN KEY (paciente_id) REFERENCES pacientes(id)
	);

	-- Catálogo normalizado de condiciones clínicas (reemplaza las columnas
	-- apto_diabetes/apto_hipertension/apto_sobrepeso para poder escalar a más
	-- condiciones sin migraciones de schema)
	CREATE TABLE IF NOT EXISTS casos (
		id TEXT PRIMARY KEY,
		slug TEXT UNIQUE NOT NULL,
		nombre TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS platillo_casos (
		platillo_id TEXT NOT NULL REFERENCES platillos(id),
		caso_id TEXT NOT NULL REFERENCES casos(id),
		PRIMARY KEY (platillo_id, caso_id)
	);

	-- Condiciones diagnosticadas de un paciente (comorbilidades)
	CREATE TABLE IF NOT EXISTS paciente_casos (
		paciente_id TEXT NOT NULL REFERENCES pacientes(id),
		caso_id TEXT NOT NULL REFERENCES casos(id),
		created_at DATETIME NOT NULL,
		PRIMARY KEY (paciente_id, caso_id)
	);

	-- Seguimiento/evolución del paciente
	CREATE TABLE IF NOT EXISTS seguimientos (
		id TEXT PRIMARY KEY,
		paciente_id TEXT NOT NULL,
		fecha DATE NOT NULL,
		peso REAL,
		cintura_cm REAL,
		cadera_cm REAL,
		brazo_cm REAL,
		muslo_cm REAL,
		grasa_corporal REAL,
		tension_sistolica INTEGER,
		tension_diastolica INTEGER,
		glucosa REAL,
		notas TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (paciente_id) REFERENCES pacientes(id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	// Migrations — ALTER TABLE ignora error si la columna ya existe
	migrations := []string{
		"ALTER TABLE alimentos_plan ADD COLUMN platillo_id TEXT",
		"ALTER TABLE alimentos_plan ADD COLUMN nombre_libre TEXT",
		"ALTER TABLE alimentos_plan ADD COLUMN calorias REAL",
		"ALTER TABLE alimentos_plan ADD COLUMN proteinas REAL",
		"ALTER TABLE alimentos_plan ADD COLUMN carbohidratos REAL",
		"ALTER TABLE alimentos_plan ADD COLUMN grasas REAL",
		// Perfil clínico extendido del paciente
		"ALTER TABLE pacientes ADD COLUMN preferencias TEXT",
		"ALTER TABLE pacientes ADD COLUMN alimentos_evitar TEXT",
		"ALTER TABLE pacientes ADD COLUMN horario_comidas TEXT",
		"ALTER TABLE pacientes ADD COLUMN nivel_actividad TEXT",
		// Columnas nutricionales y clínicas para platillos
		"ALTER TABLE platillos ADD COLUMN fibra REAL DEFAULT 0",
		"ALTER TABLE platillos ADD COLUMN sodio_mg REAL DEFAULT 0",
		"ALTER TABLE platillos ADD COLUMN apto_diabetes INTEGER DEFAULT 0",
		"ALTER TABLE platillos ADD COLUMN apto_hipertension INTEGER DEFAULT 0",
		"ALTER TABLE platillos ADD COLUMN apto_sobrepeso INTEGER DEFAULT 0",
		"ALTER TABLE platillos ADD COLUMN alto_proteina INTEGER DEFAULT 0",
		"ALTER TABLE platillos ADD COLUMN bajo_grasa INTEGER DEFAULT 0",
		"ALTER TABLE platillos ADD COLUMN vegetariano INTEGER DEFAULT 0",
	}
	for _, m := range migrations {
		db.Exec(m)
	}

	if err := SeedData(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to seed data: %w", err)
	}

	// Always runs — adds new SMAE foods to existing databases via INSERT OR IGNORE
	SeedAlimentosSMAE(db)

	// Always runs — adds 100+ healthy platillos via INSERT OR IGNORE
	SeedPlatillosSaludables(db)

	// Always runs — catálogo de condiciones clínicas + re-etiquetado heurístico
	if err := SeedCasos(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to seed casos: %w", err)
	}

	return db, nil
}

func _oldSeedRemoved(db *sql.DB) error {
	// replaced by db/seed.go
	var count int
	db.QueryRow("SELECT COUNT(*) FROM alimentos").Scan(&count)
	if count > 0 {
		return nil
	}

	seed := `
	INSERT OR IGNORE INTO alimentos (id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('a001','Tortilla de maiz','Cereales',204,5.0,44.0,2.5,3.6,'1 tortilla 30g'),
	('a002','Tortilla de harina','Cereales',290,7.5,49.0,7.0,1.8,'1 tortilla 45g'),
	('a003','Arroz blanco cocido','Cereales',130,2.7,28.2,0.3,0.4,'1/2 taza 100g'),
	('a004','Frijol negro cocido','Leguminosas',132,8.9,23.7,0.5,8.7,'1/2 taza 90g'),
	('a005','Frijol bayo cocido','Leguminosas',127,8.7,22.8,0.5,7.4,'1/2 taza 90g'),
	('a006','Pollo pechuga sin piel','Carnes',165,31.0,0.0,3.6,0.0,'100g'),
	('a007','Res molida 90/10','Carnes',176,22.0,0.0,9.3,0.0,'100g'),
	('a008','Atun en agua','Carnes',108,23.6,0.0,1.0,0.0,'1 lata escurrida 100g'),
	('a009','Salmon fresco','Carnes',208,20.4,0.0,13.4,0.0,'100g'),
	('a010','Huevo entero','Lacteos',143,12.6,0.7,9.5,0.0,'2 piezas 100g'),
	('a011','Leche entera','Lacteos',61,3.2,4.8,3.3,0.0,'1 taza 240ml'),
	('a012','Leche descremada','Lacteos',34,3.4,5.0,0.1,0.0,'1 taza 240ml'),
	('a013','Queso Oaxaca','Lacteos',318,22.1,1.3,24.8,0.0,'30g'),
	('a014','Queso panela','Lacteos',260,18.0,3.5,19.5,0.0,'30g'),
	('a015','Queso fresco','Lacteos',264,17.8,3.4,20.0,0.0,'30g'),
	('a016','Yogurt natural sin grasa','Lacteos',59,10.3,6.5,0.4,0.0,'3/4 taza 150g'),
	('a017','Jitomate','Verduras',18,0.9,3.9,0.2,1.2,'1 pieza 100g'),
	('a018','Chile poblano','Verduras',43,2.0,9.4,0.5,3.7,'1 pieza 60g'),
	('a019','Cebolla','Verduras',40,1.1,9.3,0.1,1.7,'1/2 taza 75g'),
	('a020','Ajo','Verduras',149,6.4,33.1,0.5,2.1,'3 dientes 10g'),
	('a021','Nopales cocidos','Verduras',22,1.8,4.4,0.1,3.3,'1/2 taza 90g'),
	('a022','Calabaza zuchini','Verduras',17,1.2,3.1,0.3,1.0,'1/2 taza 90g'),
	('a023','Chayote','Verduras',19,0.8,4.5,0.1,1.7,'1/2 taza 80g'),
	('a024','Espinacas','Verduras',23,2.9,3.6,0.4,2.2,'1 taza cocida 90g'),
	('a025','Brocoli','Verduras',34,2.8,6.6,0.4,2.6,'1/2 taza 80g'),
	('a026','Zanahoria','Verduras',41,0.9,9.6,0.2,2.8,'1 pieza 80g'),
	('a027','Pepino','Verduras',16,0.7,3.6,0.1,0.5,'1/2 pepino 100g'),
	('a028','Aguacate','Aceites',160,2.0,8.5,14.7,6.7,'1/4 pieza 50g'),
	('a029','Aceite de oliva','Aceites',884,0.0,0.0,100.0,0.0,'1 cucharada 14g'),
	('a030','Aceite vegetal','Aceites',884,0.0,0.0,100.0,0.0,'1 cucharada 14g'),
	('a031','Manzana','Frutas',52,0.3,13.8,0.2,2.4,'1 pieza 150g'),
	('a032','Platano','Frutas',89,1.1,22.8,0.3,2.6,'1 pieza 120g'),
	('a033','Naranja','Frutas',47,0.9,11.8,0.1,2.4,'1 pieza 150g'),
	('a034','Guayaba','Frutas',68,2.6,14.3,1.0,5.4,'2 piezas 100g'),
	('a035','Papaya','Frutas',43,0.5,10.8,0.3,1.7,'1 taza 140g'),
	('a036','Mango','Frutas',60,0.8,15.0,0.4,1.6,'1/2 pieza 100g'),
	('a037','Sandia','Frutas',30,0.6,7.6,0.2,0.4,'1 taza 150g'),
	('a038','Melon','Frutas',34,0.8,8.2,0.2,0.9,'1 taza 160g'),
	('a039','Fresa','Frutas',32,0.7,7.7,0.3,2.0,'1 taza 150g'),
	('a040','Uvas','Frutas',67,0.6,17.2,0.4,0.9,'1 taza 150g'),
	('a041','Pan integral','Cereales',247,8.5,47.3,3.4,6.3,'2 rebanadas 60g'),
	('a042','Avena','Cereales',389,16.9,66.3,6.9,10.6,'1/2 taza seca 40g'),
	('a043','Amaranto','Cereales',371,13.6,65.3,7.0,6.7,'1/4 taza 30g'),
	('a044','Nopal tuna','Frutas',41,0.8,9.9,0.5,3.6,'1 pieza 100g'),
	('a045','Jicama','Verduras',38,0.7,8.8,0.1,4.9,'1 taza 120g'),
	('a046','Ejotes','Verduras',31,1.8,7.0,0.2,3.4,'1/2 taza 80g'),
	('a047','Chile serrano','Verduras',32,1.7,6.7,0.4,3.7,'2 chiles 30g'),
	('a048','Cilantro','Verduras',23,2.1,3.7,0.5,2.8,'2 cucharadas 5g'),
	('a049','Limon','Frutas',29,1.1,9.3,0.3,2.8,'1 pieza 50g'),
	('a050','Epazote','Verduras',32,3.3,7.3,0.5,3.8,'2 cucharadas 5g');

	INSERT OR IGNORE INTO platillos (id,nombre,categoria,region,porcion_desc,porcion_g,calorias,proteinas,carbohidratos,grasas,imagen_url) VALUES
	('p001','Chilaquiles rojos con pollo','Desayuno','Centro',  '1 plato',300, 420,28.0,38.0,14.0,'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=400'),
	('p002','Enchiladas verdes','Comida',    'Centro',  '3 piezas',350, 480,25.0,52.0,18.0,'https://images.unsplash.com/photo-1565299585323-38d6b0865b47?w=400'),
	('p003','Tacos de canasta','Comida',     'Centro',  '3 tacos', 210, 390,14.0,51.0,15.0,'https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=400'),
	('p004','Pozole rojo','Comida',          'Centro',  '1 plato', 450, 380,28.0,35.0,12.0,'https://images.unsplash.com/photo-1574484284002-952d92456975?w=400'),
	('p005','Tamales de rajas','Desayuno',   'Centro',  '2 piezas',200, 340,8.0, 48.0,12.0,'https://images.unsplash.com/photo-1551504734-5ee1c4a1479b?w=400'),
	('p006','Sopa de lima','Comida',         'Sureste', '1 plato', 400, 290,22.0,28.0,8.0, 'https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400'),
	('p007','Poc chuc','Comida',             'Sureste', '1 plato', 350, 370,32.0,12.0,20.0,'https://images.unsplash.com/photo-1529042410759-befb1204b468?w=400'),
	('p008','Cochinita pibil','Comida',      'Sureste', '1 plato', 300, 420,30.0,18.0,22.0,'https://images.unsplash.com/photo-1599974579688-8dbdd335c77f?w=400'),
	('p009','Mole negro con guajolote','Comida','Sur',  '1 plato', 400, 520,35.0,30.0,26.0,'https://images.unsplash.com/photo-1613514785940-daed07799d9b?w=400'),
	('p010','Tlayuda oaxaquena','Comida',    'Sur',     '1 pieza', 350, 580,22.0,72.0,20.0,'https://images.unsplash.com/photo-1565299507177-b0ac66763828?w=400'),
	('p011','Tacos de barbacoa','Comida',    'Norte',   '3 tacos', 240, 450,32.0,36.0,18.0,'https://images.unsplash.com/photo-1551504734-5ee1c4a1479b?w=400'),
	('p012','Caldo de pollo','Comida',       'Nacional','1 plato', 450, 220,20.0,18.0,7.0, 'https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400'),
	('p013','Arroz rojo a la mexicana','Comida','Nacional','1/2 taza',150,180,3.5,36.0,3.5,'https://images.unsplash.com/photo-1603133872878-684f208fb84b?w=400'),
	('p014','Frijoles de olla','Comida',     'Nacional','1/2 taza',200,160,9.5,26.0,2.5,'https://images.unsplash.com/photo-1588166524941-3bf61a9c41db?w=400'),
	('p015','Quesadillas de flor de calabaza','Colacion','Centro','2 piezas',180,380,14.0,46.0,14.0,'https://images.unsplash.com/photo-1565299507177-b0ac66763828?w=400'),
	('p016','Huevos rancheros','Desayuno',   'Nacional','1 plato', 300, 360,18.0,28.0,18.0,'https://images.unsplash.com/photo-1529062500409-b06d81bf74ec?w=400'),
	('p017','Omelette de nopal','Desayuno',  'Nacional','1 plato', 250, 240,16.0,8.0, 16.0,'https://images.unsplash.com/photo-1565961132236-ab6f3f23fb32?w=400'),
	('p018','Avena con frutas','Desayuno',   'Nacional','1 taza',  280, 310,10.0,52.0,6.0, 'https://images.unsplash.com/photo-1517673400267-0251440c45dc?w=400'),
	('p019','Sopa de verduras','Comida',     'Nacional','1 plato', 400, 180,6.0, 28.0,5.0, 'https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400'),
	('p020','Pechuga a la plancha con verduras','Cena','Nacional','1 plato',300,280,35.0,12.0,9.0,'https://images.unsplash.com/photo-1529042410759-befb1204b468?w=400'),
	('p021','Ceviche de camaron','Colacion', 'Costa',   '1 taza',  200, 180,18.0,12.0,5.0, 'https://images.unsplash.com/photo-1580822184713-fc5400e7fe10?w=400'),
	('p022','Tacos de pescado','Comida',     'Norte',   '3 tacos', 270, 420,22.0,46.0,15.0,'https://images.unsplash.com/photo-1565299585323-38d6b0865b47?w=400'),
	('p023','Sopa de lentejas','Comida',     'Nacional','1 plato', 400, 230,14.0,36.0,4.0, 'https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400'),
	('p024','Ensalada de nopales','Colacion','Nacional','1 taza',  200, 90, 3.0, 14.0,2.5, 'https://images.unsplash.com/photo-1580822184713-fc5400e7fe10?w=400'),
	('p025','Pan dulce conchas','Desayuno',  'Nacional','2 piezas',120, 380,6.0, 68.0,8.0, 'https://images.unsplash.com/photo-1585937421612-70a008356fbe?w=400'),
	('p026','Agua de jamaica','Bebida',      'Nacional','1 vaso',  300, 60, 0.0, 15.0,0.0, 'https://images.unsplash.com/photo-1560023907-5f339617ea55?w=400'),
	('p027','Atole de guayaba','Bebida',     'Nacional','1 taza',  240, 140,3.0, 30.0,1.5, 'https://images.unsplash.com/photo-1609501676725-7186f017a4b7?w=400'),
	('p028','Licuado de platano','Bebida',   'Nacional','1 vaso',  300, 220,7.0, 42.0,3.5, 'https://images.unsplash.com/photo-1553530666-ba11a90a0868?w=400'),
	('p029','Tacos de nopales con frijoles','Cena','Nacional','3 tacos',240,280,11.0,44.0,6.0,'https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=400'),
	('p030','Sopa de lima con tortillas','Cena','Sureste','1 plato',350,260,18.0,30.0,7.0,'https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400');
	`

	_, err := db.Exec(seed)
	return err
}
