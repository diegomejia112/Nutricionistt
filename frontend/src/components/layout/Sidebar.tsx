import { NavLink, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Users, ClipboardList, Salad,
  Cpu, LogOut, Settings,
} from 'lucide-react'
import { useAuth } from '@/lib/auth'
import { clsx } from 'clsx'

const nav = [
  { to: '/',          label: 'Dashboard',    icon: LayoutDashboard, end: true },
  { to: '/pacientes', label: 'Pacientes',    icon: Users },
  { to: '/planes',    label: 'Planes',       icon: ClipboardList },
  { to: '/alimentos', label: 'Alimentos',    icon: Salad },
  { to: '/ia',        label: 'Asistente IA', icon: Cpu },
]

export default function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <aside className="w-60 min-h-screen bg-white border-r border-gray-100 flex flex-col flex-shrink-0">

      <div className="px-5 py-5 border-b border-gray-100">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-rose-500 rounded-lg flex items-center justify-center flex-shrink-0">
            <span className="text-white font-bold text-sm tracking-tight">N</span>
          </div>
          <div>
            <p className="text-gray-900 font-bold text-[15px] leading-none">Nutricionist</p>
            <p className="text-gray-400 text-[11px] mt-0.5">Sistema nutricional</p>
          </div>
        </div>
      </div>

      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {nav.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              clsx(
                'flex items-center gap-3 px-3 py-2.5 rounded-lg text-[13.5px] font-medium transition-all duration-100',
                isActive
                  ? 'bg-rose-50 text-rose-600'
                  : 'text-gray-500 hover:bg-gray-50 hover:text-gray-800'
              )
            }
          >
            {({ isActive }) => (
              <>
                <Icon size={16} strokeWidth={isActive ? 2.2 : 1.8} />
                <span>{label}</span>
              </>
            )}
          </NavLink>
        ))}
      </nav>

      <div className="px-3 py-4 border-t border-gray-100 space-y-0.5">
        <NavLink
          to="/perfil"
          className={({ isActive }) =>
            clsx(
              'flex items-center gap-3 px-3 py-2.5 rounded-lg text-[13.5px] font-medium transition-all',
              isActive ? 'bg-rose-50 text-rose-600' : 'text-gray-500 hover:bg-gray-50 hover:text-gray-800'
            )
          }
        >
          <Settings size={16} strokeWidth={1.8} />
          <span>Perfil y ajustes</span>
        </NavLink>

        <div className="flex items-center gap-3 px-3 py-2.5 mt-1">
          <div className="w-7 h-7 rounded-full bg-rose-100 flex items-center justify-center flex-shrink-0">
            <span className="text-rose-600 text-xs font-bold">
              {user?.nombre?.charAt(0).toUpperCase() ?? 'N'}
            </span>
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-gray-800 text-[13px] font-semibold truncate leading-none">
              {user?.nombre ?? 'Nutricionista'}
            </p>
            <p className="text-gray-400 text-[11px] truncate mt-0.5">{user?.email}</p>
          </div>
        </div>

        <button
          onClick={handleLogout}
          className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-[13.5px] font-medium text-gray-400 hover:text-red-500 hover:bg-red-50 transition-all"
        >
          <LogOut size={16} strokeWidth={1.8} />
          <span>Cerrar sesion</span>
        </button>
      </div>
    </aside>
  )
}
