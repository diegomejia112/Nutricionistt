import { useState, FormEvent, useRef } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, X } from 'lucide-react'
import { api } from '../lib/api'

const SUGERENCIAS_ENFERMEDADES = [
  'Diabetes tipo 2','Diabetes tipo 1','Hipertensión arterial','Obesidad','Sobrepeso',
  'Hipotiroidismo','Hipertiroidismo','Colesterol alto','Triglicéridos altos',
  'Síndrome metabólico','Hígado graso','Resistencia a la insulina',
  'Enfermedad renal crónica','Gastritis','Colitis','Enfermedad celiaca',
  'Síndrome de intestino irritable','Anemia','Osteoporosis','Artritis',
  'Gota','Cálculos renales','Reflujo gastroesofágico','Hipercolesterolemia',
]

const SUGERENCIAS_ALERGIAS = [
  'Lactosa','Gluten','Cacahuate','Mariscos','Camarón','Huevo','Soya',
  'Maíz','Trigo','Nueces','Leche de vaca','Fructosa','Sorbitol',
]

const SUGERENCIAS_EVITAR = [
  'Cerdo','Res','Pollo','Pescado','Mariscos','Lácteos','Huevo','Gluten',
  'Picante','Grasas fritas','Azúcar','Harinas refinadas','Embutidos',
  'Café','Alcohol','Refresco','Comida rápida','Pan dulce',
]

const SUGERENCIAS_PREFERENCIAS = [
  'Comida mexicana','Sopas y caldos','Leguminosas','Verduras al vapor',
  'Carnes a la plancha','Frutas','Ensaladas','Tacos','Tamales',
  'Arroz','Frijoles','Avena','Licuados','Agua fresca',
]

const OBJETIVOS = [
  'Bajar de peso','Subir de peso','Mantener peso','Control glucémico',
  'Reducir colesterol','Reducir triglicéridos','Control de hipertensión',
  'Ganar masa muscular','Mejorar digestión','Alimentación saludable general',
  'Plan posparto','Plan para embarazo','Control de gota','Dieta renal',
]

const NIVELES_ACTIVIDAD = [
  { val: 'sedentario', label: 'Sedentario — sin ejercicio' },
  { val: 'ligero', label: 'Ligero — 1-2 veces/semana' },
  { val: 'moderado', label: 'Moderado — 3-4 veces/semana' },
  { val: 'activo', label: 'Activo — 5+ veces/semana' },
  { val: 'muy_activo', label: 'Muy activo — atleta / trabajo físico' },
]

