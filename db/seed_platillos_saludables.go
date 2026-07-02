package db

import "database/sql"

// SeedPlatillosSaludables adds 100+ healthy Mexican platillos. Safe to call on existing
// databases — uses INSERT OR IGNORE so existing records are never overwritten.
func SeedPlatillosSaludables(db *sql.DB) {
	seedDesayunosSaludables(db)
	seedComidasSaludables(db)
	seedCenasSaludables(db)
	seedColacionesSaludables(db)
	seedBebSaludables(db)
}

// imagen genérica por categoría (Unsplash)
const (
	imgDesayuno = "https://images.unsplash.com/photo-1529062500409-b06d81bf74ec?w=400"
	imgSopa     = "https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400"
	imgPollo    = "https://images.unsplash.com/photo-1529042410759-befb1204b468?w=400"
	imgEnsalada = "https://images.unsplash.com/photo-1580822184713-fc5400e7fe10?w=400"
	imgPescado  = "https://images.unsplash.com/photo-1559847844-d721426d0a3a?w=400"
	imgTacos    = "https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=400"
	imgFrutas   = "https://images.unsplash.com/photo-1517673400267-0251440c45dc?w=400"
	imgLicuado  = "https://images.unsplash.com/photo-1553530666-ba11a90a0868?w=400"
	imgNopal    = "https://images.unsplash.com/photo-1580822184713-fc5400e7fe10?w=400"
	imgAvena    = "https://images.unsplash.com/photo-1517673400267-0251440c45dc?w=400"
)

