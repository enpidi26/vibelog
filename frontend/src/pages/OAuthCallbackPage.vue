<template>
  <div class="frame-page">
    <div class="frame frame-narrow">
      <main class="not-found anim-pop" aria-label="Signing in">
        <template v-if="error">
          <p class="pixel-small">OOPS_</p>
          <h1>GitHub sign-in failed</h1>
          <p>{{ error }}</p>
          <RouterLink to="/login" class="btn primary">Back to sign in</RouterLink>
        </template>
        <template v-else>
          <p class="pixel-small">SIGNING_IN_</p>
          <h1>Signing you in with GitHub...</h1>
          <p>This should only take a moment.</p>
        </template>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { setToken, api } from '../lib/api'
import type { ApiError } from '../types'

const router = useRouter()
const error = ref('')

onMounted(async () => {
  const params = new URLSearchParams(window.location.search)
  const token = params.get('token')
  const authError = params.get('error')

  if (authError) {
    window.history.replaceState({}, '', '/login')
    error.value = authError
    return
  }
  if (!token) {
    error.value = 'No token was returned. Please sign in again.'
    return
  }

  // Clean the URL so the token doesn't stay visible.
  window.history.replaceState({}, '', '/app')
  setToken(token)
  try {
    await api.me()
    // Force the auth guard to re-run against the fresh token.
    await router.replace('/app')
  } catch (e) {
    setToken(null)
    error.value = (e as ApiError).message || 'Could not complete sign in. Please try again.'
  }
})
</script>

<style scoped>
.not-found h1 {
  margin: 0.8rem 0 0.4rem;
}
.not-found p {
  color: #5c5c54;
}
</style>
