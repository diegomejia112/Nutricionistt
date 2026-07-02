import { createContext, useContext } from 'react'
import { api } from './api'

export interface User {
  id: string
  email: string
  nombre: string
  apellidos: string
  cedula?: string
}

export interface AuthCtx {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  reload: () => Promise<void>
}

export const AuthContext = createContext<AuthCtx>({
  user: null,
  loading: true,
  login: async () => {},
  logout: () => {},
  reload: async () => {},
})

export function useAuth() {
  return useContext(AuthContext)
}

export async function fetchUser(): Promise<User | null> {
  if (!localStorage.getItem('access_token')) return null
  try {
    return await api.getMe()
  } catch {
    return null
  }
}
