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

// Macro-based SMAE estimation for composite foods (platillos / IA)
function macroEstimate(pro: number, cho: number, fat: number): SmaeResult {
  const r = emptyResult()
  // 1. Protein → AOA (7g pro = 1 equiv)
  r.AOA = Math.max(0, pro / 7)
  // 2. Fat → Aceites after subtracting fat already accounted in AOA
  const fatRem = Math.max(0, fat - r.AOA * SMAE_META.AOA.fatEq)
  r.Aceites = fatRem / 5
  // 3. Carbs → Cereales (15g = 1 equiv)
  r.Cereales = Math.max(0, cho) / 15
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

    // Fallback: estimate from macros
    const est = macroEstimate(m.pro, m.cho, m.fat)
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