func seedDesayunosSaludables(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO platillos
		(id,nombre,categoria,region,porcion_desc,porcion_g,calorias,proteinas,carbohidratos,grasas,fibra,sodio_mg,imagen_url,
		 apto_diabetes,apto_hipertension,apto_sobrepeso,alto_proteina,bajo_grasa,vegetariano) VALUES

	('ds001','Claras de huevo con espinacas y jitomate','Desayuno','Nacional','1 plato',200,160,22.0,6.0,4.0,2.0,320,'`+imgDesayuno+`',1,1,1,1,1,1),
	('ds002','Omelette de claras con nopales y chile poblano','Desayuno','Nacional','1 plato',180,145,18.0,7.0,4.0,3.0,280,'`+imgDesayuno+`',1,1,1,1,1,1),
	('ds003','Avena con chia, manzana y canela','Desayuno','Nacional','1 taza',300,280,9.0,48.0,5.0,8.0,55,'`+imgAvena+`',0,1,0,0,1,1),
	('ds004','Licuado verde de espinacas con proteina','Desayuno','Nacional','1 vaso grande',350,220,24.0,22.0,4.0,5.0,180,'`+imgLicuado+`',1,1,1,1,1,1),
	('ds005','Tostadas de frijol negro con aguacate y pico de gallo','Desayuno','Nacional','2 piezas',160,240,8.0,28.0,10.0,7.0,220,'`+imgDesayuno+`',1,1,1,0,0,1),
	('ds006','Parfait de yogurt griego con fruta y linaza','Desayuno','Nacional','1 vasito',200,190,14.0,24.0,3.0,4.0,70,'`+imgFrutas+`',1,1,1,0,1,1),
	('ds007','Nopalitos con huevo a la mexicana','Desayuno','Nacional','1 plato',220,185,13.0,9.0,10.0,4.0,310,'`+imgNopal+`',1,1,1,0,0,1),
	('ds008','Chilaquiles verdes con tortilla horneada sin crema','Desayuno','Centro','1 plato',250,260,20.0,28.0,7.0,4.0,370,'`+imgDesayuno+`',1,1,1,1,1,0),
	('ds009','Pan integral con aguacate y huevo pochado','Desayuno','Nacional','2 rebanadas',150,280,14.0,26.0,14.0,7.0,240,'`+imgDesayuno+`',1,1,0,0,0,1),
	('ds010','Molletes ligeros con frijoles y pico de gallo','Desayuno','Nacional','2 piezas',200,280,11.0,44.0,5.0,6.0,330,'`+imgDesayuno+`',0,1,1,0,1,1),
	('ds011','Avena proteica con platano y mantequilla de mani','Desayuno','Nacional','1 taza',280,340,18.0,44.0,8.0,6.0,120,'`+imgAvena+`',0,1,0,1,0,1),
	('ds012','Tortilla de claras con quelites y salsa verde','Desayuno','Nacional','1 plato',200,140,20.0,5.0,4.0,2.0,260,'`+imgDesayuno+`',1,1,1,1,1,1),
	('ds013','Tacos de nopales con huevo y frijoles','Desayuno','Nacional','3 tacos',240,230,12.0,30.0,6.0,5.5,300,'`+imgTacos+`',1,1,1,0,1,1),
	('ds014','Enfrijoladas de pollo sin crema','Desayuno','Centro','3 piezas',280,300,22.0,38.0,6.0,6.0,370,'`+imgDesayuno+`',1,1,1,1,1,0),
	('ds015','Huevo escalfado sobre nopales asados','Desayuno','Nacional','1 plato',200,160,13.0,6.0,9.0,3.0,270,'`+imgDesayuno+`',1,1,1,0,0,1),
	('ds016','Licuado de nopal con piña y chia','Desayuno','Nacional','1 vaso',350,120,4.0,24.0,2.0,5.0,40,'`+imgLicuado+`',0,1,1,0,1,1),
	('ds017','Hotcakes de avena y platano sin harina blanca','Desayuno','Nacional','3 piezas',200,290,12.0,44.0,6.0,5.0,160,'`+imgDesayuno+`',0,1,0,0,1,1),
	('ds018','Tamales de acelgas y queso panela sin manteca','Desayuno','Centro','2 piezas',180,240,10.0,38.0,5.0,5.0,275,'`+imgDesayuno+`',1,1,1,0,1,1),
	('ds019','Enchiladas verdes de espinacas sin crema','Desayuno','Centro','3 piezas',260,280,16.0,36.0,7.0,5.0,370,'`+imgDesayuno+`',1,1,1,0,1,1),
	('ds020','Soya texturizada a la mexicana con verduras','Desayuno','Nacional','1 plato',200,200,18.0,16.0,5.0,4.0,280,'`+imgDesayuno+`',1,1,1,1,1,1),
	('ds021','Caldo de pollo matutino con chayote','Desayuno','Nacional','1 taza',350,150,16.0,10.0,4.0,3.0,330,'`+imgSopa+`',1,1,1,0,1,0),
	('ds022','Huevos revueltos con champiñones y epazote','Desayuno','Nacional','1 plato',200,190,15.0,5.0,12.0,1.5,280,'`+imgDesayuno+`',1,1,1,0,0,1),
	('ds023','Licuado de amaranto con platano y leche','Desayuno','Nacional','1 vaso',300,280,12.0,44.0,5.0,4.0,120,'`+imgLicuado+`',0,1,0,0,1,1),
	('ds024','Quesadillas de frijol y acelgas con tortilla de maiz','Desayuno','Nacional','2 piezas',160,250,12.0,32.0,7.0,6.0,300,'`+imgDesayuno+`',1,1,1,0,1,1),
	('ds025','Bowl de frutas con yogurt y miel de agave','Desayuno','Nacional','1 tazón',250,180,8.0,32.0,2.5,3.5,65,'`+imgFrutas+`',0,1,1,0,1,1),
	('ds026','Huevo a la mexicana con nopales y frijoles','Desayuno','Nacional','1 plato',260,280,16.0,24.0,12.0,7.0,360,'`+imgDesayuno+`',1,1,1,0,0,1),
	('ds027','Avena con semillas de chia y fresas','Desayuno','Nacional','1 taza',280,260,9.0,44.0,5.0,9.0,50,'`+imgAvena+`',0,1,0,0,1,1),
	('ds028','Proteina de chayote con claras y epazote','Desayuno','Nacional','1 plato',220,150,18.0,8.0,4.0,2.5,240,'`+imgDesayuno+`',1,1,1,1,1,1),
	('ds029','Smoothie de espinaca mango y limon','Desayuno','Nacional','1 vaso grande',300,160,5.0,32.0,2.0,4.0,60,'`+imgLicuado+`',0,1,1,0,1,1),
	('ds030','Tostadas de atun con aguacate y jitomate','Desayuno','Nacional','2 piezas',140,220,18.0,18.0,8.0,3.5,380,'`+imgDesayuno+`',1,1,1,1,0,0)`)
}

func seedComidasSaludables(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO platillos
		(id,nombre,categoria,region,porcion_desc,porcion_g,calorias,proteinas,carbohidratos,grasas,fibra,sodio_mg,imagen_url,
		 apto_diabetes,apto_hipertension,apto_sobrepeso,alto_proteina,bajo_grasa,vegetariano) VALUES

	('cs001','Pollo en salsa de tomatillo asado con calabacitas','Comida','Nacional','1 plato',350,275,32.0,12.0,8.0,4.0,310,'`+imgPollo+`',1,1,1,1,1,0),
	('cs002','Sopa de lentejas con espinacas y jitomate','Comida','Nacional','1 plato',400,220,14.0,34.0,3.0,9.0,270,'`+imgSopa+`',1,1,1,0,1,1),
	('cs003','Tilapia al vapor con ejotes y limon','Comida','Nacional','1 plato',300,190,26.0,8.0,5.0,3.0,175,'`+imgPescado+`',1,1,1,1,1,0),
	('cs004','Milanesa de pechuga al horno con ensalada de nopales','Comida','Nacional','1 plato',300,240,30.0,14.0,5.0,3.5,355,'`+imgPollo+`',1,1,1,1,1,0),
	('cs005','Salmon al horno con verduras rostizadas','Comida','Nacional','1 plato',300,300,30.0,10.0,15.0,4.0,215,'`+imgPescado+`',1,1,1,1,0,0),
	('cs006','Ceviche de atun con aguacate y jitomate','Comida','Costa','1 plato',200,220,24.0,8.0,10.0,3.0,355,'`+imgPescado+`',1,1,1,1,0,0),
	('cs007','Calabacitas rellenas de pavo molido y verduras','Comida','Nacional','2 piezas',300,240,24.0,14.0,8.0,4.0,335,'`+imgPollo+`',1,1,1,1,1,0),
	('cs008','Ensalada de pollo con vinagreta de limon y aguacate','Comida','Nacional','1 plato',300,275,28.0,10.0,14.0,5.0,235,'`+imgEnsalada+`',1,1,1,1,0,0),
	('cs009','Arroz integral con pollo y verduras de temporada','Comida','Nacional','1 plato',350,320,28.0,36.0,6.0,4.0,255,'`+imgPollo+`',1,1,0,1,1,0),
	('cs010','Sopa de quinoa con verduras y cilantro','Comida','Nacional','1 plato',400,200,8.0,34.0,4.0,5.0,235,'`+imgSopa+`',1,1,1,0,1,1),
	('cs011','Pechuga rellena de espinacas y queso panela','Comida','Nacional','1 pieza',250,280,36.0,5.0,12.0,2.0,315,'`+imgPollo+`',1,1,1,1,0,0),
	('cs012','Ensalada de betabel con queso de cabra y nuez','Comida','Nacional','1 plato',200,180,6.0,20.0,9.0,4.0,155,'`+imgEnsalada+`',1,1,1,0,0,1),
	('cs013','Pollo rostizado con hierbas y limon sin piel','Comida','Nacional','1 pieza',280,255,30.0,2.0,14.0,0.5,275,'`+imgPollo+`',1,1,1,1,0,0),
	('cs014','Nopales con camarones en salsa verde','Comida','Nacional','1 plato',300,200,24.0,12.0,5.0,5.0,375,'`+imgNopal+`',1,1,1,1,1,0),
	('cs015','Ensalada de quinoa con garbanzo y pepino','Comida','Nacional','1 plato',300,260,10.0,38.0,7.0,7.0,175,'`+imgEnsalada+`',1,1,1,0,1,1),
	('cs016','Caldo tlalpeno bajo en sodio','Comida','Centro','1 plato',450,240,20.0,24.0,6.0,7.0,375,'`+imgSopa+`',1,1,1,1,1,0),
	('cs017','Bacalao en salsa verde con verduras','Comida','Nacional','1 plato',300,215,28.0,8.0,7.0,2.0,415,'`+imgPescado+`',1,1,1,1,1,0),
	('cs018','Tacos de coliflor rostizada con frijoles negros','Comida','Nacional','3 tacos',240,240,9.0,40.0,5.0,8.0,255,'`+imgTacos+`',1,1,1,0,1,1),
	('cs019','Pollo en pipian verde ligero','Comida','Centro','1 plato',350,315,32.0,16.0,12.0,4.0,375,'`+imgPollo+`',1,1,0,1,0,0),
	('cs020','Camarones en caldo de jitomate y epazote','Comida','Costa','1 plato',350,180,22.0,10.0,4.0,2.0,415,'`+imgPescado+`',1,1,1,1,1,0),
	('cs021','Ensalada de nopales con zanahoria y pepino','Comida','Nacional','1 plato',250,80,3.0,14.0,2.0,5.0,115,'`+imgEnsalada+`',1,1,1,0,1,1),
	('cs022','Sopa de chicharos con hierbabuena','Comida','Nacional','1 plato',350,200,9.0,34.0,3.0,7.0,255,'`+imgSopa+`',1,1,1,0,1,1),
	('cs023','Filete de res magro con esparragos a la plancha','Comida','Nacional','1 plato',280,260,28.0,6.0,12.0,2.0,255,'`+imgPollo+`',1,1,1,1,0,0),
	('cs024','Arroz verde con pollo y cilantro','Comida','Nacional','1 plato',350,300,26.0,30.0,7.0,3.0,275,'`+imgPollo+`',1,1,0,1,1,0),
	('cs025','Cocido de res magro con verduras y garbanzos','Comida','Nacional','1 plato',500,280,26.0,20.0,9.0,5.0,375,'`+imgSopa+`',1,1,1,1,0,0),
	('cs026','Flautas de pollo al horno sin freir','Comida','Centro','3 piezas',240,300,22.0,36.0,6.0,3.0,375,'`+imgTacos+`',1,1,1,1,1,0),
	('cs027','Tinga de pechuga de pavo','Comida','Centro','1 taza',200,240,28.0,10.0,8.0,2.0,375,'`+imgPollo+`',1,1,1,1,1,0),
	('cs028','Chiles anchos rellenos de quinoa y verduras','Comida','Nacional','2 piezas',250,260,10.0,44.0,6.0,8.0,275,'`+imgEnsalada+`',1,1,1,0,1,1),
	('cs029','Caldo de camaron ligero con verduras','Comida','Costa','1 plato',450,195,20.0,16.0,5.0,3.0,495,'`+imgSopa+`',1,0,1,1,1,0),
	('cs030','Ensalada de espinacas con camaron y vinagreta de limon','Comida','Nacional','1 plato',300,200,24.0,8.0,8.0,3.0,375,'`+imgEnsalada+`',1,1,1,1,1,0),
	('cs031','Pollo a la naranja sin azucar con brócoli','Comida','Nacional','1 plato',300,255,30.0,12.0,9.0,3.5,275,'`+imgPollo+`',1,1,1,1,0,0),
	('cs032','Sopa de verduras con pollo deshebrado','Comida','Nacional','1 plato',450,220,20.0,20.0,5.0,5.0,295,'`+imgSopa+`',1,1,1,1,1,0),
	('cs033','Frijoles negros con verduras y epazote','Comida','Nacional','1.5 tazas',300,210,12.0,36.0,3.0,12.0,195,'`+imgSopa+`',1,1,1,0,1,1),
	('cs034','Pechugas en salsa de jitomate casera','Comida','Nacional','1 plato',300,255,30.0,12.0,8.0,2.5,315,'`+imgPollo+`',1,1,1,1,1,0),
	('cs035','Atun en salsa de habanero tatemado','Comida','Sureste','1 plato',200,180,24.0,6.0,5.0,1.0,395,'`+imgPescado+`',1,1,1,1,1,0),
	('cs036','Mole de olla sin grasa con verduras','Comida','Centro','1 plato',450,220,20.0,22.0,6.0,5.0,355,'`+imgSopa+`',1,1,1,1,1,0),
	('cs037','Enchiladas de espinacas y pollo sin crema','Comida','Centro','3 piezas',280,290,22.0,36.0,7.0,5.0,375,'`+imgTacos+`',1,1,1,1,1,0),
	('cs038','Estofado de pollo con jitomate y aceitunas','Comida','Nacional','1 plato',350,280,28.0,14.0,12.0,3.0,375,'`+imgPollo+`',1,0,1,1,0,0),
	('cs039','Tacos de res magra con aguacate y salsa verde','Comida','Nacional','3 tacos',240,320,24.0,32.0,10.0,3.5,375,'`+imgTacos+`',0,1,0,1,0,0),
	('cs040','Sopa de champiñones al ajillo con caldo de verduras','Comida','Nacional','1 plato',350,120,5.0,12.0,5.0,3.0,255,'`+imgSopa+`',1,1,1,0,1,1),
	('cs041','Tostadas de camaron con aguacate y jitomate','Comida','Costa','2 piezas',160,260,18.0,22.0,10.0,3.5,375,'`+imgPescado+`',1,1,1,1,0,0),
	('cs042','Ensalada de pollo con naranja y zanahoria rallada','Comida','Nacional','1 plato',280,260,26.0,18.0,9.0,3.0,215,'`+imgEnsalada+`',1,1,1,1,0,0),
	('cs043','Sopa azteca con tortilla horneada sin crema','Comida','Centro','1 plato',300,220,10.0,28.0,7.0,4.5,355,'`+imgSopa+`',1,1,1,0,1,1),
	('cs044','Pollo en salsa de chile chipotle ligera','Comida','Nacional','1 plato',300,260,30.0,10.0,10.0,2.0,375,'`+imgPollo+`',1,0,1,1,0,0),
	('cs045','Lentejas con verduras y chorizo de pavo','Comida','Nacional','1 plato',400,240,16.0,34.0,5.0,9.0,335,'`+imgSopa+`',1,1,1,0,1,0),
	('cs046','Robalo a la veracruzana sin aceite','Comida','Costa','1 plato',300,220,28.0,10.0,7.0,2.5,415,'`+imgPescado+`',1,1,1,1,1,0),
	('cs047','Pollo con verduras al vapor estilo oriental','Comida','Nacional','1 plato',300,250,28.0,12.0,8.0,4.0,295,'`+imgPollo+`',1,1,1,1,1,0),
	('cs048','Tortitas de atun con nopales al horno','Comida','Nacional','3 piezas',250,240,22.0,14.0,8.0,3.5,375,'`+imgPescado+`',1,1,1,1,1,0),
	('cs049','Ensalada de garbanzos tostados con verduras','Comida','Nacional','1 plato',280,280,10.0,38.0,9.0,8.0,175,'`+imgEnsalada+`',1,1,1,0,0,1),
	('cs050','Pozole de garbanzo y verduras sin maiz extra','Comida','Centro','1 plato',400,200,8.0,32.0,4.0,9.0,295,'`+imgSopa+`',1,1,1,0,1,1)`)
}

