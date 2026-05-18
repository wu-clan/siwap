<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AssistantLogo from '../../components/common/AssistantLogo.vue'
import { Input } from '../../components/ui/input'
import { Switch } from '../../components/ui/switch'
import type { Harness } from '../../domain/types'

const props = defineProps<{
  harnesses: Harness[]
}>()

const emit = defineEmits<{
  save: [harness: Harness]
  reorder: [ids: string[]]
}>()

const { t } = useI18n({ useScope: 'global' })

const localItems = ref<Harness[]>([...props.harnesses])
watch(
  () => props.harnesses,
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
    const originalIds = props.harnesses.map((i) => i.id)
    if (ids.join(',') !== originalIds.join(',')) {
      emit('reorder', ids)
    }
  }
  draggedId.value = ''
  lastSwapTarget = ''
}

function toggleEnabled(harness: Harness, value: boolean) {
  emit('save', { ...harness, enabled: value })
}
</script>

<template>
  <section class="settings-page settings-page-layout">
    <TransitionGroup tag="div" name="drag-list" class="native-list">
      <article
        v-for="harness in localItems"
        :key="harness.id"
        draggable="true"
        :class="[
          'native-row assistant-config',
          draggedId === harness.id ? 'opacity-30 scale-[0.97]' : '',
        ]"
        @dragstart="onDragStart($event, harness.id)"
        @dragover="onDragOver($event, harness.id)"
        @dragend="onDragEnd"
      >
        <span class="drag-handle" aria-hidden="true">⋮⋮</span>
        <AssistantLogo :harness="harness" mini />
        <div class="ai-row-content">
          <strong class="ai-row-name">{{ harness.label }}</strong>
          <label class="field-label"
            >{{ t('common.command') }}
            <Input v-model="harness.command" @change="emit('save', harness)" />
          </label>
        </div>
        <div class="row-actions row-actions-layout">
          <Switch
            :model-value="harness.enabled"
            @update:model-value="(v: boolean) => toggleEnabled(harness, v)"
          />
        </div>
      </article>
    </TransitionGroup>
  </section>
</template>
