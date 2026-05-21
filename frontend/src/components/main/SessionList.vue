<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '../ui/context-menu'
import type { Project, Session } from '../../domain/types'

const props = defineProps<{
  sessions: Session[]
  projects: Project[]
  selectedSessionId: string
  groupByProject: boolean
  projectName: (project: Project) => string
  harnessName: (id: string) => string
  terminalName: (id: string) => string
}>()

const emit = defineEmits<{
  focus: [id: string]
  close: [id: string]
  'clear-sessions': []
}>()

const { t } = useI18n({ useScope: 'global' })
const fallbackGroupId = '__unknown_project'

const sessionGroups = computed(() => {
  if (props.sessions.length === 0) return []

  if (!props.groupByProject) {
    return [{ id: '__all_sessions', label: '', showLabel: false, sessions: props.sessions }]
  }

  const groups = new Map<string, Session[]>()
  for (const session of props.sessions) {
    const groupId = session.projectId || fallbackGroupId
    const sessions = groups.get(groupId) ?? []
    sessions.push(session)
    groups.set(groupId, sessions)
  }

  const orderedGroups = props.projects
    .map((project) => ({
      id: project.id,
      label: props.projectName(project),
      showLabel: true,
      sessions: groups.get(project.id) ?? [],
    }))
    .filter((group) => group.sessions.length > 0)

  for (const [id, sessions] of groups) {
    if (id === fallbackGroupId || props.projects.some((project) => project.id === id)) continue
    orderedGroups.push({
      id,
      label: projectLabel(id, sessions),
      showLabel: true,
      sessions,
    })
  }

  const unknownSessions = groups.get(fallbackGroupId)
  if (unknownSessions?.length) {
    orderedGroups.push({
      id: fallbackGroupId,
      label: t('project.unknownProject'),
      showLabel: true,
      sessions: unknownSessions,
    })
  }

  return orderedGroups
})

function sessionMeta(session: Session) {
  return props.terminalName(session.adapterId)
}

function projectLabel(projectId: string, sessions: Session[]) {
  if (projectId === fallbackGroupId) return t('project.unknownProject')
  return sessions[0]?.projectName || t('project.unknownProject')
}

function sessionA11yLabel(session: Session, groupLabel: string) {
  const baseLabel = `${props.harnessName(session.harnessId)}, ${props.terminalName(session.adapterId)}`
  return groupLabel ? `${groupLabel}, ${baseLabel}` : baseLabel
}

function requestClearAll() {
  emit('clear-sessions')
}
</script>

<template>
  <section class="session-section session-layout" aria-labelledby="sessions-title">
    <div class="section-heading section-heading-layout session-heading-layout">
      <h2 id="sessions-title">{{ t('nav.sessions') }}</h2>
    </div>
    <div
      class="session-list session-list-layout"
      role="listbox"
      :aria-label="t('session.currentSessions')"
    >
      <div
        v-for="group in sessionGroups"
        :key="group.id"
        class="session-project-group"
        role="group"
        :aria-label="group.label || t('session.currentSessions')"
      >
        <div v-if="group.showLabel" class="session-project-title">
          {{ group.label }}
        </div>
        <ContextMenu v-for="session in group.sessions" :key="session.id">
          <ContextMenuTrigger as-child>
            <div
              :class="['session-row', session.id === selectedSessionId ? 'selected' : '']"
              role="option"
              tabindex="0"
              :aria-label="sessionA11yLabel(session, group.label)"
              :aria-selected="session.id === selectedSessionId"
              @click="emit('focus', session.id)"
              @keydown.enter="emit('focus', session.id)"
              @keydown.space.prevent="emit('focus', session.id)"
            >
              <span class="session-main">
                <span class="session-title-line">
                  <strong>{{ harnessName(session.harnessId) }}</strong>
                </span>
                <small>{{ sessionMeta(session) }}</small>
              </span>
            </div>
          </ContextMenuTrigger>
          <ContextMenuContent class="w-36">
            <ContextMenuItem variant="destructive" @select="emit('close', session.id)">
              {{ t('common.remove') }}
            </ContextMenuItem>
            <ContextMenuSeparator />
            <ContextMenuItem
              variant="destructive"
              @click="requestClearAll"
              @select="requestClearAll"
            >
              {{ t('session.removeAll') }}
            </ContextMenuItem>
          </ContextMenuContent>
        </ContextMenu>
      </div>
      <div v-if="sessions.length === 0" class="session-empty-card" role="status">
        <span aria-hidden="true">⌁</span>
        <strong>{{ t('session.emptyTitle') }}</strong>
        <p>{{ t('session.emptyMain') }}</p>
      </div>
    </div>
  </section>
</template>
