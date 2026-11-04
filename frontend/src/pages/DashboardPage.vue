<template>
  <div class="frame-page">
    <div class="frame">
      <div class="dash-inner">
    <header class="topbar">
      <Logo />
      <p class="micro-label dash-crumb">Journal / Dashboard</p>
      <div class="top-actions">
        <span v-if="authUser" class="user-name">{{ authUser.name }}</span>
        <button class="btn small" type="button" :disabled="exporting" @click="exportPack('md')">
          {{ exporting ? 'Exporting...' : 'Export Markdown' }}
        </button>
        <button class="btn small" type="button" :disabled="exporting" @click="exportPack('json')">Export JSON</button>
        <button class="btn small" type="button" @click="onLogout">Sign out</button>
      </div>
    </header>

    <section class="dash-hero">
      <h1 class="pixel-headline hero-title">Journal_</h1>
      <p class="intro">Record the prompts you send, rate results, and iterate on what works.</p>
    </section>

    <StatsCards :summary="summaryQuery.data.value" :loading="summaryQuery.isPending.value" />

    <section class="panel" aria-label="Projects">
      <div class="panel-head">
        <h2>Projects <span class="project-count">{{ String((projectsQuery.data.value ?? []).length).padStart(2, '0') }}</span></h2>
        <button class="btn small" type="button" @click="openNewProject">New project</button>
      </div>
      <p v-if="projectsQuery.isPending.value" class="muted">Loading projects...</p>
      <p v-else-if="projectsQuery.isError.value" class="error" role="alert">Failed to load projects. <button class="link" type="button" @click="projectsQuery.refetch()">Try again</button></p>
      <div v-else class="project-chips">
        <button
          class="chip"
          :class="{ active: filters.project_id === 'all' }"
          type="button"
          @click="filters.project_id = 'all'"
        >
          All projects
        </button>
        <button
          v-for="p in projectsQuery.data.value"
          :key="p.id"
          class="chip"
          :class="{ active: filters.project_id === String(p.id) }"
          type="button"
          @click="filters.project_id = String(p.id)"
        >
          <span class="color-dot" :style="{ background: projectColorHex(p.color) }" aria-hidden="true"></span>
          {{ p.name }} ({{ p.entry_count }})
        </button>
        <button class="chip" :class="{ active: filters.project_id === 'none' }" type="button" @click="filters.project_id = 'none'">
          No project
        </button>
      </div>
      <p class="hint">Note: deleting a project doesn't delete its journal entries. Entries move to No project.</p>
      <div v-if="selectedProject" class="row">
        <button class="btn small" type="button" @click="openEditProject(selectedProject)">Edit this project</button>
        <button class="btn small danger" type="button" @click="deleteProject(selectedProject)">Delete this project</button>
      </div>
      <div v-if="showProjectForm" class="anim-pop">
        <ProjectForm
          v-if="showProjectForm"
          :initial="editingProject"
          :saving="projectSaving"
          :error="projectError"
          @submit="saveProject"
          @cancel="showProjectForm = false"
        />
      </div>
    </section>

    <section class="panel" aria-label="Filter entries">
      <div class="filters">
        <div class="field">
          <label for="filter-q">Search title, prompt, notes, or tags</label>
          <input id="filter-q" v-model="filters.q" type="search" placeholder="Search title, prompt, tags..." />
        </div>
        <div class="field">
          <label for="filter-model">Model</label>
          <select id="filter-model" v-model="filters.model">
            <option value="">All models</option>
            <option v-for="m in modelOptions" :key="m" :value="m">{{ m }}</option>
          </select>
        </div>
        <div class="field">
          <label for="filter-tool">Tool</label>
          <select id="filter-tool" v-model="filters.tool">
            <option value="">All tools</option>
            <option v-for="t in toolOptions" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
        <div class="field">
          <label for="filter-outcome">Outcome</label>
          <select id="filter-outcome" v-model="filters.outcome">
            <option value="">All outcomes</option>
            <option value="success">Success</option>
            <option value="partial">Partial</option>
            <option value="failed">Failed</option>
            <option value="in_progress">In Progress</option>
          </select>
        </div>
      </div>
      <div v-if="summaryQuery.data.value?.top_tags.length" class="tags-row">
        <span class="muted">Popular tags:</span>
        <button
          v-for="t in summaryQuery.data.value.top_tags"
          :key="t.name"
          class="tag"
          :class="{ active: filters.tag === t.name }"
          type="button"
          @click="filters.tag = filters.tag === t.name ? '' : t.name"
        >
          #{{ t.name }} ({{ t.count }})
        </button>
        <button v-if="filters.tag" class="link" type="button" @click="filters.tag = ''">Clear tag filter</button>
      </div>
      <div class="row">
        <button class="btn small" type="button" @click="resetFilters">Reset filters</button>
        <span class="muted">{{ entriesQuery.data.value?.length ?? 0 }} entries shown</span>
      </div>
    </section>

    <section aria-label="Journal entries">
      <div class="panel-head">
        <h2>Entries</h2>
        <button class="btn primary" type="button" @click="openNewEntry">New entry</button>
      </div>

      <div v-if="showEntryForm" class="panel form-panel anim-pop">
        <h3 class="pixel-small">{{ editingEntry ? 'Edit entry' : 'New entry' }}</h3>
        <EntryForm
          :initial="editingEntry"
          :projects="projectsQuery.data.value ?? []"
          :saving="entrySaving"
          :error="entryError"
          @submit="saveEntry"
          @cancel="showEntryForm = false"
        />
      </div>

      <p v-if="entriesQuery.isPending.value" class="muted">Loading entries...</p>
      <p v-else-if="entriesQuery.isError.value" class="error" role="alert">
        Failed to load entries. <button class="link" type="button" @click="entriesQuery.refetch()">Try again</button>
      </p>
      <div v-else-if="(entriesQuery.data.value ?? []).length === 0" class="empty anim-fade-up">
        <p><strong>No matching entries yet.</strong></p>
        <p>Create your first entry with a prompt you just sent to AI, then rate the result.</p>
        <button class="btn primary" type="button" @click="openNewEntry">Create first entry</button>
      </div>
      <div v-else class="card-grid entry-grid">
        <div
          v-for="(entry, i) in entriesQuery.data.value"
          :key="entry.id"
          class="card-anim"
          :style="{ animationDelay: (i * 50) + 'ms' }"
        >
          <EntryCard
            :entry="entry"
            :duplicating="duplicatingId === entry.id"
            :deleting="deletingId === entry.id"
            @edit="openEditEntry"
            @duplicate="duplicateEntry"
            @delete="deleteEntry"
            @tag-click="(tag) => { filters.tag = tag }"
          />
        </div>
      </div>
    </section>

    <section class="cta-row" aria-label="Export prompt pack">
      <div class="cta-left">
        <p class="pixel-small"><span class="bullet" aria-hidden="true">►</span> Ready to export your prompt pack?</p>
        <p>Export follows the active filters above.</p>
      </div>
      <div class="cta-right">
        <button class="btn primary" type="button" :disabled="exporting" @click="exportPack('md')">Export Markdown</button>
        <button class="btn" type="button" :disabled="exporting" @click="exportPack('json')">Export JSON</button>
      </div>
    </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import EntryCard from '../components/EntryCard.vue'
