import { ref, watch, type Ref } from 'vue'

/** Ref yang nilainya mengikuti `source`, tapi tertunda `delay` ms setelah
 *  perubahan terakhir. Dipakai untuk kolom cari supaya request tidak
 *  dikirim tiap ketikan. */
export function useDebouncedValue<T>(source: Ref<T>, delay = 300): Ref<T> {
  const debounced = ref(source.value) as Ref<T>
  let timer: ReturnType<typeof setTimeout> | undefined
  watch(source, (value) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      debounced.value = value
    }, delay)
  })
  return debounced
}
