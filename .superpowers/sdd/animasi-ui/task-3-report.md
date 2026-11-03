# Task 3 Report: Stagger landing cards + blink headline

## Status

**DONE**

## Commit

```
7f5fdb8 feat(ui): add stagger cards + blink headline on landing page
```

## Summary

Applied 3 changes to `frontend/src/pages/LandingPage.vue`:

1. **Blink on headline underscore** — Added `class="hero-title"` to `<h1>` and added CSS for `::after` pseudo-element with `blink` animation (1.1s step-end infinite) in `--orange` color, plus `fade-up` entrance animation.

2. **Stagger cards entrance** — Changed `v-for="card in cards"` to `v-for="(card, i) in cards"` and added `:style="{ animationDelay: (i * 70) + 'ms' }"` for a 70ms stagger per card, plus `card-anim` class.

3. **Card hover lift** — `.card-anim:hover` translates Y -3px with a 0 4px 0 box-shadow on `--line-strong`, with fast ease-out transitions.

4. **Stack pills hover lift** — `.stack-row li` has a transition and `.stack-row li:hover` translates Y -2px with a 0 2px 0 box-shadow.

## Verification

- `npm run typecheck` in `frontend/` — passed (no errors)
- `git add + git commit` — committed successfully

## Diff

```diff
- <h1 class="pixel-headline">Vibe_Log</h1>
+ <h1 class="pixel-headline hero-title">Vibe_Log</h1>

- <article v-for="card in cards" :key="card.title" class="ref-card">
+ <article v-for="(card, i) in cards" :key="card.title" class="ref-card card-anim" :style="{ animationDelay: (i * 70) + 'ms' }">
```

Plus scoped CSS additions for `.hero-title`, `.hero-title::after`, `.card-anim`, `.card-anim:hover`, `.stack-row li`, `.stack-row li:hover`.