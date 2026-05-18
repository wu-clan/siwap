import type { ComputedRef, Ref } from 'vue'
import { ClearSessions, CloseSession, FocusSession, LaunchSession, ListSessions } from '../../bindings/siwap/internal/desktop/app'
import type { ActionResult, Harness, LaunchRequest, Preferences, Project, Session } from '../domain/types'
import type { SettingsSection } from '../domain/settings'

type Run = <T>(label: string, fn: () => Promise<T>) => Promise<T | undefined>
type Translate = (key: string, named?: Record<string, unknown>) => string

/** useSessionActions 封装会话启动、聚焦、关闭和清空等用户操作 */
export function useSessionActions(options: {
  selectedProject: ComputedRef<Project | undefined>
  selectedWorktreePath: Ref<string>
  selectedSessionId: Ref<string>
  sessions: Ref<Session[]>
  preferences: Ref<Preferences>
  actionMessage: Ref<string>
  run: Run
  t: Translate
  openSettings: (section?: SettingsSection) => Promise<void>
  preserveSessionSelection: () => void
}) {
  const { selectedProject, selectedWorktreePath, selectedSessionId, sessions, preferences, actionMessage, run, t, openSettings, preserveSessionSelection } = options
  // 防止连续点击时重复发送同一个启动或聚焦请求
  const launchKeys = new Set<string>()
  const focusIds = new Set<string>()

  function launchAssistantLabel(name: string) {
    return t('action.launchAssistant', { name })
  }

  async function launchHarness(harness: Harness) {
    if (!selectedProject.value) {
      void openSettings('projects')
      actionMessage.value = t('project.selectFirst')
      return
    }
    const launchKey = [
      harness.id,
      selectedProject.value.id,
      selectedWorktreePath.value,
      preferences.value.defaultAdapterId || 'auto',
    ].join('|')
    if (launchKeys.has(launchKey)) return
    launchKeys.add(launchKey)
    const request: LaunchRequest = {
      harnessId: harness.id,
      projectId: selectedProject.value.id,
      adapterId: preferences.value.defaultAdapterId || 'auto',
      workingDir: selectedWorktreePath.value || selectedProject.value.path,
      title: harness.label,
      flagOverrides: harness.flags ?? {},
      worktreePath: selectedWorktreePath.value,
    }
    try {
      const created = await run(launchAssistantLabel(harness.label), () => LaunchSession(request as never) as unknown as Promise<Session>)
      if (!created) return
      // 以后端会话列表为准，失败的启动也会保留在列表中便于查看错误和重试
      sessions.value = await ListSessions() as unknown as Session[]
      selectedSessionId.value = created.id
      if (created.status === 'failed') actionMessage.value = created.error || t('session.launchFailedKept')
    } finally {
      launchKeys.delete(launchKey)
    }
  }

  async function focusSession(id: string) {
    selectedSessionId.value = id
    if (focusIds.has(id)) return
    focusIds.add(id)
    try {
      const result = await run('action.focusSession', () => FocusSession(id) as unknown as Promise<ActionResult>)
      if (result) actionMessage.value = result.message
      // 聚焦可能触发终端重开，因此刷新后端状态而不是直接改本地对象
      sessions.value = await ListSessions() as unknown as Session[]
      preserveSessionSelection()
    } finally {
      focusIds.delete(id)
    }
  }

  async function closeSession(id: string) {
    const result = await run('session.closeSession', () => CloseSession(id) as unknown as Promise<ActionResult>)
    if (result) actionMessage.value = result.message
    sessions.value = await ListSessions() as unknown as Session[]
    preserveSessionSelection()
  }

  async function clearSessions() {
    const result = await run('action.clearAllSessions', () => ClearSessions() as unknown as Promise<ActionResult>)
    if (result) actionMessage.value = result.message
    sessions.value = await ListSessions() as unknown as Session[]
    preserveSessionSelection()
  }

  return { launchHarness, focusSession, closeSession, clearSessions }
}
