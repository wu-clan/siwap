import type { Ref } from 'vue'
import { ReorderHarnesses, UpdateHarness } from '../../bindings/siwap/internal/desktop/app'
import { normalizeHarness } from '../domain/defaults'
import type { Harness } from '../domain/types'

type Run = <T>(label: string, fn: () => Promise<T>) => Promise<T | undefined>

export function useAssistantActions(options: {
  harnesses: Ref<Harness[]>
  run: Run
}) {
  const { harnesses, run } = options

  async function saveAssistant(harness: Harness) {
    const updated = await run('ai.saveAssistant', () => UpdateHarness(harness as never) as unknown as Promise<Harness>)
    if (!updated) return
    const index = harnesses.value.findIndex((item) => item.id === updated.id)
    if (index >= 0) harnesses.value[index] = normalizeHarness(updated)
  }

  async function reorderHarnesses(ids: string[]) {
    const updated = await run('action.reorderAssistants', () => ReorderHarnesses(ids) as unknown as Promise<Harness[]>)
    if (updated) harnesses.value = updated.map(normalizeHarness)
  }

  return { saveAssistant, reorderHarnesses }
}
