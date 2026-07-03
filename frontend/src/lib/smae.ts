// Sistema Mexicano de Alimentos Equivalentes (SMAE)
// Equivalents calculation for nutritional plan analysis

export type SmaeKey =
  | 'Cereales'
  | 'AOA'
  | 'Verduras'
  | 'Frutas'
  | 'Lacteos'
  | 'Leguminosas'
  | 'Aceites'
  | 'Azucares'

export type SmaeResult = Record<SmaeKey, number>

export const SMAE_ORDER: SmaeKey[] = [
  'Cereales', 'AOA', 'Verduras', 'Frutas', 'Lacteos', 'Leguminosas', 'Aceites', 'Azucares',
]

export const SMAE_META: Record<SmaeKey, {
  label: string
  kcalEq: number   // kcal per 1 equivalent
  proEq: number    // g protein per equiv
  choEq: number    // g carbs per equiv
  fatEq: number    // g fat per equiv
  color: string    // text color class
  bg: string       // background class
  bar: string      // progress bar class
}> = {
  Cereales:    { label: 'Cereales',    kcalEq: 70,  proEq: 2, choEq: 15, fatEq: 0, color: 'text-amber-700',  bg: 'bg-amber-100',  bar: 'bg-amber-400'  },
  AOA:         { label: 'AOA',         kcalEq: 55,  proEq: 7, choEq: 0,  fatEq: 3, color: 'text-rose-700',   bg: 'bg-rose-100',   bar: 'bg-rose-400'   },
  Verduras:    { label: 'Verduras',    kcalEq: 25,  proEq: 2, choEq: 4,  fatEq: 0, color: 'text-green-700',  bg: 'bg-green-100',  bar: 'bg-green-400'  },
  Frutas:      { label: 'Frutas',      kcalEq: 60,  proEq: 0, choEq: 15, fatEq: 0, color: 'text-red-600',    bg: 'bg-red-100',    bar: 'bg-red-400'    },
  Lacteos:     { label: 'Lácteos',     kcalEq: 110, proEq: 9, choEq: 12, fatEq: 4, color: 'text-blue-700',   bg: 'bg-blue-100',   bar: 'bg-blue-400'   },
  Leguminosas: { label: 'Leguminosas', kcalEq: 120, proEq: 9, choEq: 20, fatEq: 1, color: 'text-orange-700', bg: 'bg-orange-100', bar: 'bg-orange-400' },
  Aceites:     { label: 'Aceites',     kcalEq: 45,  proEq: 0, choEq: 0,  fatEq: 5, color: 'text-yellow-700', bg: 'bg-yellow-100', bar: 'bg-yellow-400' },
  Azucares:    { label: 'Azúcares',    kcalEq: 40,  proEq: 0, choEq: 10, fatEq: 0, color: 'text-purple-700', bg: 'bg-purple-100', bar: 'bg-purple-400' },
}

// Maps alimentos.categoria → SMAE group
export const CAT_MAP: Record<string, SmaeKey> = {
  Verduras:    'Verduras',
  Frutas:      'Frutas',
  Cereales:    'Cereales',
  Leguminosas: 'Leguminosas',
  Carnes:      'AOA',
  Lacteos:     'Lacteos',
  Aceites:     'Aceites',
  Azucares:    'Azucares',
  Condimentos: 'Verduras', // condimentos → negligible, treat as free/verduras
}

function emptyResult(): SmaeResult {
  return { Cereales: 0, AOA: 0, Verduras: 0, Frutas: 0, Lacteos: 0, Leguminosas: 0, Aceites: 0, Azucares: 0 }
}

// Returns effective macros for one alimentos_plan item
function effectiveMacros(ap: any) {
  if (ap.tipo === 'ia') {
    // IA foods: macros stored directly as totals
    return {
      kcal: ap.calorias ?? 0,
      pro:  ap.proteinas ?? 0,
      cho:  ap.carbohidratos ?? 0,
      fat:  ap.grasas ?? 0,
    }
  }
  // DB alimentos/platillos: values are per 100g, scale by cantidad
  const f = (ap.cantidad ?? 0) / 100
  return {
    kcal: (ap.calorias ?? 0) * f,
    pro:  (ap.proteinas ?? 0) * f,
    cho:  (ap.carbohidratos ?? 0) * f,
    fat:  (ap.grasas ?? 0) * f,
  }
}

// Macro-based SMAE estimation (fallback cuando el nombre no da ninguna pista).
// Sólo puede reconocer AOA/Aceites/Cereales — no distingue Verduras, Frutas,
// Lácteos, Leguminosas ni Azúcares a partir de macros solos.
function macroEstimate(pro: number, cho: number, fat: number): SmaeResult {
  const r = emptyResult()
  r.AOA = Math.max(0, pro / 7)
  const fatRem = Math.max(0, fat - r.AOA * SMAE_META.AOA.fatEq)
  r.Aceites = fatRem / 5
  r.Cereales = Math.max(0, cho) / 15
  return r
}

