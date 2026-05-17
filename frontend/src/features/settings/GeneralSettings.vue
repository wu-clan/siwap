<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../components/ui/select'
import { Switch } from '../../components/ui/switch'
import type { Preferences } from '../../domain/types'

defineProps<{
  preferences: Preferences
}>()

const emit = defineEmits<{
  'change-preference': [key: keyof Preferences, value: Preferences[keyof Preferences]]
  'toggle-always-on-top': []
  'reset-window': []
  'reset-preferences': []
}>()

const { t } = useI18n({ useScope: 'global' })

function changeStringPreference(key: keyof Preferences, event: Event) {
  emit('change-preference', key, (event.target as HTMLInputElement | HTMLSelectElement).value)
}
</script>

<template>
  <section class="settings-page settings-page-layout">
    <div class="form-grid two">
      <label class="field-label">{{ t('settings.language') }}
        <Select :model-value="preferences.language" @update:model-value="emit('change-preference', 'language', String($event))">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="zh-CN">{{ t('language.chinese') }}</SelectItem>
            <SelectItem value="en">English</SelectItem>
          </SelectContent>
        </Select>
      </label>
      <label class="field-label">{{ t('settings.appearance') }}
        <Select :model-value="preferences.appearance" @update:model-value="emit('change-preference', 'appearance', String($event))">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="system">{{ t('appearance.system') }}</SelectItem>
            <SelectItem value="light">{{ t('appearance.light') }}</SelectItem>
            <SelectItem value="dark">{{ t('appearance.dark') }}</SelectItem>
          </SelectContent>
        </Select>
      </label>
      <label class="field-label">{{ t('settings.summonShortcut') }}
        <Input :value="preferences.globalShortcut" placeholder="Control+Command+S" @change="changeStringPreference('globalShortcut', $event)" />
      </label>
    </div>
    <div class="toggle-list">
      <label><Switch :model-value="preferences.alwaysOnTop" @update:model-value="emit('toggle-always-on-top')" /> {{ t('settings.alwaysOnTop') }}</label>
      <label><Switch :model-value="preferences.autohideOnBlur" @update:model-value="emit('change-preference', 'autohideOnBlur', $event)" /> {{ t('settings.hideOnBlur') }}</label>
    </div>
    <div class="settings-actions">
      <Button @click="emit('reset-window')">{{ t('settings.resetWindowPosition') }}</Button>
      <Button @click="emit('reset-preferences')">{{ t('settings.restoreDefaults') }}</Button>
    </div>
  </section>
</template>
