import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Calendar, Cpu } from 'lucide-react'
import { api } from '../lib/api'

export default function Planes() {
  const [planes, setPlanes] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getPlanes().then(d => { setPlanes(d); setLoading(false) }).catch(() => setLoading(false))
  }, [])

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Planes alimenticios</h1>
          <p className="text-sm text-gray-500 mt-0.5">Planes semanales de nutricion</p>
        </div>
        <div className="flex gap-2">
          <Link to="/ia" className="btn-secondary flex items-center gap-2">
            <Cpu className="w-4 h-4" /> Generar con IA
          </Link>
          <Link to="/planes/nuevo" className="btn-primary flex items-center gap-2">
            <Plus className="w-4 h-4" /> Nuevo plan
          </Link>
        </div>
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1,2,3].map(i => <div key={i} className="h-20 bg-gray-100 rounded-xl animate-pulse" />)}
        </div>
      ) : planes.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <Calendar className="w-10 h-10 mx-auto mb-3 opacity-30" />
          <p className="font-medium">Sin planes creados</p>
          <p className="text-sm mt-1">Crea tu primer plan nutricional</p>
        </div>
      ) : (
        <div className="space-y-2">
          {planes.map(p => (
            <Link key={p.id} to={`/planes/${p.id}`}
              className="card px-5 py-4 flex items-center gap-4 hover:shadow-card-hover transition-shadow">
              <div className="w-10 h-10 rounded-xl bg-rose-50 flex items-center justify-center flex-shrink-0">
                <Calendar className="w-5 h-5 text-rose-500" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <p className="font-medium text-gray-900 truncate">{p.nombre}</p>
                  {p.generadoIA && <span className="badge-rose text-xs">IA</span>}
                </div>
                <p className="text-sm text-gray-400">{p.paciente?.nombre} {p.paciente?.apellidos}</p>
              </div>
              <div className="text-right flex-shrink-0">
                {p.caloriasObj && <p className="text-sm font-medium text-gray-600">{p.caloriasObj} kcal</p>}
                <p className="text-xs text-gray-400">{p.fechaInicio}</p>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
