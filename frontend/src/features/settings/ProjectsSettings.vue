<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '../../components/ui/button'
import type { Project } from '../../domain/types'

const props = defineProps<{
  projects: Project[]
  projectName: (project: Project) => string
}>()

const emit = defineEmits<{
  choose: []
  reorder: [ids: string[]]
  'set-default': [id: string]
  remove: [project: Project]
}>()

const { t } = useI18n({ useScope: 'global' })

const localItems = ref<Project[]>([...props.projects])
watch(
  () => props.projects,
  (v) => {
    localItems.value = [...v]
  },
)

const draggedId = ref('')
let lastSwapTarget = ''

function onDragStart(e: DragEvent, id: string) {
  draggedId.value = id
  lastSwapTarget = ''
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', id)
  }
}

function onDragOver(e: DragEvent, targetId: string) {
  e.preventDefault()
  if (!draggedId.value || draggedId.value === targetId || lastSwapTarget === targetId) return
  lastSwapTarget = targetId
  const items = localItems.value
  const fromIdx = items.findIndex((i) => i.id === draggedId.value)
  const toIdx = items.findIndex((i) => i.id === targetId)
  if (fromIdx < 0 || toIdx < 0 || fromIdx === toIdx) return
  const [dragged] = items.splice(fromIdx, 1)
  items.splice(toIdx, 0, dragged)
}

function onDragEnd() {
  if (draggedId.value) {
    const ids = localItems.value.map((i) => i.id)
    const originalIds = props.projects.map((i) => i.id)
    if (ids.join(',') !== originalIds.join(',')) {
      emit('reorder', ids)
    }
  }
  draggedId.value = ''
  lastSwapTarget = ''
}
</script>

<template>
  <section class="settings-page settings-page-layout">
    <div class="settings-actions">
      <Button variant="default" @click="emit('choose')">{{
        t('project.chooseLocalFolder')
      }}</Button>
    </div>
    <div class="native-list">
      <article
        v-for="project in localItems"
        :key="project.id"
        draggable="true"
        :class="[
          'native-row project-item',
          project.isDefault ? 'is-default' : '',
          draggedId === project.id ? 'opacity-30 scale-[0.97]' : '',
        ]"
        @dragstart="onDragStart($event, project.id)"
        @dragover="onDragOver($event, project.id)"
        @dragend="onDragEnd"
      >
        <span class="drag-handle" aria-hidden="true">⋮⋮</span>
        <div>
          <strong>{{ projectName(project) }}</strong>
          <small>{{ project.path }}</small>
        </div>
        <div class="row-actions row-actions-layout">
          <Button v-if="!project.isDefault" @click="emit('set-default', project.id)">{{
            t('project.setDefault')
          }}</Button>
          <Button variant="destructive" @click="emit('remove', project)">{{
            t('common.remove')
          }}</Button>
        </div>
      </article>
    </div>
    <p v-if="projects.length === 0" class="settings-empty">{{ t('project.empty') }}</p>
  </section>
</template>
