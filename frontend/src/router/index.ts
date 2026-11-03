import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '../lib/api'
import { useAuth } from '../composables/useAuth'
import DashboardPage from '../pages/DashboardPage.vue'
import LandingPage from '../pages/LandingPage.vue'
import LoginPage from '../pages/LoginPage.vue'
import NotFoundPage from '../pages/NotFoundPage.vue'
import RegisterPage from '../pages/RegisterPage.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', component: LandingPage, meta: { guestOnly: true } },
  { path: '/login', component: LoginPage, meta: { guestOnly: true } },
  { path: '/register', component: RegisterPage, meta: { guestOnly: true } },
  { path: '/app', component: DashboardPage, meta: { requiresAuth: true } },
  { path: '/:pathMatch(.*)*', component: NotFoundPage },
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
