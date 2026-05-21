<script setup lang="ts">
import MainContextStack from '../components/main/MainContextStack.vue'
import AssistantLauncher from '../components/main/AssistantLauncher.vue'
import SessionList from '../components/main/SessionList.vue'
import type {
  Harness,
  Preferences,
  Project,
  Session,
  TerminalAdapter,
  Worktree,
} from '../domain/types'

defineProps<{
  preferences: Preferences
  projects: Project[]
  selectedProject?: Project
  selectedWorktreePath: string
  currentWorktrees: Worktree[]
  launchableAdapters: TerminalAdapter[]
  enabledHarnesses: Harness[]
  displayedSessions: Session[]
  selectedSessionId: string
  isAllProjectsSelected: boolean
  projectName: (project: Project) => string
  basename: (path: string) => string
  harnessName: (id: string) => string
  terminalName: (id: string) => string
}>()

const emit = defineEmits<{
  'select-project': [id: string]
  'add-project': []
  'update:selected-worktree-path': [path: string]
  'add-worktree': []
  'change-default-adapter': [id: string]
  'launch-assistant': [harness: Harness]
  'open-settings-ai': []
  'focus-session': [id: string]
  'close-session': [id: string]
  'clear-sessions': []
}>()
</script>

<template>
  <MainContextStack
    :preferences="preferences"
    :projects="projects"
    :selected-project="selectedProject"
    :selected-worktree-path="selectedWorktreePath"
    :current-worktrees="currentWorktrees"
    :launchable-adapters="launchableAdapters"
    :project-name="projectName"
    :basename="basename"
    @select-project="emit('select-project', $event)"
    @add-project="emit('add-project')"
    @update:selected-worktree-path="emit('update:selected-worktree-path', $event)"
    @add-worktree="emit('add-worktree')"
    @change-default-adapter="emit('change-default-adapter', $event)"
  />

  <AssistantLauncher
    :enabled-harnesses="enabledHarnesses"
    @launch="emit('launch-assistant', $event)"
    @open-settings-ai="emit('open-settings-ai')"
  />

  <SessionList
    :sessions="displayedSessions"
    :projects="projects"
    :selected-session-id="selectedSessionId"
    :group-by-project="isAllProjectsSelected"
    :project-name="projectName"
    :harness-name="harnessName"
    :terminal-name="terminalName"
    @focus="emit('focus-session', $event)"
    @close="emit('close-session', $event)"
    @clear-sessions="emit('clear-sessions')"
  />
</template>
