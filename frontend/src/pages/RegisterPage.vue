<template>
  <AuthShell title="Create VibeLog account" subtitle="One account for all your projects and prompt journals.">
    <form @submit.prevent="onSubmit">
      <div class="field">
        <label for="register-name">Name</label>
        <input id="register-name" v-model="name" type="text" maxlength="80" required autocomplete="name" />
      </div>
      <div class="field">
        <label for="register-email">Email</label>
        <input id="register-email" v-model="email" type="email" required autocomplete="email" />
      </div>
      <div class="field">
        <label for="register-password">Password (minimum 8 characters)</label>
        <input id="register-password" v-model="password" type="password" required minlength="8" autocomplete="new-password" />
      </div>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button class="btn primary block" type="submit" :disabled="loading">{{ loading ? 'Creating account...' : 'Sign up' }}</button>
    </form>
    <template #foot>
      Already have an account? <RouterLink to="/login">Sign in here</RouterLink>.
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
const { register } = useAuth()
const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await register(name.value.trim(), email.value.trim(), password.value)
    // Redirect ke halaman verifikasi OTP
    await router.push(`/verify-otp?email=${encodeURIComponent(email.value.trim())}`)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to create account. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>
