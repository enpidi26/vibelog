<template>
  <RouterLink to="/" class="logo" aria-label="VibeLog, kembali ke beranda">
    <span class="logo-mark" aria-hidden="true">V</span>
    <span class="logo-words">
      <span class="logo-name">Vibe</span>
      <span class="logo-sub">LOG</span>
    </span>
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
.logo-words {
  display: flex;
  flex-direction: column;
  line-height: 1;
}
.logo-name {
  font-weight: 700;
  font-size: 1.05rem;
  letter-spacing: -0.01em;
}
.logo-sub {
  font-family: 'Space Mono', monospace;
  font-size: 0.62rem;
  letter-spacing: 0.34em;
  color: #5c5c54;
}
</style>
