export const settingsSections = [
  { id: 'general', label: 'nav.general' },
  { id: 'projects', label: 'nav.projects' },
  { id: 'worktrees', label: 'nav.worktrees' },
  { id: 'terminal', label: 'nav.terminal' },
  { id: 'ai', label: 'nav.aiAssistants' },
] as const

export type SettingsSection = (typeof settingsSections)[number]['id']
export type SettingsSectionItem = (typeof settingsSections)[number]
export type DragKind = 'project' | 'terminal' | 'harness'
