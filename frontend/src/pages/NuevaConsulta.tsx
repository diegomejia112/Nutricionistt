import { useState, FormEvent } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { api } from '../lib/api'

export default function NuevaConsulta() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    fecha: new Date().toISOString().split('T')[0],
    peso: '', presionArterial: '', glucosa: '',
    motivo: '', observaciones: '', recomendaciones: '',
  })

  function set(k: string, v: string) { setForm(f => ({ ...f, [k]: v })) }

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!id) return
    setSaving(true)
    setError('')
    try {
      const body = {
        ...form,
        peso: form.peso ? parseFloat(form.peso) : undefined,
        glucosa: form.glucosa ? parseFloat(form.glucosa) : undefined,
      }
      await api.createConsulta(id, body)
      navigate(`/pacientes/${id}`)
    } catch (err: any) {
      setError(err.message ?? 'Error al guardar')
      setSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <Link to={`/pacientes/${id}`} className="text-gray-400 hover:text-gray-600">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Nueva consulta</h1>
          <p className="text-sm text-gray-500 mt-0.5">Registro de consulta</p>
        </div>
      </div>

      <form onSubmit={submit} className="space-y-5">
        {error && <p className="text-sm text-red-600 bg-red-50 px-4 py-3 rounded-lg">{error}</p>}

        <div className="card p-5 space-y-4">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400">Mediciones</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Fecha *</label>
              <input type="date" className="input" required value={form.fecha} onChange={e => set('fecha', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Peso (kg)</label>
              <input type="number" step="0.1" className="input" value={form.peso} onChange={e => set('peso', e.target.value)} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Presion arterial</label>
              <input className="input" placeholder="120/80" value={form.presionArterial} onChange={e => set('presionArterial', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Glucosa (mg/dL)</label>
              <input type="number" step="0.1" className="input" value={form.glucosa} onChange={e => set('glucosa', e.target.value)} />
            </div>
          </div>
        </div>

        <div className="card p-5 space-y-4">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400">Notas clinicas</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Motivo de consulta</label>
            <input className="input" value={form.motivo} onChange={e => set('motivo', e.target.value)} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Observaciones</label>
            <textarea className="input resize-none" rows={3} value={form.observaciones} onChange={e => set('observaciones', e.target.value)} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Recomendaciones</label>
            <textarea className="input resize-none" rows={3} value={form.recomendaciones} onChange={e => set('recomendaciones', e.target.value)} />
          </div>
        </div>

        <div className="flex gap-3 justify-end">
          <Link to={`/pacientes/${id}`} className="btn-secondary">Cancelar</Link>
          <button type="submit" className="btn-primary" disabled={saving}>
            {saving ? 'Guardando...' : 'Registrar consulta'}
          </button>
        </div>
      </form>
    </div>
  )
}
