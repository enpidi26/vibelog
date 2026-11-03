# Task 9 Report: Auth shell + NotFound animations

## Status: DONE

## Commit: 05c3f7d23d5016a7471dcebf26a0a84b74ce1a17

## Changes Applied

### AuthShell.vue
- Added `anim-pop` class to `<main class="auth-main anim-pop">`
- Added scoped CSS animation: `.auth-main { animation: pop-in var(--dur-mid) var(--ease-spring) both; }`

### NotFoundPage.vue
- Added `anim-pop` class to `<main class="not-found anim-pop">`
- Added scoped CSS animation: `.not-found { animation: pop-in var(--dur-mid) var(--ease-spring) both; }`

## Verification

- TypeScript typecheck: **PASS**

## Test Summary

```
2 files changed, 4 insertions(+), 2 deletions(-)
```

## Concerns

None.
