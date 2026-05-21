import type { Ref } from 'vue'
import {
  ChooseProjectDirectory,
  RemoveProject,
  ReorderProjects,
  SelectProject,
  SetDefaultProject,
} from '../../bindings/siwap/internal/desktop/app'
import type { Preferences, Project } from '../domain/types'
import type { SettingsSection } from '../domain/settings'
import { ALL_PROJECTS_SCOPE_ID, isAllProjectsScope } from '../domain/projectScope'

type Run = <T>(label: string, fn: () => Promise<T>) => Promise<T | undefined>
type Translate = (key: string, named?: Record<string, unknown>) => string
type Confirm = (description: string) => Promise<boolean>

/** useProjectActions 封装项目选择、默认项目、排序和删除等用户操作 */
export function useProjectActions(options: {
  preferences: Ref<Preferences>
  projects: Ref<Project[]>
  selectedWorktreePath: Ref<string>
  selectedSessionId: Ref<string>
  settingsSection: Ref<SettingsSection>
  run: Run
  t: Translate
  preserveSessionSelection: () => void
  openSettings: (section?: SettingsSection) => Promise<void>
  projectName: (project: Project) => string
  confirm: Confirm
}) {
  const {
    preferences,
    projects,
    selectedWorktreePath,
    selectedSessionId,
    settingsSection,
    run,
    t,
    preserveSessionSelection,
    openSettings,
    projectName,
    confirm,
  } = options

  function removeProjectPrompt(name: string) {
    return t('confirm.removeProject', { name })
  }

  async function chooseProjectDirectory() {
    const created = await run(
      'action.addProject',
      () => ChooseProjectDirectory() as unknown as Promise<Project>,
    )
    if (!created) return
    settingsSection.value = 'projects'
  }

  async function selectProject(id: string) {
    if (id === '__settings') {
      void openSettings('projects')
      return
    }
    const selected = await run(
      'action.switchProject',
      () => SelectProject(id) as unknown as Promise<Project>,
    )
    if (isAllProjectsScope(id)) {
      preferences.value.selectedProjectId = ALL_PROJECTS_SCOPE_ID
    } else {
      if (!selected) return
      preferences.value.selectedProjectId = selected.id
    }
    // 项目切换后 worktree 和会话选择都不再沿用，具体列表由后端事件同步
    selectedWorktreePath.value = ''
    selectedSessionId.value = ''
    preserveSessionSelection()
  }

  async function setDefaultProject(id: string) {
    await run(
      'action.setDefaultProject',
      () => SetDefaultProject(id) as unknown as Promise<Project>,
    )
  }

  async function removeProject(project: Project) {
    if (!(await confirm(removeProjectPrompt(projectName(project))))) return
    await run('action.removeProject', () => RemoveProject(project.id))
    if (preferences.value.selectedProjectId === project.id) {
      selectedWorktreePath.value = ''
      selectedSessionId.value = ''
    }
  }

  async function reorderProjects(ids: string[]) {
    const updated = await run(
      'action.reorderProjects',
      () => ReorderProjects(ids) as unknown as Promise<Project[]>,
    )
    if (updated) projects.value = updated
  }

  return {
    chooseProjectDirectory,
    selectProject,
    setDefaultProject,
    removeProject,
    reorderProjects,
  }
}
