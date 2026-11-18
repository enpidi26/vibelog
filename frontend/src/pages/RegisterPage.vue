<template>
  <AuthShell :title="registrationOpen ? 'Create VibeLog account' : 'Registration closed'" :subtitle="registrationOpen ? 'One account for all your projects and prompt journals.' : 'This instance does not accept new accounts.'">
    <template v-if="registrationOpen">
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
          <label for="register-password">Password</label>
          <input id="register-password" v-model="password" type="password" required minlength="8" autocomplete="new-password" placeholder="At least 8 characters" />
        </div>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
        <button class="btn primary block" type="submit" :disabled="loading">{{ loading ? 'Creating account...' : 'Sign up' }}</button>
        <a v-if="meta.data.value?.github_login_enabled === true" class="btn block github-btn" href="/api/auth/github" rel="noopener">
          <GithubIcon />
          Continue with GitHub
        </a>
      </form>
    </template>
    <div v-else class="closed-state">
      <p>If you already have an account on this instance, sign in below.</p>
      <RouterLink to="/login" class="btn primary block">Back to sign in</RouterLink>
    </div>
    <template #foot>
      <template v-if="registrationOpen">
        Already have an account? <RouterLink to="/login">Sign in here</RouterLink>.
      </template>
    </template>
  </AuthShell>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthShell from '../components/AuthShell.vue'
import GithubIcon from '../components/GithubIcon.vue'
import { useAuth } from '../composables/useAuth'
import { useInstanceMeta } from '../composables/useInstanceMeta'
import { ApiError } from '../types'

const router = useRouter()
const { register } = useAuth()
const meta = useInstanceMeta()
const registrationOpen = computed(() => meta.data.value?.registration_enabled ?? true)
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
    await router.push('/app')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to create account. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.closed-state {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  text-align: center;
  padding: 2rem 0;
}
.closed-state p {
  margin: 0;
  color: var(--muted);
}
</style>
