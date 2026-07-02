import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Users, ClipboardList, Calendar, TrendingUp } from 'lucide-react'
import { api } from '../lib/api'
import { useAuth } from '../lib/auth'

export default function Dashboard() {
  const { user } = useAuth()
  const [pacientes, setPacientes] = useState<any[]>([])
  const [planes, setPlanes] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      api.getPacientes('', 5),
      api.getPlanes(),
    ]).then(([p, pl]) => {
      setPacientes(p)
      setPlanes(pl.slice(0, 5))
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [])

  const hora = new Date().getHours()
  const saludo = hora < 12 ? 'Buenos dias' : hora < 18 ? 'Buenas tardes' : 'Buenas noches'

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">
          {saludo}, {user?.nombre}
        </h1>
        <p className="text-gray-500 text-sm mt-1">
          {new Date().toLocaleDateString('es-MX', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-8">
        <div className="card p-5">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-rose-50 flex items-center justify-center">
              <Users className="w-5 h-5 text-rose-500" />
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">{loading ? '—' : pacientes.length}</p>
              <p className="text-xs text-gray-500">Pacientes recientes</p>
            </div>
          </div>
        </div>
        <div className="card p-5">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-rose-50 flex items-center justify-center">
              <ClipboardList className="w-5 h-5 text-rose-500" />
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">{loading ? '—' : planes.length}</p>
              <p className="text-xs text-gray-500">Planes activos</p>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="card p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-gray-800 flex items-center gap-2">
              <Users className="w-4 h-4 text-gray-400" />
              Pacientes recientes
            </h2>
            <Link to="/pacientes" className="text-xs text-rose-600 hover:underline">Ver todos</Link>
          </div>
          {loading ? (
            <div className="space-y-3">
              {[1,2,3].map(i => <div key={i} className="h-10 bg-gray-100 rounded-lg animate-pulse" />)}
            </div>
          ) : pacientes.length === 0 ? (
            <p className="text-sm text-gray-400 text-center py-6">Sin pacientes registrados</p>
          ) : (
            <ul className="space-y-2">
              {pacientes.map(p => (
                <li key={p.id}>
                  <Link to={`/pacientes/${p.id}`}
                    className="flex items-center justify-between p-2 rounded-lg hover:bg-gray-50 transition-colors">
                    <div>
                      <p className="text-sm font-medium text-gray-800">{p.nombre} {p.apellidos}</p>
                      <p className="text-xs text-gray-400">{p.edad ? `${p.edad} anos` : ''}</p>
                    </div>
                    <span className="text-xs text-gray-300">&rsaquo;</span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="card p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-gray-800 flex items-center gap-2">
              <Calendar className="w-4 h-4 text-gray-400" />
              Planes recientes
            </h2>
            <Link to="/planes" className="text-xs text-rose-600 hover:underline">Ver todos</Link>
          </div>
          {loading ? (
            <div className="space-y-3">
              {[1,2,3].map(i => <div key={i} className="h-10 bg-gray-100 rounded-lg animate-pulse" />)}
            </div>
          ) : planes.length === 0 ? (
            <p className="text-sm text-gray-400 text-center py-6">Sin planes creados</p>
          ) : (
            <ul className="space-y-2">
              {planes.map(p => (
                <li key={p.id}>
                  <Link to={`/planes/${p.id}`}
                    className="flex items-center justify-between p-2 rounded-lg hover:bg-gray-50 transition-colors">
                    <div>
                      <p className="text-sm font-medium text-gray-800">{p.nombre}</p>
                      <p className="text-xs text-gray-400">{p.paciente?.nombre} {p.paciente?.apellidos}</p>
                    </div>
                    {p.generadoIA && <span className="badge-rose text-xs">IA</span>}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      <div className="mt-6 card p-5">
        <h2 className="font-semibold text-gray-800 flex items-center gap-2 mb-4">
          <TrendingUp className="w-4 h-4 text-gray-400" />
          Acciones rapidas
        </h2>
        <div className="flex gap-3 flex-wrap">
          <Link to="/pacientes/nuevo" className="btn-primary">Nuevo paciente</Link>
          <Link to="/planes/nuevo" className="btn-secondary">Nuevo plan</Link>
          <Link to="/ia" className="btn-secondary">Generar plan con IA</Link>
        </div>
      </div>
    </div>
  )
}
