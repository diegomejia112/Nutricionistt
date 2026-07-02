import { useEffect, useState, useRef } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, Trash2, Edit2, Save, X, Cpu, TrendingDown, TrendingUp, Minus } from 'lucide-react'
import { api } from '../lib/api'

const SUGERENCIAS_ENFERMEDADES = [
  'Diabetes tipo 2','Diabetes tipo 1','Hipertensión arterial','Obesidad','Sobrepeso',
  'Hipotiroidismo','Hipertiroidismo','Colesterol alto','Triglicéridos altos',
  'Síndrome metabólico','Hígado graso','Resistencia a la insulina',
  'Enfermedad renal crónica','Gastritis','Colitis','Enfermedad celiaca',
  'Síndrome de intestino irritable','Anemia','Osteoporosis','Gota',
]
const SUGERENCIAS_ALERGIAS = [
  'Lactosa','Gluten','Cacahuate','Mariscos','Camarón','Huevo','Soya',
  'Maíz','Trigo','Nueces','Leche de vaca','Fructosa',
]
const SUGERENCIAS_EVITAR = [
  'Cerdo','Res','Pollo','Pescado','Mariscos','Lácteos','Huevo','Gluten',
  'Picante','Grasas fritas','Azúcar','Harinas refinadas','Embutidos','Café','Alcohol',
]
const SUGERENCIAS_PREFERENCIAS = [
  'Comida mexicana','Sopas y caldos','Leguminosas','Verduras al vapor',
  'Carnes a la plancha','Frutas','Ensaladas','Tacos','Arroz','Frijoles','Avena',
]
const NIVELES_ACTIVIDAD = [
  { val: 'sedentario', label: 'Sedentario' },
  { val: 'ligero', label: 'Ligero (1-2/sem)' },
  { val: 'moderado', label: 'Moderado (3-4/sem)' },
  { val: 'activo', label: 'Activo (5+/sem)' },
  { val: 'muy_activo', label: 'Muy activo / atleta' },
]

