package db

import "database/sql"

// SeedAlimentosSMAE inserts the full SMAE food catalog. Uses INSERT OR IGNORE so
// it is safe to call on existing databases — new items are added, existing ones kept.
func SeedAlimentosSMAE(db *sql.DB) {
	seedVerduras(db)
	seedCerealesTuberculos(db)
	seedLeguminosasExtra(db)
	seedCarnesExtra(db)
	seedLacteosExtra(db)
	seedAceitesGrasas(db)
	seedFrutasExtra(db)
	seedAzucares(db)
	seedCondimentos(db)
}

func seedVerduras(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('v001','Acelgas cocidas','Verduras',20,1.9,3.6,0.2,2.1,'1 taza 120g'),
	('v002','Apio','Verduras',16,0.7,3.0,0.2,1.6,'1 taza 100g'),
	('v003','Berros','Verduras',11,2.3,0.5,0.1,0.5,'1 taza 70g'),
	('v004','Betabel cocido','Verduras',44,1.7,9.6,0.2,2.8,'1/2 taza 85g'),
	('v005','Berenjena cocida','Verduras',25,1.0,5.9,0.2,3.0,'1/2 taza 80g'),
	('v006','Cebollín (cebolla cambray)','Verduras',32,1.8,7.3,0.2,2.6,'1/4 taza 50g'),
	('v007','Champiñones frescos','Verduras',22,3.1,3.3,0.3,1.0,'1 taza 70g'),
	('v008','Chicharos frescos','Verduras',81,5.4,14.5,0.4,5.1,'1/2 taza 80g'),
	('v009','Chile jalapeño','Verduras',27,0.9,5.9,0.4,2.8,'2 chiles 30g'),
	('v010','Col (repollo) cruda','Verduras',25,1.3,5.8,0.1,2.5,'1 taza 70g'),
	('v011','Coliflor cocida','Verduras',25,1.9,5.0,0.3,2.0,'1 taza 80g'),
	('v012','Esparragos cocidos','Verduras',20,2.2,3.9,0.1,2.1,'6 piezas 90g'),
	('v013','Flor de calabaza','Verduras',20,1.2,4.3,0.2,1.5,'1 taza cruda 33g'),
	('v014','Hongos portobello','Verduras',26,2.5,3.9,0.5,1.3,'1 taza 70g'),
	('v015','Huitlacoche cocido','Verduras',42,3.7,7.7,0.4,2.5,'1/2 taza 70g'),
	('v016','Pimiento rojo','Verduras',31,1.0,6.0,0.3,2.1,'1/2 pieza 75g'),
	('v017','Pimiento verde','Verduras',20,0.9,4.6,0.2,1.7,'1/2 pieza 75g'),
	('v018','Quelites (quintoniles)','Verduras',27,2.5,4.4,0.3,2.4,'1 taza 60g'),
	('v019','Rabano','Verduras',16,0.7,3.4,0.1,1.6,'5 piezas 50g'),
	('v020','Verdolaga','Verduras',20,2.0,3.4,0.3,0.7,'1 taza 40g'),
	('v021','Xoconostle','Verduras',36,0.8,8.2,0.3,4.5,'1 pieza 80g'),
	('v022','Camote amarillo cocido','Verduras',90,2.0,20.7,0.1,3.3,'1/2 taza 100g'),
	('v023','Platano macho cocido','Verduras',122,1.0,31.9,0.4,2.3,'1/2 pieza 75g'),
	('v024','Yuca cocida','Verduras',112,1.0,26.8,0.2,1.8,'1/2 taza 80g'),
	('v025','Elote desgranado cocido','Verduras',96,3.4,21.0,1.4,2.4,'1/2 taza 77g'),
	('v026','Chayote cocido','Verduras',19,0.8,4.5,0.1,1.7,'1/2 taza 80g'),
	('v027','Nopalitos en vinagre','Verduras',14,1.0,2.8,0.1,2.0,'1/2 taza 80g'),
	('v028','Flor de Jamaica seca','Verduras',44,1.6,11.3,0.1,3.2,'2 cucharadas 10g'),
	('v029','Hierbabuena fresca','Verduras',44,3.3,8.4,0.7,6.8,'2 cucharadas 5g'),
	('v030','Chile chipotle en adobo','Verduras',58,2.7,9.2,1.7,3.4,'2 cucharadas 30g'),
	('v031','Ajo rostizado','Verduras',136,5.8,30.2,0.5,1.9,'3 dientes 12g'),
	('v032','Cebolla morada','Verduras',38,1.0,8.8,0.1,1.5,'1/2 taza 75g'),
	('v033','Poro (puerro)','Verduras',61,1.5,14.2,0.3,1.8,'1 pieza 90g'),
	('v034','Huauzontle','Verduras',28,2.8,4.5,0.3,2.2,'1 taza 50g'),
	('v035','Verduras mixtas cocidas','Verduras',65,3.2,13.0,0.3,4.0,'1/2 taza 80g'),
	('v036','Lechuga orejona','Verduras',13,1.3,2.2,0.3,1.4,'2 tazas 80g'),
	('v037','Espinacas crudas','Verduras',23,2.9,3.6,0.4,2.2,'2 tazas 60g'),
	('v038','Apio nabo (tuberosa)','Verduras',42,1.5,9.2,0.3,1.8,'1/2 taza 80g'),
	('v039','Chiles anchos secos','Verduras',281,12.0,49.8,7.3,17.3,'1 pieza 20g'),
	('v040','Chile guajillo seco','Verduras',318,14.4,55.5,9.8,21.6,'2 piezas 20g')`)
}

func seedCerealesTuberculos(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('c001','Bolillo (pan de sal)','Cereales',270,8.0,54.0,2.0,2.0,'1 pieza 80g'),
	('c002','Pan blanco de caja','Cereales',265,8.5,50.0,3.2,2.5,'2 rebanadas 50g'),
	('c003','Pan dulce (concha)','Cereales',360,7.0,58.0,10.0,1.0,'1 pieza 70g'),
	('c004','Pan dulce (cuernito)','Cereales',350,8.0,52.0,12.0,1.0,'1 pieza 65g'),
	('c005','Tostadas de maiz','Cereales',420,8.5,76.0,11.0,6.0,'2 piezas 28g'),
	('c006','Galletas Maria','Cereales',429,6.6,74.0,11.5,1.7,'5 piezas 25g'),
	('c007','Papa cocida con cáscara','Cereales',87,1.9,20.1,0.1,1.8,'1 pieza mediana 100g'),
	('c008','Papa cambray cocida','Cereales',69,1.7,15.9,0.1,1.5,'5 piezas 100g'),
	('c009','Camote morado cocido','Cereales',86,1.6,20.1,0.1,3.0,'1/2 taza 100g'),
	('c010','Pasta espagueti cocida','Cereales',131,5.0,25.0,1.1,1.8,'1/2 taza 100g'),
	('c011','Fideo seco cocido','Cereales',130,4.8,24.5,1.0,1.6,'1/2 taza 100g'),
	('c012','Macarron cocido','Cereales',131,4.9,25.1,1.1,1.8,'1/2 taza 100g'),
	('c013','Maiz cacahuazintle (pozole)','Cereales',362,9.2,74.1,4.7,9.2,'1/4 taza seco 50g'),
	('c014','Masa de maiz nixtamalizada','Cereales',195,5.3,40.6,2.1,4.0,'1/4 kg 100g'),
	('c015','Granola con miel','Cereales',471,10.5,64.2,19.6,7.0,'1/3 taza 40g'),
	('c016','Quinoa cocida','Cereales',120,4.4,21.3,1.9,2.8,'1/2 taza 93g'),
	('c017','Cereal de maiz (cornflakes)','Cereales',357,7.5,79.4,0.4,3.0,'1 taza 30g'),
	('c018','Cereal de avena inflada','Cereales',367,12.8,66.6,6.9,10.0,'1 taza 30g'),
	('c019','Pan pita integral','Cereales',275,9.1,55.7,1.2,7.9,'1 pieza 60g'),
	('c020','Arroz integral cocido','Cereales',111,2.6,23.0,0.9,1.8,'1/2 taza 100g'),
	('c021','Polenta (harina de maiz gruesa)','Cereales',362,8.1,76.8,3.6,7.3,'1/4 taza seca 40g'),
	('c022','Pinole','Cereales',393,9.0,74.3,7.2,7.0,'2 cucharadas 30g'),
	('c023','Galletas integrales','Cereales',380,8.0,62.0,12.0,6.0,'5 galletas 25g'),
	('c024','Crepas de harina','Cereales',196,6.0,26.0,7.0,1.0,'2 piezas 60g'),
	('c025','Memela de maiz','Cereales',220,5.0,44.0,3.5,3.0,'1 pieza 80g'),
	('c026','Gordita de masa','Cereales',195,4.8,39.0,2.5,2.5,'1 pieza 70g'),
	('c027','Tlayuda (tortilla grande)','Cereales',330,9.0,66.0,4.0,5.0,'1 pieza 90g'),
	('c028','Tlacoyo de frijol','Cereales',210,7.0,38.0,4.5,4.5,'1 pieza 80g'),
	('c029','Sopes de maiz','Cereales',228,5.5,43.5,4.5,3.5,'1 pieza 80g'),
	('c030','Harinas para atole (Maseca)','Cereales',361,6.4,76.5,3.8,5.3,'3 cucharadas 30g')`)
}

