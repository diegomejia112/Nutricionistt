const BASE = '/api'

function getToken(): string {
  return localStorage.getItem('access_token') ?? ''
}

function setTokens(access: string, refresh: string) {
  localStorage.setItem('access_token', access)
  localStorage.setItem('refresh_token', refresh)
}

function clearTokens() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

async function refreshAccessToken(): Promise<boolean> {
  const refresh = localStorage.getItem('refresh_token')
  if (!refresh) return false
  try {
    const res = await fetch(`${BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh }),
    })
    if (!res.ok) { clearTokens(); return false }
    const data = await res.json()
    setTokens(data.access_token, data.refresh_token)
    return true
  } catch {
    return false
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  retry = true
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401 && retry) {
    const ok = await refreshAccessToken()
    if (ok) return request<T>(method, path, body, false)
    clearTokens()
    if (!window.location.pathname.startsWith('/login')) {
      window.location.href = '/login'
    }
    throw new Error('session expired')
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: `HTTP ${res.status}` }))
    throw new Error(err.error ?? `HTTP ${res.status}`)
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  // Auth
  login: (email: string, password: string) =>
    request<{ access_token: string; refresh_token: string; expires_in: number }>(
      'POST', '/auth/login', { email, password }
    ).then(d => { setTokens(d.access_token, d.refresh_token); return d }),

  registro: (body: { email: string; password: string; nombre: string; apellidos: string; cedula?: string }) =>
    request('POST', '/auth/registro', body),

  logout: () => clearTokens(),

  // Perfil
  getMe: () => request<{ id: string; email: string; nombre: string; apellidos: string; cedula?: string }>('GET', '/me'),
  updatePerfil: (body: object) => request('PATCH', '/perfil', body),
  changePassword: (body: object) => request('POST', '/perfil/password', body),

  // Pacientes
  getPacientes: (q = '', limit = 50) =>
    request<any[]>('GET', `/pacientes?q=${encodeURIComponent(q)}&limit=${limit}`),
  createPaciente: (body: object) => request<any>('POST', '/pacientes', body),
  getPaciente: (id: string) => request<any>('GET', `/pacientes/${id}`),
  updatePaciente: (id: string, body: object) => request<any>('PATCH', `/pacientes/${id}`, body),
  deletePaciente: (id: string) => request('DELETE', `/pacientes/${id}`),

  // Consultas
  getConsultas: (pacienteId: string) => request<any[]>('GET', `/pacientes/${pacienteId}/consultas`),
  createConsulta: (pacienteId: string, body: object) =>
    request<any>('POST', `/pacientes/${pacienteId}/consultas`, body),

  // Planes
  getPlanes: () => request<any[]>('GET', '/planes'),
  createPlan: (body: object) => request<any>('POST', '/planes', body),
  getPlan: (id: string) => request<any>('GET', `/planes/${id}`),
  updatePlan: (id: string, body: object) => request<any>('PATCH', `/planes/${id}`, body),
  deletePlan: (id: string) => request('DELETE', `/planes/${id}`),
  addAlimentoPlan: (planId: string, body: object) =>
    request<any>('POST', `/planes/${planId}/alimentos`, body),
  updateAlimentoPlan: (planId: string, alimentoId: string, body: object) =>
    request<any>('PATCH', `/planes/${planId}/alimentos/${alimentoId}`, body),
  removeAlimentoPlan: (planId: string, alimentoId: string) =>
    request('DELETE', `/planes/${planId}/alimentos/${alimentoId}`),

  // Restricciones de ingredientes por paciente
  getRestricciones: (pacienteId: string) =>
    request<any[]>('GET', `/pacientes/${pacienteId}/restricciones`),
  addRestriccion: (pacienteId: string, body: { ingrediente: string; motivo?: string }) =>
    request<{ id: string }>('POST', `/pacientes/${pacienteId}/restricciones`, body),
  deleteRestriccion: (pacienteId: string, restriccionId: string) =>
    request('DELETE', `/pacientes/${pacienteId}/restricciones/${restriccionId}`),

  // Alimentos y Platillos
  getAlimentos: (q = '', categoria = '') =>
    request<any[]>('GET', `/alimentos?q=${encodeURIComponent(q)}&categoria=${encodeURIComponent(categoria)}`),
  getPlatillos: (params?: { q?: string; categoria?: string; region?: string; caso?: string }) => {
    const p = new URLSearchParams()
    if (params?.q) p.set('q', params.q)
    if (params?.categoria) p.set('categoria', params.categoria)
    if (params?.region) p.set('region', params.region)
    if (params?.caso) p.set('caso', params.caso)
    return request<any[]>('GET', `/platillos?${p.toString()}`)
  },
  getVariantes: (platilloId: string) =>
    request<any[]>('GET', `/platillos/${platilloId}/variantes`),
  getPlatillosCompatibles: (pacienteId: string, params?: { q?: string; categoria?: string; caso?: string }) => {
    const p = new URLSearchParams()
    if (params?.q) p.set('q', params.q)
    if (params?.categoria) p.set('categoria', params.categoria)
    if (params?.caso) p.set('caso', params.caso)
    return request<any[]>('GET', `/platillos/compatibles/${pacienteId}?${p.toString()}`)
  },

  // IA
  generarPlan: (body: object) => request<any>('POST', '/ia/generar', body),
  getIAConfig: () => request<{ configurada: boolean }>('GET', '/ia/config'),
  saveIAConfig: (apiKey: string) => request<{ ok: boolean }>('POST', '/ia/config', { apiKey }),

  // Seguimiento
  listSeguimientos: (pacienteId: string) =>
    request<any[]>('GET', `/pacientes/${pacienteId}/seguimiento`),
  addSeguimiento: (pacienteId: string, body: object) =>
    request<{ id: string }>('POST', `/pacientes/${pacienteId}/seguimiento`, body),
  deleteSeguimiento: (pacienteId: string, sid: string) =>
    request<{ ok: boolean }>('DELETE', `/pacientes/${pacienteId}/seguimiento/${sid}`),
}

export { setTokens, clearTokens, getToken }
