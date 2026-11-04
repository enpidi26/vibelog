/// <reference types="vite/client" />

export interface User {
  id: number
  name: string
  email: string
  email_verified: boolean
  created_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface RegisterResponse {
  user: User
  message: string
}

export interface Project {
  id: number
  name: string
  description: string
  color: string
  entry_count: number
  created_at: string
  updated_at: string
}

export interface ProjectInput {
  name: string
  description: string
  color: string
}

export type Outcome = 'success' | 'partial' | 'failed' | 'in_progress'

export interface Entry {
  id: number
  project_id: number | null
  project_name: string
  title: string
  prompt_text: string
  model: string
  tool: string
  outcome: Outcome
  rating: number | null
  result_notes: string
  tags: string[]
  iteration_no: number
  parent_id: number | null
  created_at: string
  updated_at: string
}

export interface EntryInput {
  project_id: number | null
  title: string
  prompt_text: string
  model: string
  tool: string
  outcome: Outcome
  rating: number | null
  result_notes: string
  tags: string[]
}

export interface EntryFilters {
  q: string
  project_id: string
  model: string
  tool: string
  outcome: string
  tag: string
  limit: number
}

export interface CountItem {
  name: string
  count: number
}

export interface Summary {
  total: number
  success: number
  partial: number
  failed: number
  in_progress: number
  average_rating: number | null
  rated_count: number
  by_model: CountItem[]
  by_tool: CountItem[]
  by_project: CountItem[]
  top_tags: CountItem[]
}

export class ApiError extends Error {
  code: string
  status: number
  constructor(message: string, code = 'unknown', status = 0) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}
