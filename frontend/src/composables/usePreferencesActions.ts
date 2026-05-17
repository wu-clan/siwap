import type { Ref } from 'vue'
import { ToggleAlwaysOnTop, UpdatePreferences } from '../../bindings/siwap/internal/desktop/app'
import { fallbackPreferences, normalizePreferences } from '../domain/defaults'
import type { Preferences, WindowState } from '../domain/types'

type Run = <T>(label: string, fn: () => Promise<T>) => Promise<T | undefined>
type Translate = (key: string) => string
type Confirm = (description: string) => Promise<boolean>

export function usePreferencesActions(options: {
  preferences: Ref<Preferences>
  run: Run
  t: Translate
  confirm: Confirm
}) {
  const { preferences, run, t, confirm } = options

  async function savePreferences(next: Preferences = preferences.value) {
    const updated = await run('action.saveSettings', () => UpdatePreferences(next as never) as unknown as Promise<Preferences>)
    if (updated) preferences.value = normalizePreferences(updated)
  }

  async function changePreference(key: keyof Preferences, value: Preferences[keyof Preferences]) {
    ;(preferences.value as Record<string, unknown>)[key] = value
    await savePreferences()
  }

  async function toggleAlwaysOnTop() {
    const state = await run('action.pinWindow', () => ToggleAlwaysOnTop() as unknown as Promise<WindowState>)
    if (state) preferences.value.alwaysOnTop = state.alwaysOnTop
  }

  async function resetPreferences() {
    if (!await confirm(t('confirm.restoreDefaults'))) return
    preferences.value = {
      ...fallbackPreferences,
      selectedProjectId: preferences.value.selectedProjectId,
      defaultProjectId: preferences.value.defaultProjectId,
      terminalOrder: preferences.value.terminalOrder ?? [],
      harnessOrder: preferences.value.harnessOrder ?? [],
    }
    await savePreferences()
  }

  return { savePreferences, changePreference, toggleAlwaysOnTop, resetPreferences }
}
