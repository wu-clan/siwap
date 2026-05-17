import type { Ref } from 'vue'
import type { Harness, Project, TerminalAdapter, Worktree } from '../domain/types'

type Translate = (key: string) => string

export function useDisplayFormatters(options: {
  t: Translate
  projects: Ref<Project[]>
  harnesses: Ref<Harness[]>
  adapters: Ref<TerminalAdapter[]>
  worktrees: Ref<Worktree[]>
}) {
  const { t, harnesses, adapters } = options

  function basename(path: string) {
    const clean = path.replace(/[\\/]+$/, '')
    return clean.split(/[\\/]/).pop() || clean || t('common.untitled')
  }

  function projectName(project: Project) {
    return basename(project.label || project.path)
  }

  function harnessName(id: string) {
    return harnesses.value.find((item) => item.id === id)?.label || id
  }

  function terminalName(id: string) {
    return adapters.value.find((item) => item.id === id)?.label || id
  }

  function terminalDisplayName(path: string) {
    return basename(path).replace(/\.app$/i, '')
  }

  function availabilityText(adapter: TerminalAdapter) {
    if (adapter.installed) return t('terminal.available')
    if (adapter.id === 'auto') return t('terminal.available')
    return t('terminal.notInstalled')
  }

  return {
    basename,
    projectName,
    harnessName,
    terminalName,
    terminalDisplayName,
    availabilityText,
  }
}
