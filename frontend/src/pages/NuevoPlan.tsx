import { useState, FormEvent, useEffect } from 'react'
import { useNavigate, Link, useSearchParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { api } from '../lib/api'

export default function NuevoPlan() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [pacientes, setPacientes] = useState<any[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    pacienteId: searchParams.get('pacienteId') ?? '',
    nombre: '',
    descripcion: '',
    fechaInicio: new Date().toISOString().split('T')[0],
    caloriasObj: '',
    proteinasObj: '',
    carbsObj: '',
    grasasObj: '',
  })

  useEffect(() => {
    api.getPacientes('', 100).then(setPacientes).catch(() => {})
  }, [])

  function set(k: string, v: string) { setForm(f => ({ ...f, [k]: v })) }

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!form.pacienteId || !form.nombre) {
      setError('Selecciona un paciente y escribe un nombre')
      return
    }
    setSaving(true)
    setError('')
    try {
      const body = {
        pacienteId: form.pacienteId,
        nombre: form.nombre,
        descripcion: form.descripcion,
        fechaInicio: form.fechaInicio,
        caloriasObj: form.caloriasObj ? parseFloat(form.caloriasObj) : undefined,
        proteinasObj: form.proteinasObj ? parseFloat(form.proteinasObj) : undefined,
        carbsObj: form.carbsObj ? parseFloat(form.carbsObj) : undefined,
        grasasObj: form.grasasObj ? parseFloat(form.grasasObj) : undefined,
      }
      const p = await api.createPlan(body)
      navigate(`/planes/${p.id}`)
    } catch (err: any) {
      setError(err.message ?? 'Error al crear plan')
      setSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <Link to="/planes" className="text-gray-400 hover:text-gray-600">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Nuevo plan</h1>
          <p className="text-sm text-gray-500 mt-0.5">Plan semanal de alimentacion</p>
        </div>
      </div>

      <form onSubmit={submit} className="space-y-5">
        {error && <p className="text-sm text-red-600 bg-red-50 px-4 py-3 rounded-lg">{error}</p>}

        <div className="card p-5 space-y-4">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400">Datos del plan</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Paciente *</label>
            <select className="input" required value={form.pacienteId} onChange={e => set('pacienteId', e.target.value)}>
              <option value="">Seleccionar paciente</option>
              {pacientes.map(p => (
                <option key={p.id} value={p.id}>{p.nombre} {p.apellidos}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nombre del plan *</label>
            <input className="input" required placeholder="Ej: Plan de reduccion de peso" value={form.nombre} onChange={e => set('nombre', e.target.value)} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Descripcion</label>
            <textarea className="input resize-none" rows={2} value={form.descripcion} onChange={e => set('descripcion', e.target.value)} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Fecha de inicio</label>
            <input type="date" className="input" value={form.fechaInicio} onChange={e => set('fechaInicio', e.target.value)} />
          </div>
        </div>

        <div className="card p-5 space-y-4">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400">Objetivos nutricionales (opcional)</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Calorias (kcal)</label>
              <input type="number" className="input" value={form.caloriasObj} onChange={e => set('caloriasObj', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Proteinas (g)</label>
              <input type="number" className="input" value={form.proteinasObj} onChange={e => set('proteinasObj', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Carbohidratos (g)</label>
              <input type="number" className="input" value={form.carbsObj} onChange={e => set('carbsObj', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Grasas (g)</label>
              <input type="number" className="input" value={form.grasasObj} onChange={e => set('grasasObj', e.target.value)} />
            </div>
          </div>
        </div>

        <div className="flex gap-3 justify-end">
          <Link to="/planes" className="btn-secondary">Cancelar</Link>
          <button type="submit" className="btn-primary" disabled={saving}>
            {saving ? 'Creando...' : 'Crear plan'}
          </button>
        </div>
      </form>
    </div>
  )
}
