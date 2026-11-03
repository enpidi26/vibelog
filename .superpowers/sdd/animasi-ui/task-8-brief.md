# Task 8 Brief: Logo wiggle animation

**File:**
- Modify: `frontend/src/components/Logo.vue`

## Changes

Add to the existing `<style scoped>` block:

```css
.logo:hover .logo-mark {
  animation: wiggle 0.4s ease-in-out;
}
```

`wiggle` keyframe was added globally in Task 1.

## Verification

Run `npm run typecheck` in `frontend/`. Must pass.

## Commit

```
git add frontend/src/components/Logo.vue
git commit -m "feat(ui): add wiggle animation on logo hover"
```