func seedCenasSaludables(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO platillos
		(id,nombre,categoria,region,porcion_desc,porcion_g,calorias,proteinas,carbohidratos,grasas,fibra,sodio_mg,imagen_url,
		 apto_diabetes,apto_hipertension,apto_sobrepeso,alto_proteina,bajo_grasa,vegetariano) VALUES

	('cesh001','Caldo de pollo con chayote y calabacita','Cena','Nacional','1 plato',400,160,16.0,12.0,4.0,3.0,275,'`+imgSopa+`',1,1,1,0,1,0),
	('cesh002','Claras de huevo con jitomate y nopales','Cena','Nacional','1 plato',200,130,20.0,6.0,2.0,3.0,255,'`+imgDesayuno+`',1,1,1,1,1,1),
	('cesh003','Pechuga a la plancha con ensalada verde','Cena','Nacional','1 plato',280,220,30.0,7.0,7.0,3.0,195,'`+imgPollo+`',1,1,1,1,1,0),
	('cesh004','Sopa de verduras con garbanzo','Cena','Nacional','1 plato',400,180,7.0,30.0,3.0,7.0,255,'`+imgSopa+`',1,1,1,0,1,1),
	('cesh005','Quesadillas de nopales con queso panela','Cena','Nacional','2 piezas',160,240,12.0,32.0,7.0,4.0,315,'`+imgDesayuno+`',1,1,1,0,1,1),
	('cesh006','Tilapia a la plancha con jitomate asado','Cena','Nacional','1 plato',280,180,24.0,6.0,5.0,2.0,195,'`+imgPescado+`',1,1,1,1,1,0),
	('cesh007','Atun con ensalada de pepino y jitomate','Cena','Nacional','1 plato',250,175,22.0,7.0,5.0,2.0,335,'`+imgEnsalada+`',1,1,1,1,1,0),
	('cesh008','Nopales con huevo revuelto y salsa roja casera','Cena','Nacional','1 plato',220,180,13.0,8.0,9.0,4.0,275,'`+imgNopal+`',1,1,1,0,0,1),
	('cesh009','Tacos de verduras rostizadas con guacamole','Cena','Nacional','3 tacos',240,280,6.0,38.0,12.0,8.0,215,'`+imgTacos+`',1,1,1,0,0,1),
	('cesh010','Pavo molido con calabacitas y jitomate','Cena','Nacional','1 plato',280,220,26.0,10.0,7.0,3.0,295,'`+imgPollo+`',1,1,1,1,1,0),
	('cesh011','Ensalada cesar ligera con pollo sin crotones','Cena','Nacional','1 plato',280,235,26.0,8.0,12.0,3.0,395,'`+imgEnsalada+`',1,1,1,1,0,0),
	('cesh012','Frijoles negros con nopales y epazote','Cena','Nacional','1.5 tazas',280,180,11.0,30.0,2.0,11.0,155,'`+imgNopal+`',1,1,1,0,1,1),
	('cesh013','Salmon al limon con esparragos y brocoli','Cena','Nacional','1 plato',250,290,26.0,6.0,17.0,3.0,175,'`+imgPescado+`',1,1,1,1,0,0),
	('cesh014','Ensalada de lentejas con jitomate y pepino','Cena','Nacional','1 plato',250,200,12.0,32.0,3.0,8.0,195,'`+imgEnsalada+`',1,1,1,0,1,1),
	('cesh015','Pollo en caldo con hierbas y verduras de hoja','Cena','Nacional','1 plato',400,200,22.0,14.0,5.0,4.0,295,'`+imgSopa+`',1,1,1,1,1,0),
	('cesh016','Tortillas de maiz con frijoles y chile de agua','Cena','Sur','3 tortillas',200,260,10.0,48.0,3.0,7.0,215,'`+imgTacos+`',1,1,0,0,1,1),
	('cesh017','Revoltijo de romeritos con camaron sin manteca','Cena','Centro','1 plato',300,200,18.0,18.0,6.0,5.0,315,'`+imgSopa+`',1,1,1,0,1,0),
	('cesh018','Sopa de calabaza con caldo de verduras','Cena','Nacional','1 plato',350,110,3.0,18.0,3.0,4.0,175,'`+imgSopa+`',1,1,1,0,1,1),
	('cesh019','Tacos de picadillo de res magro','Cena','Nacional','3 tacos',230,280,22.0,26.0,9.0,3.5,355,'`+imgTacos+`',1,1,1,1,0,0),
	('cesh020','Ceviche de camaron con jitomate y limon','Cena','Costa','1 taza',200,160,18.0,10.0,3.5,2.5,355,'`+imgPescado+`',1,1,1,1,1,0),
	('cesh021','Ensalada de zanahoria con atun y limon','Cena','Nacional','1 plato',200,180,20.0,12.0,5.0,3.0,375,'`+imgEnsalada+`',1,1,1,1,1,0),
	('cesh022','Calabacitas con pollo a la mexicana','Cena','Nacional','1 plato',280,220,22.0,12.0,8.0,3.5,295,'`+imgPollo+`',1,1,1,1,1,0),
	('cesh023','Caldo de verduras con huevo pochado','Cena','Nacional','1 plato',350,140,10.0,14.0,6.0,3.0,215,'`+imgSopa+`',1,1,1,0,1,1),
	('cesh024','Ensalada de quinoa con verduras y limon','Cena','Nacional','1 plato',250,220,8.0,34.0,5.0,5.0,155,'`+imgEnsalada+`',1,1,1,0,1,1),
	('cesh025','Sopa de pollo con arroz integral y verduras','Cena','Nacional','1 plato',400,230,20.0,24.0,5.0,3.5,275,'`+imgSopa+`',1,1,1,1,1,0)`)
}

func seedColacionesSaludables(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO platillos
		(id,nombre,categoria,region,porcion_desc,porcion_g,calorias,proteinas,carbohidratos,grasas,fibra,sodio_mg,imagen_url,
		 apto_diabetes,apto_hipertension,apto_sobrepeso,alto_proteina,bajo_grasa,vegetariano) VALUES

	('colh001','Pepino en rebanadas con limon y chile piquin','Colacion','Nacional','1 pepino',150,28,1.0,6.0,0.2,1.5,20,'`+imgFrutas+`',1,1,1,0,1,1),
	('colh002','Jicama en bastones con limon y tajin','Colacion','Nacional','1 taza',120,58,1.0,14.0,0.2,5.0,40,'`+imgFrutas+`',1,1,1,0,1,1),
	('colh003','Yogurt griego con nueces y canela','Colacion','Nacional','1 vasito',150,165,10.0,8.0,9.0,1.0,60,'`+imgFrutas+`',1,1,1,0,0,1),
	('colh004','Guacamole ligero con pepino y zanahoria','Colacion','Nacional','1 taza',150,130,2.0,10.0,9.0,5.0,60,'`+imgNopal+`',1,1,1,0,0,1),
	('colh005','Elote cocido al natural sin mantequilla','Colacion','Nacional','1 pieza',100,96,3.4,21.0,1.4,2.4,20,'`+imgFrutas+`',0,1,1,0,1,1),
	('colh006','Almendras naturales tostadas sin sal','Colacion','Nacional','23 piezas',28,162,6.0,6.0,14.0,3.5,0,'`+imgFrutas+`',1,1,0,0,0,1),
	('colh007','Licuado verde de espinacas pepino y limon','Colacion','Nacional','1 vaso',300,70,4.0,12.0,1.0,4.0,55,'`+imgLicuado+`',1,1,1,0,1,1),
	('colh008','Manzana con canela al natural','Colacion','Nacional','1 pieza',150,78,0.4,21.0,0.2,3.6,2,'`+imgFrutas+`',0,1,1,0,1,1),
	('colh009','Queso panela con tomate cherry y oregano','Colacion','Nacional','50g',80,130,9.0,2.0,9.0,0.5,240,'`+imgFrutas+`',1,0,1,0,0,1),
	('colh010','Edamame cocido con sal de limon','Colacion','Nacional','1/2 taza',80,98,9.5,7.0,5.2,4.2,40,'`+imgFrutas+`',1,1,1,0,0,1),
	('colh011','Nopalitos curtidos con limon','Colacion','Nacional','1 taza',100,22,1.8,4.4,0.1,3.3,20,'`+imgNopal+`',1,1,1,0,1,1),
	('colh012','Pepita de calabaza tostada sin sal','Colacion','Nacional','2 cucharadas',28,155,8.5,3.0,13.5,1.7,0,'`+imgFrutas+`',1,1,0,0,0,1),
	('colh013','Totopos de maiz horneados con salsa verde','Colacion','Nacional','10 piezas',28,120,2.0,20.0,3.5,1.5,80,'`+imgTacos+`',0,1,0,0,1,1),
	('colh014','Fruta fresca de temporada con limon','Colacion','Nacional','1 taza',150,65,1.0,16.0,0.3,2.5,5,'`+imgFrutas+`',0,1,1,0,1,1),
	('colh015','Tzatziki de jocoque con pepino y ajo','Colacion','Nacional','3 cucharadas',60,55,3.5,4.0,2.5,0.5,80,'`+imgFrutas+`',1,1,1,0,1,1),
	('colh016','Zanahoria rallada con limon y chile','Colacion','Nacional','1 taza',80,36,0.9,8.5,0.2,2.5,20,'`+imgFrutas+`',1,1,1,0,1,1),
	('colh017','Cacahuates naturales tostados sin sal','Colacion','Nacional','30 piezas',30,170,7.7,4.8,14.8,2.5,0,'`+imgFrutas+`',1,1,0,0,0,1),
	('colh018','Tostadas ligeras con frijoles y aguacate','Colacion','Nacional','2 piezas',60,185,6.0,22.0,8.5,5.5,200,'`+imgTacos+`',1,1,1,0,0,1),
	('colh019','Smoothie de frutas sin azucar','Colacion','Nacional','1 vaso',250,120,2.5,28.0,0.5,3.0,20,'`+imgLicuado+`',0,1,1,0,1,1),
	('colh020','Bowl de nopales con naranja y chile','Colacion','Nacional','1 taza',150,55,2.0,10.0,1.0,4.0,30,'`+imgNopal+`',1,1,1,0,1,1)`)
}

