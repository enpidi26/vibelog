<template>
  <form class="entry-form" @submit.prevent="onSubmit">
    <div class="form-grid">
      <div class="field span-2">
        <label for="entry-title">Judul</label>
        <input id="entry-title" v-model="title" type="text" maxlength="120" required placeholder="Perbaiki login yang kadang gagal" />
      </div>
      <div class="field">
        <label for="entry-project">Proyek</label>
        <select id="entry-project" v-model="projectId">
          <option value="">Tanpa proyek</option>
          <option v-for="p in projects" :key="p.id" :value="String(p.id)">{{ p.name }}</option>
        </select>
      </div>
      <div class="field">
        <label for="entry-outcome">Hasil</label>
        <select id="entry-outcome" v-model="outcome">
          <option value="in_progress">Berjalan</option>
          <option value="success">Berhasil</option>
          <option value="partial">Sebagian</option>
          <option value="failed">Gagal</option>
        </select>
      </div>
      <div class="field">
        <label for="entry-model">Model AI</label>
        <input id="entry-model" v-model="model" type="text" maxlength="60" required list="model-list" placeholder="Claude" />
        <datalist id="model-list">
          <option v-for="m in MODEL_SUGGESTIONS" :key="m" :value="m" />
        </datalist>
      </div>
      <div class="field">
        <label for="entry-tool">Tool</label>
        <input id="entry-tool" v-model="tool" type="text" maxlength="60" required list="tool-list" placeholder="Cursor" />
        <datalist id="tool-list">
          <option v-for="t in TOOL_SUGGESTIONS" :key="t" :value="t" />
        </datalist>
      </div>
      <div class="field">
        <label for="entry-rating">Rating (opsional)</label>
        <select id="entry-rating" v-model="rating">
          <option value="">Belum dinilai</option>
          <option value="1">1, buruk</option>
          <option value="2">2, kurang</option>
          <option value="3">3, cukup</option>
          <option value="4">4, bagus</option>
          <option value="5">5, sangat bagus</option>
        </select>
      </div>
      <div class="field">
        <label for="entry-tags">Tag, pisahkan koma (opsional)</label>
        <input id="entry-tags" v-model="tagsText" type="text" placeholder="auth, refactor, bug" />
      </div>
      <div class="field span-2">
        <label for="entry-prompt">Prompt yang dikirim</label>
        <textarea id="entry-prompt" v-model="promptText" rows="6" maxlength="10000" required placeholder="Tempel prompt persis seperti yang dikirim ke AI..."></textarea>
        <p class="hint">{{ promptText.length }} / 10000 karakter</p>
      </div>
      <div class="field span-2">
        <label for="entry-notes">Hasil dan catatan (opsional)</label>
        <textarea id="entry-notes" v-model="resultNotes" rows="4" maxlength="10000" placeholder="Apa yang berhasil, apa yang gagal, dan apa yang mau dicoba di iterasi berikutnya..."></textarea>
      </div>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="row">
      <button class="btn primary" type="submit" :disabled="saving">{{ saving ? 'Menyimpan...' : initial ? 'Simpan perubahan' : 'Simpan entri' }}</button>
      <button class="btn" type="button" @click="$emit('cancel')">Batal</button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { MODEL_SUGGESTIONS, TOOL_SUGGESTIONS } from '../lib/format'
import type { Entry, EntryInput, Outcome, Project } from '../types'

const props = defineProps<{
  initial?: Entry | null
  projects: Project[]
  saving: boolean
  error: string
}>()
const emit = defineEmits<{ submit: [input: EntryInput]; cancel: [] }>()

const title = ref(props.initial?.title ?? '')
const projectId = ref(props.initial?.project_id ? String(props.initial.project_id) : '')
const outcome = ref<Outcome>(props.initial?.outcome ?? 'in_progress')
const model = ref(props.initial?.model ?? '')
const tool = ref(props.initial?.tool ?? '')
const rating = ref(props.initial?.rating ? String(props.initial.rating) : '')
const tagsText = ref((props.initial?.tags ?? []).join(', '))
const promptText = ref(props.initial?.prompt_text ?? '')
const resultNotes = ref(props.initial?.result_notes ?? '')

function onSubmit() {
  emit('submit', {
    project_id: projectId.value ? Number(projectId.value) : null,
    title: title.value.trim(),
    prompt_text: promptText.value,
    model: model.value.trim(),
    tool: tool.value.trim(),
    outcome: outcome.value,
    rating: rating.value ? Number(rating.value) : null,
    result_notes: resultNotes.value,
    tags: tagsText.value.split(',').map((t) => t.trim()).filter(Boolean),
  })
}
</script>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.9rem;
}
.form-grid .field {
  margin-bottom: 0;
}
.span-2 {
  grid-column: span 2;
}
@media (max-width: 640px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  .span-2 {
    grid-column: span 1;
  }
}
</style>
