import { useEffect, useState, useRef } from 'react'
import { Link } from 'react-router-dom'
import { Search, UserPlus, User } from 'lucide-react'
import { api } from '../lib/api'

export default function Pacientes() {
  const [pacientes, setPacientes] = useState<any[]>([])
  const [q, setQ] = useState('')
  const [loading, setLoading] = useState(true)
  const debounce = useRef<ReturnType<typeof setTimeout>>()

  const load = (query: string) => {
    setLoading(true)
    api.getPacientes(query, 50).then(data => {
      setPacientes(data)
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  useEffect(() => { load('') }, [])

  function handleSearch(val: string) {
    setQ(val)
    clearTimeout(debounce.current)
    debounce.current = setTimeout(() => load(val), 300)
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Pacientes</h1>
          <p className="text-sm text-gray-500 mt-0.5">Expedientes clinicos</p>
        </div>
        <Link to="/pacientes/nuevo" className="btn-primary flex items-center gap-2">
          <UserPlus className="w-4 h-4" />
          Nuevo paciente
        </Link>
      </div>

      <div className="relative mb-4">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input
          type="text"
          className="input pl-9"
          placeholder="Buscar por nombre o correo..."
          value={q}
          onChange={e => handleSearch(e.target.value)}
        />
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1,2,3,4].map(i => <div key={i} className="h-16 bg-gray-100 rounded-xl animate-pulse" />)}
        </div>
      ) : pacientes.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <User className="w-10 h-10 mx-auto mb-3 opacity-30" />
          <p className="font-medium">{q ? 'Sin resultados' : 'Sin pacientes'}</p>
          <p className="text-sm mt-1">
            {q ? 'Intenta con otro nombre' : 'Agrega tu primer paciente'}
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {pacientes.map(p => (
            <Link key={p.id} to={`/pacientes/${p.id}`}
              className="card px-5 py-4 flex items-center gap-4 hover:shadow-card-hover transition-shadow">
              <div className="w-10 h-10 rounded-full bg-rose-100 flex items-center justify-center flex-shrink-0">
                <span className="text-rose-600 font-semibold text-sm">
                  {p.nombre?.[0]}{p.apellidos?.[0]}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="font-medium text-gray-900 truncate">{p.nombre} {p.apellidos}</p>
                <p className="text-sm text-gray-400 truncate">{p.email}</p>
              </div>
              <div className="text-right flex-shrink-0">
                {p.edad && <p className="text-sm text-gray-600">{p.edad} anos</p>}
                {p.telefono && <p className="text-xs text-gray-400">{p.telefono}</p>}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