func seedBebSaludables(db *sql.DB) {
	db.Exec(`INSERT OR IGNORE INTO platillos
		(id,nombre,categoria,region,porcion_desc,porcion_g,calorias,proteinas,carbohidratos,grasas,fibra,sodio_mg,imagen_url,
		 apto_diabetes,apto_hipertension,apto_sobrepeso,alto_proteina,bajo_grasa,vegetariano) VALUES

	('bh001','Agua de jamaica sin azucar con stevia','Bebida','Nacional','1 vaso',300,8,0.2,1.5,0.0,1.0,8,'`+imgLicuado+`',1,1,1,0,1,1),
	('bh002','Agua de pepino con limon y menta','Bebida','Nacional','1 vaso',300,15,0.4,3.5,0.1,0.5,10,'`+imgLicuado+`',1,1,1,0,1,1),
	('bh003','Agua de tamarindo sin azucar','Bebida','Nacional','1 vaso',300,18,0.5,4.0,0.1,0.5,12,'`+imgLicuado+`',1,1,1,0,1,1),
	('bh004','Te verde sin azucar','Bebida','Nacional','1 taza',240,2,0.0,0.5,0.0,0.0,0,'`+imgLicuado+`',1,1,1,0,1,1),
	('bh005','Agua de chia con limon','Bebida','Nacional','1 vaso',300,40,1.5,6.0,1.5,3.0,15,'`+imgLicuado+`',1,1,1,0,1,1),
	('bh006','Licuado de nopal con piña sin azucar','Bebida','Nacional','1 vaso',300,80,2.5,16.0,0.5,3.5,35,'`+imgLicuado+`',0,1,1,0,1,1),
	('bh007','Agua fresca de sandia sin azucar','Bebida','Nacional','1 vaso',300,28,0.5,7.0,0.1,0.4,5,'`+imgLicuado+`',0,1,1,0,1,1),
	('bh008','Atole de maiz azul sin azucar','Bebida','Nacional','1 taza',240,100,3.0,20.0,1.0,2.0,50,'`+imgLicuado+`',0,1,0,0,1,1),
	('bh009','Te de canela y jengibre sin azucar','Bebida','Nacional','1 taza',240,8,0.1,2.0,0.0,0.3,2,'`+imgLicuado+`',1,1,1,0,1,1),
	('bh010','Agua de horchata light con leche descremada','Bebida','Nacional','1 vaso',300,80,3.5,14.0,0.5,0.5,50,'`+imgLicuado+`',0,1,0,0,1,1)`)
}
