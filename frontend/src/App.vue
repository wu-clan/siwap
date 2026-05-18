<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  CloseSettingsWindow,
  GetBootstrap,
  GetInitialSettingsSection,
  GetWindowRole,
  ListWorktreeBranches,
  ListWorktrees,
  OpenSettingsSection,
  ResetSidebarWindow,
} from '../bindings/siwap/internal/desktop/app'
import { Events } from '@wailsio/runtime'
import { useI18n } from 'vue-i18n'
import ConfirmDialog from './components/common/ConfirmDialog.vue'
import MainWindow from './windows/MainWindow.vue'
import SettingsWindow from './windows/SettingsWindow.vue'
import { settingsSections, type SettingsSection } from './domain/settings'
import type {
  ActionResult,
  Bootstrap,
  Harness,
  Preferences,
  Project,
  Session,
  TerminalAdapter,
  TerminalProfile,
  Worktree,
} from './domain/types'
import {
  emptyProfile,
  fallbackPreferences,
  normalizeHarness,
  normalizePreferences,
} from './domain/defaults'
import { useActionRunner } from './composables/useActionRunner'
import { useAppearance } from './composables/useAppearance'
import { useAssistantActions } from './composables/useAssistantActions'
import { useDisplayFormatters } from './composables/useDisplayFormatters'
import { usePreferencesActions } from './composables/usePreferencesActions'
import { useProjectActions } from './composables/useProjectActions'
import { useSessionActions } from './composables/useSessionActions'
import { useSessionSelection } from './composables/useSessionSelection'
import { useTerminalActions } from './composables/useTerminalActions'
import { useWorktreeActions } from './composables/useWorktreeActions'
import { ALL_PROJECTS_SCOPE_ID, isAllProjectsScope } from './domain/projectScope'

const preferences = ref<Preferences>({ ...fallbackPreferences })
const harnesses = ref<Harness[]>([])
const projects = ref<Project[]>([])
const terminalProfiles = ref<TerminalProfile[]>([])
const adapters = ref<TerminalAdapter[]>([])
const sessions = ref<Session[]>([])
const worktrees = ref<Worktree[]>([])
const worktreeBranches = ref<string[]>([])
const settingsWorktreeProjectId = ref('')
const selectedWorktreePath = ref('')
const selectedSessionId = ref('')
const settingsOpen = ref(false)
const isSettingsWindow = ref(false)
const settingsSection = ref<SettingsSection>('general')
const branchDraft = ref('')
const baseBranchDraft = ref('')
const worktreePathDraft = ref('')
const worktreeCreateOpen = ref(false)
const terminalProfileOpen = ref(false)
const profileDraft = ref<TerminalProfile>(emptyProfile())
const confirmDialog = ref({ open: false, description: '', title: '', confirmLabel: '' })

const { t } = useI18n({ useScope: 'global' })
const { actionMessage, run } = useActionRunner(t)
const { resolvedAppearance } = useAppearance(preferences)

let offSettings: (() => void) | undefined
let offPreferences: (() => void) | undefined
let offProjects: (() => void) | undefined
let offHarnesses: (() => void) | undefined
let offAdapters: (() => void) | undefined
let offTerminalProfiles: (() => void) | undefined
let offSessions: (() => void) | undefined
let offWorktrees: (() => void) | undefined
let preserveSessionSelection: () => void = () => {}
let confirmResolver: ((value: boolean) => void) | undefined