func seedLeguminosasExtra(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('lg001','Soya texturizada seca','Leguminosas',300,52.0,33.5,1.0,3.5,'1/4 taza seca 28g'),
	('lg002','Edamame cocido','Leguminosas',122,11.9,8.9,5.2,5.2,'1/2 taza 78g'),
	('lg003','Alverjon seco cocido','Leguminosas',118,8.3,21.1,0.4,8.3,'1/2 taza 80g'),
	('lg004','Frijol peruano (mayo) cocido','Leguminosas',130,8.5,23.5,0.5,7.5,'1/2 taza 90g'),
	('lg005','Frijol flor de mayo cocido','Leguminosas',128,8.8,22.9,0.6,8.1,'1/2 taza 90g'),
	('lg006','Haba verde cocida','Leguminosas',88,8.0,15.8,0.6,5.4,'1/2 taza 85g'),
	('lg007','Cacahuate tostado sin sal','Leguminosas',567,25.8,16.1,49.2,8.5,'2 cucharadas 28g'),
	('lg008','Pasta de garbanzo (hummus)','Leguminosas',177,4.9,20.1,8.6,4.0,'3 cucharadas 50g'),
	('lg009','Frijol negro en lata','Leguminosas',120,7.6,22.0,0.5,7.5,'1/2 taza 130g'),
	('lg010','Soya leche sin azucar','Leguminosas',33,2.9,1.7,1.8,0.4,'1 taza 240ml')`)
}

func seedCarnesExtra(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	-- Pescados y mariscos
	('p001','Bacalao seco desalado cocido','Carnes',167,36.0,0.0,1.5,0.0,'100g cocido'),
	('p002','Ostion en agua cocido','Carnes',68,7.1,3.9,2.5,0.0,'6 piezas 90g'),
	('p003','Jaiba cocida','Carnes',87,18.0,0.0,1.1,0.0,'100g cocida'),
	('p004','Pulpo cocido','Carnes',82,14.9,2.2,1.0,0.0,'100g cocido'),
	('p005','Charales secos','Carnes',260,48.0,0.0,7.0,0.0,'20g secos'),
	('p006','Huachinango al horno','Carnes',105,22.2,0.0,1.3,0.0,'100g cocido'),
	('p007','Merluza cocida','Carnes',86,17.8,0.0,1.2,0.0,'100g cocida'),
	('p008','Robalo cocido','Carnes',124,23.7,0.0,2.6,0.0,'100g cocido'),
	('p009','Trucha arcoiris cocida','Carnes',149,20.7,0.0,6.6,0.0,'100g cocida'),
	('p010','Mahi-mahi (dorado) cocido','Carnes',109,23.7,0.0,0.9,0.0,'100g cocido'),
	('p011','Calamar cocido','Carnes',92,15.6,3.1,1.4,0.0,'100g cocido'),
	('p012','Mejillon cocido','Carnes',86,11.9,3.7,2.2,0.0,'10 piezas 85g'),
	('p013','Anguila ahumada','Carnes',184,18.4,0.0,11.7,0.0,'50g'),
	-- Aves
	('p014','Pollo entero asado sin piel','Carnes',185,24.5,0.0,9.2,0.0,'100g cocido'),
	('p015','Ala de pollo sin piel cocida','Carnes',203,26.0,0.0,10.6,0.0,'2 alas 60g'),
	('p016','Pechuga de pavo a la plancha','Carnes',135,29.9,0.0,1.0,0.0,'100g cocida'),
	('p017','Muslo de pavo sin piel','Carnes',195,25.0,0.0,10.0,0.0,'100g cocido'),
	-- Carnes rojas
	('p018','Res carne molida 80-20 cocida','Carnes',254,17.2,0.0,20.0,0.0,'100g cocida'),
	('p019','Cerdo costilla cocida','Carnes',292,17.1,0.0,24.8,0.0,'100g cocida'),
	('p020','Cerdo falda asada','Carnes',242,22.8,0.0,16.0,0.0,'100g cocida'),
	('p021','Res carne maciza de bola','Carnes',158,26.0,0.0,5.5,0.0,'100g cocida'),
	('p022','Cordero pierna asada','Carnes',206,22.0,0.0,13.0,0.0,'100g cocido'),
	-- Frios y embutidos
	('p023','Jamon de pavo sin grasa','Carnes',105,17.6,2.5,2.8,0.0,'2 rebanadas 60g'),
	('p024','Jamon de cerdo ahumado','Carnes',215,16.0,1.5,16.3,0.0,'2 rebanadas 60g'),
	('p025','Chorizo de cerdo crudo','Carnes',455,17.8,2.9,40.6,0.0,'50g crudo'),
	('p026','Chorizo de pavo cocido','Carnes',218,17.5,3.8,15.0,0.0,'50g cocido'),
	('p027','Salchicha de puerco cocida','Carnes',290,12.0,1.3,26.0,0.0,'2 piezas 60g'),
	('p028','Salchicha de pavo','Carnes',196,14.0,3.5,14.0,0.0,'2 piezas 60g'),
	('p029','Longaniza cocida','Carnes',310,14.5,3.8,27.0,0.0,'50g cocida'),
	('p030','Tocino de cerdo frito','Carnes',541,37.0,1.4,42.0,0.0,'2 rebanadas 20g'),
	('p031','Mortadela','Carnes',261,11.0,3.0,23.0,0.0,'2 rebanadas 50g'),
	-- Otros
	('p032','Atun en aceite escurrido','Carnes',184,25.5,0.0,9.0,0.0,'1 lata 100g'),
	('p033','Sardinas en aceite escurridas','Carnes',208,24.6,0.0,11.5,0.0,'1 lata 100g'),
	('p034','Salmon ahumado','Carnes',177,23.0,0.0,9.3,0.0,'50g'),
	('p035','Huevo de codorniz cocido','Carnes',158,13.0,0.4,11.1,0.0,'5 piezas 50g')`)
}

