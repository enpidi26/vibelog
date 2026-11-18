import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '../lib/api'
import { useAuth } from '../composables/useAuth'
import DashboardPage from '../pages/DashboardPage.vue'
import LandingPage from '../pages/LandingPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import NotFoundPage from '../pages/NotFoundPage.vue'
import RegisterPage from '../pages/RegisterPage.vue'
import OAuthCallbackPage from '../pages/OAuthCallbackPage.vue'
import EmailLoginPage from '../pages/EmailLoginPage.vue'
import MagicLinkVerifyPage from '../pages/MagicLinkVerifyPage.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', component: LandingPage, meta: { guestOnly: true, title: 'Prompt journal' } },
  { path: '/login', component: LoginPage, meta: { guestOnly: true, title: 'Sign in' } },
  { path: '/login/email', component: EmailLoginPage, meta: { guestOnly: true, title: 'Sign in with email' } },
  { path: '/register', component: RegisterPage, meta: { guestOnly: true, title: 'Create account' } },
  { path: '/auth/github/callback', component: OAuthCallbackPage, meta: { title: 'Signing in' } },
  { path: '/auth/verify', component: MagicLinkVerifyPage, meta: { title: 'Signing in' } },
  { path: '/app', component: DashboardPage, meta: { requiresAuth: true, title: 'Journal' } },
  { path: '/:pathMatch(.*)*', component: NotFoundPage, meta: { title: 'Page not found' } },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuth()
  if (getToken() && !auth.restored.value) {
    await auth.restore()
  }
  const authenticated = auth.user.value !== null || getToken() !== null
  if (to.meta.requiresAuth && !authenticated) {
    return '/login'
  }
  if (to.meta.guestOnly && authenticated) {
    return '/app'
  }
  return true
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} — VibeLog` : 'VibeLog — Prompt journal'
})