const isAllProjectsSelected = computed(() =>
  isAllProjectsScope(preferences.value.selectedProjectId),
)
const selectedProject = computed(() =>
  isAllProjectsSelected.value
    ? undefined
    : projects.value.find((project) => project.id === preferences.value.selectedProjectId),
)
const selectedProjectId = computed(() => selectedProject.value?.id ?? '')
const settingsWorktreeProject = computed(() =>
  projects.value.find((project) => project.id === settingsWorktreeProjectId.value),
)
const currentWorktrees = computed(() =>
  isAllProjectsSelected.value
    ? []
    : worktrees.value.filter((item) => item.projectId === selectedProjectId.value),
)
const canCreateWorktree = computed(() =>
  Boolean(settingsWorktreeProject.value && worktreeBranches.value.length > 0),
)
const enabledHarnesses = computed(() => harnesses.value.filter((h) => h.enabled))
const launchableAdapters = computed(() =>
  adapters.value.filter((adapter) => adapter.id === 'auto' || adapter.enabled),
)
const displayedSessions = computed(() =>
  sessions.value.filter((session) => {
    if (isAllProjectsSelected.value) return true
    if (!selectedProjectId.value) return false
    if (selectedProjectId.value && session.projectId !== selectedProjectId.value) return false
    const sessionWorktree = session.worktreePath || ''
    return sessionWorktree === selectedWorktreePath.value
  }),
)
const selectedSession = computed(() =>
  displayedSessions.value.find((session) => session.id === selectedSessionId.value),
)

const { basename, projectName, harnessName, terminalName, terminalDisplayName, availabilityText } =
  useDisplayFormatters({ t, projects, harnesses, adapters, worktrees })

function confirmAction(description: string, title = '', confirmLabel = '') {
  confirmResolver?.(false)
  confirmDialog.value = { open: true, description, title, confirmLabel }
  return new Promise<boolean>((resolve) => {
    confirmResolver = resolve
  })
}

function resolveConfirm(value: boolean) {
  confirmDialog.value.open = false
  confirmResolver?.(value)
  confirmResolver = undefined
}

async function refreshBootstrap() {
  const data = await run('action.refresh', () => GetBootstrap() as unknown as Promise<Bootstrap>)
  if (!data) return
  preferences.value = normalizePreferences(data.preferences)
  projects.value = data.projects ?? []
  harnesses.value = (data.harnesses ?? []).map(normalizeHarness)
  terminalProfiles.value = data.terminalProfiles ?? []
  adapters.value = data.adapters ?? []
  sessions.value = data.sessions ?? []
  if (!preferences.value.selectedProjectId)
    preferences.value.selectedProjectId = ALL_PROJECTS_SCOPE_ID
  if (!preferences.value.defaultAdapterId) preferences.value.defaultAdapterId = 'auto'
  await refreshWorktrees()
  await refreshWorktreeBranches()
  preserveSessionSelection()
}

async function refreshWorktrees() {
  if (projects.value.length === 0) {
    worktrees.value = []
    selectedWorktreePath.value = ''
    return
  }
  const lists = await Promise.all(
    projects.value.map(async (project) => {
      try {
        return (await ListWorktrees(project.id)) as unknown as Worktree[]
      } catch {
        return []
      }
    }),
  )
  worktrees.value = lists.flat()
  if (
    selectedWorktreePath.value &&
    !currentWorktrees.value.some((item) => item.path === selectedWorktreePath.value)
  ) {
    selectedWorktreePath.value = ''
  }
}

async function refreshWorktreeBranches() {
  if (!settingsWorktreeProjectId.value) {
    worktreeBranches.value = []
    baseBranchDraft.value = ''
    return
  }
  worktreeBranches.value = await listWorktreeBranches(settingsWorktreeProjectId.value)
  baseBranchDraft.value = defaultBaseBranch(worktreeBranches.value)
}

async function syncProjects(next: Project[]) {
  projects.value = next || []
  if (
    settingsWorktreeProjectId.value &&
    !projects.value.some((project) => project.id === settingsWorktreeProjectId.value)
  ) {
    settingsWorktreeProjectId.value = ''
  }
  await refreshWorktrees()
  await refreshWorktreeBranches()
  preserveSessionSelection()
}

async function syncWorktrees() {
  await refreshWorktrees()
  await refreshWorktreeBranches()
  preserveSessionSelection()
}

function syncSessions(next: Session[]) {
  sessions.value = next || []
  preserveSessionSelection()
}

async function listWorktreeBranches(projectId: string) {
  if (!projectId) return []
  try {
    return (await ListWorktreeBranches(projectId)) as unknown as string[]
  } catch {
    return []
  }
}

