# Task 5 Brief: Count-up composable + StatsCards

**Files:**
- Create: `frontend/src/composables/useCountUp.ts`
- Modify: `frontend/src/components/StatsCards.vue`

## Step 1: Create useCountUp composable

Create `frontend/src/composables/useCountUp.ts` with this EXACT content:

```ts
import { ref, onMounted, watch } from 'vue'
import type { Ref } from 'vue'

export function useCountUp(target: number, duration = 800): Ref<number> {
  const value = ref(0)
  let rafId: number | null = null

  function animate() {
    const start = performance.now()
    function step(now: number) {
      const elapsed = now - start
      const progress = Math.min(elapsed / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      value.value = Math.round(eased * target)
      if (progress < 1) {
        rafId = requestAnimationFrame(step)
      }
    }
    rafId = requestAnimationFrame(step)
  }

  onMounted(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      value.value = target
      return
    }
    animate()
  })

  return value
}
```

## Step 2: Modify StatsCards.vue

In the `<script setup>` block, replace `defineProps` with:

```ts
import { computed } from 'vue'
import { useCountUp } from '../composables/useCountUp'
import type { Summary } from '../types'

const props = defineProps<{ summary?: Summary; loading: boolean }>()

const total = useCountUp(props.summary?.total ?? 0)
const success = useCountUp(props.summary?.success ?? 0)
const partial = useCountUp(props.summary?.partial ?? 0)
const failed = useCountUp(props.summary?.failed ?? 0)
```

In the template, change:
- `{{ summary.total }}` → `{{ total }}`
- `{{ summary.success }}` → `{{ success }}`
- `{{ summary.partial }}` → `{{ partial }}`
- `{{ summary.failed }}` → `{{ failed }}`

Note: keep the average_rating and rated_count lines unchanged (they stay as `summary.average_rating` etc).

Add to the `<style scoped>` block:

```css
.stats-strip {
  animation: fade-up var(--dur-mid) var(--ease-out) both;
}
```

## Verification

Run `npm run typecheck` in `frontend/`. Must pass with no errors.

## Commit

```
git add frontend/src/composables/useCountUp.ts frontend/src/components/StatsCards.vue
git commit -m "feat(ui): add count-up animation for stats cards numbers"
```
