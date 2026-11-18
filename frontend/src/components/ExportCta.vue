<template>
  <section class="cta-row" aria-label="Export prompt pack">
    <div class="cta-left">
      <p class="pixel-small"><span class="bullet" aria-hidden="true">►</span> Ready to export your prompt pack?</p>
      <p>Export follows the active filters above.</p>
    </div>
    <div class="cta-right">
      <button class="btn primary" type="button" :disabled="exporting" @click="$emit('export', 'md')">Export Markdown</button>
      <button class="btn" type="button" :disabled="exporting" @click="$emit('export', 'json')">Export JSON</button>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{ exporting: boolean }>()
defineEmits<{ export: [format: 'md' | 'json'] }>()
</script>

<style scoped>
.cta-row {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 1rem;
  align-items: center;
  border-top: 1px solid #d2d2c8;
  margin-top: 1.25rem;
  padding-top: 1.25rem;
}
.cta-left p {
  margin: 0.3rem 0 0;
  color: #5c5c54;
}
.cta-left .pixel-small {
  color: #111;
}
.bullet {
  color: #ff5c1a;
}
.cta-right {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
@media (max-width: 720px) {
  .cta-row {
    grid-template-columns: 1fr;
  }
}
</style>
