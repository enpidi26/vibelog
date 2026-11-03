<template>
  <div class="frame-page">
    <div class="frame">
      <div class="dash-inner">
    <header class="topbar">
      <Logo />
      <p class="micro-label dash-crumb">Jurnal / Dashboard</p>
      <div class="top-actions">
        <span v-if="authUser" class="user-name">{{ authUser.name }}</span>
        <button class="btn small" type="button" :disabled="exporting" @click="exportPack('md')">
          {{ exporting ? 'Mengekspor...' : 'Ekspor Markdown' }}
        </button>
        <button class="btn small" type="button" :disabled="exporting" @click="exportPack('json')">Ekspor JSON</button>
        <button class="btn small" type="button" @click="onLogout">Keluar</button>
      </div>
    </header>

    <section class="dash-hero">
      <h1 class="pixel-headline">Jurnal_</h1>
      <p class="intro">Catat prompt yang kamu kirim, nilai hasilnya, dan ulangi yang berhasil di iterasi berikutnya.</p>
    </section>

    <StatsCards :summary="summaryQuery.data.value" :loading="summaryQuery.isPending.value" />

    <section class="panel" aria-label="Proyek">
      <div class="panel-head">
        <h2>Proyek <span class="project-count">{{ String((projectsQuery.data.value ?? []).length).padStart(2, '0') }}</span></h2>
        <button class="btn small" type="button" @click="openNewProject">Proyek baru</button>
      </div>
      <p v-if="projectsQuery.isPending.value" class="muted">Memuat proyek...</p>
      <p v-else-if="projectsQuery.isError.value" class="error" role="alert">Gagal memuat proyek. <button class="link" type="button" @click="projectsQuery.refetch()">Coba lagi</button></p>
      <div v-else class="project-chips">
        <button
          class="chip"
          :class="{ active: filters.project_id === 'all' }"
          type="button"
          @click="filters.project_id = 'all'"
        >
          Semua proyek
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
          Tanpa proyek
        </button>
      </div>
      <p class="hint">Catatan: menghapus proyek tidak menghapus entri jurnalnya. Entri pindah ke Tanpa proyek.</p>
      <div v-if="selectedProject" class="row">
        <button class="btn small" type="button" @click="openEditProject(selectedProject)">Ubah proyek ini</button>
        <button class="btn small danger" type="button" @click="deleteProject(selectedProject)">Hapus proyek ini</button>
      </div>
      <ProjectForm
        v-if="showProjectForm"
        :initial="editingProject"
        :saving="projectSaving"
        :error="projectError"
        @submit="saveProject"
        @cancel="showProjectForm = false"
      />
    </section>

    <section class="panel" aria-label="Filter entri">
      <div class="filters">
        <div class="field">
          <label for="filter-q">Cari judul, prompt, catatan, atau tag</label>
          <input id="filter-q" v-model="filters.q" type="search" placeholder="Contoh: auth, refactor..." />
        </div>
        <div class="field">
          <label for="filter-model">Model</label>
          <select id="filter-model" v-model="filters.model">
            <option value="">Semua model</option>
            <option v-for="m in modelOptions" :key="m" :value="m">{{ m }}</option>
          </select>
        </div>
        <div class="field">
          <label for="filter-tool">Tool</label>
          <select id="filter-tool" v-model="filters.tool">
            <option value="">Semua tool</option>
            <option v-for="t in toolOptions" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
        <div class="field">
          <label for="filter-outcome">Hasil</label>
          <select id="filter-outcome" v-model="filters.outcome">
            <option value="">Semua hasil</option>
            <option value="success">Berhasil</option>
            <option value="partial">Sebagian</option>
            <option value="failed">Gagal</option>
            <option value="in_progress">Berjalan</option>
          </select>
        </div>
      </div>
      <div v-if="summaryQuery.data.value?.top_tags.length" class="tags-row">
        <span class="muted">Tag populer:</span>
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
        <button v-if="filters.tag" class="link" type="button" @click="filters.tag = ''">Hapus filter tag</button>
      </div>
      <div class="row">
        <button class="btn small" type="button" @click="resetFilters">Reset filter</button>
        <span class="muted">{{ entriesQuery.data.value?.length ?? 0 }} entri ditampilkan</span>
      </div>
    </section>

    <section aria-label="Entri jurnal">
      <div class="panel-head">
        <h2>Entri</h2>
        <button class="btn primary" type="button" @click="openNewEntry">Entri baru</button>
      </div>

      <div v-if="showEntryForm" class="panel form-panel">
        <h3 class="pixel-small">{{ editingEntry ? 'Ubah entri' : 'Entri baru' }}</h3>
        <EntryForm
          :initial="editingEntry"
          :projects="projectsQuery.data.value ?? []"
          :saving="entrySaving"
          :error="entryError"
          @submit="saveEntry"
          @cancel="showEntryForm = false"
        />
      </div>

      <p v-if="entriesQuery.isPending.value" class="muted">Memuat entri...</p>
      <p v-else-if="entriesQuery.isError.value" class="error" role="alert">
        Gagal memuat entri. <button class="link" type="button" @click="entriesQuery.refetch()">Coba lagi</button>
      </p>
      <div v-else-if="(entriesQuery.data.value ?? []).length === 0" class="empty">
        <p><strong>Belum ada entri yang cocok.</strong></p>
        <p>Buat entri pertama dengan prompt yang baru kamu kirim ke AI, lalu nilai hasilnya.</p>
        <button class="btn primary" type="button" @click="openNewEntry">Buat entri pertama</button>
      </div>
      <div v-else class="card-grid entry-grid">
        <EntryCard
          v-for="entry in entriesQuery.data.value"
          :key="entry.id"
          :entry="entry"
          :duplicating="duplicatingId === entry.id"
          :deleting="deletingId === entry.id"
          @edit="openEditEntry"
          @duplicate="duplicateEntry"
          @delete="deleteEntry"
          @tag-click="(tag) => { filters.tag = tag }"
        />
      </div>
    </section>

    <section class="cta-row" aria-label="Ekspor prompt pack">
      <div class="cta-left">
        <p class="pixel-small"><span class="bullet" aria-hidden="true">►</span> Siap mengekspor prompt pack?</p>
        <p>Ekspor mengikuti filter yang sedang aktif di atas.</p>
      </div>
      <div class="cta-right">
        <button class="btn primary" type="button" :disabled="exporting" @click="exportPack('md')">Ekspor Markdown</button>
        <button class="btn" type="button" :disabled="exporting" @click="exportPack('json')">Ekspor JSON</button>
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
    projectError.value = e instanceof ApiError ? e.message : 'Gagal menyimpan proyek.'
  } finally {
    projectSaving.value = false
  }
}
async function deleteProject(p: Project) {
  if (!window.confirm(`Hapus proyek "${p.name}"? Entri jurnalnya tetap tersimpan sebagai Tanpa proyek.`)) return
  try {
    await deleteProjectMutation.mutateAsync(p.id)
    filters.project_id = 'all'
  } catch (e) {
    window.alert(e instanceof ApiError ? e.message : 'Gagal menghapus proyek.')
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
    entryError.value = e instanceof ApiError ? e.message : 'Gagal menyimpan entri.'
  } finally {
    entrySaving.value = false
  }
}
async function duplicateEntry(entry: Entry) {
  duplicatingId.value = entry.id
  try {
    await duplicateMutation.mutateAsync(entry.id)
  } catch (e) {
    window.alert(e instanceof ApiError ? e.message : 'Gagal membuat iterasi berikutnya.')
  } finally {
    duplicatingId.value = null
  }
}
async function deleteEntry(entry: Entry) {
  if (!window.confirm(`Hapus entri "${entry.title}"? Tindakan ini tidak bisa dibatalkan.`)) return
  deletingId.value = entry.id
  try {
    await deleteMutation.mutateAsync(entry.id)
  } catch (e) {
    window.alert(e instanceof ApiError ? e.message : 'Gagal menghapus entri.')
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
    window.alert(e instanceof ApiError ? e.message : 'Gagal mengekspor entri.')
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
