<template>
  <article class="ref-card">
    <div class="ref-card-top">
      <span class="tag-chip" :class="{ orange: entry.outcome === 'success' }">{{ outcomeLabel(entry.outcome) }}</span>
      <span class="micro-label">Iterasi {{ entry.iteration_no }}</span>
    </div>

    <div class="ref-card-body">
      <h3>{{ entry.title }}</h3>
      <p class="meta">
        {{ entry.project_name || 'Tanpa proyek' }}
        <span v-if="entry.parent_id"> · lanjutan dari entri #{{ entry.parent_id }}</span>
        · {{ formatDate(entry.created_at) }}
      </p>
      <p class="chips">
        <span>Model: <strong>{{ entry.model }}</strong></span>
        <span>Tool: <strong>{{ entry.tool }}</strong></span>
        <span>{{ stars(entry.rating) }}</span>
      </p>

      <button class="prompt-toggle" type="button" @click="promptOpen = !promptOpen">
        {{ promptOpen ? 'Sembunyikan prompt' : 'Lihat prompt' }}
      </button>
      <Transition name="prompt">
        <div v-if="promptOpen" class="prompt-box">
          <pre>{{ entry.prompt_text }}</pre>
          <button class="btn small" type="button" @click="copyPrompt">{{ copied ? 'Prompt tersalin' : 'Salin prompt' }}</button>
        </div>
      </Transition>

      <p v-if="entry.result_notes" class="notes">{{ entry.result_notes }}</p>

      <p v-if="entry.tags.length" class="tags">
        <button
          v-for="tag in entry.tags"
          :key="tag"
          class="tag"
          type="button"
          @click="$emit('tag-click', tag)"
        >
          #{{ tag }}
        </button>
      </p>
    </div>

    <div class="ref-card-stats">
      <div class="ref-stat"><b>{{ entry.model }}</b><span>Model</span></div>
      <div class="ref-stat"><b>{{ entry.tool }}</b><span>Tool</span></div>
      <div class="ref-stat"><b>{{ entry.iteration_no }}</b><span>Iterasi</span></div>
      <div class="ref-stat"><b>{{ entry.rating === null ? 'Belum dinilai' : entry.rating + ' / 5' }}</b><span>Rating</span></div>
      <button class="arrow-cell" type="button" :aria-label="promptOpen ? 'Sembunyikan prompt' : 'Lihat prompt'" @click="promptOpen = !promptOpen">→</button>
    </div>

    <footer class="card-actions">
      <button class="btn small" type="button" @click="$emit('edit', entry)">Ubah</button>
      <button class="btn small" type="button" :disabled="duplicating" @click="$emit('duplicate', entry)">
        {{ duplicating ? 'Membuat...' : 'Iterasi berikutnya' }}
      </button>
      <button class="btn small danger" type="button" :disabled="deleting" @click="$emit('delete', entry)">
        {{ deleting ? 'Menghapus...' : 'Hapus' }}
      </button>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { formatDate, outcomeLabel, stars } from '../lib/format'
import type { Entry } from '../types'

const props = defineProps<{
  entry: Entry
  duplicating: boolean
  deleting: boolean
}>()
defineEmits<{
  edit: [entry: Entry]
  duplicate: [entry: Entry]
  delete: [entry: Entry]
  'tag-click': [tag: string]
}>()

const copied = ref(false)
const promptOpen = ref(false)

async function copyPrompt() {
  try {
    await navigator.clipboard.writeText(props.entry.prompt_text)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch {
    copied.value = false
  }
}
</script>

<style scoped>
h3 {
  margin: 0;
  font-size: 1.15rem;
  line-height: 1.3;
  text-align: left;
}
.meta {
  margin: 0.3rem 0 0;
  font-size: 0.85rem;
  color: #5c5c54;
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 1rem;
  font-size: 0.9rem;
  color: #5c5c54;
  margin: 0.6rem 0 0;
}
.prompt-toggle {
  margin-top: 0.8rem;
  padding: 0;
  border: 0;
  background: none;
  font-weight: 700;
  color: #e64e12;
  cursor: pointer;
  text-align: left;
}
.prompt-box {
  margin-top: 0.5rem;
}
pre {
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Space Mono', monospace;
  font-size: 0.85rem;
  background: #f6f6f1;
  border: 1px solid #d2d2c8;
  border-radius: 6px;
  padding: 0.75rem;
  font-size: 0.9rem;
  max-height: 16rem;
  overflow: auto;
  margin: 0 0 0.5rem;
}
.notes {
  white-space: pre-wrap;
  word-break: break-word;
  color: #24352f;
  margin: 0.8rem 0 0;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin: 0.8rem 0 0;
}
.card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.7rem 0.8rem;
  border-top: 1px solid #d2d2c8;
}
.prompt-enter-active,
.prompt-leave-active {
  transition: opacity var(--dur-mid) var(--ease-out), transform var(--dur-mid) var(--ease-out);
}
.prompt-enter-from,
.prompt-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.prompt-toggle {
  transition: color var(--dur-fast) var(--ease-out);
}
.prompt-toggle:hover {
  color: #b83c0c;
}

.arrow-cell {
  transition: background-color var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out), transform var(--dur-fast) var(--ease-out);
}
</style>
