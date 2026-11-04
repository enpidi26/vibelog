import { computed, ref } from 'vue'
import { api, getToken, setToken } from '../lib/api'
import type { User } from '../types'

const user = ref<User | null>(null)
const restoring = ref(false)
const restored = ref(false)

export function useAuth() {
  const isAuthenticated = computed(() => user.value !== null || getToken() !== null)

  async function restore(): Promise<void> {
    if (restored.value || restoring.value) return
    if (!getToken()) {
      restored.value = true
      return
    }
    restoring.value = true
    try {
      user.value = await api.me()
    } catch {
      setToken(null)
      user.value = null
    } finally {
      restoring.value = false
      restored.value = true
    }
  }

  async function login(email: string, password: string): Promise<void> {
    const res = await api.login(email, password)
    setToken(res.token)
    user.value = res.user
    restored.value = true
  }

  async function register(name: string, email: string, password: string): Promise<User> {
    const res = await api.register(name, email, password)
    setToken(res.token)
    user.value = res.user
    restored.value = true
    return res.user
  }

  function logout(): void {
    setToken(null)
    user.value = null
    restored.value = true
  }

  return { user, restoring, restored, isAuthenticated, restore, login, register, logout }
}