import EntryForm from '../components/EntryForm.vue'
import Logo from '../components/Logo.vue'
import ProjectForm from '../components/ProjectForm.vue'
import StatsCards from '../components/StatsCards.vue'
import { useAuth } from '../composables/useAuth'
import { useDebouncedValue } from '../composables/useDebouncedValue'
import { useCreateEntry, useDeleteEntry, useDuplicateEntry, useEntries, useSummary, useUpdateEntry } from '../composables/useEntries'
import { useCreateProject, useDeleteProject, useProjects, useUpdateProject } from '../composables/useProjects'
import { api } from '../lib/api'
import { downloadBlob, projectColorHex } from '../lib/format'
import { ApiError, type Entry, type EntryFilters, type EntryInput, type Project, type ProjectInput } from '../types'

const router = useRouter()
const { user: authUser, logout } = useAuth()

const filters = reactive<EntryFilters>({
  q: '',
  project_id: 'all',
  model: '',
  tool: '',
  outcome: '',
  tag: '',
  limit: 100,
})
const debouncedQ = useDebouncedValue(computed(() => filters.q))
const effectiveFilters = computed<EntryFilters>(() => ({
  ...filters,
  q: debouncedQ.value,
  // 'none' dipetakan ke project_id=0, yang di backend berarti entri
  // tanpa proyek (project_id IS NULL).
  project_id: filters.project_id === 'none' ? '0' : filters.project_id,
}))

const projectsQuery = useProjects()
const entriesQuery = useEntries(() => effectiveFilters.value)
const summaryQuery = useSummary(() => {
  if (filters.project_id === 'all') return 'all'
  if (filters.project_id === 'none') return '0'
  return filters.project_id
})

const modelOptions = computed(() => {
  const names = (summaryQuery.data.value?.by_model ?? []).map((i) => i.name)
  return filters.model && !names.includes(filters.model) ? [...names, filters.model] : names
})
const toolOptions = computed(() => {
  const names = (summaryQuery.data.value?.by_tool ?? []).map((i) => i.name)
  return filters.tool && !names.includes(filters.tool) ? [...names, filters.tool] : names
})
const selectedProject = computed<Project | undefined>(() =>
  (projectsQuery.data.value ?? []).find((p) => String(p.id) === filters.project_id),
)

function resetFilters() {
  filters.q = ''
  filters.project_id = 'all'
  filters.model = ''
  filters.tool = ''
  filters.outcome = ''
  filters.tag = ''
}

// Proyek
const showProjectForm = ref(false)
const editingProject = ref<Project | null>(null)
const projectError = ref('')
const projectSaving = ref(false)
const createProject = useCreateProject()
const updateProject = useUpdateProject()
const deleteProjectMutation = useDeleteProject()

