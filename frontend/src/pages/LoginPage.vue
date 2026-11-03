<template>
  <AuthShell title="Masuk ke VibeLog" subtitle="Lanjutkan jurnal prompt kamu.">
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
      <button class="btn primary block" type="submit" :disabled="loading">{{ loading ? 'Masuk...' : 'Masuk' }}</button>
    </form>
    <template #foot>
      Belum punya akun? <RouterLink to="/register">Daftar di sini</RouterLink>.
    </template>
  </AuthShell>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthShell from '../components/AuthShell.vue'
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
    error.value = e instanceof ApiError ? e.message : 'Gagal masuk. Coba lagi.'
  } finally {
    loading.value = false
  }
}
</script>
