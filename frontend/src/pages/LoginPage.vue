<template>
  <AuthShell title="Sign in to VibeLog" subtitle="Continue your prompt journal.">
    <form @submit.prevent="onSubmit">
      <div class="field">
        <label for="login-email">Email</label>
        <input id="login-email" v-model="email" type="email" required autocomplete="email" />
      </div>
      <div class="field">
        <label for="login-password">Password</label>
        <input id="login-password" v-model="password" type="password" required autocomplete="current-password" />
      </div>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button class="btn primary block" type="submit" :disabled="loading">{{ loading ? 'Signing in...' : 'Sign in' }}</button>
      <a class="btn block github-btn" href="/api/auth/github" rel="noopener">
        <GithubIcon />
        Continue with GitHub
      </a>
      <RouterLink to="/login/email" class="btn block email-btn">
        Continue with Email (Magic Link)
      </RouterLink>
    </form>
    <template #foot>
      Don't have an account? <RouterLink to="/register">Sign up here</RouterLink>.
    </template>
  </AuthShell>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthShell from '../components/AuthShell.vue'
import GithubIcon from '../components/GithubIcon.vue'
import { useAuth } from '../composables/useAuth'
import { ApiError } from '../types'

const router = useRouter()
const { login } = useAuth()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await login(email.value.trim(), password.value)
    await router.push('/app')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to sign in. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>
