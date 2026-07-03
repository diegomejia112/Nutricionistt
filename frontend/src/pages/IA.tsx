import { useState, FormEvent, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Cpu, Loader, Key, Eye, EyeOff, CheckCircle, User } from 'lucide-react'
import { api } from '../lib/api'
import CalculadoraMifflin from '../components/ui/CalculadoraMifflin'

export default function IA() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [pacientes, setPacientes] = useState<any[]>([])
  const [pacienteInfo, setPacienteInfo] = useState<any>(null)
  const [generating, setGenerating] = useState(false)
  const [error, setError] = useState('')
  const [keyConfigurada, setKeyConfigurada] = useState<boolean | null>(null)
  const [apiKey, setApiKey] = useState('')
  const [showKey, setShowKey] = useState(false)
  const [savingKey, setSavingKey] = useState(false)
  const [keySaved, setKeySaved] = useState(false)
  const [form, setForm] = useState({
    pacienteId: searchParams.get('pacienteId') ?? '',
    nombrePlan: '',
    caloriasObj: '',
    proteinasObj: '',
    restriccionesExtra: '',
    preferenciaRegion: '',
  })

  useEffect(() => {
    api.getPacientes('', 100).then(setPacientes).catch(() => {})
    api.getIAConfig().then(c => setKeyConfigurada(c.configurada)).catch(() => setKeyConfigurada(false))
  }, [])

  // Cargar info del paciente cuando se selecciona
  useEffect(() => {
    if (!form.pacienteId) { setPacienteInfo(null); return }
    api.getPaciente(form.pacienteId).then(d => setPacienteInfo(d)).catch(() => {})
  }, [form.pacienteId])

  function set(k: string, v: string) { setForm(f => ({ ...f, [k]: v })) }

  const p = pacienteInfo?.paciente
  const restricciones: any[] = pacienteInfo?.restricciones ?? []

  async function guardarKey(e: FormEvent) {
    e.preventDefault()
    if (!apiKey.trim()) return
    setSavingKey(true)
    try {
      await api.saveIAConfig(apiKey.trim())
      setKeyConfigurada(true)
      setKeySaved(true)
      setApiKey('')
    } catch (err: any) {
      setError(err.message ?? 'Error guardando la key')
    } finally {
      setSavingKey(false)
    }
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!form.pacienteId || !form.nombrePlan) {
      setError('Selecciona un paciente y escribe un nombre para el plan')
      return
    }
    setGenerating(true)
    setError('')
    try {
      const result = await api.generarPlan({
        pacienteId: form.pacienteId,
        nombrePlan: form.nombrePlan,
        caloriasObj: form.caloriasObj ? parseFloat(form.caloriasObj) : undefined,
        proteinasObj: form.proteinasObj ? parseFloat(form.proteinasObj) : undefined,
        restriccionesExtra: form.restriccionesExtra,
        preferenciaRegion: form.preferenciaRegion,
      })
      if (result.planId) navigate(`/planes/${result.planId}`)
    } catch (err: any) {
      setError(err.message ?? 'Error al generar el plan')
      setGenerating(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <div className="mb-6">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-rose-50 flex items-center justify-center">
            <Cpu className="w-5 h-5 text-rose-500" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Generar plan con IA</h1>
            <p className="text-sm text-gray-500 mt-0.5">Plan 100% personalizado basado en el expediente del paciente</p>
          </div>
        </div>
      </div>

      {/* API Key */}
      <div className="card p-4 mb-5">
        <div className="flex items-center gap-2 mb-2">
          <Key className="w-4 h-4 text-gray-400" />
          <h2 className="text-sm font-semibold text-gray-700">API Key DeepSeek</h2>
          {keyConfigurada && (
            <span className="ml-auto flex items-center gap-1 text-xs text-green-600">
              <CheckCircle className="w-3.5 h-3.5" /> Configurada
            </span>
          )}
        </div>
        {keyConfigurada === false && (
          <p className="text-xs text-amber-600 bg-amber-50 px-3 py-2 rounded-lg mb-2">
            Necesitas una API Key de DeepSeek. Obtenla en <strong>platform.deepseek.com</strong>
          </p>
        )}
        <form onSubmit={guardarKey} className="flex gap-2">
          <div className="relative flex-1">
            <input type={showKey ? 'text' : 'password'} className="input pr-9"
              placeholder={keyConfigurada ? 'Cambiar key...' : 'sk-...'}
              value={apiKey} onChange={e => setApiKey(e.target.value)} />
            <button type="button" onClick={() => setShowKey(v => !v)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">
              {showKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
          <button type="submit" className={keyConfigurada ? 'btn-secondary' : 'btn-primary'}
            disabled={savingKey || !apiKey.trim()}>
            {savingKey ? '...' : keySaved ? 'Guardada ✓' : keyConfigurada ? 'Cambiar' : 'Guardar'}
          </button>
        </form>
      </div>

      <form onSubmit={submit} className="space-y-4">
        {error && <p className="text-sm text-red-600 bg-red-50 px-4 py-3 rounded-lg">{error}</p>}

        {/* Seleccionar paciente */}
        <div className="card p-5 space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Paciente</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Paciente *</label>
            <select className="input" required value={form.pacienteId} onChange={e => set('pacienteId', e.target.value)}>
              <option value="">Seleccionar paciente...</option>
              {pacientes.map(pac => (
                <option key={pac.id} value={pac.id}>{pac.nombre} {pac.apellidos}</option>
              ))}
            </select>
          </div>

          {/* Resumen del expediente del paciente seleccionado */}
          {p && (
            <div className="bg-gray-50 rounded-xl p-4 space-y-2 text-sm">
              <div className="flex items-center gap-2 mb-3">
                <User className="w-4 h-4 text-gray-400" />
                <span className="font-medium text-gray-700">Perfil cargado automáticamente</span>
              </div>
              <div className="grid grid-cols-2 gap-x-6 gap-y-1 text-xs">
                {p.edad && <InfoRow label="Edad" value={`${p.edad} años`} />}
                {p.sexo && <InfoRow label="Sexo" value={p.sexo === 'F' ? 'Femenino' : 'Masculino'} />}
                {p.peso && <InfoRow label="Peso" value={`${p.peso} kg`} />}
                {p.altura && <InfoRow label="Altura" value={`${p.altura} cm`} />}
                {p.nivelActividad && <InfoRow label="Actividad" value={p.nivelActividad} />}
              </div>
              {p.objetivo && (
                <div className="mt-2">
                  <span className="text-xs text-gray-400">Objetivo: </span>
                  <span className="text-xs font-medium text-gray-700">{p.objetivo}</span>
                </div>
              )}
              {p.enfermedades && (
                <div className="flex flex-wrap gap-1 mt-1">
                  {p.enfermedades.split(',').map((e: string) => e.trim()).filter(Boolean).map((e: string) => (
                    <span key={e} className="text-xs bg-amber-100 text-amber-700 px-2 py-0.5 rounded-full">{e}</span>
                  ))}
                </div>
              )}
              {p.alergias && (
                <div className="flex flex-wrap gap-1 mt-1">
                  {p.alergias.split(',').map((a: string) => a.trim()).filter(Boolean).map((a: string) => (
                    <span key={a} className="text-xs bg-red-100 text-red-600 px-2 py-0.5 rounded-full">⚠ {a}</span>
                  ))}
                </div>
              )}
              {p.alimentosEvitar && (
                <div className="flex flex-wrap gap-1 mt-1">
                  <span className="text-xs text-gray-400">No come: </span>
                  {p.alimentosEvitar.split(',').map((a: string) => a.trim()).filter(Boolean).map((a: string) => (
                    <span key={a} className="text-xs bg-gray-200 text-gray-600 px-2 py-0.5 rounded-full">{a}</span>
                  ))}
                </div>
              )}
              {restricciones.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-1">
                  <span className="text-xs text-gray-400">Ingredientes restringidos: </span>
                  {restricciones.map((r: any) => (
                    <span key={r.id} className="text-xs bg-orange-100 text-orange-700 px-2 py-0.5 rounded-full">{r.ingrediente}</span>
                  ))}
                </div>
              )}
              {p.preferencias && (
                <div className="text-xs text-gray-500 mt-1">
                  Prefiere: <span className="text-gray-700">{p.preferencias}</span>
                </div>
              )}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nombre del plan *</label>
            <input className="input" required placeholder="Ej: Plan junio 2026"
              value={form.nombrePlan} onChange={e => set('nombrePlan', e.target.value)} />
          </div>
        </div>

        {/* Parámetros adicionales */}
        <div className="card p-5 space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Ajustes adicionales (opcionales)</h2>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Calorías objetivo (kcal/día)</label>
              <input type="number" className="input" placeholder="Automático"
                value={form.caloriasObj} onChange={e => set('caloriasObj', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Proteína objetivo (g/día)</label>
              <input type="number" className="input" placeholder="Automático"
                value={form.proteinasObj} onChange={e => set('proteinasObj', e.target.value)} />
            </div>
          </div>

          {p && (
            <CalculadoraMifflin paciente={p} aplicarLabel="Usar estos valores"
              onAplicar={v => setForm(f => ({ ...f, caloriasObj: String(v.caloriasObj), proteinasObj: String(v.proteinas) }))} />
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Preferencia regional</label>
            <select className="input" value={form.preferenciaRegion} onChange={e => set('preferenciaRegion', e.target.value)}>
              <option value="">Sin preferencia</option>
              <option value="norte">Norte de México</option>
              <option value="centro">Centro</option>
              <option value="sur">Sur / Sureste</option>
              <option value="costa">Costa / Golfo</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Indicaciones extra para la IA</label>
            <textarea className="input resize-none" rows={2}
              placeholder="Ej: Que incluya muchos nopales, que el presupuesto sea bajo, que evite cocinar complicado..."
              value={form.restriccionesExtra} onChange={e => set('restriccionesExtra', e.target.value)} />
          </div>
        </div>

        {generating ? (
          <div className="card p-8 flex flex-col items-center gap-4 text-center">
            <Loader className="w-8 h-8 text-rose-400 animate-spin" />
            <div>
              <p className="font-semibold text-gray-800">Generando plan personalizado...</p>
              <p className="text-sm text-gray-400 mt-1">DeepSeek está analizando el expediente completo. Puede tomar hasta 2 minutos.</p>
            </div>
          </div>
        ) : (
          <button type="submit" disabled={!keyConfigurada || !form.pacienteId}
            className="btn-primary w-full flex items-center justify-center gap-2 py-3 disabled:opacity-50 disabled:cursor-not-allowed">
            <Cpu className="w-4 h-4" />
            {!keyConfigurada ? 'Configura la API Key primero'
              : !form.pacienteId ? 'Selecciona un paciente'
              : 'Generar plan personalizado con IA'}
          </button>
        )}
      </form>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex gap-2">
      <span className="text-gray-400">{label}:</span>
      <span className="text-gray-700 font-medium">{value}</span>
    </div>
  )
}
