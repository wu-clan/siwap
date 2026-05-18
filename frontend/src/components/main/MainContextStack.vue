<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from '../ui/select'
import type { Preferences, Project, TerminalAdapter, Worktree } from '../../domain/types'
import { ALL_PROJECTS_SCOPE_ID } from '../../domain/projectScope'

const props = defineProps<{
  preferences: Preferences
  projects: Project[]
  selectedProject?: Project
  selectedWorktreePath: string
  currentWorktrees: Worktree[]
  launchableAdapters: TerminalAdapter[]
  projectName: (project: Project) => string
  basename: (path: string) => string
}>()

const emit = defineEmits<{
  'select-project': [id: string]
  'add-project': []
  'update:selected-worktree-path': [path: string]
  'add-worktree': []
  'change-default-adapter': [id: string]
}>()

const { t } = useI18n({ useScope: 'global' })

function onProjectChange(value: string) {
  if (value === '__add_project') {
    emit('add-project')
    return
  }
  emit('select-project', value)
}

function onWorktreeChange(value: string) {
  if (value === '__add_worktree') {
    emit('add-worktree')
    return
  }
  emit('update:selected-worktree-path', value === '__project' ? '' : value)
}
</script>

<template>
  <section class="context-stack context-stack-layout" :aria-label="t('main.launchContext')">
    <div class="control-field">
      <span id="project-select-label">{{ t('nav.projects') }}</span>
      <Select
        :model-value="preferences.selectedProjectId"
        @update:model-value="onProjectChange(String($event))"
      >
        <SelectTrigger aria-labelledby="project-select-label">
          <SelectValue :placeholder="t('project.noneSelected')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem :value="ALL_PROJECTS_SCOPE_ID">{{ t('project.allProjects') }}</SelectItem>
          <SelectItem v-for="project in projects" :key="project.id" :value="project.id">
            {{ props.projectName(project)
            }}{{ project.isDefault ? ` · ${t('common.default')}` : '' }}
          </SelectItem>
          <SelectSeparator v-if="projects.length > 0" />
          <SelectItem value="__add_project">{{ t('project.addEllipsis') }}</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="control-field">
      <span id="worktree-select-label">{{ t('nav.worktrees') }}</span>
      <Select
        :model-value="selectedWorktreePath || '__project'"
        :disabled="!selectedProject"
        @update:model-value="onWorktreeChange(String($event))"
      >
        <SelectTrigger aria-labelledby="worktree-select-label">
          <SelectValue :placeholder="t('worktree.projectDirectory')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="__project">{{ t('worktree.projectDirectory') }}</SelectItem>
          <SelectItem v-for="item in currentWorktrees" :key="item.id" :value="item.path">
            {{ item.branch || props.basename(item.path)
            }}{{ item.dirty ? ` · ${t('worktree.modified')}` : '' }}
          </SelectItem>
          <SelectSeparator />
          <SelectItem value="__add_worktree">{{ t('worktree.addEllipsis') }}</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="control-field">
      <span id="terminal-select-label">{{ t('nav.terminal') }}</span>
      <Select
        :model-value="preferences.defaultAdapterId"
        @update:model-value="emit('change-default-adapter', String($event))"
      >
        <SelectTrigger aria-labelledby="terminal-select-label">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem
            v-for="adapter in launchableAdapters"
            :key="adapter.id"
            :value="adapter.id"
            :disabled="adapter.id !== 'auto' && !adapter.installed"
          >
            {{ adapter.label
            }}{{
              adapter.id !== 'auto' && !adapter.installed ? ` · ${t('terminal.notInstalled')}` : ''
            }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>
  </section>
</template>
