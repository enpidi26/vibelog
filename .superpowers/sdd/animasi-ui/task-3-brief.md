# Task 3 Brief: Stagger landing cards + blink headline

**File:**
- Modify: `frontend/src/pages/LandingPage.vue`

## Changes

### 1. Blink on headline underscore

The headline `<h1 class="pixel-headline">Vibe_Log</h1>` already has an underscore in the text. Add `class="hero-title"` to it, then add CSS to animate the underscore blink.

### 2. Stagger cards entrance

The `v-for` loop for feature cards already exists. Add inline style with animation delay based on index.

### 3. Card hover lift

### 4. Stack pills hover lift

## Exact code changes

### Template changes

In `<h1 class="pixel-headline">`, add `class="hero-title"`:
```html
<h1 class="pixel-headline hero-title">Vibe_Log</h1>
```

Change the article tag from:
```html
<article v-for="card in cards" :key="card.title" class="ref-card">
```
to:
```html
<article v-for="(card, i) in cards" :key="card.title" class="ref-card card-anim" :style="{ animationDelay: (i * 70) + 'ms' }">
```

### Scoped style additions

Add these rules to the existing `<style scoped>` block:

```css
.hero-title {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
  display: inline-block;
}
.hero-title::after {
  content: '_';
  animation: blink 1.1s step-end infinite;
  color: var(--orange);
}

.card-anim {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}

.card-anim:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 0 var(--line-strong);
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
}

.stack-row li {
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
}
.stack-row li:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 0 var(--line-strong);
}
```

## Verification

Run `npm run typecheck` in `frontend/`. Must pass.

## Commit

```
git add frontend/src/pages/LandingPage.vue
git commit -m "feat(ui): add stagger cards + blink headline on landing page"
```
