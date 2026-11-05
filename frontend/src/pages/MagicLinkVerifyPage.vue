<template>
  <AuthShell title="Verifying..." subtitle="Please wait while we sign you in.">
    <div class="verify-status">
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>Verifying your magic link...</p>
      </div>
      <div v-else-if="error" class="error-state">
        <div class="icon">❌</div>
        <h3>Verification failed</h3>
        <p class="error">{{ error }}</p>
        <RouterLink to="/login/email" class="btn primary block">Request a new link</RouterLink>
        <RouterLink to="/login" class="btn block">Back to login</RouterLink>
      </div>
    </div>
  </AuthShell>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import AuthShell from '../components/AuthShell.vue'
import { apiCall } from '../lib/api'
import { ApiError } from '../types'

const router = useRouter()
const route = useRoute()
const error = ref('')
const loading = ref(true)

onMounted(async () => {
  const token = route.query.token as string
  
  if (!token) {
    error.value = 'No verification token found in URL'
    loading.value = false
    return
  }

  try {
    const data = await apiCall(`/api/auth/email/verify?token=${encodeURIComponent(token)}`, {
      method: 'GET',
    })
    
    // Store token in localStorage
    localStorage.setItem('token', data.token)
    
    // Redirect to app
    await router.push('/app')
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
