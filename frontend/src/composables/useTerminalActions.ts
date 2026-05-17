import type { Ref } from 'vue'
import { ChooseTerminalExecutable, RemoveTerminalProfile, ReorderTerminalAdapters, UpsertTerminalProfile } from '../../bindings/siwap/internal/desktop/app'
import { emptyProfile } from '../domain/defaults'
import type { Preferences, TerminalAdapter, TerminalProfile } from '../domain/types'

type Run = <T>(label: string, fn: () => Promise<T>) => Promise<T | undefined>
type Translate = (key: string) => string
type Confirm = (description: string) => Promise<boolean>

export function useTerminalActions(options: {
  preferences: Ref<Preferences>
  adapters: Ref<TerminalAdapter[]>
  profileDraft: Ref<TerminalProfile>
  terminalProfileOpen: Ref<boolean>
  run: Run
  t: Translate
  savePreferences: () => Promise<void>
  refreshBootstrap: () => Promise<void>
  terminalDisplayName: (path: string) => string
  confirm: Confirm
}) {
  const { preferences, adapters, profileDraft, terminalProfileOpen, run, t, savePreferences, refreshBootstrap, terminalDisplayName, confirm } = options

  async function changeDefaultAdapter(id: string) {
    preferences.value.defaultAdapterId = id
    await savePreferences()
  }

  async function toggleAdapter(id: string, enabled: boolean) {
    const disabled = preferences.value.disabledTerminalIds.filter((tid) => tid !== id)
    if (!enabled) disabled.push(id)
    preferences.value.disabledTerminalIds = disabled
    if (!enabled && preferences.value.defaultAdapterId === id) preferences.value.defaultAdapterId = 'auto'
    await savePreferences()
    await refreshBootstrap()
  }

  function updateProfileField(key: keyof TerminalProfile, value: TerminalProfile[keyof TerminalProfile]) {
    ;(profileDraft.value as Record<string, unknown>)[key] = value
  }

  function openTerminalProfile(profile?: TerminalProfile) {
    profileDraft.value = profile ? { ...profile } : emptyProfile()
    terminalProfileOpen.value = true
  }

  function closeTerminalProfile() {
    terminalProfileOpen.value = false
    profileDraft.value = emptyProfile()
  }

  async function chooseTerminalExecutable() {
    const path = await run('action.chooseExecutablePath', () => ChooseTerminalExecutable() as unknown as Promise<string>)
    if (!path) return
    profileDraft.value.executablePath = path
    if (!profileDraft.value.label.trim()) profileDraft.value.label = terminalDisplayName(path)
  }

  async function saveProfile(profile?: TerminalProfile) {
    const draft = profile ?? profileDraft.value
    const updated = await run('terminal.save', () => UpsertTerminalProfile({ ...draft, enabled: true } as never) as unknown as Promise<TerminalProfile>)
    if (!updated) return
    closeTerminalProfile()
    await refreshBootstrap()
  }

  async function removeProfile(id: string) {
    if (!await confirm(t('confirm.removeTerminalProfile'))) return
    await run('action.removeTerminal', () => RemoveTerminalProfile(id))
    await refreshBootstrap()
  }

  async function reorderTerminals(ids: string[]) {
    const updated = await run('action.reorderTerminals', () => ReorderTerminalAdapters(ids) as unknown as Promise<TerminalAdapter[]>)
    if (updated) adapters.value = updated
  }

  return {
    changeDefaultAdapter,
    toggleAdapter,
    updateProfileField,
    openTerminalProfile,
    closeTerminalProfile,
    chooseTerminalExecutable,
    saveProfile,
    removeProfile,
    reorderTerminals,
  }
}