function TagInput({
  label, placeholder, value, onChange, sugerencias,
}: {
  label: string; placeholder: string; value: string
  onChange: (v: string) => void; sugerencias: string[]
}) {
  const [input, setInput] = useState('')
  const [open, setOpen] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const tags = value ? value.split(',').map(t => t.trim()).filter(Boolean) : []

  const filtradas = sugerencias.filter(s =>
    !tags.includes(s) && s.toLowerCase().includes(input.toLowerCase())
  )

  function add(tag: string) {
    const t = tag.trim()
    if (!t || tags.includes(t)) return
    onChange([...tags, t].join(', '))
    setInput('')
    setOpen(false)
    inputRef.current?.focus()
  }

  function remove(tag: string) {
    onChange(tags.filter(t => t !== tag).join(', '))
  }

  function handleKey(e: React.KeyboardEvent) {
    if ((e.key === 'Enter' || e.key === ',') && input.trim()) {
      e.preventDefault()
      add(input)
    }
    if (e.key === 'Backspace' && !input && tags.length) {
      remove(tags[tags.length - 1])
    }
  }

  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <div
        className="min-h-[42px] border border-gray-200 rounded-xl px-3 py-2 flex flex-wrap gap-1.5 cursor-text bg-white focus-within:ring-2 focus-within:ring-rose-300 focus-within:border-rose-400"
        onClick={() => inputRef.current?.focus()}>
        {tags.map(t => (
          <span key={t} className="inline-flex items-center gap-1 text-sm bg-rose-50 text-rose-700 px-2 py-0.5 rounded-full">
            {t}
            <button type="button" onClick={() => remove(t)} className="hover:text-rose-900">
              <X className="w-3 h-3" />
            </button>
          </span>
        ))}
        <div className="relative flex-1 min-w-[120px]">
          <input
            ref={inputRef}
            className="w-full outline-none text-sm bg-transparent"
            placeholder={tags.length === 0 ? placeholder : 'Agregar más...'}
            value={input}
            onChange={e => { setInput(e.target.value); setOpen(true) }}
            onFocus={() => setOpen(true)}
            onBlur={() => setTimeout(() => setOpen(false), 150)}
            onKeyDown={handleKey}
          />
          {open && filtradas.length > 0 && (
            <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-xl shadow-lg z-20 max-h-48 overflow-y-auto min-w-[200px]">
              {filtradas.map(s => (
                <button key={s} type="button"
                  className="w-full text-left px-3 py-2 text-sm hover:bg-rose-50 hover:text-rose-700"
                  onMouseDown={() => add(s)}>
                  {s}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
      <p className="text-xs text-gray-400 mt-1">Escribe y presiona Enter, o selecciona una sugerencia</p>
    </div>
  )
}

export default function NuevoPaciente() {
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState({
    nombre: '', apellidos: '', email: '', telefono: '',
    fechaNacimiento: '', sexo: '', peso: '', altura: '',
    objetivo: '', enfermedades: '', alergias: '',
    preferencias: '', alimentosEvitar: '',
    horarioComidas: '', nivelActividad: '', notas: '',
  })

  function set(k: string, v: string) { setForm(f => ({ ...f, [k]: v })) }

  async function submit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      const body = {
        ...form,
        peso: form.peso ? parseFloat(form.peso) : undefined,
        altura: form.altura ? parseFloat(form.altura) : undefined,
      }
      const p = await api.createPaciente(body)
      navigate(`/pacientes/${p.id}`)
    } catch (err: any) {
      setError(err.message ?? 'Error al guardar')
      setSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <Link to="/pacientes" className="text-gray-400 hover:text-gray-600">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Nuevo paciente</h1>
          <p className="text-sm text-gray-500 mt-0.5">Expediente clínico completo</p>
        </div>
      </div>

      <form onSubmit={submit} className="space-y-5">
        {error && <p className="text-sm text-red-600 bg-red-50 px-4 py-3 rounded-lg">{error}</p>}

        {/* Datos personales */}
        <div className="card p-5 space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Datos personales</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Nombre *</label>
              <input className="input" required value={form.nombre} onChange={e => set('nombre', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Apellidos *</label>
              <input className="input" required value={form.apellidos} onChange={e => set('apellidos', e.target.value)} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Correo</label>
              <input type="email" className="input" value={form.email} onChange={e => set('email', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Teléfono</label>
              <input className="input" value={form.telefono} onChange={e => set('telefono', e.target.value)} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Fecha de nacimiento</label>
              <input type="date" className="input" value={form.fechaNacimiento} onChange={e => set('fechaNacimiento', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Sexo</label>
              <select className="input" value={form.sexo} onChange={e => set('sexo', e.target.value)}>
                <option value="">Seleccionar</option>
                <option value="F">Femenino</option>
                <option value="M">Masculino</option>
              </select>
            </div>
          </div>
        </div>

        {/* Antropometría */}
        <div className="card p-5 space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Antropometría</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Peso actual (kg)</label>
              <input type="number" step="0.1" className="input" placeholder="70.5" value={form.peso} onChange={e => set('peso', e.target.value)} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Altura (cm)</label>
              <input type="number" step="0.1" className="input" placeholder="165" value={form.altura} onChange={e => set('altura', e.target.value)} />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nivel de actividad física</label>
            <select className="input" value={form.nivelActividad} onChange={e => set('nivelActividad', e.target.value)}>
              <option value="">Seleccionar</option>
              {NIVELES_ACTIVIDAD.map(n => (
                <option key={n.val} value={n.val}>{n.label}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Objetivos y condición clínica */}
        <div className="card p-5 space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Condición clínica</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Objetivo principal</label>
            <div className="flex flex-wrap gap-2 mb-2">
              {OBJETIVOS.map(o => (
                <button key={o} type="button"
                  onClick={() => set('objetivo', o)}
                  className={`text-xs px-3 py-1.5 rounded-full border transition-colors ${
                    form.objetivo === o
                      ? 'bg-rose-500 text-white border-rose-500'
                      : 'border-gray-200 text-gray-600 hover:border-rose-300 hover:text-rose-600'
                  }`}>
                  {o}
                </button>
              ))}
            </div>
            <input className="input" placeholder="O escribe un objetivo específico..."
              value={form.objetivo} onChange={e => set('objetivo', e.target.value)} />
          </div>

          <TagInput
            label="Enfermedades / condiciones médicas"
            placeholder="Diabetes tipo 2, hipertensión..."
            value={form.enfermedades}
            onChange={v => set('enfermedades', v)}
            sugerencias={SUGERENCIAS_ENFERMEDADES}
          />

          <TagInput
            label="Alergias e intolerancias"
            placeholder="Lactosa, gluten, cacahuate..."
            value={form.alergias}
            onChange={v => set('alergias', v)}
            sugerencias={SUGERENCIAS_ALERGIAS}
          />
        </div>

        {/* Hábitos alimenticios */}
        <div className="card p-5 space-y-4">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Hábitos alimenticios</h2>
          <TagInput
            label="Alimentos que NO come / no le gustan"
            placeholder="Cerdo, mariscos, picante..."
            value={form.alimentosEvitar}
            onChange={v => set('alimentosEvitar', v)}
            sugerencias={SUGERENCIAS_EVITAR}
          />

          <TagInput
            label="Alimentos preferidos / que le gustan"
            placeholder="Frijoles, verduras, pollo..."
            value={form.preferencias}
            onChange={v => set('preferencias', v)}
            sugerencias={SUGERENCIAS_PREFERENCIAS}
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Horario de comidas</label>
            <input className="input" placeholder="Ej: Desayuna 8am, come 2pm, cena 8pm. Trabaja de noche."
              value={form.horarioComidas} onChange={e => set('horarioComidas', e.target.value)} />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Notas adicionales</label>
            <textarea className="input resize-none" rows={2}
              placeholder="Contexto extra: situación económica, familia, cocina propia..."
              value={form.notas} onChange={e => set('notas', e.target.value)} />
          </div>
        </div>

        <div className="flex gap-3 justify-end pb-4">
          <Link to="/pacientes" className="btn-secondary">Cancelar</Link>
          <button type="submit" className="btn-primary" disabled={saving}>
            {saving ? 'Guardando...' : 'Guardar paciente'}
          </button>
        </div>
      </form>
    </div>
  )
}
