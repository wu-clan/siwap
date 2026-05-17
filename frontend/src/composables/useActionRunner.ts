import { ref } from 'vue'

type Translate = (key: string, named?: Record<string, unknown>) => string

export function useActionRunner(t: Translate) {
  const actionMessage = ref('')
  const busy = ref(false)

  async function run<T>(label: string, fn: () => Promise<T>): Promise<T | undefined> {
    busy.value = true
    const displayLabel = t(label)
    actionMessage.value = `${displayLabel}...`
    try {
      const result = await fn()
      actionMessage.value = t('action.done', { label: displayLabel })
      return result
    } catch (error) {
      actionMessage.value = error instanceof Error ? error.message : String(error)
      return undefined
    } finally {
      busy.value = false
    }
  }

  return { actionMessage, busy, run }
}
