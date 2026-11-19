<template>
  <div class="frame-page">
    <div class="frame">
      <header class="topbar">
        <Logo />
        <nav class="center-nav" aria-label="Navigation">
          <a href="#features">Features</a>
        </nav>
        <nav class="right-nav" aria-label="Account">
          <RouterLink to="/login" class="nav-link">Sign in</RouterLink>
          <RouterLink v-if="registrationOpen" to="/register" class="btn primary small">Sign up free</RouterLink>
        </nav>
      </header>

      <main>
        <section class="hero">
          <h1 class="pixel-headline hero-title">Vibe_Log</h1>
          <p class="hero-lead">A prompt journal for vibe coders. Save the prompt, rate the result, iterate.</p>
          <div class="hero-actions">
            <RouterLink :to="registrationOpen ? '/register' : '/login'" class="btn primary">
              {{ registrationOpen ? 'Start journal' : 'Sign in' }}
            </RouterLink>
          </div>
        </section>

        <section class="features" id="features">
          <p class="micro-label">What it does</p>
          <div class="feature-grid">
            <article v-for="feature in features" :key="feature.title" class="feature">
              <h2>{{ feature.title }}</h2>
              <p>{{ feature.description }}</p>
            </article>
          </div>
        </section>

      </main>

      <footer class="frame-foot">
        <p class="micro-label">© 2026 VibeLog. AGPL-3.0 licensed. · <a href="https://github.com/enpidi26/vibelog" rel="noopener">GitHub</a></p>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Logo from '../components/Logo.vue'
import { useInstanceMeta } from '../composables/useInstanceMeta'

const meta = useInstanceMeta()
const registrationOpen = computed(() => meta.data.value?.registration_enabled === true)

const features = [
  {
    title: 'Prompt journal',
    description: 'Save the exact prompt, model, tool, result, rating, and tags.',
  },
  {
    title: 'Iterations',
    description: 'Duplicate an entry as the next attempt, without losing the old one.',
  },
  {
    title: 'Projects and filters',
    description: 'Group entries by project. Search text, outcome, model, tool, or tag.',
  },
  {
    title: 'Prompt packs',
    description: 'Export filtered entries as Markdown or JSON.',
  },
]
</script>

<style scoped>
.hero {
  padding: 2rem 1.25rem 0.5rem;
}
.hero-lead {
  max-width: 38rem;
  margin: 1rem 0 0;
  font-size: 1.1rem;
  color: #5c5c54;
}
.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin-top: 1.2rem;
}
.features {
  padding: 2rem 1.25rem 0;
}
.feature-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1px;
  margin-top: 0.8rem;
  border: 1px solid #d2d2c8;
  background: #d2d2c8;
}
.feature {
  padding: 1rem;
  background: #fff;
}
.feature h2 {
  margin: 0 0 0.35rem;
  font-size: 1rem;
  line-height: 1.3;
}
.feature p {
  margin: 0;
  color: #5c5c54;
  font-size: 0.9rem;
}
.frame-foot {
  padding: 1rem 1.25rem 1.25rem;
  border-top: 1px solid #d2d2c8;
}
.frame-foot p {
  margin: 0;
}
@media (max-width: 900px) {
  .feature-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 620px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
  .hero-actions .btn {
    width: 100%;
  }
}
</style>
