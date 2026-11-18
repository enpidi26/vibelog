<template>
  <header class="topbar">
    <Logo />
    <p class="micro-label dash-crumb">Journal / Dashboard</p>
    <div class="top-actions">
      <span v-if="userName" class="user-name">{{ userName }}</span>
      <button class="btn small" type="button" :disabled="exporting" title="Uses the active filters" @click="$emit('export', 'md')">
        {{ exporting ? 'Exporting...' : 'Export Markdown' }}
      </button>
      <button class="btn small" type="button" :disabled="exporting" title="Uses the active filters" @click="$emit('export', 'json')">Export JSON</button>
      <button class="btn small" type="button" @click="$emit('signout')">Sign out</button>
    </div>
  </header>
</template>

<script setup lang="ts">
import Logo from './Logo.vue'

defineProps<{ userName?: string; exporting: boolean }>()
defineEmits<{ export: [format: 'md' | 'json']; signout: [] }>()
</script>

<style scoped>
.topbar {
  margin: 0 -1.25rem;
  padding-left: 1.25rem;
  padding-right: 1.25rem;
}
.dash-crumb {
  margin: 0;
}
</style>
