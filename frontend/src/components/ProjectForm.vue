<template>
  <form class="inline-form" @submit.prevent="onSubmit">
    <div class="field">
      <label for="project-name">Nama proyek</label>
      <input id="project-name" v-model="name" type="text" maxlength="80" required placeholder="Toko online" />
    </div>
    <div class="field">
      <label for="project-desc">Deskripsi (opsional)</label>
      <input id="project-desc" v-model="description" type="text" maxlength="500" placeholder="Apa yang sedang dibangun di proyek ini?" />
    </div>
    <div class="field">
      <label for="project-color">Warna</label>
      <select id="project-color" v-model="color">
        <option v-for="c in PROJECT_COLORS" :key="c" :value="c">{{ c }}</option>
      </select>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="row">
      <button class="btn primary" type="submit" :disabled="saving">{{ saving ? 'Menyimpan...' : initial ? 'Simpan perubahan' : 'Buat proyek' }}</button>
      <button class="btn" type="button" @click="$emit('cancel')">Batal</button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { PROJECT_COLORS } from '../lib/format'
import type { Project, ProjectInput } from '../types'

const props = defineProps<{ initial?: Project | null; saving: boolean; error: string }>()
const emit = defineEmits<{ submit: [input: ProjectInput]; cancel: [] }>()

const name = ref(props.initial?.name ?? '')
const description = ref(props.initial?.description ?? '')
const color = ref(props.initial?.color ?? 'teal')

function onSubmit() {
  emit('submit', { name: name.value.trim(), description: description.value.trim(), color: color.value })
}
</script>
