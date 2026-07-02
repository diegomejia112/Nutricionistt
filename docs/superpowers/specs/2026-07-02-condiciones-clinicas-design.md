# Condiciones clínicas múltiples (comorbilidades) — Spec

## Contexto

El sistema actual filtra platillos por una sola "condición clínica" (`caso`) a la vez,
usando columnas booleanas fijas en la tabla `platillos`: `apto_diabetes`,
`apto_hipertension`, `apto_sobrepeso`. El filtro está hardcodeado en un `switch` de Go
(`casoFilter`/`casoColumn` en `handlers/platillos.go`) y en un array fijo en el frontend
(`CASOS` en `Alimentos.tsx`).

Se necesita:
1. Ampliar el catálogo de condiciones de 3 a 16 (agregar SOP, Hipotiroidismo,
   Enfermedad renal, Dislipidemia, Gastritis/Reflujo, Colon irritable, Embarazo,
   Lactancia, Anemia, Hígado graso, Gota/Ácido úrico, Osteoporosis).
2. Permitir que un **paciente** tenga varias condiciones diagnosticadas a la vez
   (comorbilidades), no solo un platillo.
3. Que la búsqueda de platillos compatibles exija cumplir **todas** las condiciones
   seleccionadas/del paciente (AND), no basta con una sola.

El esquema de columnas booleanas no escala a este tamaño (cada condición nueva futura
requeriría una migración de schema + tocar el switch de Go + el array del frontend).
Se reemplaza por un catálogo normalizado.

## Base de datos

Tablas nuevas (SQLite, `db/db.go`):

```sql
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

CREATE TABLE IF NOT EXISTS paciente_casos (
    paciente_id TEXT NOT NULL REFERENCES pacientes(id),
    caso_id TEXT NOT NULL REFERENCES casos(id),
    created_at DATETIME NOT NULL,
    PRIMARY KEY (paciente_id, caso_id)
);
```

Las columnas booleanas existentes (`apto_diabetes`, `apto_hipertension`,
`apto_sobrepeso`, `alto_proteina`, `bajo_grasa`, `vegetariano`) **no se eliminan**
(evita un `DROP COLUMN` en SQLite, no soportado de forma confiable en la versión de
`modernc.org/sqlite` que usa el proyecto). `alto_proteina`, `bajo_grasa` y
`vegetariano` no son condiciones clínicas — siguen siendo columnas booleanas sin
cambios. `apto_diabetes/hipertension/sobrepeso` quedan sin usarse una vez migrados
sus datos a `platillo_casos`.

### Semilla del catálogo (16 casos)

Diabetes, Hipertensión, Sobrepeso/Obesidad (migrados de las columnas existentes) +
SOP, Hipotiroidismo, Enfermedad renal, Dislipidemia, Gastritis/Reflujo, Colon
irritable, Embarazo, Lactancia, Anemia, Hígado graso, Gota/Ácido úrico,
Osteoporosis.

### Re-etiquetado de platillos existentes

Los ~100 platillos ya sembrados se re-etiquetan con reglas heurísticas basadas en
sus datos nutricionales existentes (ej. bajo IG/azúcares simples → SOP e
Hipotiroidismo; bajo en grasas saturadas/azúcar → Dislipidemia; platillos con
proteína animal completa y sin exceso de fósforo/potasio conocido → cuidado con
Renal). **Esto es una aproximación, no asesoría médica** — se deja una nota visible
en el código y se le advierte al usuario que un nutriólogo debe validar las
etiquetas antes de usarlas con pacientes reales.

## Backend (Go)

- `GET /api/casos` — catálogo completo `[{id, slug, nombre}]`
- `GET /api/platillos?casos=diabetes,sop&...` — reemplaza el parámetro `caso`
  (singular) por `casos` (coma-separado). Un platillo aparece solo si tiene
  **todos** los casos pedidos en `platillo_casos` (AND vía `GROUP BY` + `HAVING
  COUNT(DISTINCT caso_id) = N`).
- `GET /api/pacientes/{id}/condiciones` — lista condiciones del paciente
- `POST /api/pacientes/{id}/condiciones` — agrega `{casoId}`
- `DELETE /api/pacientes/{id}/condiciones/{casoId}` — quita una condición
- `GET /api/platillos/compatibles/{pacienteId}` — se extiende: además de excluir
  platillos con ingredientes restringidos (comportamiento actual sin cambios),
  ahora también exige que el platillo cumpla **todas** las condiciones en
  `paciente_casos` del paciente (mismo AND que el endpoint de arriba).

Compatibilidad: el parámetro viejo `caso` (singular) se elimina — no hay
consumidores externos del API, solo el frontend propio, que se actualiza en el
mismo cambio.

## Frontend

- `lib/api.ts`: nuevos métodos `getCasos()`, `getCondicionesPaciente(pacienteId)`,
  `addCondicionPaciente(pacienteId, casoId)`,
  `deleteCondicionPaciente(pacienteId, casoId)`. `getPlatillos`/
  `getPlatillosCompatibles` cambian el parámetro `caso?: string` por
  `casos?: string[]`.
- `pages/Alimentos.tsx`: el `<select>` de un solo caso se reemplaza por un
  multi-select (checkboxes en un dropdown), cargado dinámicamente desde
  `GET /api/casos` en vez del array `CASOS` hardcodeado.
- `pages/ExpedientePaciente.tsx`: nueva pestaña "Condiciones" junto a
  "Restricciones", con checkboxes del catálogo de `casos` para marcar/quitar
  diagnósticos del paciente. Mismo patrón visual que la pestaña de
  restricciones existente.

## Fuera de alcance de este spec

Las siguientes ideas se recolectaron en la misma conversación pero son
subsistemas independientes — cada una se diseña y ejecuta por separado después
de esta:
- Manejo/cache de imágenes de platillos (pendiente de aclarar con el usuario)
- Exportar plan semanal a PDF o formato editable
- Sugerencia automática de alimentos que quepan en el espacio de
  calorías/macros restante de un tiempo de comida al armar un plan
