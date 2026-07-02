import { useEffect, useState, useRef, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Search, X, Plus, ChevronDown, ChevronUp, Check, Pencil, Trash2, Printer, FileSpreadsheet, FileText } from 'lucide-react'
import { api } from '../lib/api'
import {
  calcEquivAlimentos, calcEquivTiempos, totalKcalEquiv,
  SMAE_META, SMAE_ORDER, CAT_MAP, type SmaeKey,
} from '../lib/smae'

const DIAS = ['', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo']
const UNIDADES = ['g', 'ml', 'taza', 'cucharada', 'cucharadita', 'pieza', 'porción', 'vaso', 'rebanada', 'piezas', 'gramos']
// Reparto estándar del objetivo calórico diario entre tiempos de comida
const PCT_TIEMPO: Record<string, number> = {
  'Desayuno': 0.25, 'Colacion AM': 0.10, 'Comida': 0.30, 'Colacion PM': 0.10, 'Cena': 0.25,
}

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
  const [exportando, setExportando] = useState(false)

  // Add modal state (flat, avoids nested object updates)
  const [addOpen, setAddOpen] = useState(false)
  const [addTiempoId, setAddTiempoId] = useState('')
  const [addTiempoNombre, setAddTiempoNombre] = useState('')
  const [addTiempoUsado, setAddTiempoUsado] = useState(0)
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

  function slugPlan(): string {
    return (plan?.nombre ?? 'plan')
      .toLowerCase()
      .normalize('NFD').replace(/[̀-ͯ]/g, '')
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9\-]/g, '')
  }

  function descargarBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  // Tiempos de comida únicos del plan, ordenados por el orden mínimo encontrado
  function tiemposUnicosDelPlan(): { nombre: string; orden: number }[] {
    const seen = new Map<string, number>()
    ;(plan?.dias ?? []).forEach((dia: any) => {
      ;(dia.tiempos ?? []).forEach((t: any) => {
        const existente = seen.get(t.nombre)
        if (existente === undefined || t.orden < existente) seen.set(t.nombre, t.orden)
      })
    })
    return [...seen.entries()].map(([nombre, orden]) => ({ nombre, orden })).sort((a, b) => a.orden - b.orden)
  }

  async function exportarExcel() {
    if (!plan) return
    setExportando(true)
    try {
      const ExcelJS = (await import('exceljs')).default
      const workbook = new ExcelJS.Workbook()
      const sheet = workbook.addWorksheet('Plan semanal')

      sheet.mergeCells('A1:H1')
      const titleCell = sheet.getCell('A1')
      titleCell.value = `${plan.nombre}${plan.caloriasObj ? ' — Objetivo: ' + plan.caloriasObj + ' kcal/día' : ''}`
      titleCell.font = { bold: true, size: 14 }
      titleCell.alignment = { horizontal: 'center' }

      const headerRow = sheet.getRow(2)
      headerRow.getCell(1).value = ''
      for (let col = 2; col <= 8; col++) {
        const cell = headerRow.getCell(col)
        cell.value = DIAS[col - 1]
        cell.font = { bold: true, color: { argb: 'FFFFFFFF' } }
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF43F5E' } }
        cell.alignment = { horizontal: 'center', vertical: 'middle' }
      }

      const tiemposUnicos = tiemposUnicosDelPlan()
      let rowIndex = 3
      tiemposUnicos.forEach(tiempo => {
        const row = sheet.getRow(rowIndex)
        row.getCell(1).value = tiempo.nombre
        row.getCell(1).font = { bold: true }
        for (let diaSemana = 1; diaSemana <= 7; diaSemana++) {
          const dia = plan.dias.find((d: any) => d.diaSemana === diaSemana)
          const tiempoDia = dia?.tiempos.find((t: any) => t.nombre === tiempo.nombre)
          const alimentos = tiempoDia?.alimentos ?? []
          if (alimentos.length === 0) { row.getCell(diaSemana + 1).value = ''; continue }
          row.getCell(diaSemana + 1).value = alimentos
            .map((a: any) => `${a.nombre} — ${a.cantidad}${a.unidad} (${kcalItem(a)} kcal)`)
            .join('\n')
          row.getCell(diaSemana + 1).alignment = { wrapText: true, vertical: 'top' }
        }
        rowIndex++
      })

      const totalRow = sheet.getRow(rowIndex)
      totalRow.getCell(1).value = 'Total kcal'
      totalRow.getCell(1).font = { bold: true }
      for (let diaSemana = 1; diaSemana <= 7; diaSemana++) {
        const dia = plan.dias.find((d: any) => d.diaSemana === diaSemana)
        totalRow.getCell(diaSemana + 1).value = dia ? totalCalDia(dia) : 0
        totalRow.getCell(diaSemana + 1).font = { bold: true }
      }

      sheet.getColumn(1).width = 14
      for (let col = 2; col <= 8; col++) sheet.getColumn(col).width = 26

      const buffer = await workbook.xlsx.writeBuffer()
      descargarBlob(
        new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }),
        `plan-${slugPlan()}.xlsx`
      )
    } finally {
      setExportando(false)
    }
  }

  function exportarTexto() {
    if (!plan) return
    const tiemposUnicos = tiemposUnicosDelPlan()
    const lineas: string[] = []
    lineas.push(plan.nombre)
    if (plan.caloriasObj) lineas.push(`Objetivo: ${plan.caloriasObj} kcal/día`)
    lineas.push('')

    for (let diaSemana = 1; diaSemana <= 7; diaSemana++) {
      const dia = plan.dias.find((d: any) => d.diaSemana === diaSemana)
      lineas.push(`== ${DIAS[diaSemana]} ==`)
      if (!dia) { lineas.push('(sin datos)', ''); continue }
      tiemposUnicos.forEach(tiempo => {
        const tiempoDia = dia.tiempos.find((t: any) => t.nombre === tiempo.nombre)
        const alimentos = tiempoDia?.alimentos ?? []
        if (alimentos.length === 0) return
        lineas.push(`${tiempo.nombre}:`)
        alimentos.forEach((a: any) => {
          lineas.push(`  - ${a.nombre} — ${a.cantidad}${a.unidad} (${kcalItem(a)} kcal)`)
        })
      })
      lineas.push(`Total: ${totalCalDia(dia)} kcal`, '')
    }

    descargarBlob(new Blob([lineas.join('\n')], { type: 'text/plain;charset=utf-8' }), `plan-${slugPlan()}.txt`)
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
  function openAdd(tiempo: any) {
    setAddTiempoId(tiempo.id)
    setAddTiempoNombre(tiempo.nombre)
    setAddTiempoUsado(Math.round((tiempo.alimentos ?? []).reduce((s: number, a: any) => s + kcalItem(a), 0)))
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

  // Espacio restante de kcal en el tiempo de comida que se está editando,
  // repartiendo plan.caloriasObj entre tiempos con PCT_TIEMPO
  const presupuestoTiempo = plan?.caloriasObj
    ? Math.round(plan.caloriasObj * (PCT_TIEMPO[addTiempoNombre] ?? 0.2))
    : null
  const espacioRestante = presupuestoTiempo != null ? presupuestoTiempo - addTiempoUsado : null

  function previewKcalItem(item: any): number {
    const cant = addTab === 'alimentos' ? 100 : 1
    return Math.round((item.calorias ?? 0) * cant / 100)
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
    <div className="p-6 max-w-6xl mx-auto">
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
        <button onClick={exportarExcel} disabled={exportando}
          className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 border border-gray-200 hover:border-gray-300 px-3 py-1.5 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
          <FileSpreadsheet className="w-4 h-4" />
          {exportando ? 'Exportando...' : 'Excel'}
        </button>
        <button onClick={exportarTexto}
          className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 border border-gray-200 hover:border-gray-300 px-3 py-1.5 rounded-lg transition-colors">
          <FileText className="w-4 h-4" />
          Nota (.txt)
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

            {espacioRestante != null && (
              <div className={`px-5 py-2 text-xs font-medium border-b border-gray-100 ${
                espacioRestante < 0 ? 'bg-red-50 text-red-600' : 'bg-gray-50 text-gray-500'
              }`}>
                {espacioRestante >= 0
                  ? `Espacio restante en ${addTiempoNombre}: ${espacioRestante} kcal`
                  : `Ya pasaste el presupuesto de ${addTiempoNombre} por ${Math.abs(espacioRestante)} kcal`}
              </div>
            )}
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
                      {addResultados.map((item: any) => {
                        const previewKcal = previewKcalItem(item)
                        const cabe = espacioRestante == null ? null : previewKcal <= espacioRestante
                        return (
                          <button key={item.id} onClick={() => selectItem(item)}
                            className="w-full text-left px-3 py-2.5 rounded-xl hover:bg-rose-50 transition-colors flex items-center justify-between gap-2">
                            <div className="min-w-0">
                              <p className="text-sm font-medium text-gray-800 truncate">{item.nombre}</p>
                              <p className="text-xs text-gray-400">
                                {item.calorias ?? '?'} kcal/100g
                                {item.categoria && ` · ${item.categoria}`}
                              </p>
                            </div>
                            {cabe != null && (
                              <span className={`shrink-0 text-[10px] font-semibold px-2 py-0.5 rounded-full ${
                                cabe ? 'bg-green-100 text-green-700' : 'bg-amber-100 text-amber-700'
                              }`}>
                                {cabe ? 'Cabe' : `+${previewKcal - espacioRestante!} kcal`}
                              </span>
                            )}
                          </button>
                        )
                      })}
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

      <div className="flex gap-6 items-start">
      <div className="flex-1 min-w-0">
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
                          <button onClick={() => openAdd(t)}
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
      <PanelSMAEReferencia />
      </div>
    </div>
  )
}

// ── Catálogo de referencia SMAE (alimentos activos agrupados por categoría) ──
function PanelSMAEReferencia() {
  const [porCategoria, setPorCategoria] = useState<Record<SmaeKey, any[]> | null>(null)
  const [abierta, setAbierta] = useState<SmaeKey | null>(null)

  useEffect(() => {
    api.getAlimentos('', '', false, 2000).then(alimentos => {
      const grupos = {} as Record<SmaeKey, any[]>
      for (const k of SMAE_ORDER) grupos[k] = []
      for (const a of alimentos) {
        const key = CAT_MAP[a.categoria]
        if (key) grupos[key].push(a)
      }
      setPorCategoria(grupos)
    }).catch(() => setPorCategoria({} as Record<SmaeKey, any[]>))
  }, [])

  if (!porCategoria) return null

  // Aproximación: gramos del alimento que equivalen a 1 equivalente SMAE de
  // su categoría, estimado por calorías (100g de referencia). No sustituye
  // la tabla SMAE oficial, es una guía rápida.
  function gramosPorEquivalente(a: any, k: SmaeKey): string {
    if (!a.calorias || a.calorias <= 0) return ''
    const g = (SMAE_META[k].kcalEq / a.calorias) * 100
    return `${Math.round(g / 5) * 5}g ≈ 1 equiv`
  }

  return (
    <aside className="hidden lg:block w-80 shrink-0 no-print sticky top-6">
      <div className="card p-4">
        <h3 className="text-xs font-bold text-gray-600 uppercase tracking-wider mb-3">
          Catálogo SMAE
        </h3>
        <div className="space-y-1.5 max-h-[75vh] overflow-y-auto pr-1">
          {SMAE_ORDER.map(k => {
            const m = SMAE_META[k]
            const alimentos = porCategoria[k] ?? []
            const open = abierta === k
            return (
              <div key={k}>
                <button onClick={() => setAbierta(open ? null : k)}
                  className={`w-full flex items-center justify-between text-xs font-semibold px-2 py-1.5 rounded-lg ${m.bg} ${m.color}`}>
                  <span>{m.label}</span>
                  <span className="tabular-nums">{alimentos.length}</span>
                </button>
                {open && (
                  <ul className="pl-2 py-1.5 space-y-1">
                    {alimentos.length === 0 ? (
                      <li className="text-[11px] text-gray-300 italic px-1">Sin alimentos</li>
                    ) : alimentos.map(a => (
                      <li key={a.id} className="flex items-baseline justify-between gap-2 text-[11px] text-gray-600 px-1 py-0.5">
                        <span className="truncate" title={a.nombre}>{a.nombre}</span>
                        <span className="text-gray-400 tabular-nums shrink-0">{gramosPorEquivalente(a, k)}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )
          })}
        </div>
      </div>
    </aside>
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

  // Per-day totals — siempre las 8 categorías SMAE, aunque estén en 0
  const totalEquiv = calcEquivTiempos(tiempos)
  const totalKcal  = Math.round(totalKcalEquiv(totalEquiv))
  const hasAny     = SMAE_ORDER.some(k => totalEquiv[k] >= 0.05)
  const maxEquiv   = Math.max(...SMAE_ORDER.map(k => totalEquiv[k]), 0.01)

  // Per-meal breakdown
  const tiempoEquivs = tiempos.map(t => {
    const equiv = calcEquivAlimentos(t.alimentos ?? [])
    return {
      nombre: t.nombre,
      equiv,
      kcal: Math.round(totalKcalEquiv(equiv)),
      maxEquiv: Math.max(...SMAE_ORDER.map(k => equiv[k]), 0.01),
    }
  })

  if (!hasAny) return null

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
            {SMAE_ORDER.map(k => {
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
          /* ── Desglose por comida (barras) ── */
          <div className="space-y-4">
            {tiempoEquivs.map((t, i) => (
              <div key={i}>
                <div className="flex items-center justify-between mb-1.5">
                  <h6 className="text-xs font-bold text-gray-700">{t.nombre}</h6>
                  <span className="text-xs text-gray-400 tabular-nums">{t.kcal} kcal</span>
                </div>
                <div className="space-y-1">
                  {SMAE_ORDER.map(k => {
                    const m   = SMAE_META[k]
                    const eq  = t.equiv[k]
                    const pct = (eq / t.maxEquiv) * 100
                    return (
                      <div key={k} className="flex items-center gap-3">
                        <span className={`inline-flex items-center text-[11px] font-semibold px-2 py-0.5 rounded-full w-36 shrink-0 ${m.bg} ${m.color}`}>
                          {m.label}
                        </span>
                        <div className="flex-1 h-1.5 bg-gray-200 rounded-full overflow-hidden">
                          <div className={`h-full rounded-full ${m.bar}`} style={{ width: `${pct}%` }} />
                        </div>
                        <span className="text-xs font-mono font-bold text-gray-700 w-10 text-right tabular-nums">
                          {eq.toFixed(1)}
                        </span>
                      </div>
                    )
                  })}
                </div>
              </div>
            ))}
            <div className="flex justify-end pt-2 border-t border-gray-200">
              <span className="text-xs text-gray-500">
                Total día:{' '}
                <span className="font-bold text-gray-700 tabular-nums">{totalKcal} kcal</span>
              </span>
            </div>
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
