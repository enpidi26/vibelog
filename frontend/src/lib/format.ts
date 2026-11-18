export const OUTCOME_LABELS: Record<string, string> = {
  success: 'Success',
  partial: 'Partial',
  failed: 'Failed',
  in_progress: 'In Progress',
}

export const MODEL_SUGGESTIONS = ['Claude', 'GPT-4', 'ChatGPT', 'Gemini', 'Codex']
export const TOOL_SUGGESTIONS = ['Cursor', 'Claude Code', 'Copilot', 'Windsurf', 'VS Code']
export const PROJECT_COLORS = ['teal', 'amber', 'violet', 'rose', 'sky', 'lime', 'slate']

export const PROJECT_COLOR_HEX: Record<string, string> = {
  teal: '#0f766e',
  amber: '#d97706',
  violet: '#7c3aed',
  rose: '#e11d48',
  sky: '#0284c7',
  lime: '#65a30d',
  slate: '#475569',
}

export function projectColorHex(color: string): string {
  return PROJECT_COLOR_HEX[color] ?? '#0f766e'
}

export function outcomeLabel(outcome: string): string {
  return OUTCOME_LABELS[outcome] ?? outcome
}

export function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  // Waktu lokal pembaca. Self-host bisa berada di zona waktu mana pun.
  return d.toLocaleString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  // Give the browser time to start the download before dropping the blob.
  setTimeout(() => {
    a.remove()
    URL.revokeObjectURL(url)
  }, 1000)
}
