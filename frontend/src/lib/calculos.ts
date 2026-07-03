// Cálculos nutricionales estándar: TMB (Mifflin-St Jeor), GET y macros.

export const FACTOR_ACTIVIDAD: Record<string, { label: string; factor: number }> = {
  sedentario:  { label: 'Sedentario (poco o ningún ejercicio)', factor: 1.2 },
  ligero:      { label: 'Ligero (1-2 días/sem)',                factor: 1.375 },
  moderado:    { label: 'Moderado (3-4 días/sem)',              factor: 1.55 },
  activo:      { label: 'Activo (5+ días/sem)',                 factor: 1.725 },
  muy_activo:  { label: 'Muy activo / atleta',                  factor: 1.9 },
}

// Mifflin-St Jeor: TMB en kcal/día
// Hombres: 10*peso + 6.25*altura - 5*edad + 5
// Mujeres: 10*peso + 6.25*altura - 5*edad - 161
export function calcularTMB(pesoKg: number, alturaCm: number, edad: number, sexo: string): number {
  const base = 10 * pesoKg + 6.25 * alturaCm - 5 * edad
  return sexo === 'M' ? base + 5 : base - 161
}

export function detectarAjusteObjetivo(objetivo: string): { ajuste: number; motivo: string } {
  const o = (objetivo || '').toLowerCase()
  if (/baja|perd|reduc|adelgaz/.test(o)) return { ajuste: -500, motivo: 'déficit de 500 kcal para bajar de peso' }
  if (/sub|gana|aument|masa/.test(o)) return { ajuste: 400, motivo: 'superávit de 400 kcal para ganar peso/masa' }
  return { ajuste: 0, motivo: 'mantenimiento, sin ajuste' }
}

// Proteína en g/kg de peso — así se calcula en la práctica real (no % fijo):
// 1.2 g/kg es el estándar; condiciones renales bajan a 0.8 g/kg para no
// sobrecargar el riñón. El resto de calorías se reparte entre carbos/grasas.
export function sugerirProteinaGkg(enfermedades: string): { gkg: number; motivo: string } {
  if (/renal|ri[ñn]on/i.test(enfermedades || '')) {
    return { gkg: 0.8, motivo: 'restricción por condición renal' }
  }
  return { gkg: 1.2, motivo: 'estándar' }
}

// Macros a partir de proteína en g/kg + reparto de lo restante entre
// carbohidratos y grasas (porcentajes del remanente, no del total).
export function calcularMacrosPorProteinaGkg(
  caloriasObj: number, pesoKg: number, proteinaGkg: number, pctCarbRestante: number
): { proteinas: number; carbos: number; grasas: number } {
  const proteinas = Math.round(proteinaGkg * pesoKg)
  const kcalProteina = proteinas * 4
  const kcalRestante = Math.max(0, caloriasObj - kcalProteina)
  const carbos = Math.round((kcalRestante * (pctCarbRestante / 100)) / 4)
  const grasas = Math.round((kcalRestante * (1 - pctCarbRestante / 100)) / 9)
  return { proteinas, carbos, grasas }
}