func seedLacteosExtra(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('lc001','Leche semidescremada 1pct','Lacteos',42,3.4,4.9,1.0,0.0,'1 taza 240ml'),
	('lc002','Leche deslactosada entera','Lacteos',59,3.2,5.0,3.0,0.0,'1 taza 240ml'),
	('lc003','Leche de almendra sin azucar','Lacteos',15,0.6,0.6,1.2,0.4,'1 taza 240ml'),
	('lc004','Leche de avena sin azucar','Lacteos',46,1.0,8.0,1.0,0.5,'1 taza 240ml'),
	('lc005','Leche de coco sin azucar','Lacteos',19,0.2,1.9,1.8,0.2,'1 taza 240ml'),
	('lc006','Yogurt griego natural entero','Lacteos',100,9.0,3.6,5.0,0.0,'150g'),
	('lc007','Yogurt griego 0pct grasa','Lacteos',59,10.0,3.6,0.4,0.0,'150g'),
	('lc008','Yogurt natural bebible','Lacteos',63,3.0,8.0,2.0,0.0,'1 vasito 200ml'),
	('lc009','Queso manchego mexicano','Lacteos',370,25.0,1.0,29.8,0.0,'30g'),
	('lc010','Queso manchego rallado','Lacteos',380,26.0,1.0,30.5,0.0,'30g'),
	('lc011','Queso crema Philadelphia','Lacteos',342,6.2,4.1,34.0,0.0,'2 cucharadas 30g'),
	('lc012','Queso cottage bajo grasa','Lacteos',98,11.0,3.4,4.3,0.0,'1/2 taza 113g'),
	('lc013','Jocoque seco','Lacteos',228,14.5,7.4,16.2,0.0,'3 cucharadas 50g'),
	('lc014','Jocoque liquido','Lacteos',90,7.5,9.0,2.5,0.0,'1/2 taza 120ml'),
	('lc015','Requeson bajo grasa','Lacteos',113,11.4,5.6,5.0,0.0,'1/2 taza 113g'),
	('lc016','Crema light 10pct grasa','Lacteos',100,2.8,4.0,8.0,0.0,'2 cucharadas 30g'),
	('lc017','Queso amarillo (american)','Lacteos',304,16.7,5.2,25.0,0.0,'2 rebanadas 40g'),
	('lc018','Queso de cabra suave','Lacteos',268,18.5,2.5,21.4,0.0,'30g'),
	('lc019','Natilla sin azucar','Lacteos',98,3.6,12.0,4.0,0.5,'1/2 taza 125g'),
	('lc020','Yogurt con fruta azucarado','Lacteos',120,4.0,22.0,2.0,0.5,'1 vasito 150g')`)
}

func seedAceitesGrasas(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('g001','Aceite de coco extra virgen','Aceites',862,0.0,0.0,100.0,0.0,'1 cucharada 14g'),
	('g002','Aceite de aguacate','Aceites',884,0.0,0.0,100.0,0.0,'1 cucharada 14g'),
	('g003','Aceite de ajonjoli','Aceites',884,0.0,0.0,100.0,0.0,'1 cucharada 14g'),
	('g004','Aceite de girasol','Aceites',884,0.0,0.0,100.0,0.0,'1 cucharada 14g'),
	('g005','Mantequilla sin sal','Aceites',717,0.9,0.1,81.1,0.0,'1 cucharada 14g'),
	('g006','Mantequilla de mani natural','Aceites',588,25.1,20.0,50.4,6.0,'2 cucharadas 32g'),
	('g007','Margarina vegetal','Aceites',718,0.2,0.7,80.4,0.0,'1 cucharada 14g'),
	('g008','Manteca de cerdo','Aceites',898,0.0,0.0,99.5,0.0,'1 cucharada 13g'),
	('g009','Mayonesa regular','Aceites',680,1.0,0.6,74.8,0.0,'1 cucharada 15g'),
	('g010','Mayonesa light','Aceites',340,0.5,6.0,34.0,0.0,'1 cucharada 15g'),
	('g011','Almendras naturales','Aceites',579,21.2,21.7,49.9,12.5,'23 piezas 28g'),
	('g012','Nuez de Castilla','Aceites',654,15.2,13.7,65.2,6.7,'7 mitades 28g'),
	('g013','Semilla de girasol sin cascara','Aceites',584,20.8,20.0,51.5,8.6,'3 cucharadas 28g'),
	('g014','Pepita de calabaza sin cáscara','Aceites',559,30.2,10.7,49.1,6.0,'3 cucharadas 28g'),
	('g015','Ajonjoli (semillas de sesamo)','Aceites',573,17.7,23.4,49.7,11.8,'3 cucharadas 28g'),
	('g016','Nuez de la India (anacardo)','Aceites',553,18.2,30.2,43.9,3.3,'18 piezas 28g'),
	('g017','Piñon tostado','Aceites',673,13.7,13.1,68.4,3.7,'2 cucharadas 28g'),
	('g018','Nuez de macadamia','Aceites',718,7.9,13.8,75.8,8.6,'10-12 piezas 28g'),
	('g019','Aguacate Hass maduro','Aceites',160,2.0,8.5,14.7,6.7,'1/2 pieza 68g'),
	('g020','Aceite de linaza','Aceites',884,0.0,0.0,100.0,0.0,'1 cucharada 14g')`)
}

