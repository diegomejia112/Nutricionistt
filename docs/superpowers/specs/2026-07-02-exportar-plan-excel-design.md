# Exportar plan a Excel (calendario semanal editable) — Spec

## Contexto

El plan semanal (`EditorPlan.tsx`) ya tiene un botón "Exportar" que solo abre
el diálogo de impresión del navegador (`window.print()`), útil para PDF pero
no editable. Se necesita una exportación real a un archivo que el
nutriólogo pueda abrir y modificar después (Excel/LibreOffice/Google
Sheets), con el mismo layout de calendario semanal que ya ve en pantalla.

## Alcance

- Nuevo botón "Exportar Excel" en `EditorPlan.tsx`, junto al botón de
  imprimir existente (no lo reemplaza).
- Genera un `.xlsx` real en el navegador con la librería `exceljs`
  (`npm install exceljs`), usando los datos ya cargados en el estado `plan`
  (`plan.dias[].tiempos[].alimentos[]`) — sin llamadas nuevas al backend.
- Descarga inmediata vía blob (`URL.createObjectURL` + `<a download>`),
  nombre de archivo `plan-{nombre-del-plan-slugificado}.xlsx`.

## Layout del archivo

Una sola hoja ("Plan semanal"):

- Fila 1: encabezado — nombre del plan y objetivo calórico
  (`plan.caloriasObj`). El nombre del paciente NO se incluye: el endpoint
  `GET /api/planes/{id}` solo devuelve `pacienteId`, no el nombre, y no se
  agrega una llamada nueva solo para esto (YAGNI).
- Fila 2: encabezados de columna — columna A vacía (para los nombres de
  tiempo de comida), columnas B–H = "Lunes".."Domingo" (mismo orden que
  `plan.dias[].diaSemana` 1–7, usando el array `DIAS` que ya existe en
  `EditorPlan.tsx`).
- Filas siguientes: una fila por cada tiempo de comida distinto presente en
  el plan, ordenadas por `tiempo.orden`. Columna A = nombre del tiempo
  (`tiempo.nombre`). Cada celda de día contiene, en líneas separadas dentro
  de la misma celda (salto de línea `\n` + `alignment.wrapText = true`):
  `{nombre} — {cantidad}{unidad} ({calorias} kcal)` por cada alimento de
  ese día+tiempo.
- Última fila: "Total kcal" con la suma de calorías de todos los alimentos
  de cada día, para comparar contra el objetivo calórico a simple vista.
- Estilo: encabezados de columna con fondo de color (tono rosa acorde a la
  paleta de la app, `#f43f5e`-ish) y texto en negrita/blanco; ancho de
  columna suficiente para leer el contenido; `wrapText` activado en las
  celdas de contenido para que el texto multilínea no se corte.

## Manejo de datos faltantes

- Si un tiempo de comida no tiene alimentos en un día dado, la celda queda
  vacía (no se agrega texto placeholder).
- Si el plan no tiene ningún día/tiempo cargado (plan recién creado sin
  contenido), el botón exporta igual un archivo con solo los encabezados —
  no se deshabilita el botón, evita un estado de carga especial innecesario.

## Fuera de alcance

- No se toca el backend Go — cero endpoints nuevos para este feature.
- No se reemplaza el botón de imprimir/PDF existente.
- No se resuelve aquí la exportación de imágenes de platillos ni la
  sugerencia de alimentos para espacio restante de calorías — son
  subsistemas independientes pendientes de diseño aparte.