function defaultBaseBranch(branches: string[]) {
  return (
    branches.find((branch) => branch === 'main') ??
    branches.find((branch) => branch.endsWith('/main')) ??
    branches.find((branch) => branch === 'master') ??
    branches.find((branch) => branch.endsWith('/master')) ??
    branches[0] ??
    ''
  )
}

async function changeSettingsWorktreeProject(projectId: string) {
  settingsWorktreeProjectId.value = projectId
  worktreeCreateOpen.value = false
  await refreshWorktreeBranches()
}

const { savePreferences, changePreference, toggleAlwaysOnTop, resetPreferences } =
  usePreferencesActions({
    preferences,
    run,
    t,
    confirm: confirmAction,
  })

const {
  changeDefaultAdapter,
  toggleAdapter,
  updateProfileField,
  openTerminalProfile,
  closeTerminalProfile,
  chooseTerminalExecutable,
  saveProfile,
  removeProfile,
  reorderTerminals,
} = useTerminalActions({
  preferences,
  adapters,
  profileDraft,
  terminalProfileOpen,
  run,
  t,
  savePreferences,
  refreshBootstrap,
  terminalDisplayName,
  confirm: confirmAction,
})

const { saveAssistant, reorderHarnesses } = useAssistantActions({
  harnesses,
  run,
})

const { chooseProjectDirectory, selectProject, setDefaultProject, removeProject, reorderProjects } =
  useProjectActions({
    preferences,
    projects,
    selectedWorktreePath,
    selectedSessionId,
    settingsSection,
    run,
    t,
    refreshBootstrap,
    refreshWorktrees,
    preserveSessionSelection: () => preserveSessionSelection(),
    openSettings,
    projectName,
    confirm: confirmAction,
  })

const { openWorktreeCreate, closeWorktreeCreate, createWorktree, deleteWorktree } =
  useWorktreeActions({
    selectedProject: settingsWorktreeProject,
    selectedProjectId,
    selectedWorktreePath,
    branchDraft,
    baseBranchDraft,
    worktreePathDraft,
    worktreeCreateOpen,
    preferences,
    actionMessage,
    run,
    t,
    refreshWorktrees,
    canCreateWorktree,
    confirm: confirmAction,
  })

const { launchHarness, focusSession, closeSession, clearSessions } = useSessionActions({
  selectedProject,
  selectedWorktreePath,
  selectedSessionId,
  sessions,
  preferences,
  actionMessage,
  run,
  t,
  openSettings,
  preserveSessionSelection: () => preserveSessionSelection(),
})

async function openSettings(section: SettingsSection = settingsSection.value) {
  settingsSection.value = section
  if (!isSettingsWindow.value) {
    await run(
      'action.openSettings',
      () => OpenSettingsSection(section) as unknown as Promise<ActionResult>,
    )
    return
  }
  settingsOpen.value = true
}

function closeSettings() {
  if (isSettingsWindow.value) {
    void CloseSettingsWindow()
    return
  }
  settingsOpen.value = false
}

async function openWorktreeCreateFromMain() {
  if (!selectedProject.value) {
    actionMessage.value = t('project.selectFirst')
    return
  }
  const branches = await listWorktreeBranches(selectedProjectId.value)
  if (branches.length === 0) {
    actionMessage.value = t('worktree.gitRequired')
    return
  }
  settingsWorktreeProjectId.value = selectedProjectId.value
  worktreeBranches.value = branches
  baseBranchDraft.value = defaultBaseBranch(branches)
  if (isSettingsWindow.value || !hasWailsRuntime()) {
    settingsSection.value = 'worktrees'
    settingsOpen.value = true
    worktreeCreateOpen.value = true
    return
  }
  await run(
    'action.openSettings',
    () => OpenSettingsSection('worktrees:create') as unknown as Promise<ActionResult>,
  )
}

