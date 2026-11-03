import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { api } from '../lib/api'
import type { ProjectInput } from '../types'

export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: () => api.listProjects(),
  })
}

function invalidate(queryClient: ReturnType<typeof useQueryClient>) {
  void queryClient.invalidateQueries({ queryKey: ['projects'] })
  void queryClient.invalidateQueries({ queryKey: ['entries'] })
  void queryClient.invalidateQueries({ queryKey: ['summary'] })
}

export function useCreateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: ProjectInput) => api.createProject(input),
    onSuccess: () => invalidate(queryClient),
  })
}

export function useUpdateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: ProjectInput }) => api.updateProject(id, input),
    onSuccess: () => invalidate(queryClient),
  })
}

export function useDeleteProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.deleteProject(id),
    onSuccess: () => invalidate(queryClient),
  })
}