function TagInput({ value, onChange, sugerencias, placeholder }: {
  value: string; onChange: (v: string) => void
  sugerencias: string[]; placeholder: string
}) {
  const [input, setInput] = useState('')
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLInputElement>(null)
  const tags = value ? value.split(',').map(t => t.trim()).filter(Boolean) : []
  const filtradas = sugerencias.filter(s => !tags.includes(s) && s.toLowerCase().includes(input.toLowerCase()))

  function add(t: string) {
    const tag = t.trim()
    if (!tag || tags.includes(tag)) return
    onChange([...tags, tag].join(', '))
    setInput(''); setOpen(false); ref.current?.focus()
  }
  function remove(t: string) { onChange(tags.filter(x => x !== t).join(', ')) }
  function handleKey(e: React.KeyboardEvent) {
    if ((e.key === 'Enter' || e.key === ',') && input.trim()) { e.preventDefault(); add(input) }
    if (e.key === 'Backspace' && !input && tags.length) remove(tags[tags.length - 1])
  }

  return (
    <div className="min-h-[42px] border border-gray-200 rounded-xl px-3 py-2 flex flex-wrap gap-1.5 cursor-text bg-white focus-within:ring-2 focus-within:ring-rose-300 focus-within:border-rose-400"
      onClick={() => ref.current?.focus()}>
      {tags.map(t => (
        <span key={t} className="inline-flex items-center gap-1 text-xs bg-rose-50 text-rose-700 px-2 py-0.5 rounded-full">
          {t}
          <button type="button" onClick={() => remove(t)}><X className="w-3 h-3" /></button>
        </span>
      ))}
      <div className="relative flex-1 min-w-[120px]">
        <input ref={ref} className="w-full outline-none text-sm bg-transparent"
          placeholder={tags.length === 0 ? placeholder : ''}
          value={input}
          onChange={e => { setInput(e.target.value); setOpen(true) }}
          onFocus={() => setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          onKeyDown={handleKey} />
        {open && filtradas.length > 0 && (
          <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-xl shadow-lg z-20 max-h-40 overflow-y-auto min-w-[180px]">
            {filtradas.map(s => (
              <button key={s} type="button" onMouseDown={() => add(s)}
                className="w-full text-left px-3 py-2 text-xs hover:bg-rose-50 hover:text-rose-700">{s}</button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default function ExpedientePaciente() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<'info' | 'consultas' | 'planes' | 'seguimiento'>('info')
  const [editando, setEditando] = useState(false)
  const [form, setForm] = useState<any>({})
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  // Restricciones
  const [newIngrediente, setNewIngrediente] = useState('')
  const [newMotivo, setNewMotivo] = useState('')

  const reload = () => {
    if (!id) return
    api.getPaciente(id).then(d => {
      setData(d)
      setLoading(false)
      const p = d.paciente
      setForm({
        nombre: p.nombre ?? '', apellidos: p.apellidos ?? '',
        email: p.email ?? '', telefono: p.telefono ?? '',
        fechaNacimiento: p.fecha_nacimiento ?? '', sexo: p.sexo ?? '',
        peso: p.peso ?? '', altura: p.altura ?? '',
        objetivo: p.objetivo ?? '', enfermedades: p.enfermedades ?? '',
        alergias: p.alergias ?? '', preferencias: p.preferencias ?? '',
        alimentosEvitar: p.alimentosEvitar ?? '',
        horarioComidas: p.horarioComidas ?? '',
        nivelActividad: p.nivelActividad ?? '', notas: p.notas ?? '',
      })
    }).catch(() => navigate('/pacientes'))
  }

  useEffect(() => { reload() }, [id])

  async function guardar() {
    if (!id) return
    setSaving(true)
    await api.updatePaciente(id, {
      ...form,
      peso: form.peso ? parseFloat(form.peso) : undefined,
      altura: form.altura ? parseFloat(form.altura) : undefined,
    })
    setSaving(false)
    setEditando(false)
    reload()
  }

  async function eliminar() {
    if (!id || !window.confirm('¿Eliminar paciente permanentemente?')) return
    setDeleting(true)
    await api.deletePaciente(id)
    navigate('/pacientes')
  }

  async function agregarRestriccion() {
    if (!id || !newIngrediente.trim()) return
    await api.addRestriccion(id, { ingrediente: newIngrediente.trim(), motivo: newMotivo.trim() || undefined })
    setNewIngrediente('')
    setNewMotivo('')
    reload()
  }

  async function quitarRestriccion(rid: string) {
    if (!id) return
    await api.deleteRestriccion(id, rid)
    reload()
  }

  function setF(k: string, v: any) { setForm((f: any) => ({ ...f, [k]: v })) }

  if (loading) return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="h-8 w-48 bg-gray-100 rounded-lg animate-pulse mb-4" />
      <div className="card h-64 animate-pulse" />
    </div>
  )

  const p = data?.paciente
  const restricciones: any[] = data?.restricciones ?? []
  const imc = p?.imc ? parseFloat(p.imc).toFixed(1) : null

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-start justify-between mb-5">
        <div className="flex items-center gap-3">
          <Link to="/pacientes" className="text-gray-400 hover:text-gray-600">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{p.nombre} {p.apellidos}</h1>
            <p className="text-sm text-gray-400 mt-0.5">
              {p.edad ? `${p.edad} años` : ''}
              {imc ? ` · IMC ${imc} (${imcLabel(parseFloat(imc))})` : ''}
              {p.nivelActividad ? ` · ${p.nivelActividad}` : ''}
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Link to={`/ia?pacienteId=${id}`} className="btn-primary flex items-center gap-1.5 text-sm">
            <Cpu className="w-4 h-4" /> Generar plan IA
          </Link>
          {editando ? (
            <>
              <button onClick={guardar} disabled={saving}
                className="btn-secondary flex items-center gap-1.5 text-green-600 hover:bg-green-50">
                <Save className="w-4 h-4" /> {saving ? 'Guardando...' : 'Guardar'}
              </button>
              <button onClick={() => { setEditando(false); reload() }} className="btn-secondary">
                Cancelar
              </button>
            </>
          ) : (
            <>
              <button onClick={() => setEditando(true)} className="btn-secondary flex items-center gap-1.5">
                <Edit2 className="w-4 h-4" /> Editar
              </button>
              <button onClick={eliminar} disabled={deleting}
                className="btn-secondary text-red-500 hover:bg-red-50 flex items-center gap-1.5">
                <Trash2 className="w-4 h-4" />
              </button>
            </>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-5 border-b border-gray-100 overflow-x-auto">
        {(['info', 'consultas', 'planes', 'seguimiento'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-2.5 text-sm font-medium transition-colors border-b-2 -mb-px whitespace-nowrap ${
              tab === t ? 'border-rose-500 text-rose-600' : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}>
            {t === 'info' ? 'Expediente'
              : t === 'consultas' ? `Consultas (${data?.consultas?.length ?? 0})`
              : t === 'planes' ? `Planes (${data?.planes?.length ?? 0})`
              : 'Seguimiento'}
          </button>
        ))}
      </div>

      {tab === 'info' && (
        <div className="space-y-4">
          {/* Datos personales */}
          <div className="card p-5 space-y-4">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Datos personales</h2>
            {editando ? (
              <div className="grid grid-cols-2 gap-4">
                <Field label="Nombre"><input className="input" value={form.nombre} onChange={e => setF('nombre', e.target.value)} /></Field>
                <Field label="Apellidos"><input className="input" value={form.apellidos} onChange={e => setF('apellidos', e.target.value)} /></Field>
                <Field label="Correo"><input type="email" className="input" value={form.email} onChange={e => setF('email', e.target.value)} /></Field>
                <Field label="Teléfono"><input className="input" value={form.telefono} onChange={e => setF('telefono', e.target.value)} /></Field>
                <Field label="Nacimiento"><input type="date" className="input" value={form.fechaNacimiento} onChange={e => setF('fechaNacimiento', e.target.value)} /></Field>
                <Field label="Sexo">
                  <select className="input" value={form.sexo} onChange={e => setF('sexo', e.target.value)}>
                    <option value="">—</option>
                    <option value="F">Femenino</option>
                    <option value="M">Masculino</option>
                  </select>
                </Field>
              </div>
            ) : (
              <div className="grid grid-cols-2 gap-x-8 gap-y-2">
                <Row label="Correo" value={p.email} />
                <Row label="Teléfono" value={p.telefono} />
                <Row label="Nacimiento" value={p.fecha_nacimiento} />
                <Row label="Sexo" value={p.sexo === 'M' ? 'Masculino' : p.sexo === 'F' ? 'Femenino' : ''} />
              </div>
            )}
          </div>

          {/* Antropometría */}
          <div className="card p-5 space-y-4">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Antropometría</h2>
            {editando ? (
              <div className="grid grid-cols-3 gap-4">
                <Field label="Peso (kg)"><input type="number" step="0.1" className="input" value={form.peso} onChange={e => setF('peso', e.target.value)} /></Field>
                <Field label="Altura (cm)"><input type="number" step="0.1" className="input" value={form.altura} onChange={e => setF('altura', e.target.value)} /></Field>
                <Field label="Actividad">
                  <select className="input" value={form.nivelActividad} onChange={e => setF('nivelActividad', e.target.value)}>
                    <option value="">—</option>
                    {NIVELES_ACTIVIDAD.map(n => <option key={n.val} value={n.val}>{n.label}</option>)}
                  </select>
                </Field>
              </div>
            ) : (
              <div className="grid grid-cols-3 gap-4">
                <StatBox label="Peso" value={p.peso ? `${p.peso} kg` : '—'} />
                <StatBox label="Altura" value={p.altura ? `${p.altura} cm` : '—'} />
                <StatBox label="IMC" value={imc ? `${imc} · ${imcLabel(parseFloat(imc))}` : '—'} highlight={imc ? imcColor(parseFloat(imc)) : ''} />
              </div>
            )}
          </div>

          {/* Condición clínica */}
          <div className="card p-5 space-y-4">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Condición clínica</h2>
            {editando ? (
              <div className="space-y-4">
                <Field label="Objetivo">
                  <input className="input" value={form.objetivo} onChange={e => setF('objetivo', e.target.value)} />
                </Field>
                <Field label="Enfermedades / condiciones">
                  <TagInput value={form.enfermedades} onChange={v => setF('enfermedades', v)} sugerencias={SUGERENCIAS_ENFERMEDADES} placeholder="Diabetes tipo 2..." />
                </Field>
                <Field label="Alergias e intolerancias">
                  <TagInput value={form.alergias} onChange={v => setF('alergias', v)} sugerencias={SUGERENCIAS_ALERGIAS} placeholder="Lactosa, gluten..." />
                </Field>
              </div>
            ) : (
              <div className="space-y-2">
                <Row label="Objetivo" value={p.objetivo} />
                <div className="flex gap-2 mt-1 flex-wrap">
                  {p.enfermedades?.split(',').map((e: string) => e.trim()).filter(Boolean).map((e: string) => (
                    <span key={e} className="text-xs bg-amber-50 text-amber-700 px-2 py-0.5 rounded-full">{e}</span>
                  ))}
                </div>
                <div className="flex gap-2 mt-1 flex-wrap">
                  {p.alergias?.split(',').map((a: string) => a.trim()).filter(Boolean).map((a: string) => (
                    <span key={a} className="text-xs bg-red-50 text-red-600 px-2 py-0.5 rounded-full">⚠ {a}</span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Hábitos alimenticios */}
          <div className="card p-5 space-y-4">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Hábitos alimenticios</h2>
            {editando ? (
              <div className="space-y-4">
                <Field label="No come / no le gusta">
                  <TagInput value={form.alimentosEvitar} onChange={v => setF('alimentosEvitar', v)} sugerencias={SUGERENCIAS_EVITAR} placeholder="Cerdo, mariscos..." />
                </Field>
                <Field label="Alimentos preferidos">
                  <TagInput value={form.preferencias} onChange={v => setF('preferencias', v)} sugerencias={SUGERENCIAS_PREFERENCIAS} placeholder="Frijoles, verduras..." />
                </Field>
                <Field label="Horario de comidas">
                  <input className="input" value={form.horarioComidas} onChange={e => setF('horarioComidas', e.target.value)} placeholder="Desayuna 8am, come 2pm..." />
                </Field>
                <Field label="Notas adicionales">
                  <textarea className="input resize-none" rows={2} value={form.notas} onChange={e => setF('notas', e.target.value)} />
                </Field>
              </div>
            ) : (
              <div className="space-y-3">
                {p.alimentosEvitar && (
                  <div>
                    <p className="text-xs text-gray-400 mb-1">No come</p>
                    <div className="flex flex-wrap gap-1.5">
                      {p.alimentosEvitar.split(',').map((a: string) => a.trim()).filter(Boolean).map((a: string) => (
                        <span key={a} className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full line-through">{a}</span>
                      ))}
                    </div>
                  </div>
                )}
                {p.preferencias && (
                  <div>
                    <p className="text-xs text-gray-400 mb-1">Prefiere</p>
                    <div className="flex flex-wrap gap-1.5">
                      {p.preferencias.split(',').map((a: string) => a.trim()).filter(Boolean).map((a: string) => (
                        <span key={a} className="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded-full">✓ {a}</span>
                      ))}
                    </div>
                  </div>
                )}
                {p.horarioComidas && <Row label="Horario" value={p.horarioComidas} />}
                {p.notas && <Row label="Notas" value={p.notas} />}
              </div>
            )}
          </div>

          {/* Restricciones de ingredientes */}
          <div className="card p-5 space-y-3">
            <h2 className="text-xs font-semibold uppercase tracking-wide text-gray-400">Ingredientes restringidos</h2>
            <p className="text-xs text-gray-400">Los platillos con estos ingredientes no aparecerán en el plan compatible.</p>
            <div className="flex gap-2">
              <input className="input flex-1 text-sm" placeholder="Ingrediente o grupo (ej: gluten, cerdo)"
                value={newIngrediente} onChange={e => setNewIngrediente(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && agregarRestriccion()} />
              <input className="input w-36 text-sm" placeholder="Motivo (opcional)"
                value={newMotivo} onChange={e => setNewMotivo(e.target.value)} />
              <button onClick={agregarRestriccion} className="btn-primary px-3">
                <Plus className="w-4 h-4" />
              </button>
            </div>
            {restricciones.length === 0 ? (
              <p className="text-xs text-gray-300 italic">Sin restricciones registradas</p>
            ) : (
              <div className="flex flex-wrap gap-2">
                {restricciones.map((r: any) => (
                  <span key={r.id} className="inline-flex items-center gap-1.5 text-xs bg-orange-50 text-orange-700 px-2.5 py-1 rounded-full">
                    {r.ingrediente}
                    {r.motivo && <span className="text-orange-400">· {r.motivo}</span>}
                    <button onClick={() => quitarRestriccion(r.id)} className="hover:text-orange-900 ml-0.5">
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {tab === 'consultas' && (
        <div className="space-y-3">
          <div className="flex justify-end">
            <Link to={`/pacientes/${id}/consultas/nueva`} className="btn-primary flex items-center gap-2">
              <Plus className="w-4 h-4" /> Nueva consulta
            </Link>
          </div>
          {!data?.consultas?.length ? (
            <div className="text-center py-16 text-gray-400">Sin consultas registradas</div>
          ) : data.consultas.map((c: any) => (
            <div key={c.id} className="card p-4">
              <div className="flex items-start justify-between">
                <div>
                  <p className="font-medium text-gray-800">{c.fecha}</p>
                  {c.motivo && <p className="text-sm text-gray-500 mt-0.5">{c.motivo}</p>}
                </div>
                <div className="text-right text-sm">
                  {c.peso && <p className="font-medium text-gray-700">{c.peso} kg</p>}
                  {c.imc && <p className="text-xs text-gray-400">IMC {c.imc.toFixed ? c.imc.toFixed(1) : c.imc}</p>}
                  {c.glucosa && <p className="text-xs text-amber-600">Glucosa {c.glucosa}</p>}
                </div>
              </div>
              {c.observaciones && <p className="text-sm text-gray-500 mt-2 border-t border-gray-50 pt-2">{c.observaciones}</p>}
            </div>
          ))}
        </div>
      )}

      {tab === 'planes' && (
        <div className="space-y-3">
          <div className="flex justify-end">
            <Link to={`/ia?pacienteId=${id}`} className="btn-primary flex items-center gap-2">
              <Cpu className="w-4 h-4" /> Generar plan con IA
            </Link>
          </div>
          {!data?.planes?.length ? (
            <div className="text-center py-16 text-gray-400">Sin planes asignados</div>
          ) : data.planes.map((pl: any) => (
            <Link key={pl.id} to={`/planes/${pl.id}`}
              className="card p-4 flex items-center justify-between hover:shadow-md transition-shadow">
              <div>
                <p className="font-medium text-gray-800">{pl.nombre}</p>
                <p className="text-sm text-gray-400">Desde {pl.fecha_inicio}</p>
              </div>
              {pl.generado_ia && <span className="badge-rose text-xs">IA</span>}
            </Link>
          ))}
        </div>
      )}

      {tab === 'seguimiento' && id && (
        <SeguimientoTab pacienteId={id} alturaInicial={data?.paciente?.altura} />
      )}
    </div>
  )
}

// ── Seguimiento / Evolución ─────────────────────────────────────
function SeguimientoTab({ pacienteId, alturaInicial }: { pacienteId: string; alturaInicial?: number }) {
  const [list, setList]       = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving]   = useState(false)

  // Form state (flat)
  const today = new Date().toISOString().split('T')[0]
  const [fecha, setFecha]         = useState(today)
  const [peso, setPeso]           = useState('')
  const [cintura, setCintura]     = useState('')
  const [cadera, setCadera]       = useState('')
  const [brazo, setBrazo]         = useState('')
  const [muslo, setMuslo]         = useState('')
  const [grasa, setGrasa]         = useState('')
  const [sistolica, setSistolica] = useState('')
  const [diastolica, setDiastolica] = useState('')
  const [glucosa, setGlucosa]     = useState('')
  const [notas, setNotas]         = useState('')
  const [showExtra, setShowExtra] = useState(false)

  const reload = () => {
    api.listSeguimientos(pacienteId)
      .then(d => { setList(d); setLoading(false) })
      .catch(() => setLoading(false))
  }

  useEffect(() => { reload() }, [pacienteId])

  async function guardar() {
    if (!peso && !cintura && !cadera && !brazo && !muslo && !grasa && !sistolica && !glucosa && !notas) return
    setSaving(true)
    const body: any = { fecha }
    if (peso)      body.peso           = parseFloat(peso)
    if (cintura)   body.cinturaCm      = parseFloat(cintura)
    if (cadera)    body.caderaCm       = parseFloat(cadera)
    if (brazo)     body.brazoCm        = parseFloat(brazo)
    if (muslo)     body.musloCm        = parseFloat(muslo)
    if (grasa)     body.grasaCorporal  = parseFloat(grasa)
    if (sistolica) body.tensionSistolica  = parseInt(sistolica)
    if (diastolica) body.tensionDiastolica = parseInt(diastolica)
    if (glucosa)   body.glucosa        = parseFloat(glucosa)
    if (notas)     body.notas          = notas
    await api.addSeguimiento(pacienteId, body)
    // Reset form
    setPeso(''); setCintura(''); setCadera(''); setBrazo(''); setMuslo('')
    setGrasa(''); setSistolica(''); setDiastolica(''); setGlucosa(''); setNotas('')
    setSaving(false)
    reload()
  }

  // Weight chart data (chronological order)
  const pesoEntries = [...list]
    .filter(e => e.peso != null)
    .sort((a, b) => a.fecha.localeCompare(b.fecha))

  // IMC helper
  const calcIMC = (p: number) =>
    alturaInicial ? p / ((alturaInicial / 100) ** 2) : null

  if (loading) return <div className="py-16 text-center text-gray-300">Cargando...</div>

  return (
    <div className="space-y-5">
      {/* ── Quick add form ── */}
      <div className="card p-5 space-y-4">
        <h3 className="font-semibold text-gray-800 text-sm">Nuevo registro</h3>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div>
            <label className="text-xs text-gray-500 mb-1 block">Fecha</label>
            <input type="date" className="input text-sm" value={fecha} onChange={e => setFecha(e.target.value)} />
          </div>
          <div>
            <label className="text-xs text-gray-500 mb-1 block">Peso (kg)</label>
            <input type="number" step="0.1" placeholder="70.5" className="input text-sm"
              value={peso} onChange={e => setPeso(e.target.value)} />
          </div>
          <div>
            <label className="text-xs text-gray-500 mb-1 block">Cintura (cm)</label>
            <input type="number" step="0.5" placeholder="90" className="input text-sm"
              value={cintura} onChange={e => setCintura(e.target.value)} />
          </div>
          <div>
            <label className="text-xs text-gray-500 mb-1 block">Cadera (cm)</label>
            <input type="number" step="0.5" placeholder="100" className="input text-sm"
              value={cadera} onChange={e => setCadera(e.target.value)} />
          </div>
        </div>

        {showExtra && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Brazo (cm)</label>
              <input type="number" step="0.5" className="input text-sm"
                value={brazo} onChange={e => setBrazo(e.target.value)} />
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Muslo (cm)</label>
              <input type="number" step="0.5" className="input text-sm"
                value={muslo} onChange={e => setMuslo(e.target.value)} />
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">% Grasa corporal</label>
              <input type="number" step="0.1" className="input text-sm"
                value={grasa} onChange={e => setGrasa(e.target.value)} />
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Glucosa (mg/dL)</label>
              <input type="number" step="1" className="input text-sm"
                value={glucosa} onChange={e => setGlucosa(e.target.value)} />
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Tensión sistólica</label>
              <input type="number" step="1" placeholder="120" className="input text-sm"
                value={sistolica} onChange={e => setSistolica(e.target.value)} />
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Tensión diastólica</label>
              <input type="number" step="1" placeholder="80" className="input text-sm"
                value={diastolica} onChange={e => setDiastolica(e.target.value)} />
            </div>
            <div className="col-span-2">
              <label className="text-xs text-gray-500 mb-1 block">Notas</label>
              <input className="input text-sm" placeholder="Observaciones..."
                value={notas} onChange={e => setNotas(e.target.value)} />
            </div>
          </div>
        )}

        <div className="flex items-center justify-between">
          <button
            onClick={() => setShowExtra(v => !v)}
            className="text-xs text-gray-400 hover:text-gray-600 underline underline-offset-2">
            {showExtra ? 'Menos campos' : '+ Medidas adicionales y signos vitales'}
          </button>
          <button onClick={guardar} disabled={saving}
            className="btn-primary text-sm px-5 disabled:opacity-50">
            {saving ? 'Guardando...' : 'Guardar registro'}
          </button>
        </div>
      </div>

      {/* ── Weight chart ── */}
      {pesoEntries.length >= 2 && (
        <div className="card p-5">
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-semibold text-gray-800 text-sm">Evolución del peso</h3>
            <span className="text-xs text-gray-400">
              {pesoEntries[0].peso} → {pesoEntries[pesoEntries.length - 1].peso} kg
              {' '}
              <span className={
                pesoEntries[pesoEntries.length - 1].peso < pesoEntries[0].peso
                  ? 'text-green-600 font-semibold'
                  : 'text-red-500 font-semibold'
              }>
                ({((pesoEntries[pesoEntries.length - 1].peso - pesoEntries[0].peso) >= 0 ? '+' : '')}
                {(pesoEntries[pesoEntries.length - 1].peso - pesoEntries[0].peso).toFixed(1)} kg)
              </span>
            </span>
          </div>
          <PesoChart entries={pesoEntries} />
        </div>
      )}

      {/* ── History ── */}
      {list.length === 0 ? (
        <div className="text-center py-16 text-gray-300 text-sm">
          Sin registros aun. Agrega el primero arriba.
        </div>
      ) : (
        <div className="space-y-2">
          {list.map((s, i) => {
            const prev = list[i + 1]
            const deltaPeso = s.peso != null && prev?.peso != null ? s.peso - prev.peso : null
            const imc = s.peso ? calcIMC(s.peso) : null
            return (
              <div key={s.id} className="card px-5 py-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 flex-wrap mb-2">
                      <span className="text-sm font-semibold text-gray-800">{s.fecha}</span>
                      {s.peso != null && (
                        <span className="text-sm font-bold text-rose-600 tabular-nums">{s.peso} kg</span>
                      )}
                      {deltaPeso != null && (
                        <span className={`flex items-center gap-0.5 text-xs font-semibold ${
                          deltaPeso < -0.1 ? 'text-green-600' : deltaPeso > 0.1 ? 'text-red-500' : 'text-gray-400'
                        }`}>
                          {deltaPeso < -0.1 ? <TrendingDown className="w-3.5 h-3.5" />
                            : deltaPeso > 0.1 ? <TrendingUp className="w-3.5 h-3.5" />
                            : <Minus className="w-3.5 h-3.5" />}
                          {deltaPeso >= 0 ? '+' : ''}{deltaPeso.toFixed(1)} kg
                        </span>
                      )}
                      {imc != null && (
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                          imc < 18.5 ? 'bg-blue-50 text-blue-700'
                          : imc < 25  ? 'bg-green-50 text-green-700'
                          : imc < 30  ? 'bg-amber-50 text-amber-700'
                          : 'bg-red-50 text-red-700'
                        }`}>
                          IMC {imc.toFixed(1)}
                        </span>
                      )}
                    </div>
                    {/* Medidas */}
                    <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500">
                      {s.cinturaCm   && <span>Cintura {s.cinturaCm} cm</span>}
                      {s.caderaCm    && <span>Cadera {s.caderaCm} cm</span>}
                      {s.brazoCm     && <span>Brazo {s.brazoCm} cm</span>}
                      {s.musloCm     && <span>Muslo {s.musloCm} cm</span>}
                      {s.grasaCorporal && <span>Grasa {s.grasaCorporal}%</span>}
                      {s.tensionSistolica && (
                        <span>T/A {s.tensionSistolica}/{s.tensionDiastolica} mmHg</span>
                      )}
                      {s.glucosa     && <span>Glucosa {s.glucosa} mg/dL</span>}
                    </div>
                    {s.notas && (
                      <p className="text-xs text-gray-400 mt-1.5 italic">{s.notas}</p>
                    )}
                  </div>
                  <button
                    onClick={async () => {
                      await api.deleteSeguimiento(pacienteId, s.id)
                      reload()
                    }}
                    className="shrink-0 p-1.5 rounded hover:bg-red-50 text-gray-300 hover:text-red-400 transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

function PesoChart({ entries }: { entries: { fecha: string; peso: number }[] }) {
  const W = 500, H = 90, PX = 8, PY = 10
  const pesos = entries.map(e => e.peso)
  const minP  = Math.min(...pesos)
  const maxP  = Math.max(...pesos)
  const range = Math.max(maxP - minP, 1)
  const n     = entries.length

  const cx = (i: number) => PX + (i / (n - 1)) * (W - PX * 2)
  const cy = (p: number) => PY + ((maxP - p) / range) * (H - PY * 2)

  const d = entries.map((e, i) => `${i === 0 ? 'M' : 'L'}${cx(i).toFixed(1)},${cy(e.peso).toFixed(1)}`).join(' ')

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="w-full" style={{ height: 90 }}>
      <path d={d} fill="none" stroke="#f43f5e" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      {entries.map((e, i) => (
        <g key={i}>
          <circle cx={cx(i)} cy={cy(e.peso)} r="4" fill="white" stroke="#f43f5e" strokeWidth="2" />
          {(i === 0 || i === n - 1) && (
            <text x={cx(i)} y={cy(e.peso) - 6} fontSize="9" fill="#6b7280" textAnchor="middle">
              {e.peso}
            </text>
          )}
        </g>
      ))}
      <text x={PX} y={H - 2} fontSize="8" fill="#d1d5db">{entries[0].fecha}</text>
      <text x={W - PX} y={H - 2} fontSize="8" fill="#d1d5db" textAnchor="end">
        {entries[n - 1].fecha}
      </text>
    </svg>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-500 mb-1">{label}</label>
      {children}
    </div>
  )
}

function Row({ label, value }: { label: string; value?: string | null }) {
  if (!value) return null
  return (
    <div className="flex gap-3">
      <span className="text-sm text-gray-400 w-28 flex-shrink-0">{label}</span>
      <span className="text-sm text-gray-800">{value}</span>
    </div>
  )
}

function StatBox({ label, value, highlight }: { label: string; value: string; highlight?: string }) {
  return (
    <div className={`rounded-xl p-3 text-center ${highlight || 'bg-gray-50'}`}>
      <p className="text-lg font-semibold text-gray-800">{value}</p>
      <p className="text-xs text-gray-400 mt-0.5">{label}</p>
    </div>
  )
}

function imcLabel(imc: number) {
  if (imc < 18.5) return 'Bajo peso'
  if (imc < 25) return 'Normal'
  if (imc < 30) return 'Sobrepeso'
  if (imc < 35) return 'Obesidad I'
  return 'Obesidad II'
}

function imcColor(imc: number) {
  if (imc < 18.5) return 'bg-blue-50'
  if (imc < 25) return 'bg-green-50'
  if (imc < 30) return 'bg-amber-50'
  return 'bg-red-50'
}
