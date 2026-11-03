import { ApiError, type AuthResponse, type Entry, type EntryFilters, type EntryInput, type Project, type ProjectInput, type Summary, type User } from '../types'

export const TOKEN_KEY = 'vibelog_token'

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function setToken(token: string | null): void {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    // localStorage tidak tersedia: sesi hanya di memori lewat useAuth.
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body) headers.set('Content-Type', 'application/json')
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)

  let res: Response
  try {
    res = await fetch(path, { ...init, headers })
  } catch {
    throw new ApiError('Tidak bisa terhubung ke server. Periksa koneksi, lalu coba lagi.', 'network', 0)
  }

  let data: unknown = null
  const text = await res.text()
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = null
    }
  }
  if (!res.ok) {
    const err = (data ?? {}) as { error?: { code?: string; message?: string } }
    throw new ApiError(
      err.error?.message ?? `Server mengembalikan ${res.status}.`,
      err.error?.code ?? 'unknown',
      res.status,
    )
  }
  return data as T
}

function entryQuery(filters: EntryFilters): string {
  const params = new URLSearchParams()
  if (filters.q) params.set('q', filters.q)
  if (filters.project_id && filters.project_id !== 'all') params.set('project_id', filters.project_id)
  if (filters.model) params.set('model', filters.model)
  if (filters.tool) params.set('tool', filters.tool)
  if (filters.outcome) params.set('outcome', filters.outcome)
  if (filters.tag) params.set('tag', filters.tag)
  if (filters.limit) params.set('limit', String(filters.limit))
  const qs = params.toString()
  return qs ? `?${qs}` : ''
}

export const api = {
  register(name: string, email: string, password: string): Promise<AuthResponse> {
    return request<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    })
  },
  login(email: string, password: string): Promise<AuthResponse> {
    return request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  },
  me(): Promise<User> {
    return request<User>('/api/auth/me')
  },
  listProjects(): Promise<Project[]> {
    return request<Project[]>('/api/projects')
  },
  createProject(input: ProjectInput): Promise<Project> {
    return request<Project>('/api/projects', { method: 'POST', body: JSON.stringify(input) })
  },
  updateProject(id: number, input: ProjectInput): Promise<Project> {
    return request<Project>(`/api/projects/${id}`, { method: 'PATCH', body: JSON.stringify(input) })
  },
  deleteProject(id: number): Promise<{ deleted: boolean; id: number }> {
    return request<{ deleted: boolean; id: number }>(`/api/projects/${id}`, { method: 'DELETE' })
  },
  listEntries(filters: EntryFilters): Promise<Entry[]> {
    return request<Entry[]>(`/api/entries${entryQuery(filters)}`)
  },
  createEntry(input: EntryInput): Promise<Entry> {
    return request<Entry>('/api/entries', { method: 'POST', body: JSON.stringify(input) })
  },
  updateEntry(id: number, input: EntryInput): Promise<Entry> {
    return request<Entry>(`/api/entries/${id}`, { method: 'PATCH', body: JSON.stringify(input) })
  },
  deleteEntry(id: number): Promise<{ deleted: boolean; id: number }> {
    return request<{ deleted: boolean; id: number }>(`/api/entries/${id}`, { method: 'DELETE' })
  },
  duplicateEntry(id: number): Promise<Entry> {
    return request<Entry>(`/api/entries/${id}/duplicate`, { method: 'POST' })
  },
  summary(projectId: string): Promise<Summary> {
    const qs = projectId && projectId !== 'all' ? `?project_id=${projectId}` : ''
    return request<Summary>(`/api/entries/summary${qs}`)
  },
  async exportEntries(filters: EntryFilters, format: 'md' | 'json'): Promise<{ blob: Blob; filename: string }> {
    const headers = new Headers()
    const token = getToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
    const qs = entryQuery(filters)
    const sep = qs ? '&' : '?'
    const res = await fetch(`/api/entries/export${qs}${sep}format=${format}`, { headers })
    if (!res.ok) {
      throw new ApiError('Gagal mengekspor entri.', 'export_failed', res.status)
    }
    const blob = await res.blob()
    const disposition = res.headers.get('Content-Disposition') ?? ''
    const match = /filename="([^"]+)"/.exec(disposition)
    return { blob, filename: match?.[1] ?? `vibelog-export.${format}` }
  },
}
