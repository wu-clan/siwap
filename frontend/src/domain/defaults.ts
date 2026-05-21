import type { Harness, Preferences, TerminalProfile } from './types'
import { ALL_PROJECTS_SCOPE_ID } from './projectScope'

/** fallbackPreferences 是前端启动阶段的兜底配置，真实配置以后端返回为准 */
export const fallbackPreferences: Preferences = {
  selectedProjectId: ALL_PROJECTS_SCOPE_ID,
  defaultProjectId: '',
  language: 'zh-CN',
  appearance: 'system',
  defaultAdapterId: 'auto',
  terminalCommandTemplate: '{{command}}',
  terminalOrder: [],
  disabledTerminalIds: [],
  harnessOrder: [],
  globalShortcut: 'Control+Command+S',
  launchInBackground: false,
  worktreeBaseDir: '',
  worktreeLocation: 'project-parent',
  showDockIcon: false,
  panelWidth: 0,
  windowWidth: 320,
  windowHeight: 900,
  windowX: 0,
  windowY: 0,
  alwaysOnTop: false,
}

export function normalizePreferences(input?: Partial<Preferences>): Preferences {
  return {
    ...fallbackPreferences,
    ...input,
    terminalOrder: input?.terminalOrder ?? [],
    disabledTerminalIds: input?.disabledTerminalIds ?? [],
    harnessOrder: input?.harnessOrder ?? [],
  }
}

export function normalizeHarness(harness: Harness): Harness {
  return {
    ...harness,
    enabled: harness.enabled !== false,
    flags: harness.flags ?? {},
    flagOptions: harness.flagOptions ?? [],
    iconSource: harness.iconSource || 'builtin',
    builtIn: Boolean(harness.builtIn),
  }
}

export function emptyProfile(): TerminalProfile {
  return {
    id: '',
    label: '',
    executablePath: '',
    argumentTemplate: '{{command}}',
    workingDirFlag: '',
    commandMode: 'shell',
    platform: '',
    enabled: true,
  }
}
