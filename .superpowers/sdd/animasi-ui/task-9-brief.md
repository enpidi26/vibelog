# Task 9 Brief: Auth shell + NotFound animations

**Files:**
- Modify: `frontend/src/components/AuthShell.vue`
- Modify: `frontend/src/pages/NotFoundPage.vue`

## AuthShell.vue changes

Add class `anim-pop` to the `<main>` element. Change:
```html
<main class="auth-main">
```
to:
```html
<main class="auth-main anim-pop">
```

Add to the existing `<style scoped>` block:
```css
.auth-main {
  animation: pop-in var(--dur-mid) var(--ease-spring) both;
}
```

Note: `anim-pop` is the global utility class from Task 1. The scoped `auth-main` animation acts as a secondary entrance.

## NotFoundPage.vue changes

Add class `anim-pop` to the `<main>` element. Change:
```html
<main class="not-found">
```
to:
```html
<main class="not-found anim-pop">
```

Add to the existing `<style scoped>` block:
```css
.not-found {
  animation: pop-in var(--dur-mid) var(--ease-spring) both;
}
```

## Verification

Run `npm run typecheck` in `frontend/`. Must pass.

## Commit

```
git add frontend/src/components/AuthShell.vue frontend/src/pages/NotFoundPage.vue
git commit -m "feat(ui): add pop-in animation on auth pages and 404"
```