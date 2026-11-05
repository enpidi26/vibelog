<template>
  <AuthShell title="Sign in with Email" subtitle="We'll send you a magic link to sign in.">
    <form v-if="!sent" @submit.prevent="onSubmit">
      <div class="field">
        <label for="magic-email">Email</label>
        <input id="magic-email" v-model="email" type="email" required autocomplete="email" placeholder="your@email.com" />
      </div>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button class="btn primary block" type="submit" :disabled="loading">
        {{ loading ? 'Sending...' : 'Send magic link' }}
      </button>
      <RouterLink to="/login" class="btn block">Back to login</RouterLink>
    </form>
    
    <div v-else class="success-message">
      <div class="icon">✉️</div>
      <h3>Check your email</h3>
      <p>We've sent a magic link to <strong>{{ email }}</strong></p>
      <p class="hint">Click the link in the email to sign in. The link expires in 15 minutes.</p>
      <button class="btn primary block" @click="sent = false">Send another link</button>
      <RouterLink to="/login" class="btn block">Back to login</RouterLink>
    </div>
  </AuthShell>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import AuthShell from '../components/AuthShell.vue'
import { apiCall } from '../lib/api'
import { ApiError } from '../types'

const email = ref('')
const error = ref('')
const loading = ref(false)
const sent = ref(false)

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await apiCall('/api/auth/email/request', {
      method: 'POST',
      body: JSON.stringify({ email: email.value.trim() }),
    })
    sent.value = true
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to send magic link. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.success-message {
  text-align: center;
  padding: 2rem 0;
}

.success-message .icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.success-message h3 {
  margin-bottom: 1rem;
  color: var(--text);
}

.success-message p {
  margin-bottom: 1rem;
  color: var(--text-muted);
}

.success-message .hint {
  font-size: 0.9rem;
  margin-bottom: 2rem;
}

.success-message strong {
  color: var(--text);
}
</style>
