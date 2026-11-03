<template>
  <div class="frame-page">
    <div class="frame">
      <header class="topbar">
        <Logo />
        <nav class="center-nav" aria-label="Navigasi">
          <a href="#fitur">Fitur</a>
          <a href="#mulai">Mulai</a>
        </nav>
        <nav aria-label="Akun">
          <RouterLink to="/login" class="nav-link">Masuk</RouterLink>
          <RouterLink to="/register" class="btn primary small">Daftar gratis</RouterLink>
        </nav>
      </header>

      <main>
        <section class="hero">
          <h1 class="pixel-headline hero-title">Vibe_Log</h1>
          <div class="hero-grid anim-fade-up" style="animation-delay: 80ms">
            <div class="hero-cell intro-cell">
              <p><span class="bullet" aria-hidden="true">►</span> Catat prompt, nilai hasilnya, dan ulangi yang berhasil di iterasi berikutnya.</p>
            </div>
            <div class="hero-cell deco-cell">
              <span class="plus-motif pulse-motif" aria-hidden="true">+</span>
              <span class="micro-label">Prompt Pack</span>
            </div>
            <div class="hero-cell featured-cell">
              <span class="tag-chip">Unggulan</span>
              <h2>Prompt pack siap tempel</h2>
              <p>Prompt yang berhasil bisa diekspor dan ditempel lagi ke AI agent atau editor, tanpa menulis ulang dari riwayat chat.</p>
              <div class="mini-stats">
                <div><b>1</b><span>Klik iterasi</span></div>
                <div><b>2</b><span>Format ekspor</span></div>
                <div><b>5</b><span>Rating maks</span></div>
                <div><b>10</b><span>Tag maks</span></div>
              </div>
            </div>
          </div>
        </section>

        <section class="count-row" id="fitur">
          <div>
            <p class="micro-label">Fitur</p>
            <p class="pixel-small count">06</p>
          </div>
        </section>

        <section class="card-grid" aria-label="Fitur VibeLog">
          <article v-for="(card, i) in cards" :key="card.title" class="ref-card card-anim" :style="{ animationDelay: (i * 70) + 'ms' }">
            <div class="ref-card-top">
              <span class="tag-chip">{{ card.tag }}</span>
              <span class="micro-label">{{ card.index }}</span>
            </div>
            <div class="ref-card-body feature-body">
              <h2>{{ card.title }}</h2>
              <p>{{ card.description }}</p>
            </div>
            <div class="ref-card-stats">
              <div class="ref-stat"><b>{{ card.stats[0].value }}</b><span>{{ card.stats[0].label }}</span></div>
              <div class="ref-stat"><b>{{ card.stats[1].value }}</b><span>{{ card.stats[1].label }}</span></div>
              <div class="ref-stat"><b>{{ card.stats[2].value }}</b><span>{{ card.stats[2].label }}</span></div>
              <div class="ref-stat"><b>{{ card.stats[3].value }}</b><span>{{ card.stats[3].label }}</span></div>
              <RouterLink to="/register" class="arrow-cell arrow-link" aria-label="Coba fitur ini">→</RouterLink>
            </div>
          </article>
        </section>

        <section class="cta-row" id="mulai">
          <div class="cta-left">
            <p class="pixel-small"><span class="bullet" aria-hidden="true">►</span> Siap mencatat prompt pertamamu?</p>
            <p>Daftar, buat proyek, dan simpan prompt yang baru saja kamu kirim ke AI.</p>
          </div>
          <div class="cta-right">
            <RouterLink to="/register" class="btn primary">Mulai jurnal saya</RouterLink>
            <RouterLink to="/login" class="btn">Saya sudah punya akun</RouterLink>
          </div>
        </section>
      </main>

      <footer class="frame-foot">
        <p class="micro-label">© 2026 VibeLog. Semua hak dilindungi.</p>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import Logo from '../components/Logo.vue'

interface FeatureStat {
  value: string
  label: string
}
interface FeatureCard {
  index: string
  tag: string
  title: string
  description: string
  stats: [FeatureStat, FeatureStat, FeatureStat, FeatureStat]
}

const cards: FeatureCard[] = [
  {
    index: '01',
    tag: 'Iterasi',
    title: 'Iterasi berikutnya',
    description: 'Duplikasi entri lama jadi percobaan berikutnya.',
    stats: [
      { value: '1 klik', label: 'Duplikasi' },
      { value: '+1', label: 'Nomor iterasi' },
      { value: 'Berjalan', label: 'Hasil awal' },
      { value: 'Kosong', label: 'Rating awal' },
    ],
  },
  {
    index: '02',
    tag: 'Ringkasan',
    title: 'Pola yang berhasil',
    description: 'Ringkasan per model, tool, proyek, dan tag.',
    stats: [
      { value: 'Total', label: 'Dan per hasil' },
      { value: '1-5', label: 'Rata-rata rating' },
      { value: 'Model', label: 'Per model' },
      { value: 'Tag', label: 'Teratas' },
    ],
  },
  {
    index: '03',
    tag: 'Ekspor',
    title: 'Prompt pack',
    description: 'Ekspor entri terfilter untuk dipakai ulang.',
    stats: [
      { value: 'Markdown', label: 'Format 1' },
      { value: 'JSON', label: 'Format 2' },
      { value: 'Filter', label: 'Mengikuti filter' },
      { value: 'Siap', label: 'Tempel ke AI' },
    ],
  },
  {
    index: '04',
    tag: 'Filter',
    title: 'Cari dan saring',
    description: 'Temukan prompt lama tanpa menggulir riwayat chat.',
    stats: [
      { value: 'Cari', label: 'Debounce' },
      { value: 'Proyek', label: 'Termasuk kosong' },
      { value: '4', label: 'Pilihan hasil' },
      { value: 'Tag', label: 'Klik untuk filter' },
    ],
  },
  {
    index: '05',
    tag: 'Proyek',
    title: 'Proyek terpisah',
    description: 'Satu jurnal untuk banyak proyek yang sedang dibangun.',
    stats: [
      { value: 'CRUD', label: 'Lengkap' },
      { value: '7', label: 'Warna' },
      { value: 'Aman', label: 'Entri bertahan' },
      { value: 'Hitung', label: 'Jumlah entri' },
    ],
  },
  {
    index: '06',
    tag: 'Isolasi',
    title: 'Akun terpisah',
    description: 'Jurnalmu hanya bisa dibaca oleh akunmu sendiri.',
    stats: [
      { value: 'JWT', label: 'HS256' },
      { value: 'Bcrypt', label: 'Password' },
      { value: 'Per user', label: 'Isolasi data' },
      { value: '32', label: 'Karakter secret min' },
    ],
  },
]
</script>