function hasWailsRuntime() {
  return (
    typeof window !== 'undefined' &&
    Boolean(
      (window as Window & { _wails?: unknown; runtime?: unknown })._wails ||
      (window as Window & { runtime?: unknown }).runtime,
    )
  )
}

function applySettingsPayload(payload?: string) {
  const [section, action] = (payload || '').split(':')
  if (section && settingsSections.some((item) => item.id === section)) {
    settingsSection.value = section as SettingsSection
  }
  if (section === 'worktrees' && action === 'create') {
    settingsWorktreeProjectId.value = selectedProjectId.value
    worktreeCreateOpen.value = true
    void refreshWorktreeBranches()
  }
}

const sessionSelection = useSessionSelection({
  displayedSessions,
  selectedSession,
  selectedSessionId,
  settingsOpen,
  focusSession,
  closeSession,
  closeSettings,
  openSettings,
})
preserveSessionSelection = sessionSelection.preserveSessionSelection

async function initializeApp() {
  const urlParams = new URLSearchParams(window.location.search)
  const urlSection = urlParams.get('section') as SettingsSection | null
  const urlAction = urlParams.get('action')
  const isUrlSettingsWindow = urlParams.get('window') === 'settings'
  if (hasWailsRuntime()) {
    if (isUrlSettingsWindow) {
      isSettingsWindow.value = true
      if (urlSection && settingsSections.some((item) => item.id === urlSection))
        settingsSection.value = urlSection
      if (urlSection === 'worktrees' && urlAction === 'create') worktreeCreateOpen.value = true
    } else {
      const role = await GetWindowRole()
      isSettingsWindow.value = role === 'settings'
      const initialSection = (await GetInitialSettingsSection()) as SettingsSection
      if (settingsSections.some((item) => item.id === initialSection))
        settingsSection.value = initialSection
    }
    if (isSettingsWindow.value) settingsOpen.value = true
  }
  await refreshBootstrap()
  if (worktreeCreateOpen.value && selectedProjectId.value) {
    settingsWorktreeProjectId.value = selectedProjectId.value
    await refreshWorktreeBranches()
  }
}

onMounted(() => {
  void initializeApp()
  if (hasWailsRuntime()) {
    offSettings = Events.On('ui:open-settings', (event) => {
      applySettingsPayload(event.data as string | undefined)
      if (isSettingsWindow.value) settingsOpen.value = true
    })
    offPreferences = Events.On('preferences:updated', (event) => {
      preferences.value = normalizePreferences(event.data as Partial<Preferences>)
    })
    offProjects = Events.On('projects:updated', (event) => {
      void syncProjects((event.data as Project[]) || [])
    })
    offHarnesses = Events.On('harnesses:updated', (event) => {
      harnesses.value = ((event.data as Harness[]) || []).map(normalizeHarness)
    })
    offAdapters = Events.On('adapters:updated', (event) => {
      adapters.value = (event.data as TerminalAdapter[]) || []
    })
    offTerminalProfiles = Events.On('terminalProfiles:updated', (event) => {
      terminalProfiles.value = (event.data as TerminalProfile[]) || []
    })
    offSessions = Events.On('sessions:updated', (event) => {
      syncSessions((event.data as Session[]) || [])
    })
    offWorktrees = Events.On('worktrees:updated', () => {
      void syncWorktrees()
    })
  }
})

watch(
  () => selectedProjectId.value,
  async () => {
    selectedWorktreePath.value = ''
    await refreshWorktrees()
    preserveSessionSelection()
  },
)

watch(settingsSection, (section) => {
  if (section !== 'worktrees') worktreeCreateOpen.value = false
  if (section !== 'terminal') terminalProfileOpen.value = false
})

onBeforeUnmount(() => {
  offSettings?.()
  offPreferences?.()
  offProjects?.()
  offHarnesses?.()
  offAdapters?.()
  offTerminalProfiles?.()
  offSessions?.()
  offWorktrees?.()
  confirmResolver?.(false)
})
</script>

