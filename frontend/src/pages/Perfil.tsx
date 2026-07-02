import { useState, FormEvent } from 'react'
import { useAuth } from '../lib/auth'
import { api } from '../lib/api'
import { User, Lock } from 'lucide-react'

export default function Perfil() {
  const { user, reload } = useAuth()
  const [tab, setTab] = useState<'info' | 'password'>('info')

  const [perfil, setPerfil] = useState({
    nombre: user?.nombre ?? '',
    apellidos: user?.apellidos ?? '',
    cedula: user?.cedula ?? '',
  })
  const [perfilMsg, setPerfilMsg] = useState('')
  const [perfilErr, setPerfilErr] = useState('')
  const [perfilSaving, setPerfilSaving] = useState(false)

  const [pw, setPw] = useState({ actual: '', nueva: '', confirma: '' })
  const [pwMsg, setPwMsg] = useState('')
  const [pwErr, setPwErr] = useState('')
  const [pwSaving, setPwSaving] = useState(false)

  async function savePerfil(e: FormEvent) {
    e.preventDefault()
    setPerfilErr('')
    setPerfilMsg('')
    setPerfilSaving(true)
    try {
      await api.updatePerfil(perfil)
      await reload()
      setPerfilMsg('Perfil actualizado correctamente')
    } catch (err: any) {
      setPerfilErr(err.message ?? 'Error al actualizar')
    } finally {
      setPerfilSaving(false)
    }
  }

  async function savePassword(e: FormEvent) {
    e.preventDefault()
    if (pw.nueva !== pw.confirma) { setPwErr('Las contrasenas no coinciden'); return }
    if (pw.nueva.length < 8) { setPwErr('Minimo 8 caracteres'); return }
    setPwErr('')
    setPwMsg('')
    setPwSaving(true)
    try {
      await api.changePassword({ passwordActual: pw.actual, passwordNuevo: pw.nueva })
      setPwMsg('Contrasena actualizada')
      setPw({ actual: '', nueva: '', confirma: '' })
    } catch (err: any) {
      setPwErr(err.message ?? 'Error al cambiar contrasena')
    } finally {
      setPwSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Mi perfil</h1>
        <p className="text-sm text-gray-500 mt-0.5">{user?.email}</p>
      </div>

      <div className="flex gap-1 mb-6 border-b border-gray-100">
        {(['info', 'password'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-2.5 text-sm font-medium transition-colors border-b-2 -mb-px ${
              tab === t ? 'border-rose-500 text-rose-600' : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}>
            {t === 'info' ? 'Datos personales' : 'Contrasena'}
          </button>
        ))}
      </div>

      {tab === 'info' && (
        <form onSubmit={savePerfil} className="card p-5 space-y-4">
          {perfilErr && <p className="text-sm text-red-600 bg-red-50 px-4 py-3 rounded-lg">{perfilErr}</p>}
          {perfilMsg && <p className="text-sm text-green-700 bg-green-50 px-4 py-3 rounded-lg">{perfilMsg}</p>}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nombre</label>
            <input className="input" value={perfil.nombre}
              onChange={e => setPerfil(f => ({ ...f, nombre: e.target.value }))} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Apellidos</label>
            <input className="input" value={perfil.apellidos}
              onChange={e => setPerfil(f => ({ ...f, apellidos: e.target.value }))} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Cedula profesional</label>
            <input className="input" value={perfil.cedula}
              onChange={e => setPerfil(f => ({ ...f, cedula: e.target.value }))} />
          </div>
          <div className="pt-1">
            <button type="submit" className="btn-primary" disabled={perfilSaving}>
              {perfilSaving ? 'Guardando...' : 'Guardar cambios'}
            </button>
          </div>
        </form>
      )}

      {tab === 'password' && (
        <form onSubmit={savePassword} className="card p-5 space-y-4">
          {pwErr && <p className="text-sm text-red-600 bg-red-50 px-4 py-3 rounded-lg">{pwErr}</p>}
          {pwMsg && <p className="text-sm text-green-700 bg-green-50 px-4 py-3 rounded-lg">{pwMsg}</p>}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Contrasena actual</label>
            <input type="password" className="input" value={pw.actual}
              onChange={e => setPw(f => ({ ...f, actual: e.target.value }))} required />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Nueva contrasena</label>
            <input type="password" className="input" value={pw.nueva}
              onChange={e => setPw(f => ({ ...f, nueva: e.target.value }))} required minLength={8} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Confirmar nueva contrasena</label>
            <input type="password" className="input" value={pw.confirma}
              onChange={e => setPw(f => ({ ...f, confirma: e.target.value }))} required />
          </div>
          <div className="pt-1">
            <button type="submit" className="btn-primary" disabled={pwSaving}>
              {pwSaving ? 'Guardando...' : 'Cambiar contrasena'}
            </button>
          </div>
        </form>
      )}
    </div>
  )
}
