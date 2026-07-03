import { useState, useEffect } from 'react'
import { Calculator } from 'lucide-react'
import {
  calcularTMB, detectarAjusteObjetivo, sugerirProteinaGkg,
  calcularMacrosPorProteinaGkg, FACTOR_ACTIVIDAD,
} from '../../lib/calculos'

type Paciente = {
  peso?: number | null
  altura?: number | null
  edad?: number | null
  sexo?: string | null
  nivelActividad?: string | null
  objetivo?: string | null
  enfermedades?: string | null
}

export default function CalculadoraMifflin({ paciente, onAplicar, aplicarLabel = 'Usar estos valores' }: {
  paciente: Paciente
  onAplicar?: (v: { caloriasObj: number; proteinas: number; carbos: number; grasas: number }) => void
  aplicarLabel?: string
}) {
  const [nivelActividad, setNivelActividad] = useState('sedentario')
  const [ajusteKcal, setAjusteKcal] = useState(0)
  const [proteinaGkg, setProteinaGkg] = useState(1.2)
  const [proteinaMotivo, setProteinaMotivo] = useState('estándar')
  const [pctCarbRestante, setPctCarbRestante] = useState(65) // del kcal restante tras la proteína

  // Inicializar con los valores sugeridos automáticamente al cargar/cambiar de paciente
  useEffect(() => {
    setNivelActividad(paciente.nivelActividad ?? 'sedentario')
    setAjusteKcal(detectarAjusteObjetivo(paciente.objetivo ?? '').ajuste)
    const sugerido = sugerirProteinaGkg(paciente.enfermedades ?? '')
    setProteinaGkg(sugerido.gkg)
    setProteinaMotivo(sugerido.motivo)
    setPctCarbRestante(65)
  }, [paciente.peso, paciente.altura, paciente.edad, paciente.sexo, paciente.nivelActividad, paciente.objetivo, paciente.enfermedades])

  if (!paciente.peso || !paciente.altura || !paciente.edad || !paciente.sexo) {
    return (
      <p className="text-sm text-amber-600 bg-amber-50 px-3 py-2 rounded-lg">
        Faltan datos en el expediente del paciente (peso, altura, fecha de nacimiento o sexo) para calcular.
      </p>
    )
  }

  const tmb = calcularTMB(paciente.peso, paciente.altura, paciente.edad, paciente.sexo)
  const factor = FACTOR_ACTIVIDAD[nivelActividad]?.factor ?? 1.2
  const get = tmb * factor
  const caloriasObj = Math.max(1200, Math.round(get + ajusteKcal))
  const macros = calcularMacrosPorProteinaGkg(caloriasObj, paciente.peso, proteinaGkg, pctCarbRestante)

  return (
    <div className="bg-gray-50 rounded-xl p-4 space-y-3">
      <div className="flex items-center gap-2">
        <Calculator className="w-4 h-4 text-gray-400" />
        <span className="text-xs font-semibold text-gray-600 uppercase tracking-wide">
          Cálculo automático (Mifflin-St Jeor)
        </span>
      </div>

      <div className="flex items-center justify-between text-xs">
        <span className="text-gray-500">
          TMB ({paciente.sexo === 'M' ? '10×peso+6.25×altura−5×edad+5' : '10×peso+6.25×altura−5×edad−161'})
        </span>
        <span className="font-medium text-gray-700 tabular-nums">{Math.round(tmb)} kcal</span>
      </div>

      <div className="grid grid-cols-2 gap-3 items-end">
        <div>
          <label className="text-xs text-gray-500 mb-1 block">Factor de actividad</label>
          <select className="input text-sm py-1.5" value={nivelActividad} onChange={e => setNivelActividad(e.target.value)}>
            {Object.entries(FACTOR_ACTIVIDAD).map(([key, v]) => (
              <option key={key} value={key}>{v.label} (×{v.factor})</option>
            ))}
          </select>
        </div>
        <div>
          <label className="text-xs text-gray-500 mb-1 block">Ajuste (kcal, +/-)</label>
          <input type="number" step="50" className="input text-sm py-1.5"
            value={ajusteKcal} onChange={e => setAjusteKcal(parseFloat(e.target.value) || 0)} />
        </div>
      </div>

      <div className="flex items-center justify-between text-xs text-gray-500">
        <span>GET = TMB × actividad</span>
        <span className="font-medium text-gray-700 tabular-nums">{Math.round(get)} kcal</span>
      </div>

      <div className="grid grid-cols-2 gap-3 items-end">
        <div>
          <label className="text-xs text-gray-500 mb-1 block">
            Proteína (g/kg peso) {proteinaMotivo !== 'estándar' && <span className="text-amber-600">— {proteinaMotivo}</span>}
          </label>
          <input type="number" step="0.1" className="input text-sm py-1.5"
            value={proteinaGkg} onChange={e => setProteinaGkg(parseFloat(e.target.value) || 0)} />
        </div>
        <div>
          <label className="text-xs text-gray-500 mb-1 block">% carbos del resto (vs grasas)</label>
          <input type="number" className="input text-sm py-1.5"
            value={pctCarbRestante} onChange={e => setPctCarbRestante(parseFloat(e.target.value) || 0)} />
        </div>
      </div>
      <p className="text-[10px] text-gray-400">
        Proteína = g/kg × peso. Lo que sobra de las calorías se reparte entre carbohidratos y grasas según el % de arriba.
      </p>

      <div className="flex items-center justify-between pt-2 border-t border-gray-200">
        <span className="text-sm font-semibold text-gray-800">Calorías objetivo</span>
        <span className="text-sm font-bold text-rose-600 tabular-nums">{caloriasObj} kcal</span>
      </div>
      <div className="flex items-center justify-between text-xs text-gray-400">
        <span>Macros resultantes</span>
        <span className="tabular-nums">P {macros.proteinas}g · C {macros.carbos}g · G {macros.grasas}g</span>
      </div>

      {onAplicar && (
        <button type="button"
          onClick={() => onAplicar({ caloriasObj, proteinas: macros.proteinas, carbos: macros.carbos, grasas: macros.grasas })}
          className="btn-secondary w-full text-sm">
          {aplicarLabel}
        </button>
      )}
    </div>
  )
}
