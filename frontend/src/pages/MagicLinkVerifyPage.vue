<template>
  <AuthShell :title="heading" :subtitle="subtitle">
    <div class="verify-status">
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>Verifying your magic link...</p>
      </div>
      <div v-else-if="error" class="error-state">
        <div class="icon">❌</div>
        <h3>Verification failed</h3>
        <p class="error">{{ error }}</p>
        <RouterLink v-if="meta.data.value?.email_login_enabled !== false" to="/login/email" class="btn primary block">Request a new link</RouterLink>
        <RouterLink to="/login" class="btn block">Back to login</RouterLink>
      </div>
    </div>
  </AuthShell>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import AuthShell from '../components/AuthShell.vue'
import { useInstanceMeta } from '../composables/useInstanceMeta'
import { api, setToken } from '../lib/api'
import { ApiError } from '../types'

const router = useRouter()
const route = useRoute()
const meta = useInstanceMeta()
const error = ref('')
const loading = ref(true)
const heading = computed(() => (error.value ? 'Invalid link' : 'Verifying...'))
const subtitle = computed(() =>
  error.value ? 'This sign-in link cannot be used.' : 'Please wait while we sign you in.',
)

onMounted(async () => {
  const token = route.query.token as string | undefined

  if (!token) {
    error.value = 'No verification token found in URL'
    loading.value = false
    return
  }

  try {
    const data = await api.verifyMagicLink(token)
    setToken(data.token)
    // Replace so the token never lingers in browser history.
    await router.replace('/app')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to verify magic link. The link may have expired.'
    loading.value = false
  }
})
</script>

<style scoped>
.verify-status {
  text-align: center;
  padding: 3rem 0;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading p {
  color: var(--text-muted);
  font-size: 1.1rem;
}

.error-state {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.error-state .icon {
  font-size: 4rem;
  margin-bottom: 0.5rem;
}

.error-state h3 {
  color: var(--text);
  margin-bottom: 0.5rem;
}

.error-state .error {
  color: var(--error);
  margin-bottom: 1rem;
}
</style>
