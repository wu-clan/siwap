<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '../ui/context-menu'
import type { Session } from '../../domain/types'

const props = defineProps<{
  sessions: Session[]
  selectedSessionId: string
  harnessName: (id: string) => string
  terminalName: (id: string) => string
}>()

const emit = defineEmits<{
  focus: [id: string]
  close: [id: string]
  'clear-sessions': []
}>()

const { t } = useI18n({ useScope: 'global' })

function sessionMeta(session: Session) {
  return props.terminalName(session.adapterId)
}

function sessionA11yLabel(session: Session) {
  return `${props.harnessName(session.harnessId)}, ${sessionMeta(session)}`
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
      <ContextMenu v-for="session in sessions" :key="session.id">
        <ContextMenuTrigger as-child>
          <div
            :class="['session-row', session.id === selectedSessionId ? 'selected' : '']"
            role="option"
            tabindex="0"
            :aria-label="sessionA11yLabel(session)"
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
          <ContextMenuItem variant="destructive" @click="requestClearAll" @select="requestClearAll">
            {{ t('session.removeAll') }}
          </ContextMenuItem>
        </ContextMenuContent>
      </ContextMenu>
      <div v-if="sessions.length === 0" class="session-empty-card" role="status">
        <span aria-hidden="true">⌁</span>
        <strong>{{ t('session.emptyTitle') }}</strong>
        <p>{{ t('session.emptyMain') }}</p>
      </div>
    </div>
  </section>
</template>
