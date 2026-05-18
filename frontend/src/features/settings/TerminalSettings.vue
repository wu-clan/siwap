<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { toTypedSchema } from '@vee-validate/zod'
import { useForm } from 'vee-validate'
import { useI18n } from 'vue-i18n'
import { Button } from '../../components/ui/button'
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '../../components/ui/form'
import { Input } from '../../components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../../components/ui/select'
import { Switch } from '../../components/ui/switch'
import {
  createTerminalProfileFormSchema,
  type TerminalProfileFormValues,
} from '../../domain/formValidation'
import type { Preferences, TerminalAdapter, TerminalProfile } from '../../domain/types'

const props = defineProps<{
  preferences: Preferences
  adapters: TerminalAdapter[]
  terminalProfiles: TerminalProfile[]
  terminalProfileOpen: boolean
  profileDraft: TerminalProfile
  availabilityText: (adapter: TerminalAdapter) => string
}>()

const emit = defineEmits<{
  'change-default-adapter': [id: string]
  'toggle-adapter': [id: string, enabled: boolean]
  'open-profile': [profile?: TerminalProfile]
  'close-profile': []
  'choose-executable': []
  'save-profile': [profile: TerminalProfile]
  'remove-profile': [id: string]
  reorder: [ids: string[]]
  'update-profile-field': [
    key: keyof TerminalProfile,
    value: TerminalProfile[keyof TerminalProfile],
  ]
}>()

const { t } = useI18n({ useScope: 'global' })

const draggableAdapters = computed(() => props.adapters.filter((a) => a.id !== 'auto'))
const localItems = ref<TerminalAdapter[]>([...draggableAdapters.value])
const terminalProfileSchema = computed(() => toTypedSchema(createTerminalProfileFormSchema(t)))
const terminalProfileForm = useForm<TerminalProfileFormValues>({
  validationSchema: terminalProfileSchema,
  initialValues: {
    label: props.profileDraft.label,
    executablePath: props.profileDraft.executablePath,
  },
})

watch(draggableAdapters, (v) => {
  localItems.value = [...v]
})
watch(
  () => props.terminalProfileOpen,
  (open) => {
    if (!open) return
    terminalProfileForm.resetForm({
      values: {
        label: props.profileDraft.label,
        executablePath: props.profileDraft.executablePath,
      },
    })
  },
)
watch(
  () => props.profileDraft.label,
  (label) => {
    if (props.terminalProfileOpen) terminalProfileForm.setFieldValue('label', label)
  },
)
watch(
  () => props.profileDraft.executablePath,
  (path) => {
    if (props.terminalProfileOpen) terminalProfileForm.setFieldValue('executablePath', path)
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
    const originalIds = draggableAdapters.value.map((i) => i.id)
    if (ids.join(',') !== originalIds.join(',')) {
      emit('reorder', ids)
    }
  }
  draggedId.value = ''
  lastSwapTarget = ''
}

const submitTerminalProfile = terminalProfileForm.handleSubmit((values) => {
  emit('save-profile', {
    ...props.profileDraft,
    label: values.label,
    executablePath: values.executablePath,
  })
})
</script>

<template>
  <section class="settings-page settings-page-layout">
    <template v-if="!terminalProfileOpen">
      <div class="field-label">
        {{ t('terminal.default') }}
        <Select
          :model-value="preferences.defaultAdapterId"
          @update:model-value="emit('change-default-adapter', String($event))"
        >
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem
              v-for="adapter in adapters"
              :key="adapter.id"
              :value="adapter.id"
              :disabled="adapter.id !== 'auto' && !adapter.enabled"
            >
              {{ adapter.label }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <TransitionGroup tag="div" name="drag-list" class="native-list">
        <article
          v-for="adapter in localItems"
          :key="adapter.id"
          draggable="true"
          :class="['native-row', draggedId === adapter.id ? 'opacity-30 scale-[0.97]' : '']"
          @dragstart="onDragStart($event, adapter.id)"
          @dragover="onDragOver($event, adapter.id)"
          @dragend="onDragEnd"
        >
          <span class="drag-handle" aria-hidden="true">⋮⋮</span>
          <div>
            <strong>{{ adapter.label }}</strong>
            <small
              >{{ availabilityText(adapter)
              }}{{ adapter.message ? ` · ${adapter.message}` : '' }}</small
            >
          </div>
          <div class="row-actions row-actions-layout">
            <Switch
              :model-value="adapter.enabled"
              :disabled="!adapter.installed"
              @update:model-value="(v: boolean) => emit('toggle-adapter', adapter.id, v)"
            />
          </div>
        </article>
      </TransitionGroup>
      <h3>{{ t('terminal.customTerminals') }}</h3>
      <div class="settings-actions">
        <Button variant="default" @click="emit('open-profile')">{{ t('terminal.add') }}</Button>
      </div>
      <div class="native-list compact">
        <article v-for="profile in terminalProfiles" :key="profile.id" class="native-row">
          <div>
            <strong>{{ profile.label }}</strong>
            <small>{{ profile.executablePath }}</small>
          </div>
          <div class="row-actions row-actions-layout">
            <Button @click="emit('open-profile', profile)">{{ t('common.edit') }}</Button>
            <Button variant="destructive" @click="emit('remove-profile', profile.id)">{{
              t('common.remove')
            }}</Button>
          </div>
        </article>
        <p v-if="terminalProfiles.length === 0" class="settings-empty">
          {{ t('terminal.emptyCustom') }}
        </p>
      </div>
    </template>
    <template v-else>
      <form class="settings-form" @submit.prevent="submitTerminalProfile">
        <div class="form-grid terminal-profile-form">
          <FormField v-slot="{ value, handleChange, handleBlur }" name="label">
            <FormItem>
              <FormLabel>{{ t('common.name') }}</FormLabel>
              <FormControl>
                <Input
                  :model-value="value"
                  placeholder="My Terminal"
                  @update:model-value="
                    (next) => {
                      handleChange(next)
                      emit('update-profile-field', 'label', String(next))
                    }
                  "
                  @blur="handleBlur"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
          <FormField v-slot="{ value }" name="executablePath">
            <FormItem>
              <FormLabel>{{ t('terminal.executablePath') }}</FormLabel>
              <FormControl>
                <div class="path-picker" :aria-invalid="!value">
                  <span>{{ value || t('common.notSelected') }}</span>
                  <Button type="button" @click="emit('choose-executable')">{{
                    t('common.choose')
                  }}</Button>
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
        <div class="settings-actions">
          <Button variant="default" type="submit">{{ t('terminal.save') }}</Button>
          <Button type="button" @click="emit('close-profile')">{{ t('common.cancel') }}</Button>
        </div>
      </form>
    </template>
  </section>
</template>
