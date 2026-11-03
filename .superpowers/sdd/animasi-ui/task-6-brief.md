# Task 6 Brief: Dashboard form panel animations

**File:**
- Modify: `frontend/src/pages/DashboardPage.vue`

## Changes

### 1. Form panel pop-in animation

In the template, change:
```html
<div v-if="showEntryForm" class="panel form-panel">
```
to:
```html
<div v-if="showEntryForm" class="panel form-panel anim-pop">
```

Wrap the ProjectForm with a div that has the pop animation. Change:
```html
<ProjectForm
  v-if="showProjectForm"
  ...
/>
```
to:
```html
<div v-if="showProjectForm" class="anim-pop">
  <ProjectForm ... />
</div>
```

### 2. Entry cards stagger

The `v-for` for `EntryCard` is:
```html
<EntryCard
  v-for="entry in entriesQuery.data.value"
  :key="entry.id"
  ...
/>
```

Wrap each card with a div that has stagger animation. Change to:
```html
<div
  v-for="(entry, i) in entriesQuery.data.value"
  :key="entry.id"
  class="card-anim"
  :style="{ animationDelay: (i * 50) + 'ms' }"
>
  <EntryCard
    :entry="entry"
    ...
  />
</div>
```

Note: this changes the `key` from `entry.id` to the div wrapper, so remove `:key="entry.id"` from EntryCard and keep it on the div.

### 3. Empty state fade

Add `class="anim-fade-up"` to the empty state div:
```html
<div v-else-if="(entriesQuery.data.value ?? []).length === 0" class="empty anim-fade-up">
```

### 4. Headline dash entrance animation

The headline `<h1 class="pixel-headline">Jurnal_</h1>` gets `class="hero-title"` and the underscore blink via the `.hero-title::after` rule (already defined globally from Task 3, so no scoped style needed for the blink itself).

Add class `hero-title`:
```html
<h1 class="pixel-headline hero-title">Jurnal_</h1>
```

### 5. Scoped style additions

Add to the existing `<style scoped>` block:

```css
.card-anim {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}
.empty {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}
```

## Verification

Run `npm run typecheck` in `frontend/`. Must pass.

## Commit

```
git add frontend/src/pages/DashboardPage.vue
git commit -m "feat(ui): animate dashboard form panels, entry cards stagger, and headline"
```