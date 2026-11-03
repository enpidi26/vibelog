# Task 1 Brief: Fondasi animasi global di main.css

**Files:**
- Modify: `frontend/src/assets/main.css`

## Exact values to use verbatim

### Variables (add inside `:root` block, after `--muted: #5c5c54;`)

```
  --ease-out: cubic-bezier(0.22, 1, 0.36, 1);
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
  --dur-fast: 140ms;
  --dur-mid: 240ms;
```

### Keyframes (add after the `:root` block closes)

```css
@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes fade-up {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes pop-in {
  0% { opacity: 0; transform: scale(0.94) translateY(6px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}
@keyframes blink {
  0%, 45% { opacity: 1; }
  55%, 100% { opacity: 0; }
}
@keyframes wiggle {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(-6deg); }
  75% { transform: rotate(6deg); }
}
@keyframes pulse-soft {
  0%, 100% { opacity: 0.55; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.06); }
}
```

### Utility classes (after the keyframes)

```css
.anim-fade-up {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}
.anim-pop {
  animation: pop-in var(--dur-mid) var(--ease-spring) both;
}
```

### Buttons

Add to the EXISTING `.btn` rule:
```
  transition:
    transform var(--dur-fast) var(--ease-out),
    box-shadow var(--dur-fast) var(--ease-out),
    background-color var(--dur-fast) var(--ease-out);
```

Change EXISTING `.btn:hover:not(:disabled)` to:
```css
.btn:hover:not(:disabled) {
  background: #ecece4;
  transform: translateY(-1px);
  box-shadow: 0 2px 0 #c9c9bf;
}
```

Add new rule AFTER it:
```css
.btn:active:not(:disabled) {
  transform: translateY(1px);
  box-shadow: none;
}
```

Add to EXISTING `.btn.primary:hover:not(:disabled)`: `box-shadow: 0 2px 0 #b83c0c;`

Add new rule after it:
```css
.btn.primary:active:not(:disabled) {
  transform: translateY(1px);
  box-shadow: none;
}
```

### Chips and tags

Add to EXISTING `.chip`:
```
  transition: background-color var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out), transform var(--dur-fast) var(--ease-out);
```

Add new rules after `.chip.active`:
```css
.chip:hover {
  transform: translateY(-1px);
}
.chip:active {
  transform: translateY(1px);
}
```

Add to EXISTING `.tag`:
```
  transition: background-color var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out), transform var(--dur-fast) var(--ease-out);
```

Add new rules after `.tag.active`:
```css
.tag:hover {
  transform: translateY(-1px);
}
.tag:active {
  transform: translateY(1px);
}
```

### Field focus

Add to EXISTING `.field input, .field select, .field textarea` rule:
```
  transition: border-color var(--dur-fast) var(--ease-out), outline-color var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out);
```

### Reduced motion guard (at end of file)

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
  .btn:hover:not(:disabled),
  .btn:active:not(:disabled),
  .chip:hover,
  .chip:active,
  .tag:hover,
  .tag:active {
    transform: none;
    box-shadow: none;
  }
}
```

## Verification

Run in `frontend/`: `npm run typecheck`
Expected: PASS.

## Commit

```
git add frontend/src/assets/main.css
git commit -m "feat(ui): add global animation foundation (keyframes, easing vars, button/chip press effects)"
```
