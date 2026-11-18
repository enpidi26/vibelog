<template>
  <article class="ref-card card-anim">
    <div class="ref-card-top">
      <span class="tag-chip">{{ card.tag }}</span>
      <span class="micro-label">{{ card.index }}</span>
    </div>
    <div class="ref-card-body feature-body">
      <h2>{{ card.title }}</h2>
      <p>{{ card.description }}</p>
    </div>
    <div class="ref-card-stats">
      <div v-for="stat in card.stats" :key="stat.label" class="ref-stat">
        <b>{{ stat.value }}</b><span>{{ stat.label }}</span>
      </div>
      <RouterLink to="/register" class="arrow-cell arrow-link" aria-label="Try this feature">→</RouterLink>
    </div>
  </article>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { FeatureCard } from '../types'

defineProps<{ card: FeatureCard }>()
</script>

<style scoped>
.feature-body {
  align-items: center;
  text-align: center;
  justify-content: center;
  min-height: 170px;
}
.feature-body h2 {
  margin: 0;
  font-size: 1.3rem;
  line-height: 1.25;
}
.feature-body p {
  margin: 0.4rem 0 0;
  color: #5c5c54;
  font-size: 0.92rem;
  max-width: 17rem;
}
.arrow-link {
  display: flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
}
</style>
