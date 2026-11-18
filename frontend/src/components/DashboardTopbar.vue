<template>
  <header class="topbar">
    <Logo />
    <div class="top-actions">
      <span v-if="userName" class="user-name">{{ userName }}</span>
      <details ref="exportMenu" class="export-menu">
        <summary class="btn small" :class="{ disabled: exporting }">Export</summary>
        <div class="export-options">
          <button type="button" :disabled="exporting" @click="choose('md')">Markdown</button>
          <button type="button" :disabled="exporting" @click="choose('json')">JSON</button>
        </div>
      </details>
      <button class="btn small" type="button" @click="$emit('signout')">Sign out</button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Logo from './Logo.vue'

defineProps<{ userName?: string; exporting: boolean }>()
const emit = defineEmits<{ export: [format: 'md' | 'json']; signout: [] }>()

const exportMenu = ref<HTMLDetailsElement | null>(null)

function choose(format: 'md' | 'json') {
  emit('export', format)
  if (exportMenu.value) exportMenu.value.open = false
}
</script>

<style scoped>
.topbar {
  margin: 0 -1.25rem;
  padding-left: 1.25rem;
  padding-right: 1.25rem;
}
.export-menu {
  position: relative;
}
.export-menu summary {
  list-style: none;
}
.export-menu summary::-webkit-details-marker {
  display: none;
}
.export-menu summary.disabled {
  opacity: 0.55;
  pointer-events: none;
}
.export-options {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  z-index: 20;
  display: flex;
  flex-direction: column;
  min-width: 128px;
  overflow: hidden;
  background: #fff;
  border: 1px solid #111;
  border-radius: 6px;
}
.export-options button {
  padding: 0.5rem 0.75rem;
  border: 0;
  background: #fff;
  color: #111;
  font: inherit;
  text-align: left;
  cursor: pointer;
}
.export-options button:hover {
  background: #f1f1ea;
}
@media (max-width: 620px) {
  .user-name {
    display: none;
  }
}
</style>
