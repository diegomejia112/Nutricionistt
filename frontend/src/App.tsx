import { useState, useEffect, useCallback } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthContext, fetchUser, type User } from './lib/auth'
import { api } from './lib/api'
import Sidebar from './components/layout/Sidebar'

import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Pacientes from './pages/Pacientes'
import NuevoPaciente from './pages/NuevoPaciente'
import ExpedientePaciente from './pages/ExpedientePaciente'
import NuevaConsulta from './pages/NuevaConsulta'
import Planes from './pages/Planes'
import NuevoPlan from './pages/NuevoPlan'
import EditorPlan from './pages/EditorPlan'
import Alimentos from './pages/Alimentos'
import IA from './pages/IA'
import Perfil from './pages/Perfil'

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('access_token')
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-screen bg-gray-50">
      <Sidebar />
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
    </div>
  )
}

export default function App() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const reload = useCallback(async () => {
    const u = await fetchUser()
    setUser(u)
  }, [])

  useEffect(() => {
    fetchUser().then(u => { setUser(u); setLoading(false) })
  }, [])

  const login = async (email: string, password: string) => {
    await api.login(email, password)
    const u = await fetchUser()
    setUser(u)
  }

  const logout = () => {
    api.logout()
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, reload }}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/*" element={
            <RequireAuth>
              <AppLayout>
                <Routes>
                  <Route path="/" element={<Dashboard />} />
                  <Route path="/pacientes" element={<Pacientes />} />
                  <Route path="/pacientes/nuevo" element={<NuevoPaciente />} />
                  <Route path="/pacientes/:id" element={<ExpedientePaciente />} />
                  <Route path="/pacientes/:id/consultas/nueva" element={<NuevaConsulta />} />
                  <Route path="/planes" element={<Planes />} />
                  <Route path="/planes/nuevo" element={<NuevoPlan />} />
                  <Route path="/planes/:id" element={<EditorPlan />} />
                  <Route path="/alimentos" element={<Alimentos />} />
                  <Route path="/ia" element={<IA />} />
                  <Route path="/perfil" element={<Perfil />} />
                </Routes>
              </AppLayout>
            </RequireAuth>
          } />
        </Routes>
      </BrowserRouter>
    </AuthContext.Provider>
  )
}
