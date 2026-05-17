export type Summary = { name: string; stack: string[]; scope: string[]; exclusions: string[] }
export type Preferences = {
  selectedProjectId: string
  defaultProjectId: string
  language: string
  appearance: string
  defaultAdapterId: string
  terminalCommandTemplate: string
  terminalOrder: string[]
  disabledTerminalIds: string[]
  harnessOrder: string[]
  globalShortcut: string
  launchInBackground: boolean
  worktreeBaseDir: string
  worktreeLocation: string
  autohideOnBlur: boolean
  panelWidth: number
  windowWidth: number
  windowHeight: number
  windowX: number
  windowY: number
  alwaysOnTop: boolean
}
export type TerminalProfile = {
  id: string
  label: string
  executablePath: string
  argumentTemplate: string
  workingDirFlag: string
  commandMode: string
  platform: string
  enabled: boolean
}
export type HarnessFlag = { key: string; label: string; type: string; commandFlag: string; default: string; options: string[] }
export type Harness = {
  id: string
  label: string
  command: string
  enabled: boolean
  builtIn: boolean
  icon: string
  iconSource: string
  tint: string
  flags: Record<string, string>
  flagOptions: HarnessFlag[]
}
export type Project = { id: string; path: string; label?: string; isDefault: boolean; lastUsedAt?: string }
export type TerminalCapability = { key: string; label: string; supported: boolean; description: string }
export type TerminalAdapter = {
  id: string
  label: string
  platform: string
  executable?: string
  installed: boolean
  enabled: boolean
  stability: string
  confidence: string
  message?: string
  capabilities: TerminalCapability[]
}
export type TerminalSessionRef = {
  adapterId: string
  platform: string
  pid?: number
  processTreePids?: number[]
  windowId?: string
  tabId?: string
  terminalId?: string
  title: string
  cwd: string
  identityStrategy: string
  capabilitiesSnapshot: string[]
  canFocus: boolean
  canClose: boolean
  requiresPlatformGrant: boolean
}
export type Session = {
  id: string
  harnessId: string
  projectId?: string
  adapterId: string
  title: string
  command: string
  workingDir: string
  worktreePath?: string
  status: string
  createdAt: string
  updatedAt: string
  pid?: number
  sessionEnv: string
  launchMode: string
  focusMode: string
  closeMode: string
  error?: string
  ref: TerminalSessionRef
}
export type Worktree = {
  id: string
  projectId: string
  path: string
  branch: string
  baseBranch?: string
  head?: string
  isMain: boolean
  dirty: boolean
  exists: boolean
  status: string
  createdAt?: string
}
export type ActionResult = { ok: boolean; status: string; message: string }
export type WindowState = { width: number; height: number; alwaysOnTop: boolean; mode: string }
export type Bootstrap = {
  summary: Summary
  configPath: string
  preferences: Preferences
  harnesses: Harness[]
  projects: Project[]
  terminalProfiles: TerminalProfile[]
  adapters: TerminalAdapter[]
  sessions: Session[]
}
export type LaunchRequest = {
  harnessId: string
  projectId?: string
  adapterId?: string
  command?: string
  workingDir?: string
  title?: string
  flagOverrides?: Record<string, string>
  worktreePath?: string
}
