import { useEffect, useState, useRef } from 'react'
import { Search, Apple, ChevronDown } from 'lucide-react'
import { api } from '../lib/api'
import DishImage from '../components/ui/DishImage'

const CATEGORIAS_ALIM = ['', 'Frutas', 'Verduras', 'Lacteos', 'Carnes', 'Cereales', 'Leguminosas', 'Aceites', 'Azucares', 'Bebidas']
const CATEGORIAS_PLAT = ['', 'Desayuno', 'Comida', 'Cena', 'Colacion', 'Postre', 'Bebida']
const ATRIBUTO_KEY_MAP: Record<string, string> = {
  alto_proteina: 'altoProteina',
  bajo_grasa: 'bajoGrasa',
  vegetariano: 'vegetariano',
}

type Tab = 'alimentos' | 'platillos'

const FLAG_BADGES: { key: string; label: string; color: string }[] = [
  { key: 'aptoDiabetes',      label: 'Diabetes',    color: 'bg-blue-100 text-blue-700' },
  { key: 'aptoHipertension',  label: 'Hipert.',     color: 'bg-purple-100 text-purple-700' },
  { key: 'aptoSobrepeso',     label: 'Sobrepeso',   color: 'bg-green-100 text-green-700' },
  { key: 'altoProteina',      label: 'Alto P',      color: 'bg-orange-100 text-orange-700' },
  { key: 'bajoGrasa',         label: 'Bajo G',      color: 'bg-teal-100 text-teal-700' },
  { key: 'vegetariano',       label: 'Veg',         color: 'bg-lime-100 text-lime-700' },
]

