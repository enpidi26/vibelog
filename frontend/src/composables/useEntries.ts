import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, toValue, type Ref } from 'vue'
import { api } from '../lib/api'
import type { EntryFilters, EntryInput, Summary } from '../types'

export const entriesKey = (filters: EntryFilters) =>
  ['entries', filters.q, filters.project_id, filters.model, filters.tool, filters.outcome, filters.tag, filters.limit] as const

export type FilterSource = EntryFilters | Ref<EntryFilters> | (() => EntryFilters)

export function useEntries(filters: FilterSource) {
  return useQuery({
    queryKey: computed(() => entriesKey(toValue(filters))),
    queryFn: () => api.listEntries(toValue(filters)),
  })
}

export function useSummary(projectId: string | Ref<string> | (() => string)) {
  return useQuery<Summary>({
    queryKey: computed(() => ['summary', toValue(projectId)] as const),
    queryFn: () => api.summary(toValue(projectId)),
  })
}

function invalidateAll(queryClient: ReturnType<typeof useQueryClient>) {
  void queryClient.invalidateQueries({ queryKey: ['entries'] })
  void queryClient.invalidateQueries({ queryKey: ['summary'] })
  void queryClient.invalidateQueries({ queryKey: ['projects'] })
}

export function useCreateEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: EntryInput) => api.createEntry(input),
    onSuccess: () => invalidateAll(queryClient),
  })
}

export function useUpdateEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: EntryInput }) => api.updateEntry(id, input),
    onSuccess: () => invalidateAll(queryClient),
  })
}

export function useDeleteEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.deleteEntry(id),
    onSuccess: () => invalidateAll(queryClient),
  })
}

export function useDuplicateEntry() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.duplicateEntry(id),
    onSuccess: () => invalidateAll(queryClient),
  })
}
