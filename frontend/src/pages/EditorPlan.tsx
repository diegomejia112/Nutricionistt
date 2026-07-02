import { useEffect, useState, useRef, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Search, X, Plus, ChevronDown, ChevronUp, Check, Pencil, Trash2, Printer } from 'lucide-react'
import { api } from '../lib/api'
import {
  calcEquivAlimentos, calcEquivTiempos, totalKcalEquiv,
  SMAE_META, SMAE_ORDER, type SmaeKey,
} from '../lib/smae'

const DIAS = ['', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo']
const UNIDADES = ['g', 'ml', 'taza', 'cucharada', 'cucharadita', 'pieza', 'porción', 'vaso', 'rebanada', 'piezas', 'gramos']

type EditState = {
  id: string
  tipo: string
  nombre: string
  cantidad: string
  unidad: string
  kcal: string
  proteinas: string
  carbohidratos: string
  grasas: string
  caloriasBase: number
}

export default function EditorPlan() {
  const { id } = useParams<{ id: string }>()
  const [plan, setPlan] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [diaAbierto, setDiaAbierto] = useState(1)
  const [editando, setEditando] = useState<EditState | null>(null)
  const [saving, setSaving] = useState(false)
  const [printAll, setPrintAll] = useState(false)

  // Add modal state (flat, avoids nested object updates)
  const [addOpen, setAddOpen] = useState(false)
  const [addTiempoId, setAddTiempoId] = useState('')
  const [addTab, setAddTab] = useState<'alimentos' | 'platillos' | 'libre'>('alimentos')
  const [addQuery, setAddQuery] = useState('')
  const [addResultados, setAddResultados] = useState<any[]>([])
  const [addSeleccionado, setAddSeleccionado] = useState<any>(null)
  const [addCantidad, setAddCantidad] = useState('100')
  const [addUnidad, setAddUnidad] = useState('g')
  const [libreNombre, setLibreNombre] = useState('')
  const [libreKcal, setLibreKcal] = useState('0')
  const [libreProteinas, setLibreProteinas] = useState('0')
  const [libreCarbos, setLibreCarbos] = useState('0')
  const [libreGrasas, setLibreGrasas] = useState('0')
  const [libreCantidad, setLibreCantidad] = useState('1')
  const [libreUnidad, setLibreUnidad] = useState('porción')

  const debounce = useRef<ReturnType<typeof setTimeout>>()

  function handlePrint() {
    setPrintAll(true)
    setTimeout(() => {
      window.print()
      setPrintAll(false)
    }, 150)
  }

  const reload = useCallback(() => {
    if (!id) return
    api.getPlan(id).then(d => { setPlan(d); setLoading(false) }).catch(() => {})
  }, [id])

  useEffect(() => { reload() }, [reload])

  // Calorie helpers
  function kcalItem(ap: any): number {
    if (ap.tipo === 'ia') return Math.round(ap.calorias ?? 0)
    return Math.round((ap.calorias ?? 0) * ap.cantidad / 100)
  }

  function totalCalDia(dia: any): number {
    return Math.round(dia.tiempos?.reduce((s: number, t: any) =>
      s + (t.alimentos ?? []).reduce((ss: number, a: any) => ss + kcalItem(a), 0), 0) ?? 0)
  }

  function kcalEditPreview(): number {
    if (!editando) return 0
    if (editando.tipo === 'ia') return Math.round(parseFloat(editando.kcal) || 0)
    return Math.round(editando.caloriasBase * (parseFloat(editando.cantidad) || 0) / 100)
  }

  // Inline edit
  function startEdit(ap: any) {
    setEditando({
      id: ap.id,
      tipo: ap.tipo,
      nombre: ap.nombre ?? '',
      cantidad: String(ap.cantidad ?? 1),
      unidad: ap.unidad ?? 'g',
      kcal: String(ap.calorias ?? 0),
      proteinas: String(ap.proteinas ?? 0),
      carbohidratos: String(ap.carbohidratos ?? 0),
      grasas: String(ap.grasas ?? 0),
      caloriasBase: ap.calorias ?? 0,
    })
  }

  async function saveEdit() {
    if (!editando || !id) return
    setSaving(true)
    try {
      const body: Record<string, unknown> = {
        cantidad: parseFloat(editando.cantidad) || 1,
        unidad: editando.unidad,
      }
      if (editando.tipo === 'ia') {
        body.nombreLibre = editando.nombre
        body.calorias = parseFloat(editando.kcal) || 0
        body.proteinas = parseFloat(editando.proteinas) || 0
        body.carbohidratos = parseFloat(editando.carbohidratos) || 0
        body.grasas = parseFloat(editando.grasas) || 0
      }
      await api.updateAlimentoPlan(id, editando.id, body)
      setEditando(null)
      reload()
    } catch { /* ignore */ }
    setSaving(false)
  }

  async function quitar(alimentoId: string) {
    if (!id) return
    if (editando?.id === alimentoId) setEditando(null)
    await api.removeAlimentoPlan(id, alimentoId)
    reload()
  }

  // Add modal
  function openAdd(tiempoId: string) {
    setAddTiempoId(tiempoId)
    setAddTab('alimentos')
    setAddQuery('')
    setAddResultados([])
    setAddSeleccionado(null)
    setAddCantidad('100')
    setAddUnidad('g')
    setLibreNombre('')
    setLibreKcal('0')
    setLibreProteinas('0')
    setLibreCarbos('0')
    setLibreGrasas('0')
    setLibreCantidad('1')
    setLibreUnidad('porción')
    setAddOpen(true)
  }

  function handleAddTab(tab: 'alimentos' | 'platillos' | 'libre') {
    setAddTab(tab)
    setAddQuery('')
    setAddResultados([])
    setAddSeleccionado(null)
  }

  function handleAddQuery(q: string) {
    setAddQuery(q)
    setAddSeleccionado(null)
    clearTimeout(debounce.current)
    if (!q.trim()) { setAddResultados([]); return }
    debounce.current = setTimeout(async () => {
      try {
        const r = addTab === 'alimentos'
          ? await api.getAlimentos(q)
          : await api.getPlatillos({ q })
        setAddResultados(r)
      } catch { /* ignore */ }
    }, 300)
  }

  function selectItem(item: any) {
    setAddSeleccionado(item)
    setAddCantidad(addTab === 'alimentos' ? '100' : '1')
    setAddUnidad(addTab === 'alimentos' ? 'g' : 'porción')
  }

  const addKcalPreview = addSeleccionado
    ? Math.round((addSeleccionado.calorias ?? 0) * (parseFloat(addCantidad) || 0) / 100)
    : 0

  async function guardarAdd() {
    if (!id) return
    setSaving(true)
    try {
      if (addTab === 'libre') {
        if (!libreNombre.trim()) { setSaving(false); return }
        await api.addAlimentoPlan(id, {
          tiempoId: addTiempoId,
          cantidad: parseFloat(libreCantidad) || 1,
          unidad: libreUnidad,
          nombreLibre: libreNombre,
          calorias: parseFloat(libreKcal) || 0,
          proteinas: parseFloat(libreProteinas) || 0,
          carbohidratos: parseFloat(libreCarbos) || 0,
          grasas: parseFloat(libreGrasas) || 0,
        })
      } else if (addSeleccionado) {
        const body: Record<string, unknown> = {
          tiempoId: addTiempoId,
          cantidad: parseFloat(addCantidad) || 100,
          unidad: addUnidad,
        }
        if (addTab === 'alimentos') body.alimentoId = addSeleccionado.id
        else body.platilloId = addSeleccionado.id
        await api.addAlimentoPlan(id, body)
      }
      setAddOpen(false)
      reload()
    } catch { /* ignore */ }
    setSaving(false)
  }

  const canGuardarAdd =
    (addTab !== 'libre' && !!addSeleccionado) ||
    (addTab === 'libre' && libreNombre.trim().length > 0)

  if (loading) return (
    <div className="p-6 max-w-4xl mx-auto space-y-3">
      {[1, 2, 3].map(i => <div key={i} className="h-16 bg-gray-100 rounded-xl animate-pulse" />)}
    </div>
  )
  if (!plan) return null

  return (
    <div className="p-6 max-w-4xl mx-auto">
      {/* Print styles */}
      <style>{`
        @media print {
          body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
          .no-print { display: none !important; }
          .print-expand { display: block !important; }
        }
      `}</style>

      <div className="flex items-center gap-3 mb-6 no-print">
        <Link to="/planes" className="text-gray-400 hover:text-gray-600">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div className="flex-1 min-w-0">
          <h1 className="text-2xl font-bold text-gray-900 truncate">{plan.nombre}</h1>
          {plan.caloriasObj && (
            <p className="text-sm text-gray-400">Objetivo: {plan.caloriasObj} kcal/día</p>
          )}
        </div>
        <button onClick={handlePrint}
          className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 border border-gray-200 hover:border-gray-300 px-3 py-1.5 rounded-lg transition-colors">
          <Printer className="w-4 h-4" />
          Exportar
        </button>
      </div>
      {/* Print header — visible only when printing */}
      <div className="hidden print:block mb-4">
        <h1 className="text-xl font-bold">{plan.nombre}</h1>
        {plan.caloriasObj && <p className="text-sm text-gray-500">Objetivo: {plan.caloriasObj} kcal/día</p>}
      </div>

      {/* ── Add modal ── */}
      {addOpen && (
        <div className="fixed inset-0 bg-black/30 z-50 flex items-start justify-center pt-16 px-4"
          onClick={() => setAddOpen(false)}>
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md max-h-[76vh] flex flex-col"
            onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
              <h3 className="font-semibold text-gray-800">Agregar alimento</h3>
              <button onClick={() => setAddOpen(false)}>
                <X className="w-5 h-5 text-gray-400" />
              </button>
            </div>

            <div className="flex gap-1 px-4 py-3 border-b border-gray-100">
              {(['alimentos', 'platillos', 'libre'] as const).map(t => (
                <button key={t} onClick={() => handleAddTab(t)}
                  className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                    addTab === t ? 'bg-rose-50 text-rose-600' : 'text-gray-500 hover:bg-gray-50'
                  }`}>
                  {t === 'libre' ? 'Libre / IA' : t.charAt(0).toUpperCase() + t.slice(1)}
                </button>
              ))}
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {addTab !== 'libre' ? (
                !addSeleccionado ? (
                  <>
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                      <input className="input pl-9" autoFocus
                        placeholder={addTab === 'alimentos' ? 'Buscar alimento...' : 'Buscar platillo...'}
                        value={addQuery} onChange={e => handleAddQuery(e.target.value)} />
                    </div>
                    <div className="space-y-0.5">
                      {addResultados.map((item: any) => (
                        <button key={item.id} onClick={() => selectItem(item)}
                          className="w-full text-left px-3 py-2.5 rounded-xl hover:bg-rose-50 transition-colors">
                          <p className="text-sm font-medium text-gray-800">{item.nombre}</p>
                          <p className="text-xs text-gray-400">
                            {item.calorias ?? '?'} kcal/100g
                            {item.categoria && ` · ${item.categoria}`}
                          </p>
                        </button>
                      ))}
                      {!addQuery && (
                        <p className="text-sm text-center text-gray-400 py-8">Escribe para buscar...</p>
                      )}
                      {addQuery && !addResultados.length && (
                        <p className="text-sm text-center text-gray-400 py-8">Sin resultados</p>
                      )}
                    </div>
                  </>
                ) : (
                  <div className="space-y-4">
                    <div className="bg-gray-50 rounded-xl p-3">
                      <p className="font-semibold text-gray-800">{addSeleccionado.nombre}</p>
                      <p className="text-xs text-gray-400 mt-0.5">
                        {addSeleccionado.calorias} kcal/100g
                        {addSeleccionado.categoria && ` · ${addSeleccionado.categoria}`}
                      </p>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="text-xs text-gray-500 mb-1 block">Cantidad</label>
                        <input type="number" step="0.1" className="input"
                          value={addCantidad} onChange={e => setAddCantidad(e.target.value)} />
                      </div>
                      <div>
                        <label className="text-xs text-gray-500 mb-1 block">Unidad</label>
                        <input list="unidades-add" className="input"
                          value={addUnidad} onChange={e => setAddUnidad(e.target.value)} />
                        <datalist id="unidades-add">
                          {UNIDADES.map(u => <option key={u} value={u} />)}
                        </datalist>
                      </div>
                    </div>
                    <div className="bg-rose-50 rounded-lg px-4 py-2.5 flex items-center justify-between">
                      <span className="text-sm text-gray-500">Calorías calculadas:</span>
                      <span className="text-sm font-bold text-rose-600 tabular-nums">{addKcalPreview} kcal</span>
                    </div>
                    <button
                      onClick={() => { setAddSeleccionado(null); setAddQuery(''); setAddResultados([]) }}
                      className="text-xs text-gray-400 hover:text-gray-600">
                      ← Cambiar {addTab === 'alimentos' ? 'alimento' : 'platillo'}
                    </button>
                  </div>
                )
              ) : (
                <div className="space-y-3">
                  <div>
                    <label className="text-xs text-gray-500 mb-1 block">Nombre *</label>
                    <input className="input" placeholder="Ej: Agua de jamaica, Sopa de fideos..."
                      value={libreNombre} onChange={e => setLibreNombre(e.target.value)} />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-xs text-gray-500 mb-1 block">Cantidad</label>
                      <input type="number" step="0.1" className="input"
                        value={libreCantidad} onChange={e => setLibreCantidad(e.target.value)} />
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 mb-1 block">Unidad</label>
                      <input list="unidades-libre" className="input"
                        value={libreUnidad} onChange={e => setLibreUnidad(e.target.value)} />
                      <datalist id="unidades-libre">
                        {UNIDADES.map(u => <option key={u} value={u} />)}
                      </datalist>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-xs text-gray-500 mb-1 block">Calorías (kcal)</label>
                      <input type="number" className="input"
                        value={libreKcal} onChange={e => setLibreKcal(e.target.value)} />
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 mb-1 block">Proteínas (g)</label>
                      <input type="number" className="input"
                        value={libreProteinas} onChange={e => setLibreProteinas(e.target.value)} />
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 mb-1 block">Carbohidratos (g)</label>
                      <input type="number" className="input"
                        value={libreCarbos} onChange={e => setLibreCarbos(e.target.value)} />
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 mb-1 block">Grasas (g)</label>
                      <input type="number" className="input"
                        value={libreGrasas} onChange={e => setLibreGrasas(e.target.value)} />
                    </div>
                  </div>
                </div>
              )}
            </div>

            <div className="px-4 py-4 border-t border-gray-100">
              <button onClick={guardarAdd} disabled={saving || !canGuardarAdd}
                className="btn-primary w-full disabled:opacity-40 disabled:cursor-not-allowed">
                {saving ? 'Guardando...' : 'Agregar al plan'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Plan days ── */}
      <div className="space-y-2">
        {plan.dias?.map((dia: any) => {
          const abierto = printAll || diaAbierto === dia.diaSemana
          const calDia = totalCalDia(dia)
          return (
            <div key={dia.id} className="card overflow-hidden">
              <button
                onClick={() => setDiaAbierto(abierto && !printAll ? 0 : dia.diaSemana)}
                className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors no-print">
                <div className="flex items-center gap-3">
                  <span className="font-semibold text-gray-800">{DIAS[dia.diaSemana]}</span>
                  {calDia > 0 && (
                    <span className="text-xs font-medium text-gray-500 bg-gray-100 px-2.5 py-0.5 rounded-full tabular-nums">
                      {calDia} kcal
                    </span>
                  )}
                </div>
                {abierto
                  ? <ChevronUp className="w-4 h-4 text-gray-400" />
                  : <ChevronDown className="w-4 h-4 text-gray-400" />}
              </button>
              {/* Print-only day header */}
              <div className="hidden print:flex items-center gap-3 px-5 py-3 border-b border-gray-200">
                <span className="font-bold text-gray-800">{DIAS[dia.diaSemana]}</span>
                {calDia > 0 && <span className="text-sm text-gray-500">{calDia} kcal</span>}
              </div>

              {abierto && (
                <div className="border-t border-gray-100">
                  {dia.tiempos?.map((t: any) => {
                    const calT = Math.round((t.alimentos ?? []).reduce((s: number, a: any) => s + kcalItem(a), 0))
                    return (
                      <div key={t.id} className="border-b border-gray-50 last:border-0">
                        <div className="px-5 pt-3 pb-1 flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <h4 className="text-sm font-semibold text-gray-700">{t.nombre}</h4>
                            {calT > 0 && (
                              <span className="text-xs text-gray-400 tabular-nums">{calT} kcal</span>
                            )}
                          </div>
                          <button onClick={() => openAdd(t.id)}
                            className="no-print text-xs text-rose-500 hover:text-rose-600 font-medium flex items-center gap-1 transition-colors">
                            <Plus className="w-3.5 h-3.5" /> Agregar
                          </button>
                        </div>
                        <div className="px-5 pb-2 space-y-0.5">
                          {(!t.alimentos || t.alimentos.length === 0) && (
                            <p className="text-xs text-gray-300 italic py-1.5">
                              Sin alimentos — usa Agregar para añadir
                            </p>
                          )}
                          {t.alimentos?.map((ap: any) => (
                            editando?.id === ap.id ? (
                              <EditRow
                                key={ap.id}
                                editando={editando!}
                                kcalPreview={kcalEditPreview()}
                                saving={saving}
                                onChange={(k, v) => setEditando(e => e ? { ...e, [k]: v } : null)}
                                onSave={saveEdit}
                                onCancel={() => setEditando(null)}
                                onDelete={() => quitar(ap.id)}
                              />
                            ) : (
                              <FoodRow
                                key={ap.id}
                                ap={ap}
                                kcal={kcalItem(ap)}
                                onClick={() => startEdit(ap)}
                                onDelete={() => quitar(ap.id)}
                              />
                            )
                          ))}
                        </div>
                        {/* Equivalentes por tiempo de comida */}
                        <EquivalentesTiempo alimentos={t.alimentos ?? []} />
                      </div>
                    )
                  })}
                  {/* Equivalentes SMAE del día completo */}
                  <EquivalentesDia tiempos={dia.tiempos ?? []} caloriasObj={plan.caloriasObj} />
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ── SMAE equivalents chips per meal ─────────────────────────────
function EquivalentesTiempo({ alimentos }: { alimentos: any[] }) {
  const equiv = calcEquivAlimentos(alimentos)
  const items = SMAE_ORDER.filter(k => equiv[k] >= 0.1)
  if (items.length === 0) return null
  return (
    <div className="flex flex-wrap gap-1 px-5 pb-2.5 pt-0.5">
      {items.map(k => {
        const m = SMAE_META[k]
        return (
          <span key={k}
            className={`inline-flex items-center gap-0.5 text-[10px] font-semibold px-1.5 py-0.5 rounded-full ${m.bg} ${m.color}`}>
            {m.label} {equiv[k].toFixed(1)}
          </span>
        )
      })}
    </div>
  )
}

// ── SMAE equivalents full table per day ─────────────────────────
function EquivalentesDia({ tiempos, caloriasObj }: { tiempos: any[]; caloriasObj?: number | null }) {
  const [desglose, setDesglose] = useState(false)

  // Per-day totals
  const totalEquiv = calcEquivTiempos(tiempos)
  const totalKcal  = Math.round(totalKcalEquiv(totalEquiv))
  const activeKeys = SMAE_ORDER.filter(k => totalEquiv[k] >= 0.05)
  const maxEquiv   = Math.max(...activeKeys.map(k => totalEquiv[k]), 0.01)

  // Per-meal breakdown
  const tiempoEquivs = tiempos.map(t => ({
    nombre: t.nombre,
    equiv:  calcEquivAlimentos(t.alimentos ?? []),
    kcal:   Math.round(totalKcalEquiv(calcEquivAlimentos(t.alimentos ?? []))),
  }))

  if (activeKeys.length === 0) return null

  return (
    <div className="border-t border-gray-100 bg-gray-50/60">
      <div className="px-5 py-3">
        <div className="flex items-center justify-between mb-3">
          <h5 className="text-xs font-bold text-gray-600 uppercase tracking-wider">
            Equivalentes SMAE
          </h5>
          <div className="flex items-center gap-3">
            {caloriasObj && (
              <span className={`text-xs font-medium tabular-nums ${
                Math.abs(totalKcal - caloriasObj) < caloriasObj * 0.05
                  ? 'text-green-600'
                  : Math.abs(totalKcal - caloriasObj) < caloriasObj * 0.15
                    ? 'text-amber-600'
                    : 'text-red-500'
              }`}>
                {totalKcal} / {caloriasObj} kcal
              </span>
            )}
            <button
              onClick={() => setDesglose(d => !d)}
              className="text-[10px] text-gray-400 hover:text-gray-600 underline underline-offset-2">
              {desglose ? 'Ver resumen' : 'Ver por comida'}
            </button>
          </div>
        </div>

        {!desglose ? (
          /* ── Resumen del día ── */
          <div className="space-y-2">
            {activeKeys.map(k => {
              const m   = SMAE_META[k]
              const eq  = totalEquiv[k]
              const pct = (eq / maxEquiv) * 100
              return (
                <div key={k} className="flex items-center gap-3">
                  <span className={`inline-flex items-center text-[11px] font-semibold px-2 py-0.5 rounded-full w-36 shrink-0 ${m.bg} ${m.color}`}>
                    {m.label}
                  </span>
                  <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
                    <div className={`h-full rounded-full ${m.bar}`} style={{ width: `${pct}%` }} />
                  </div>
                  <span className="text-xs font-mono font-bold text-gray-700 w-10 text-right tabular-nums">
                    {eq.toFixed(1)}
                  </span>
                  <span className="text-xs text-gray-400 w-14 text-right tabular-nums">
                    {Math.round(eq * m.kcalEq)} kcal
                  </span>
                </div>
              )
            })}
            <div className="flex justify-end pt-1 border-t border-gray-200 mt-2">
              <span className="text-xs text-gray-500">
                Total estimado:{' '}
                <span className="font-bold text-gray-700 tabular-nums">{totalKcal} kcal</span>
              </span>
            </div>
          </div>
        ) : (
          /* ── Desglose por comida ── */
          <div className="overflow-x-auto -mx-1">
            <table className="w-full text-[11px]">
              <thead>
                <tr className="text-gray-400 border-b border-gray-200">
                  <th className="text-left py-1 pr-3 font-medium whitespace-nowrap">Comida</th>
                  {activeKeys.map(k => (
                    <th key={k} className="text-center py-1 px-1 font-medium whitespace-nowrap">
                      {SMAE_META[k].label.slice(0, 3)}
                    </th>
                  ))}
                  <th className="text-right py-1 pl-2 font-medium whitespace-nowrap">kcal</th>
                </tr>
              </thead>
              <tbody>
                {tiempoEquivs.map((t, i) => (
                  <tr key={i} className="border-b border-gray-100 last:border-0">
                    <td className="py-1.5 pr-3 font-medium text-gray-700 whitespace-nowrap">{t.nombre}</td>
                    {activeKeys.map(k => (
                      <td key={k} className="text-center px-1 tabular-nums text-gray-600">
                        {t.equiv[k] >= 0.1 ? t.equiv[k].toFixed(1) : <span className="text-gray-300">—</span>}
                      </td>
                    ))}
                    <td className="text-right pl-2 tabular-nums text-gray-500 font-medium">{t.kcal}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr className="border-t-2 border-gray-300 font-bold">
                  <td className="pt-1.5 text-gray-700">Total día</td>
                  {activeKeys.map(k => (
                    <td key={k} className="text-center px-1 pt-1.5 tabular-nums text-gray-800">
                      {totalEquiv[k].toFixed(1)}
                    </td>
                  ))}
                  <td className="text-right pl-2 pt-1.5 tabular-nums text-gray-800">{totalKcal}</td>
                </tr>
                <tr>
                  <td colSpan={activeKeys.length + 2} className="pt-1 pb-0.5">
                    <div className="flex gap-2 flex-wrap pt-1">
                      {activeKeys.map(k => (
                        <span key={k} className={`text-[10px] px-1.5 py-0.5 rounded-full ${SMAE_META[k].bg} ${SMAE_META[k].color}`}>
                          {SMAE_META[k].label}
                        </span>
                      ))}
                    </div>
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

function FoodRow({ ap, kcal, onClick, onDelete }: {
  ap: any; kcal: number; onClick: () => void; onDelete: () => void
}) {
  return (
    <div
      className="group flex items-center gap-2 py-1.5 rounded-lg hover:bg-gray-50 cursor-pointer transition-colors px-1 -mx-1"
      onClick={onClick}
    >
      <div className="flex-1 min-w-0 flex items-center gap-2 flex-wrap">
        <span className="text-sm text-gray-800">{ap.nombre}</span>
        <span className="text-xs text-gray-400">{ap.cantidad} {ap.unidad}</span>
        {kcal > 0 && (
          <span className="text-xs text-gray-400 tabular-nums">{kcal} kcal</span>
        )}
        {ap.tipo === 'ia' && (
          <span className="text-[10px] font-semibold text-rose-300 uppercase tracking-wide">IA</span>
        )}
      </div>
      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
        <button
          className="p-1.5 rounded hover:bg-rose-50 text-gray-400 hover:text-rose-500 transition-colors"
          onClick={e => { e.stopPropagation(); onClick() }}>
          <Pencil className="w-3.5 h-3.5" />
        </button>
        <button
          className="p-1.5 rounded hover:bg-red-50 text-gray-400 hover:text-red-500 transition-colors"
          onClick={e => { e.stopPropagation(); onDelete() }}>
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  )
}

function EditRow({ editando, kcalPreview, saving, onChange, onSave, onCancel, onDelete }: {
  editando: EditState
  kcalPreview: number
  saving: boolean
  onChange: (k: string, v: string) => void
  onSave: () => void
  onCancel: () => void
  onDelete: () => void
}) {
  const isIA = editando.tipo === 'ia'
  return (
    <div className="bg-rose-50 border border-rose-100 rounded-xl p-3 my-2 space-y-3">
      {isIA ? (
        <input className="input text-sm py-1.5" placeholder="Nombre del alimento"
          value={editando.nombre} onChange={e => onChange('nombre', e.target.value)} />
      ) : (
        <p className="text-sm font-semibold text-gray-800 px-0.5">{editando.nombre}</p>
      )}

      <div className="flex gap-2">
        <div className="flex-1">
          <label className="text-[11px] text-gray-500 mb-0.5 block">Cantidad</label>
          <input type="number" step="0.1" className="input text-sm py-1.5"
            value={editando.cantidad} onChange={e => onChange('cantidad', e.target.value)} />
        </div>
        <div className="flex-1">
          <label className="text-[11px] text-gray-500 mb-0.5 block">Unidad</label>
          <input list="unidades-edit" className="input text-sm py-1.5"
            value={editando.unidad} onChange={e => onChange('unidad', e.target.value)} />
          <datalist id="unidades-edit">
            {UNIDADES.map(u => <option key={u} value={u} />)}
          </datalist>
        </div>
        <div className="flex-1">
          <label className="text-[11px] text-gray-500 mb-0.5 block">
            {isIA ? 'kcal (total)' : 'kcal (calc.)'}
          </label>
          {isIA ? (
            <input type="number" className="input text-sm py-1.5"
              value={editando.kcal} onChange={e => onChange('kcal', e.target.value)} />
          ) : (
            <div className="input text-sm py-1.5 bg-white text-gray-600 tabular-nums select-none">
              {kcalPreview}
            </div>
          )}
        </div>
      </div>

      {isIA && (
        <div className="grid grid-cols-3 gap-2">
          {(['proteinas', 'carbohidratos', 'grasas'] as const).map(k => (
            <div key={k}>
              <label className="text-[11px] text-gray-500 mb-0.5 block capitalize">
                {k === 'carbohidratos' ? 'Carbos g' : k === 'proteinas' ? 'Proteínas g' : 'Grasas g'}
              </label>
              <input type="number" className="input text-sm py-1.5"
                value={editando[k]} onChange={e => onChange(k, e.target.value)} />
            </div>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between pt-1">
        <button onClick={onDelete}
          className="flex items-center gap-1 text-xs text-red-400 hover:text-red-600 transition-colors">
          <Trash2 className="w-3 h-3" /> Eliminar
        </button>
        <div className="flex gap-2">
          <button onClick={onCancel} className="btn-secondary text-xs px-3 py-1.5">
            Cancelar
          </button>
          <button onClick={onSave} disabled={saving}
            className="btn-primary text-xs px-3 py-1.5 flex items-center gap-1 disabled:opacity-50">
            <Check className="w-3 h-3" />
            {saving ? 'Guardando...' : 'Guardar'}
          </button>
        </div>
      </div>
    </div>
  )
}
