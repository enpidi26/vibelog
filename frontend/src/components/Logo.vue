<template>
  <RouterLink to="/" class="logo" aria-label="VibeLog, kembali ke beranda">
    <span class="logo-mark" aria-hidden="true">V</span>
  </RouterLink>
</template>

<style scoped>
.logo {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  color: inherit;
}
.logo-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.1rem;
  height: 2.1rem;
  border-radius: 8px;
  background: #ff5c1a;
  color: #fff;
  font-family: 'Press Start 2P', monospace;
  font-size: 0.95rem;
}
.logo:hover .logo-mark {
  animation: wiggle 0.4s ease-in-out;
}
</style>
