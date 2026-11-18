import { useQuery } from '@tanstack/vue-query'
import { api } from '../lib/api'

// Fitur opsional (GitHub, magic link, pendaftaran) dibaca sekali dari server
// supaya UI tidak menampilkan tombol yang belum dikonfigurasi.
export function useInstanceMeta() {
  return useQuery({
    queryKey: ['instance-meta'],
    queryFn: () => api.meta(),
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}