func seedFrutasExtra(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('f001','Tamarindo sin semilla','Frutas',239,2.8,62.5,0.6,5.1,'1 cucharada pulpa 15g'),
	('f002','Mamey rojo','Frutas',83,1.5,20.0,0.5,5.3,'1 rebanada 100g'),
	('f003','Zapote negro','Frutas',57,0.7,12.6,0.5,6.0,'1/2 pieza 100g'),
	('f004','Tejocote','Frutas',58,0.3,14.5,0.2,4.3,'3 piezas 80g'),
	('f005','Chico zapote (nispero)','Frutas',83,0.4,19.9,1.1,5.3,'1 pieza 100g'),
	('f006','Ciruela mexicana fresca','Frutas',46,0.7,11.4,0.3,1.4,'3 piezas 80g'),
	('f007','Mandarina','Frutas',53,0.8,13.3,0.3,1.8,'2 piezas 120g'),
	('f008','Pina fresca','Frutas',50,0.5,13.1,0.1,1.4,'1 taza 165g'),
	('f009','Kiwi','Frutas',61,1.1,14.7,0.5,3.0,'2 piezas 140g'),
	('f010','Durazno fresco','Frutas',39,0.9,9.5,0.3,1.5,'1 pieza 150g'),
	('f011','Chabacano (albaricoque)','Frutas',48,1.4,11.1,0.4,2.0,'3 piezas 120g'),
	('f012','Granada roja (granada china)','Frutas',83,1.7,18.7,1.2,4.0,'1/2 pieza 87g'),
	('f013','Guanabana fresca','Frutas',66,1.0,16.8,0.3,3.3,'1 rebanada 100g'),
	('f014','Nanche','Frutas',47,1.1,10.6,0.4,5.2,'1 taza 100g'),
	('f015','Uvas rojas','Frutas',69,0.7,18.1,0.2,0.9,'1 taza 150g'),
	('f016','Pera','Frutas',57,0.4,15.2,0.1,3.1,'1 pieza 178g'),
	('f017','Cereza','Frutas',50,1.0,12.2,0.3,1.6,'1 taza 154g'),
	('f018','Platano tabasco (macho maduro)','Frutas',90,1.3,22.1,0.4,2.4,'1 pieza 120g'),
	('f019','Nectarina','Frutas',44,1.1,10.6,0.3,1.7,'1 pieza 136g'),
	('f020','Toronja (pomelo)','Frutas',42,0.8,10.7,0.1,1.6,'1/2 pieza 123g'),
	('f021','Lima persa','Frutas',30,0.7,10.5,0.2,2.8,'2 piezas 67g'),
	('f022','Zapote blanco','Frutas',68,1.2,14.6,1.2,1.9,'1 pieza 100g'),
	('f023','Jicama pelada cruda','Frutas',38,0.7,8.8,0.1,4.9,'1 taza 130g'),
	('f024','Tuna roja','Frutas',41,0.8,9.9,0.5,3.6,'2 piezas 100g'),
	('f025','Coco fresco rallado','Frutas',354,3.3,15.2,33.5,9.0,'1/4 taza 20g'),
	('f026','Maracuya (fruta de la pasion)','Frutas',97,2.2,23.4,0.7,10.4,'2 piezas 36g'),
	('f027','Pitaya (dragon fruit)','Frutas',60,1.2,13.0,0.6,3.0,'1/2 pieza 100g'),
	('f028','Mango manila','Frutas',65,0.5,17.0,0.3,1.8,'1/2 pieza 100g'),
	('f029','Sandia sin semilla','Frutas',30,0.6,7.6,0.2,0.4,'1 taza 152g'),
	('f030','Melon verde (honeydew)','Frutas',36,0.5,9.1,0.1,0.8,'1 taza 170g')`)
}

func seedAzucares(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('az001','Azucar morena','Azucares',380,0.0,98.1,0.0,0.0,'1 cucharada 12g'),
	('az002','Miel de agave organica','Azucares',310,0.1,76.0,0.0,0.0,'1 cucharada 21g'),
	('az003','Mermelada de fresa regular','Azucares',250,0.4,65.0,0.1,1.0,'1 cucharada 20g'),
	('az004','Cajeta de leche','Azucares',382,9.1,64.8,10.4,0.0,'2 cucharadas 30g'),
	('az005','Ate de membrillo','Azucares',310,0.4,79.7,0.1,5.9,'1 rebanada 30g'),
	('az006','Chocolate oscuro 70pct cacao','Azucares',598,7.8,45.9,42.6,10.9,'20g'),
	('az007','Chocolate de mesa (Abuelita)','Azucares',393,4.6,73.9,12.7,5.5,'1 tablilla 22g'),
	('az008','Paleta de tamarindo','Azucares',80,0.2,20.0,0.1,0.3,'1 pieza 25g'),
	('az009','Miel de maple','Azucares',260,0.0,67.0,0.0,0.0,'1 cucharada 20g'),
	('az010','Stevia en polvo (endulzante)','Azucares',0,0.0,0.0,0.0,0.0,'1 sobre 1g')`)
}

