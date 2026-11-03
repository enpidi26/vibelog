<template>
  <div class="frame-page">
    <div class="frame frame-narrow">
      <div class="topbar">
        <Logo />
        <RouterLink to="/" class="micro-label back-link">Kembali ke beranda</RouterLink>
      </div>
      <main class="auth-main">
        <p class="micro-label">VibeLog, Jurnal Prompt</p>
        <h1 class="pixel-small">{{ title }}</h1>
        <p class="subtitle">{{ subtitle }}</p>
        <slot />
        <p class="auth-foot"><slot name="foot" /></p>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import Logo from './Logo.vue'

defineProps<{ title: string; subtitle: string }>()
</script>

<style scoped>
.auth-main {
  padding: 1.5rem 1.5rem 1.75rem;
}
h1 {
  margin: 0.6rem 0 0.4rem;
}
.subtitle {
  margin: 0 0 1.25rem;
  color: #5c5c54;
}
.auth-foot {
  margin: 1.25rem 0 0;
  font-size: 0.9rem;
  color: #5c5c54;
}
.back-link {
  text-decoration: none;
}
</style>