<style scoped>
.hero {
  padding: 1.75rem 1.25rem 0;
}
.hero-grid {
  display: grid;
  grid-template-columns: 1fr 0.65fr 1.5fr;
  border: 1px solid #d2d2c8;
  border-radius: 4px;
  background: #fff;
  margin-top: 1.25rem;
  overflow: hidden;
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}
.hero-cell {
  padding: 1rem 1.1rem;
  border-right: 1px solid #d2d2c8;
}
.hero-cell:last-child {
  border-right: 0;
}
.intro-cell p {
  margin: 0;
  font-weight: 500;
}
.bullet {
  color: #ff5c1a;
}
.deco-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.4rem;
  justify-content: center;
}
.plus-motif {
  font-family: 'Space Grotesk', sans-serif;
  font-size: 4.6rem;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1.5px #c9c9bf;
  align-self: center;
}
.pulse-motif {
  animation: pulse-soft 2.6s ease-in-out infinite;
}
.featured-cell h2 {
  margin: 0.5rem 0 0.3rem;
  font-size: 1.25rem;
}
.featured-cell p {
  margin: 0;
  color: #5c5c54;
}
.mini-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  border: 1px solid #d2d2c8;
  margin-top: 0.9rem;
}
.mini-stats div {
  padding: 0.5rem 0.6rem;
  border-right: 1px solid #d2d2c8;
}
.mini-stats div:last-child {
  border-right: 0;
}
.mini-stats b {
  display: block;
}
.mini-stats span {
  display: block;
  font-family: 'Space Mono', monospace;
  font-size: 0.6rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: #5c5c54;
}
.count-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
  padding: 1.5rem 1.25rem 0.9rem;
  border-bottom: 1px solid #d2d2c8;
  margin: 0 0 1rem;
}
.count-row p {
  margin: 0;
}
.count {
  font-size: 1.6rem;
  margin-top: 0.3rem;
}
.count-note {
  max-width: 22rem;
  text-align: right;
  color: #5c5c54;
  font-size: 0.9rem;
}
.card-grid {
  padding: 0 1.25rem;
}
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
.cta-row {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 1rem;
  align-items: center;
  border-top: 1px solid #d2d2c8;
  border-bottom: 1px solid #d2d2c8;
  margin: 1.25rem 0 0;
  padding: 1.25rem;
}
.cta-left p {
  margin: 0.3rem 0 0;
  color: #5c5c54;
}
.cta-left .pixel-small {
  color: #111;
}
.cta-right {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.stack-row {
  padding: 1.25rem;
}
.stack-row h2 {
  margin: 0 0 0.3rem;
  font-size: 1.1rem;
}
.stack-row p {
  margin: 0;
  color: #5c5c54;
}
.stack-row ul {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  list-style: none;
  padding: 0;
  margin: 0.8rem 0 0;
}
.stack-row li {
  background: #fff;
  border: 1px solid #d2d2c8;
  border-radius: 999px;
  padding: 0.35rem 0.8rem;
  font-weight: 500;
  font-size: 0.9rem;
}
.frame-foot {
  padding: 1rem 1.25rem 1.25rem;
  border-top: 1px solid #d2d2c8;
}
.frame-foot p {
  margin: 0;
}
.center-nav a,
.nav-link {
  font-weight: 500;
  text-decoration: none;
  font-size: 0.92rem;
}
.hero-title {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
  display: inline-block;
}
.hero-title::after {
  content: '_';
  animation: blink 1.1s step-end infinite;
  color: var(--orange);
}

.card-anim {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}

.card-anim:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 0 var(--line-strong);
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
}

.stack-row li {
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
}
.stack-row li:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 0 var(--line-strong);
}
@media (max-width: 720px) {
  .hero-grid {
    grid-template-columns: 1fr;
  }
  .hero-cell {
    border-right: 0;
    border-bottom: 1px solid #d2d2c8;
  }
  .hero-cell:last-child {
    border-bottom: 0;
  }
  .cta-row {
    grid-template-columns: 1fr;
  }
  .count-note {
    text-align: left;
  }
  .center-nav {
    display: none;
  }
}
</style>
