<template>
  <div class="frame-page">
    <div class="frame">
      <header class="topbar">
        <Logo />
        <nav class="center-nav" aria-label="Navigation">
          <a href="#features">Features</a>
          <a href="#start">Get Started</a>
        </nav>
        <nav class="right-nav" aria-label="Account">
          <RouterLink to="/login" class="nav-link">Sign in</RouterLink>
          <RouterLink to="/register" class="btn primary small">Sign up free</RouterLink>
        </nav>
      </header>

      <main>
        <section class="hero">
          <h1 class="pixel-headline hero-title">Vibe_Log</h1>
          <div class="hero-grid anim-fade-up" style="animation-delay: 80ms">
            <div class="hero-cell intro-cell">
              <p><span class="bullet" aria-hidden="true">►</span> Record prompts, rate results, and iterate on what works.</p>
            </div>
            <div class="hero-cell deco-cell">
              <span class="plus-motif pulse-motif" aria-hidden="true">+</span>
              <span class="micro-label">Prompt Pack</span>
            </div>
            <div class="hero-cell featured-cell">
              <span class="tag-chip">Featured</span>
              <h2>Ready-to-paste prompt packs</h2>
              <p>Successful prompts can be exported and pasted into AI agents or editors, without rewriting from chat history.</p>
              <div class="mini-stats">
                <div><b>1</b><span>Click iteration</span></div>
                <div><b>2</b><span>Export formats</span></div>
                <div><b>5</b><span>Max rating</span></div>
                <div><b>10</b><span>Max tags</span></div>
              </div>
            </div>
          </div>
        </section>

        <section class="count-row" id="features">
          <div>
            <p class="micro-label">Features</p>
            <p class="pixel-small count">06</p>
          </div>
        </section>

        <section class="card-grid" aria-label="VibeLog Features">
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
              <RouterLink to="/register" class="arrow-cell arrow-link" aria-label="Try this feature">→</RouterLink>
            </div>
          </article>
        </section>

        <section class="cta-row" id="start">
          <div class="cta-left">
            <p class="pixel-small"><span class="bullet" aria-hidden="true">►</span> Ready to record your first prompt?</p>
            <p>Sign up, create projects, and save prompts you just sent to AI.</p>
          </div>
          <div class="cta-right">
            <RouterLink to="/register" class="btn primary">Start my journal</RouterLink>
            <RouterLink to="/login" class="btn">I already have an account</RouterLink>
          </div>
        </section>
      </main>

      <footer class="frame-foot">
        <p class="micro-label">© 2026 VibeLog. All rights reserved.</p>
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
    tag: 'Iteration',
    title: 'Next iteration',
    description: 'Duplicate old entries into next attempts.',
    stats: [
      { value: '1 click', label: 'Duplicate' },
      { value: '+1', label: 'Iteration number' },
      { value: 'In Progress', label: 'Initial outcome' },
      { value: 'Empty', label: 'Initial rating' },
    ],
  },
  {
    index: '02',
    tag: 'Summary',
    title: 'Successful patterns',
    description: 'Summary by model, tool, project, and tags.',
    stats: [
      { value: 'Total', label: 'And by outcome' },
      { value: '1-5', label: 'Average rating' },
      { value: 'Model', label: 'By model' },
      { value: 'Tag', label: 'Top tags' },
    ],
  },
  {
    index: '03',
    tag: 'Export',
    title: 'Prompt packs',
    description: 'Export filtered entries for reuse.',
    stats: [
      { value: 'Markdown', label: 'Format 1' },
      { value: 'JSON', label: 'Format 2' },
      { value: 'Filter', label: 'Follows filters' },
      { value: 'Ready', label: 'Paste to AI' },
    ],
  },
  {
    index: '04',
    tag: 'Filter',
    title: 'Search and filter',
    description: 'Find old prompts without scrolling chat history.',
    stats: [
      { value: 'Search', label: 'Debounced' },
      { value: 'Project', label: 'Including none' },
      { value: '4', label: 'Outcome options' },
      { value: 'Tag', label: 'Click to filter' },
    ],
  },
  {
    index: '05',
    tag: 'Projects',
    title: 'Separate projects',
    description: 'One journal for many projects you\'re building.',
    stats: [
      { value: 'CRUD', label: 'Complete' },
      { value: '7', label: 'Colors' },
      { value: 'Safe', label: 'Entries persist' },
      { value: 'Count', label: 'Entry count' },
    ],
  },
  {
    index: '06',
    tag: 'Isolation',
    title: 'Separate accounts',
    description: 'Your journal can only be read by your account.',
    stats: [
      { value: 'JWT', label: 'HS256' },
      { value: 'Bcrypt', label: 'Password' },
      { value: 'Per user', label: 'Data isolation' },
      { value: '32', label: 'Min secret chars' },
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
.center-nav {
  flex: 1;
  display: flex;
  justify-content: center;
  gap: 1rem;
}
.center-nav a,
.nav-link {
  font-weight: 500;
  text-decoration: none;
  font-size: 0.92rem;
  transition: color var(--dur-fast) var(--ease-out), transform var(--dur-fast) var(--ease-out);
  display: inline-block;
}
.center-nav a:hover,
.nav-link:hover {
  color: var(--orange);
  transform: translateY(-1px);
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
