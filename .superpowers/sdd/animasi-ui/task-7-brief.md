# Task 7 Brief: Prompt box transition in EntryCard

**File:**
- Modify: `frontend/src/components/EntryCard.vue`

## Changes

### 1. Replace v-show with Transition

The current code has:
```html
<button class="prompt-toggle" type="button" @click="promptOpen = !promptOpen">
  {{ promptOpen ? 'Sembunyikan prompt' : 'Lihat prompt' }}
</button>
<div v-show="promptOpen" class="prompt-box">
  <pre>{{ entry.prompt_text }}</pre>
  <button class="btn small" type="button" @click="copyPrompt">{{ copied ? 'Prompt tersalin' : 'Salin prompt' }}</button>
</div>
```

Change to:
```html
<button class="prompt-toggle" type="button" @click="promptOpen = !promptOpen">
  {{ promptOpen ? 'Sembunyikan prompt' : 'Lihat prompt' }}
</button>
<Transition name="prompt">
  <div v-if="promptOpen" class="prompt-box">
    <pre>{{ entry.prompt_text }}</pre>
    <button class="btn small" type="button" @click="copyPrompt">{{ copied ? 'Prompt tersalin' : 'Salin prompt' }}</button>
  </div>
</Transition>
```

Note: `v-show` was used for instant show/hide. Now using `v-if` inside `<Transition>` for animated show/hide.

### 2. Add transition CSS

Add to the existing `<style scoped>` block:

```css
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
```

Note: The arrow-cell already has `background: var(--orange)` on hover from global CSS. The scoped transition here adds smooth color interpolation.

## Verification

Run `npm run typecheck` in `frontend/`. Must pass.

## Commit

```
git add frontend/src/components/EntryCard.vue
git commit -m "feat(ui): animate prompt box expand/collapse on entry cards"
```