func seedCondimentos(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO alimentos
		(id,nombre,categoria,calorias,proteinas,carbohidratos,grasas,fibra,porcion_desc) VALUES
	('cd001','Chile en polvo Tajin','Condimentos',200,7.0,37.0,4.0,9.0,'1 cucharadita 3g'),
	('cd002','Salsa valentina','Condimentos',25,0.9,4.5,0.6,0.8,'1 cucharada 15ml'),
	('cd003','Salsa verde Herdez embotellada','Condimentos',20,1.0,3.8,0.3,1.0,'2 cucharadas 30g'),
	('cd004','Salsa roja Herdez embotellada','Condimentos',23,1.1,4.2,0.5,1.3,'2 cucharadas 30g'),
	('cd005','Vinagre de manzana','Condimentos',22,0.0,0.9,0.0,0.0,'1 cucharada 15ml'),
	('cd006','Jugo de limon fresco','Condimentos',22,0.4,7.0,0.2,0.3,'2 cucharadas 30ml'),
	('cd007','Canela molida','Condimentos',247,4.0,80.6,1.2,53.1,'1 cucharadita 2.6g'),
	('cd008','Chile mulato seco','Condimentos',261,11.3,46.0,7.5,16.0,'1 pieza 20g'),
	('cd009','Chile pasilla negro seco','Condimentos',281,12.0,49.8,7.3,17.3,'1 pieza 20g'),
	('cd010','Chile morita ahumado','Condimentos',258,11.2,44.6,7.9,17.5,'2 piezas 15g'),
	('cd011','Chile de arbol seco','Condimentos',282,10.5,55.0,6.1,11.1,'3 chiles 3g'),
	('cd012','Oregano seco','Condimentos',265,9.0,68.9,4.3,42.5,'1 cucharadita 1.5g'),
	('cd013','Comino molido','Condimentos',375,17.8,44.2,22.3,10.5,'1 cucharadita 2.1g'),
	('cd014','Achiote en pasta','Condimentos',100,4.0,16.0,3.0,5.0,'1 cucharada 30g'),
	('cd015','Consomé de pollo en polvo','Condimentos',300,20.0,40.0,5.0,0.0,'1 cucharadita 3g'),
	('cd016','Salsa de soya','Condimentos',53,8.1,4.9,0.6,0.8,'1 cucharada 16ml'),
	('cd017','Mostaza amarilla','Condimentos',66,4.4,6.0,3.6,3.6,'1 cucharada 15g'),
	('cd018','Catsup (ketchup)','Condimentos',112,1.7,27.2,0.5,0.3,'1 cucharada 15g'),
	('cd019','Bicarbonato de sodio','Condimentos',0,0.0,0.0,0.0,0.0,'1 cucharadita 4.6g'),
	('cd020','Sal de mesa','Condimentos',0,0.0,0.0,0.0,0.0,'1 cucharadita 6g')`)
}
