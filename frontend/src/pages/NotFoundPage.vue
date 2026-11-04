<template>
  <div class="frame-page">
    <div class="frame frame-narrow">
      <main class="not-found anim-pop">
        <p class="pixel-small">404_</p>
        <h1>Page not found</h1>
        <p>The link you opened doesn't exist or has changed.</p>
        <RouterLink to="/" class="btn primary">Back to home</RouterLink>
      </main>
    </div>
  </div>
</template>

<style scoped>
.not-found {
  animation: pop-in var(--dur-mid) var(--ease-spring) both;
  padding: 2rem 1.5rem;
}
.not-found h1 {
  margin: 0.8rem 0 0.4rem;
}
.not-found p {
  color: #5c5c54;
}
</style>
