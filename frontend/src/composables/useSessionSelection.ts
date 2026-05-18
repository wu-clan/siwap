import { onBeforeUnmount, onMounted, watch, type ComputedRef, type Ref } from 'vue'
import type { Session } from '../domain/types'

export function useSessionSelection(options: {
  displayedSessions: ComputedRef<Session[]>
  selectedSession: ComputedRef<Session | undefined>
  selectedSessionId: Ref<string>
  settingsOpen: Ref<boolean>
  focusSession: (id: string) => Promise<void>
  closeSession: (id: string) => Promise<void>
  closeSettings: () => void
  openSettings: () => Promise<void>
}) {
  const {
    displayedSessions,
    selectedSession,
    selectedSessionId,
    settingsOpen,
    focusSession,
    closeSession,
    closeSettings,
    openSettings,
  } = options

  function preserveSessionSelection() {
    if (!displayedSessions.value.some((session) => session.id === selectedSessionId.value)) {
      selectedSessionId.value = displayedSessions.value[0]?.id ?? ''
    }
  }

  function moveSelection(delta: number) {
    const items = displayedSessions.value
    if (items.length === 0) return
    const current = Math.max(
      0,
      items.findIndex((item) => item.id === selectedSessionId.value),
    )
    const next = Math.min(items.length - 1, Math.max(0, current + delta))
    selectedSessionId.value = items[next].id
  }

  function handleKeydown(event: KeyboardEvent) {
    const target = event.target as HTMLElement | null
    if (target && ['INPUT', 'SELECT', 'TEXTAREA'].includes(target.tagName)) return
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      moveSelection(1)
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault()
      moveSelection(-1)
    }
    if (event.key === 'Enter' && selectedSession.value) {
      event.preventDefault()
      void focusSession(selectedSession.value.id)
    }
    if ((event.key === 'Delete' || event.key === 'Backspace') && selectedSession.value) {
      event.preventDefault()
      void closeSession(selectedSession.value.id)
    }
    if (event.key === 'Escape' && settingsOpen.value) {
      event.preventDefault()
      closeSettings()
    }
    if ((event.metaKey || event.ctrlKey) && event.key === ',') {
      event.preventDefault()
      void openSettings()
    }
  }

  onMounted(() => window.addEventListener('keydown', handleKeydown))
  onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown))
  watch(displayedSessions, preserveSessionSelection)

  return { preserveSessionSelection, moveSelection }
}
