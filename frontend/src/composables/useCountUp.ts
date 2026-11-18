import { onMounted, onUnmounted, ref, watch } from 'vue'
import type { Ref } from 'vue'

function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * Menganimasikan angka menuju nilai ref sumber. Nilai sumber boleh datang
 * belakangan (mis. setelah query selesai): setiap perubahan memicu animasi
 * baru dari nilai yang sedang tampil, bukan dari 0.
 */
export function useCountUp(source: Ref<number>, duration = 800): Ref<number> {
  const value = ref(0)
  let rafId: number | null = null

  function stop(): void {
    if (rafId !== null) {
      cancelAnimationFrame(rafId)
      rafId = null
    }
  }

  function animate(target: number): void {
    stop()
    const from = value.value
    if (from === target) return
    const start = performance.now()

    function step(now: number): void {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      value.value = Math.round(from + (target - from) * eased)
      if (progress < 1) {
        rafId = requestAnimationFrame(step)
      } else {
        rafId = null
      }
    }
    rafId = requestAnimationFrame(step)
  }

  function apply(target: number): void {
    if (prefersReducedMotion()) {
      stop()
      value.value = target
      return
    }
    animate(target)
  }

  onMounted(() => apply(source.value))
  watch(source, (target) => apply(target))
  onUnmounted(stop)

  return value
}
