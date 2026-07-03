import { useState, FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../lib/api'
import { useAuth } from '../lib/auth'

export default function Register() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [nombre, setNombre] = useState('')
  const [apellidos, setApellidos] = useState('')
  const [cedula, setCedula] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(e: FormEvent) {
    e.preventDefault()
    setError('')
    if (password.length < 8) {
      setError('La contraseña debe tener al menos 8 caracteres')
      return
    }
    setLoading(true)
    try {
      await api.registro({ email, password, nombre, apellidos, cedula: cedula || undefined })
      await login(email, password)
      navigate('/', { replace: true })
    } catch (err: any) {
      setError(err.message ?? 'Error al registrar')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-rose-500 mb-4">
            <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Nutricionist</h1>
          <p className="text-sm text-gray-500 mt-1">Crear cuenta de nutriólogo</p>
        </div>

        <div className="card p-8">
          <h2 className="text-lg font-semibold text-gray-800 mb-6">Registro</h2>
          <form onSubmit={submit} className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Nombre</label>
                <input className="input" value={nombre} onChange={e => setNombre(e.target.value)}
                  required autoFocus placeholder="Brenda" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Apellidos</label>
                <input className="input" value={apellidos} onChange={e => setApellidos(e.target.value)}
                  required placeholder="Ramírez" />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Correo electronico</label>
              <input type="email" className="input" value={email} onChange={e => setEmail(e.target.value)}
                required placeholder="nutricionista@ejemplo.com" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Contrasena</label>
              <input type="password" className="input" value={password} onChange={e => setPassword(e.target.value)}
                required placeholder="Mínimo 8 caracteres" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Cédula profesional (opcional)</label>
              <input className="input" value={cedula} onChange={e => setCedula(e.target.value)} />
            </div>
            {error && (
              <p className="text-sm text-red-600 bg-red-50 px-3 py-2 rounded-lg">{error}</p>
            )}
            <button type="submit" className="btn-primary w-full" disabled={loading}>
              {loading ? 'Creando cuenta...' : 'Crear cuenta'}
            </button>
          </form>
          <p className="text-center text-sm text-gray-400 mt-5">
            ¿Ya tienes cuenta?{' '}
            <Link to="/login" className="text-rose-500 hover:text-rose-600 font-medium">Inicia sesión</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
