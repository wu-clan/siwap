<script setup lang="ts">
import SettingsDialog from '../components/common/SettingsDialog.vue'
import AiSettings from '../features/settings/AiSettings.vue'
import GeneralSettings from '../features/settings/GeneralSettings.vue'
import ProjectsSettings from '../features/settings/ProjectsSettings.vue'
import TerminalSettings from '../features/settings/TerminalSettings.vue'
import WorktreesSettings from '../features/settings/WorktreesSettings.vue'
import type { SettingsSection, SettingsSectionItem } from '../domain/settings'
import type {
  Harness,
  Preferences,
  Project,
  TerminalAdapter,
  TerminalProfile,
  Worktree,
} from '../domain/types'

defineProps<{
  visible: boolean
  isSettingsWindow: boolean
  settingsSection: SettingsSection
  settingsSections: readonly SettingsSectionItem[]
  preferences: Preferences
  projects: Project[]
  allWorktrees: Worktree[]
  settingsWorktreeProjectId: string
  worktreeBranches: string[]
  canCreateWorktree: boolean
  selectedWorktreePath: string
  worktreeCreateOpen: boolean
  branchDraft: string
  baseBranchDraft: string
  worktreePathDraft: string
  adapters: TerminalAdapter[]
  terminalProfiles: TerminalProfile[]
  terminalProfileOpen: boolean
  profileDraft: TerminalProfile
  harnesses: Harness[]
  projectName: (project: Project) => string
  availabilityText: (adapter: TerminalAdapter) => string
}>()

const emit = defineEmits<{
  'update:settings-section': [section: SettingsSection]
  close: []
  'change-preference': [key: keyof Preferences, value: Preferences[keyof Preferences]]
  'toggle-always-on-top': []
  'reset-window': []
  'reset-preferences': []
  'choose-project': []
  'reorder-projects': [ids: string[]]
  'set-default-project': [id: string]
  'remove-project': [project: Project]
  'update-settings-worktree-project': [id: string]
  'open-worktree-create': []
  'close-worktree-create': []
  'create-worktree': []
  'delete-worktree': [item: Worktree]
  'set-default-worktree': [path: string]
  'update-branch-draft': [value: string]
  'update-base-branch-draft': [value: string]
  'update-worktree-path-draft': [value: string]
  'change-default-adapter': [id: string]
  'toggle-adapter': [id: string, enabled: boolean]
  'reorder-terminals': [ids: string[]]
  'open-terminal-profile': [profile?: TerminalProfile]
  'close-terminal-profile': []
  'choose-terminal-executable': []
  'save-terminal-profile': [profile: TerminalProfile]
  'remove-terminal-profile': [id: string]
  'update-profile-field': [
    key: keyof TerminalProfile,
    value: TerminalProfile[keyof TerminalProfile],
  ]
  'save-assistant': [harness: Harness]
  'reorder-harnesses': [ids: string[]]
}>()
</script>

<template>
  <SettingsDialog
    :visible="visible"
    :is-settings-window="isSettingsWindow"
    :settings-section="settingsSection"
    :settings-sections="settingsSections"
    @update:settings-section="emit('update:settings-section', $event)"
    @close="emit('close')"
  >
    <GeneralSettings
      v-if="settingsSection === 'general'"
      :preferences="preferences"
      @change-preference="(key, value) => emit('change-preference', key, value)"
      @toggle-always-on-top="emit('toggle-always-on-top')"
      @reset-window="emit('reset-window')"
      @reset-preferences="emit('reset-preferences')"
    />

    <ProjectsSettings
      v-else-if="settingsSection === 'projects'"
      :projects="projects"
      :project-name="projectName"
      @choose="emit('choose-project')"
      @reorder="emit('reorder-projects', $event)"
      @set-default="emit('set-default-project', $event)"
      @remove="emit('remove-project', $event)"
    />

    <WorktreesSettings
      v-else-if="settingsSection === 'worktrees'"
      :projects="projects"
      :all-worktrees="allWorktrees"
      :settings-worktree-project-id="settingsWorktreeProjectId"
      :worktree-branches="worktreeBranches"
      :can-create-worktree="canCreateWorktree"
      :selected-worktree-path="selectedWorktreePath"
      :worktree-create-open="worktreeCreateOpen"
      :preferences="preferences"
      :branch-draft="branchDraft"
      :base-branch-draft="baseBranchDraft"
      :worktree-path-draft="worktreePathDraft"
      :project-name="projectName"
      @update-settings-worktree-project="emit('update-settings-worktree-project', $event)"
      @open-create="emit('open-worktree-create')"
      @close-create="emit('close-worktree-create')"
      @create="emit('create-worktree')"
      @delete="emit('delete-worktree', $event)"
      @set-default-worktree="emit('set-default-worktree', $event)"
      @update-branch-draft="emit('update-branch-draft', $event)"
      @update-base-branch-draft="emit('update-base-branch-draft', $event)"
      @update-worktree-path-draft="emit('update-worktree-path-draft', $event)"
      @change-preference="(key, value) => emit('change-preference', key, value)"
    />

    <TerminalSettings
      v-else-if="settingsSection === 'terminal'"
      :preferences="preferences"
      :adapters="adapters"
      :terminal-profiles="terminalProfiles"
      :terminal-profile-open="terminalProfileOpen"
      :profile-draft="profileDraft"
      :availability-text="availabilityText"
      @change-default-adapter="emit('change-default-adapter', $event)"
      @toggle-adapter="(id, enabled) => emit('toggle-adapter', id, enabled)"
      @reorder="emit('reorder-terminals', $event)"
      @open-profile="emit('open-terminal-profile', $event)"
      @close-profile="emit('close-terminal-profile')"
      @choose-executable="emit('choose-terminal-executable')"
      @save-profile="emit('save-terminal-profile', $event)"
      @remove-profile="emit('remove-terminal-profile', $event)"
      @update-profile-field="(key, value) => emit('update-profile-field', key, value)"
    />

    <AiSettings
      v-else-if="settingsSection === 'ai'"
      :harnesses="harnesses"
      @save="emit('save-assistant', $event)"
      @reorder="emit('reorder-harnesses', $event)"
    />
  </SettingsDialog>
</template>
