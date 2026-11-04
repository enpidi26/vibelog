<template>
  <form class="entry-form" @submit.prevent="onSubmit">
    <div class="form-grid">
      <div class="field span-2">
        <label for="entry-title">Title</label>
        <input id="entry-title" v-model="title" type="text" maxlength="120" required placeholder="Fix login that sometimes fails" />
      </div>
      <div class="field">
        <label for="entry-project">Project</label>
        <select id="entry-project" v-model="projectId">
          <option value="">No project</option>
          <option v-for="p in projects" :key="p.id" :value="String(p.id)">{{ p.name }}</option>
        </select>
      </div>
      <div class="field">
        <label for="entry-outcome">Outcome</label>
        <select id="entry-outcome" v-model="outcome">
          <option value="in_progress">In Progress</option>
          <option value="success">Success</option>
          <option value="partial">Partial</option>
          <option value="failed">Failed</option>
        </select>
      </div>
      <div class="field">
        <label for="entry-model">AI Model</label>
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
        <label for="entry-rating">Rating (optional)</label>
        <select id="entry-rating" v-model="rating">
          <option value="">Not rated</option>
          <option value="1">1, poor</option>
          <option value="2">2, fair</option>
          <option value="3">3, good</option>
          <option value="4">4, great</option>
          <option value="5">5, excellent</option>
        </select>
      </div>
      <div class="field">
        <label for="entry-tags">Tags, comma separated (optional)</label>
        <input id="entry-tags" v-model="tagsText" type="text" placeholder="auth, refactor, bug" />
      </div>
      <div class="field span-2">
        <label for="entry-prompt">Prompt sent</label>
        <textarea id="entry-prompt" v-model="promptText" rows="6" maxlength="10000" required placeholder="Paste the prompt exactly as sent to AI..."></textarea>
        <p class="hint">{{ promptText.length }} / 10000 characters</p>
      </div>
      <div class="field span-2">
        <label for="entry-notes">Results and notes (optional)</label>
        <textarea id="entry-notes" v-model="resultNotes" rows="4" maxlength="10000" placeholder="What worked, what failed, and what to try in the next iteration..."></textarea>
      </div>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="row">
      <button class="btn primary" type="submit" :disabled="saving">{{ saving ? 'Saving...' : initial ? 'Save changes' : 'Save entry' }}</button>
      <button class="btn" type="button" @click="$emit('cancel')">Cancel</button>
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