export default function Alimentos() {
  const [tab, setTab] = useState<Tab>('alimentos')
  const [q, setQ] = useState('')
  const [categoria, setCategoria] = useState('')
  const [casos, setCasos] = useState<string[]>([])
  const [casosCatalogo, setCasosCatalogo] = useState<{ id: string; slug: string; nombre: string }[]>([])
  const [showCasosDropdown, setShowCasosDropdown] = useState(false)
  const [atributo, setAtributo] = useState('')
  const [items, setItems] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [expanded, setExpanded] = useState<string | null>(null)
  const [variantes, setVariantes] = useState<Record<string, any[]>>({})
  const debounce = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => { api.getCasos().then(setCasosCatalogo).catch(() => {}) }, [])

  const load = (query: string, cat: string, c: string[], t: Tab, attr: string) => {
    setLoading(true)
    const fn = t === 'alimentos'
      ? api.getAlimentos(query, cat)
      : api.getPlatillos({ q: query, categoria: cat, casos: c })
    fn.then(d => {
      const filtered = attr && t === 'platillos'
        ? d.filter((p: any) => p[ATRIBUTO_KEY_MAP[attr]])
        : d
      setItems(filtered)
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  useEffect(() => { load('', '', [], tab, '') }, [tab])

  function handleSearch(val: string) {
    setQ(val)
    clearTimeout(debounce.current)
    debounce.current = setTimeout(() => load(val, categoria, casos, tab, atributo), 300)
  }

  function handleCat(val: string) {
    setCategoria(val)
    load(q, val, casos, tab, atributo)
  }

  function handleCasos(seleccionados: string[]) {
    setCasos(seleccionados)
    load(q, categoria, seleccionados, tab, atributo)
  }

  function handleAtributo(val: string) {
    setAtributo(val)
    load(q, categoria, casos, tab, val)
  }

  function changeTab(t: Tab) {
    setTab(t)
    setQ('')
    setCategoria('')
    setCasos([])
    setAtributo('')
    setItems([])
    setExpanded(null)
  }

  async function toggleExpand(id: string) {
    if (expanded === id) { setExpanded(null); return }
    setExpanded(id)
    if (!variantes[id]) {
      const vs = await api.getVariantes(id).catch(() => [])
      setVariantes(prev => ({ ...prev, [id]: vs }))
    }
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Base de datos nutricional</h1>
        <p className="text-sm text-gray-500 mt-0.5">Alimentos y platillos mexicanos</p>
      </div>

      <div className="flex gap-1 mb-5 border-b border-gray-100">
        {(['alimentos', 'platillos'] as const).map(t => (
          <button key={t} onClick={() => changeTab(t)}
            className={`px-4 py-2.5 text-sm font-medium capitalize transition-colors border-b-2 -mb-px ${
              tab === t ? 'border-rose-500 text-rose-600' : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}>
            {t === 'alimentos' ? 'Alimentos' : 'Platillos'}
          </button>
        ))}
      </div>

      <div className="flex gap-3 mb-5 flex-wrap">
        <div className="relative flex-1 min-w-[180px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input type="text" className="input pl-9" placeholder="Buscar..." value={q} onChange={e => handleSearch(e.target.value)} />
        </div>
        <select className="input w-44" value={categoria} onChange={e => handleCat(e.target.value)}>
          {(tab === 'alimentos' ? CATEGORIAS_ALIM : CATEGORIAS_PLAT).map(c => (
            <option key={c} value={c}>{c || 'Todas las categorías'}</option>
          ))}
        </select>
        {tab === 'platillos' && (
          <div className="relative">
            <button type="button" onClick={() => setShowCasosDropdown(v => !v)}
              className="input w-52 text-left flex items-center justify-between text-sm">
              <span>{casos.length === 0 ? 'Todas las condiciones' : `Condiciones (${casos.length})`}</span>
              <ChevronDown className="w-4 h-4 text-gray-400" />
            </button>
            {showCasosDropdown && (
              <div className="absolute z-10 mt-1 w-64 bg-white border border-gray-200 rounded-xl shadow-lg max-h-60 overflow-y-auto">
                {casosCatalogo.map(c => (
                  <label key={c.id} className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer text-sm">
                    <input type="checkbox" checked={casos.includes(c.slug)}
                      onChange={() => {
                        const next = casos.includes(c.slug)
                          ? casos.filter(s => s !== c.slug)
                          : [...casos, c.slug]
                        handleCasos(next)
                      }}
                      className="rounded border-gray-300 text-rose-500 focus:ring-rose-400" />
                    {c.nombre}
                  </label>
                ))}
              </div>
            )}
          </div>
        )}
        {tab === 'platillos' && (
          <select className="input w-44" value={atributo} onChange={e => handleAtributo(e.target.value)}>
            <option value="">Atributo nutricional</option>
            <option value="alto_proteina">Alto en proteína</option>
            <option value="bajo_grasa">Bajo en grasa</option>
            <option value="vegetariano">Vegetariano</option>
          </select>
        )}
      </div>

      {loading ? (
        <div className="grid grid-cols-3 gap-4">
          {[1,2,3,4,5,6].map(i => <div key={i} className="h-24 bg-gray-100 rounded-xl animate-pulse" />)}
        </div>
      ) : items.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <Apple className="w-10 h-10 mx-auto mb-3 opacity-30" />
          <p>Sin resultados</p>
        </div>
      ) : tab === 'alimentos' ? (
        <div className="grid grid-cols-1 gap-2">
          {items.map(a => (
            <div key={a.id} className="card px-4 py-3 flex items-center gap-4">
              <div className="flex-1 min-w-0">
                <p className="font-medium text-gray-800 truncate">{a.nombre}</p>
                {a.categoria && <span className="badge-gray text-xs mt-0.5">{a.categoria}</span>}
              </div>
              <div className="flex gap-4 text-sm text-gray-500 flex-shrink-0 text-right">
                <Macro label="kcal" val={a.calorias} bold />
                <Macro label="P" val={a.proteinas} unit="g" />
                <Macro label="C" val={a.carbohidratos} unit="g" />
                <Macro label="G" val={a.grasas} unit="g" />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-3">
          {items.map(p => (
            <div key={p.id} className="card overflow-hidden">
              <div className="flex gap-3">
                <div className="w-24 h-24 flex-shrink-0">
                  <DishImage id={p.id} url={p.imagen_url} categoria={p.categoria} nombre={p.nombre} />
                </div>
                <div className="py-3 pr-3 flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-2">
                    <p className="font-medium text-gray-800">{p.nombre}</p>
                    <button onClick={() => toggleExpand(p.id)}
                      className="text-xs text-rose-500 hover:text-rose-700 flex-shrink-0 mt-0.5">
                      {expanded === p.id ? 'Cerrar' : 'Variantes'}
                    </button>
                  </div>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {p.categoria && <span className="badge-rose text-xs">{p.categoria}</span>}
                    {p.region && <span className="badge-gray text-xs">{p.region}</span>}
                    {FLAG_BADGES.filter(f => p[f.key]).map(f => (
                      <span key={f.key} className={`text-xs px-1.5 py-0.5 rounded-full font-medium ${f.color}`}>{f.label}</span>
                    ))}
                  </div>
                  <div className="flex gap-3 mt-2 text-xs text-gray-400">
                    <span>{p.calorias} kcal</span>
                    <span>P {p.proteinas}g</span>
                    <span>C {p.carbohidratos}g</span>
                    <span>G {p.grasas}g</span>
                    {p.fibra > 0 && <span>F {p.fibra}g</span>}
                    {p.porcionDesc && <span>· {p.porcionDesc}</span>}
                  </div>
                </div>
              </div>

              {expanded === p.id && (
                <div className="border-t border-gray-100 px-4 py-3 bg-gray-50">
                  {!variantes[p.id] ? (
                    <p className="text-xs text-gray-400">Cargando variantes...</p>
                  ) : variantes[p.id].length === 0 ? (
                    <p className="text-xs text-gray-400">Sin variantes clínicas registradas</p>
                  ) : (
                    <div className="space-y-2">
                      <p className="text-xs font-semibold text-gray-600 uppercase tracking-wide">Variantes clínicas</p>
                      {variantes[p.id].map(v => (
                        <div key={v.id} className="flex items-start gap-3 text-sm">
                          {v.caso && (
                            <span className="text-xs px-2 py-0.5 rounded-full bg-rose-100 text-rose-700 font-medium capitalize flex-shrink-0">
                              {v.caso}
                            </span>
                          )}
                          <div className="flex-1">
                            <p className="font-medium text-gray-700">{v.nombre}</p>
                            {v.descripcion && <p className="text-xs text-gray-400">{v.descripcion}</p>}
                          </div>
                          <div className="text-xs text-gray-400 text-right flex-shrink-0">
                            <p>{v.calorias} kcal</p>
                            <p>P{v.proteinas} C{v.carbohidratos} G{v.grasas}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function Macro({ label, val, unit, bold }: { label: string; val?: number; unit?: string; bold?: boolean }) {
  if (val == null) return null
  return (
    <div className="flex flex-col items-center min-w-[40px]">
      <span className={bold ? 'font-semibold text-gray-800' : ''}>{val}{unit ?? ''}</span>
      <span className="text-xs text-gray-400">{label}</span>
    </div>
  )
}
