<template>
  <section v-if="summary" class="stats-strip" aria-label="Ringkasan jurnal">
    <div class="stat-cell">
      <span class="stat-num">{{ summary.total }}</span>
      <span class="stat-label">Total entri</span>
    </div>
    <div class="stat-cell">
      <span class="stat-num">{{ summary.success }}</span>
      <span class="stat-label">Berhasil</span>
    </div>
    <div class="stat-cell">
      <span class="stat-num">{{ summary.partial }}</span>
      <span class="stat-label">Sebagian</span>
    </div>
    <div class="stat-cell">
      <span class="stat-num">{{ summary.failed }}</span>
      <span class="stat-label">Gagal</span>
    </div>
    <div class="stat-cell">
      <span class="stat-num">{{ summary.average_rating === null ? 'Belum ada' : summary.average_rating.toFixed(1) + ' / 5' }}</span>
      <span class="stat-label">Rata-rata rating ({{ summary.rated_count }} dinilai)</span>
    </div>
  </section>
  <p v-else-if="loading" class="muted">Memuat ringkasan...</p>
</template>

<script setup lang="ts">
import type { Summary } from '../types'

defineProps<{ summary?: Summary; loading: boolean }>()
</script>

<style scoped>
.stats-strip {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  border: 1px solid #111;
  border-radius: 4px;
  background: #fff;
  overflow: hidden;
  margin: 1rem 0;
}
.stat-cell {
  padding: 0.8rem 0.9rem;
  border-right: 1px solid #d2d2c8;
}
.stat-cell:last-child {
  border-right: 0;
}
.stat-num {
  display: block;
  font-size: 1.35rem;
  font-weight: 700;
  line-height: 1.2;
}
.stat-label {
  display: block;
  font-family: 'Space Mono', monospace;
  font-size: 0.62rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #5c5c54;
}
@media (max-width: 720px) {
  .stats-strip {
    grid-template-columns: repeat(2, 1fr);
  }
  .stat-cell {
    border-bottom: 1px solid #d2d2c8;
  }
  .stat-cell:last-child {
    grid-column: span 2;
    border-bottom: 0;
  }
}
</style>