<template>
  <main
    class="app-shell app-frame"
    :class="{ 'settings-window-host': isSettingsWindow, 'settings-frame': isSettingsWindow }"
    :data-appearance="preferences.appearance"
    :data-resolved-appearance="resolvedAppearance"
  >
    <MainWindow
      v-if="!isSettingsWindow"
      :preferences="preferences"
      :projects="projects"
      :selected-project="selectedProject"
      :selected-worktree-path="selectedWorktreePath"
      :current-worktrees="currentWorktrees"
      :launchable-adapters="launchableAdapters"
      :enabled-harnesses="enabledHarnesses"
      :displayed-sessions="displayedSessions"
      :selected-session-id="selectedSessionId"
      :project-name="projectName"
      :basename="basename"
      :harness-name="harnessName"
      :terminal-name="terminalName"
      @select-project="selectProject"
      @add-project="chooseProjectDirectory"
      @update:selected-worktree-path="selectedWorktreePath = $event"
      @add-worktree="openWorktreeCreateFromMain"
      @change-default-adapter="changeDefaultAdapter"
      @launch-assistant="launchHarness"
      @open-settings-ai="openSettings('ai')"
      @focus-session="focusSession"
      @close-session="closeSession"
      @clear-sessions="clearSessions"
    />

    <SettingsWindow
      :visible="isSettingsWindow || settingsOpen"
      :is-settings-window="isSettingsWindow"
      :settings-section="settingsSection"
      :settings-sections="settingsSections"
      :preferences="preferences"
      :projects="projects"
      :all-worktrees="worktrees"
      :settings-worktree-project-id="settingsWorktreeProjectId"
      :worktree-branches="worktreeBranches"
      :can-create-worktree="canCreateWorktree"
      :selected-worktree-path="selectedWorktreePath"
      :worktree-create-open="worktreeCreateOpen"
      :branch-draft="branchDraft"
      :base-branch-draft="baseBranchDraft"
      :worktree-path-draft="worktreePathDraft"
      :adapters="adapters"
      :terminal-profiles="terminalProfiles"
      :terminal-profile-open="terminalProfileOpen"
      :profile-draft="profileDraft"
      :harnesses="harnesses"
      :project-name="projectName"
      :availability-text="availabilityText"
      @update:settings-section="settingsSection = $event"
      @close="closeSettings"
      @change-preference="changePreference"
      @toggle-always-on-top="toggleAlwaysOnTop"
      @reset-window="ResetSidebarWindow()"
      @reset-preferences="resetPreferences"
      @choose-project="chooseProjectDirectory"
      @reorder-projects="reorderProjects"
      @reorder-terminals="reorderTerminals"
      @reorder-harnesses="reorderHarnesses"
      @set-default-project="setDefaultProject"
      @remove-project="removeProject"
      @update-settings-worktree-project="changeSettingsWorktreeProject"
      @open-worktree-create="openWorktreeCreate"
      @close-worktree-create="closeWorktreeCreate"
      @create-worktree="createWorktree"
      @delete-worktree="deleteWorktree"
      @set-default-worktree="selectedWorktreePath = $event"
      @update-selected-worktree="selectedWorktreePath = $event"
      @update-branch-draft="branchDraft = $event"
      @update-base-branch-draft="baseBranchDraft = $event"
      @update-worktree-path-draft="worktreePathDraft = $event"
      @change-default-adapter="changeDefaultAdapter"
      @toggle-adapter="toggleAdapter"
      @open-terminal-profile="openTerminalProfile"
      @close-terminal-profile="closeTerminalProfile"
      @choose-terminal-executable="chooseTerminalExecutable"
      @save-terminal-profile="saveProfile"
      @remove-terminal-profile="removeProfile"
      @update-profile-field="updateProfileField"
      @save-assistant="saveAssistant"
    />

    <ConfirmDialog
      :open="confirmDialog.open"
      :title="confirmDialog.title"
      :description="confirmDialog.description"
      :confirm-label="confirmDialog.confirmLabel"
      @confirm="resolveConfirm(true)"
      @cancel="resolveConfirm(false)"
    />
  </main>
</template>
