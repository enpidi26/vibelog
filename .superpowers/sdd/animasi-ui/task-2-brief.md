# Task 2 Brief: Transisi halaman di App.vue

**File:**
- Modify: `frontend/src/App.vue`

## Changes

Replace the ENTIRE `<template>` and add a `<style>` block:

```vue
<template>
  <RouterView v-slot="{ Component, route }">
    <Transition name="page" mode="out-in">
      <component :is="Component" :key="route.path" />
    </Transition>
  </RouterView>
</template>
```

```css
<style>
.page-enter-active,
.page-leave-active {
  transition: opacity 200ms var(--ease-out), transform 200ms var(--ease-out);
}
.page-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
.page-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
```

Existing `<script setup>` block remains unchanged.

## Verification

Run `npm run typecheck` in `frontend/`. Must pass.

## Commit

```
git add frontend/src/App.vue
git commit -m "feat(ui): add page transition (fade + slide on route change)"
```