function openNewProject() {
  editingProject.value = null
  projectError.value = ''
  showProjectForm.value = true
}
function openEditProject(p: Project) {
  editingProject.value = p
  projectError.value = ''
  showProjectForm.value = true
}
async function saveProject(input: ProjectInput) {
  projectError.value = ''
  projectSaving.value = true
  try {
    if (editingProject.value) {
      await updateProject.mutateAsync({ id: editingProject.value.id, input })
    } else {
      await createProject.mutateAsync(input)
    }
    showProjectForm.value = false
  } catch (e) {
    projectError.value = e instanceof ApiError ? e.message : 'Failed to save project.'
  } finally {
    projectSaving.value = false
  }
}
async function deleteProject(p: Project) {
  if (!window.confirm(`Delete project "${p.name}"? Journal entries will remain saved as No project.`)) return
  try {
    await deleteProjectMutation.mutateAsync(p.id)
    filters.project_id = 'all'
  } catch (e) {
    window.alert(e instanceof ApiError ? e.message : 'Failed to delete project.')
  }
}

// Entri
const showEntryForm = ref(false)
const editingEntry = ref<Entry | null>(null)
const entryError = ref('')
const entrySaving = ref(false)
const createEntry = useCreateEntry()
const updateEntry = useUpdateEntry()
const deleteMutation = useDeleteEntry()
const duplicateMutation = useDuplicateEntry()
const duplicatingId = ref<number | null>(null)
const deletingId = ref<number | null>(null)

function openNewEntry() {
  editingEntry.value = null
  entryError.value = ''
  showEntryForm.value = true
}
function openEditEntry(entry: Entry) {
  editingEntry.value = entry
  entryError.value = ''
  showEntryForm.value = true
}
async function saveEntry(input: EntryInput) {
  entryError.value = ''
  entrySaving.value = true
  try {
    if (editingEntry.value) {
      await updateEntry.mutateAsync({ id: editingEntry.value.id, input })
    } else {
      await createEntry.mutateAsync(input)
    }
    showEntryForm.value = false
    editingEntry.value = null
  } catch (e) {
    entryError.value = e instanceof ApiError ? e.message : 'Failed to save entry.'
  } finally {
    entrySaving.value = false
  }
}
async function duplicateEntry(entry: Entry) {
  duplicatingId.value = entry.id
  try {
    await duplicateMutation.mutateAsync(entry.id)
  } catch (e) {
    window.alert(e instanceof ApiError ? e.message : 'Failed to create next iteration.')
  } finally {
    duplicatingId.value = null
  }
}
async function deleteEntry(entry: Entry) {
  if (!window.confirm(`Delete entry "${entry.title}"? This action cannot be undone.`)) return
  deletingId.value = entry.id
  try {
    await deleteMutation.mutateAsync(entry.id)
  } catch (e) {
    window.alert(e instanceof ApiError ? e.message : 'Failed to delete entry.')
  } finally {
    deletingId.value = null
  }
}

// Ekspor
const exporting = ref(false)
async function exportPack(format: 'md' | 'json') {
  exporting.value = true
  try {
    const { blob, filename } = await api.exportEntries(effectiveFilters.value, format)
    downloadBlob(blob, filename)
  } catch (e) {
    window.alert(e instanceof ApiError ? e.message : 'Failed to export entries.')
  } finally {
    exporting.value = false
  }
}

function onLogout() {
  logout()
  void router.push('/')
}
</script>

<style scoped>
.card-anim {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}
.empty {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}
.dash-inner {
  padding: 0 1.25rem 1.25rem;
}
.topbar {
  margin: 0 -1.25rem;
  padding-left: 1.25rem;
  padding-right: 1.25rem;
}
.dash-hero {
  padding-top: 1.5rem;
}
.dash-crumb {
  margin: 0;
}
.intro {
  margin: 0.8rem 0 0;
  color: #5c5c54;
  max-width: 36rem;
}
.panel {
  background: #fff;
  border: 1px solid #111;
  border-radius: 4px;
  padding: 1rem;
  margin: 1rem 0;
}
.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
.panel-head h2 {
  margin: 0;
  font-size: 1.15rem;
}
.project-count {
  font-family: 'Press Start 2P', monospace;
  font-size: 1.1rem;
  margin-left: 0.3rem;
}
.project-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0.75rem 0;
}
.filters {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 0.9rem;
}
.filters .field {
  margin-bottom: 0;
}
.tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
  margin: 0.9rem 0;
}
.entry-grid {
  margin-top: 1rem;
}
.empty {
  background: #fff;
  border: 1px dashed #9a9a90;
  border-radius: 4px;
  padding: 1.5rem;
  text-align: center;
  margin-top: 1rem;
}
.form-panel {
  background: #fff;
}
.form-panel h3 {
  margin: 0 0 1rem;
}
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
  .filters {
    grid-template-columns: 1fr;
  }
  .cta-row {
    grid-template-columns: 1fr;
  }
}
</style>
