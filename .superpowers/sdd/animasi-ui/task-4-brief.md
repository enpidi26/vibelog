# Task 4 Brief: Hero-grid + plus motif animation

**File:**
- Modify: `frontend/src/pages/LandingPage.vue` (same file as Task 3)

## Changes

This task adds entrance animation to the hero-grid panel and a gentle pulse to the plus motif (+) in the hero.

### Template changes

Add `class="anim-fade-up"` and `style="animation-delay: 80ms"` to the `.hero-grid` div.

Change this:
```html
<span class="plus-motif" aria-hidden="true">+</span>
```
to:
```html
<span class="plus-motif pulse-motif" aria-hidden="true">+</span>
```

### Scoped style additions

Add to the existing `<style scoped>` block:

```css
.hero-grid {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}

.pulse-motif {
  animation: pulse-soft 2.6s ease-in-out infinite;
}
```

Note: `anim-fade-up` is a global class added in Task 1, so no scoped prefix needed.
`pulse-soft` keyframe was also added in Task 1.

## Verification

Run `npm run typecheck` in `frontend/`. Must pass.

## Commit

```
git add frontend/src/pages/LandingPage.vue
git commit -m "feat(ui): animate hero-grid entrance and plus motif pulse"
```