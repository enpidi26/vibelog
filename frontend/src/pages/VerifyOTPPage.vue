<template>
  <AuthShell title="Verifikasi Email" subtitle="Masukkan kode 6 digit yang sudah dikirim ke email kamu.">
    <form @submit.prevent="onSubmit">
      <div class="field">
        <label for="verify-email">Email</label>
        <input id="verify-email" v-model="email" type="email" required autocomplete="email" readonly />
      </div>
      <div class="field">
        <label for="verify-code">Kode Verifikasi</label>
        <input 
          id="verify-code" 
          v-model="code" 
          type="text" 
          required 
          maxlength="6" 
          pattern="[0-9]{6}"
          placeholder="000000"
          autocomplete="one-time-code"
          inputmode="numeric"
        />
        <small>Cek inbox atau folder spam email kamu.</small>
      </div>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <p v-if="success" class="success" role="alert">{{ success }}</p>
      <button class="btn primary block" type="submit" :disabled="loading || code.length !== 6">
        {{ loading ? 'Memverifikasi...' : 'Verifikasi' }}
      </button>
    </form>
    <template #foot>
      <div style="text-align: center">
        Belum terima kode? 
        <button 
          type="button" 
          @click="onResend" 
          :disabled="resending || resendCooldown > 0"
          style="background: none; border: none; color: var(--primary); cursor: pointer; text-decoration: underline; padding: 0"
        >
          {{ resendCooldown > 0 ? `Kirim ulang (${resendCooldown}s)` : 'Kirim ulang' }}
        </button>
      </div>
    </template>
  </AuthShell>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AuthShell from '../components/AuthShell.vue'
import { useAuth } from '../composables/useAuth'
import { api } from '../lib/api'
import { ApiError } from '../types'

const router = useRouter()
const route = useRoute()
const { verifyOTP } = useAuth()

const email = ref(route.query.email as string || '')
const code = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)
const resending = ref(false)
const resendCooldown = ref(0)

let cooldownInterval: number | undefined

onMounted(() => {
  if (!email.value) {
    router.push('/register')
  }
})

onUnmounted(() => {
  if (cooldownInterval) {
    clearInterval(cooldownInterval)
  }
})

async function onSubmit() {
  if (code.value.length !== 6) {
    error.value = 'Kode harus 6 digit'
    return
  }
  
  error.value = ''
  success.value = ''
  loading.value = true
  
  try {
    await verifyOTP(email.value.trim(), code.value.trim())
    await router.push('/app')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Gagal memverifikasi. Coba lagi.'
  } finally {
    loading.value = false
  }
}

async function onResend() {
  if (resendCooldown.value > 0) return
  
  error.value = ''
  success.value = ''
  resending.value = true
  
  try {
    await api.resendOTP(email.value.trim())
    success.value = 'Kode verifikasi baru telah dikirim ke email kamu!'
    
    // Cooldown 60 detik
    resendCooldown.value = 60
    cooldownInterval = setInterval(() => {
      resendCooldown.value--
      if (resendCooldown.value <= 0 && cooldownInterval) {
        clearInterval(cooldownInterval)
      }
    }, 1000)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Gagal mengirim ulang kode. Coba lagi.'
  } finally {
    resending.value = false
  }
}
</script>

<style scoped>
.success {
  color: var(--success, #16a34a);
  background: var(--success-bg, #dcfce7);
  padding: 0.75rem;
  border-radius: 0.375rem;
  margin: 1rem 0;
}

small {
  display: block;
  margin-top: 0.25rem;
  color: var(--text-muted, #6b7280);
  font-size: 0.875rem;
}

input[readonly] {
  background-color: var(--bg-muted, #f3f4f6);
  cursor: not-allowed;
}
</style>