// Palabras clave por categoría SMAE, para clasificar alimentos de texto libre
// (generados por IA o "libre") que no tienen categoria de catálogo. Cubre
// nombres de comida mexicana comunes.
// Combina palabras clave manuales (comida mexicana común, para no perder
// cobertura) con ~380 palabras reales extraídas de la tabla SMAE de 2159
// alimentos de la práctica del nutriólogo (planilla "Nutriphasev"),
// clasificadas por su GRUPO real, no adivinadas.
const NOMBRE_KEYWORDS: { key: SmaeKey; pattern: RegExp }[] = [
  { key: 'Frutas', pattern: /manzana|pl[aá]tano|papaya|mango|pera\b|uvas?\b|mel[oó]n|sand[ií]a|fresas?|naranja|mandarina|guayaba|pi[ñn]a\b|kiwi|durazno|ciruela|toronja|lim[oó]n\b|cocoy|higo|frutas?|chabacano|gajos|granada|guanabana|jinicuil|mamey|maracuya|nispero|orejones|tamarindo|tuna|zapote/i },
  { key: 'Verduras', pattern: /lechuga|jitomate|tomate|pepino|zanahoria|nopal|espinaca|br[oó]coli|coliflor|calabacit|chayote|apio\b|champi[ñn]|hongo|ejote|\bcol\b|verdura|ensalada|rajas|jicama|jícama|acelga|alcachofa|betabel|chicoria|colorin|creson|guaje|pepinillos|pimiento|setas|verdolaga/i },
  { key: 'Cereales', pattern: /tortilla|arroz|\bpan\b|avena|pasta|tostada|elote|ma[ií]z|cereal|amaranto|atole|bolillo|camote|centeno|espagueti|fideo|galletas?|granola|hojuelas|integral|palomitas|papas?|sopa|tamales?|trigo/i },
  { key: 'Leguminosas', pattern: /frijol(es)?|lenteja|garbanzo|haba\b|soya|alubia|alverjon/i },
  { key: 'AOA', pattern: /pollo|\bres\b|carne|pescado|at[uú]n|huevo|camar[oó]n|cerdo|jam[oó]n|pavo|mariscos?|arrachera|bacalao|bistec|borrego|cabra|calamar|cangrejo|carnero|cecina|chuleta|conejo|cordero|costilla|filete|gallina|iguana|jaiba|langosta|lomo|milanesa|mojarra|muslo|pata\b|pechuga|puerco|pulpo|robalo|salchicha|salmon|salmón|sardinas|ternera|trucha|venado/i },
  { key: 'Lacteos', pattern: /\bleche\b|yogur|queso(?! panela)|descremada|evaporada|helado|jocoque|malteada/i },
  { key: 'Aceites', pattern: /aceite|aguacate|nueces?|almendras?|cacahuate|pistach|aceituna|avellana|cacao|chia|chorizo|manteca|margarina|oliva|pepitas|tocino/i },
  { key: 'Azucares', pattern: /az[uú]car|\bmiel\b|piloncillo|mermelada|gelatina|caramelo|chicle|condensada|flan|jarabe|mousse|paleta|refresco/i },
]

// Estima equivalentes a partir del NOMBRE del alimento (más confiable que solo
// macros, ya reconoce Verduras/Frutas/Lácteos/Leguminosas/Azúcares). Si el
// nombre no matchea ninguna categoría, usa macroEstimate como respaldo.
function nombreEstimate(nombre: string, kcal: number, pro: number, cho: number, fat: number): SmaeResult {
  const matches = NOMBRE_KEYWORDS.filter(k => k.pattern.test(nombre)).map(k => k.key)
  if (matches.length === 0) {
    return macroEstimate(pro, cho, fat)
  }
  const r = emptyResult()
  const kcalPorMatch = kcal / matches.length
  for (const key of matches) {
    r[key] += kcalPorMatch / SMAE_META[key].kcalEq
  }
  return r
}

export function calcEquivAlimentos(alimentos: any[]): SmaeResult {
  const total = emptyResult()

  for (const ap of alimentos) {
    const m = effectiveMacros(ap)
    if (m.kcal <= 0) continue

    if (ap.tipo === 'alimento' && ap.categoria) {
      // Accurate: use SMAE category from DB
      const key = CAT_MAP[ap.categoria]
      if (key) {
        total[key] += m.kcal / SMAE_META[key].kcalEq
        continue
      }
    }

    // Sin categoria de catálogo (alimento libre, IA, o platillo): clasificar
    // por el nombre del alimento, con respaldo por macros si no matchea nada
    const est = nombreEstimate(ap.nombre ?? '', m.kcal, m.pro, m.cho, m.fat)
    for (const k of SMAE_ORDER) {
      total[k] += est[k]
    }
  }

  return total
}

export function calcEquivTiempos(tiempos: any[]): SmaeResult {
  const total = emptyResult()
  for (const t of tiempos) {
    const sub = calcEquivAlimentos(t.alimentos ?? [])
    for (const k of SMAE_ORDER) total[k] += sub[k]
  }
  return total
}

export function totalKcalEquiv(equiv: SmaeResult): number {
  return SMAE_ORDER.reduce((s, k) => s + equiv[k] * SMAE_META[k].kcalEq, 0)
}
