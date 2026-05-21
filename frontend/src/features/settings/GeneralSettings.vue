<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../../components/ui/select'
import { Switch } from '../../components/ui/switch'
import type { Preferences } from '../../domain/types'

defineProps<{
  preferences: Preferences
  platform: string
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

function handleShortcutKeydown(event: KeyboardEvent) {
  const shortcut = shortcutFromKeyboardEvent(event)
  if (!shortcut) return
  event.preventDefault()
  emit('change-preference', 'globalShortcut', shortcut)
}

function shortcutFromKeyboardEvent(event: KeyboardEvent) {
  const key = normalizeShortcutKey(event.code, event.key)
  if (!key) return ''
  const modifiers: string[] = []
  if (event.ctrlKey) modifiers.push('Control')
  if (event.metaKey) modifiers.push('Command')
  if (event.altKey) modifiers.push('Alt')
  if (event.shiftKey) modifiers.push('Shift')
  if (modifiers.length === 0) return ''
  return [...modifiers, key].join('+')
}

function normalizeShortcutKey(code: string, key: string) {
  if (/^Key[A-Z]$/.test(code)) return code.slice(3)
  if (/^Digit[0-9]$/.test(code)) return code.slice(5)
  if (/^Numpad[0-9]$/.test(code)) return code.slice(7)
  if (/^F\d{1,2}$/.test(code)) return code
  if (code === 'Space') return 'Space'
  const lowered = key.toLowerCase()
  if (lowered === 'spacebar') return 'Space'
  if (['tab', 'escape', 'enter'].includes(lowered)) return ''
  if (key.length === 1 && /[a-z0-9]/i.test(key)) {
    return key.toUpperCase()
  }
  return ''
}
</script>

<template>
  <section class="settings-page settings-page-layout">
    <div class="form-grid two">
      <label class="field-label"
        >{{ t('settings.language') }}
        <Select
          :model-value="preferences.language"
          @update:model-value="emit('change-preference', 'language', String($event))"
        >
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="zh-CN">{{ t('language.chinese') }}</SelectItem>
            <SelectItem value="en">English</SelectItem>
          </SelectContent>
        </Select>
      </label>
      <label class="field-label"
        >{{ t('settings.appearance') }}
        <Select
          :model-value="preferences.appearance"
          @update:model-value="emit('change-preference', 'appearance', String($event))"
        >
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="system">{{ t('appearance.system') }}</SelectItem>
            <SelectItem value="light">{{ t('appearance.light') }}</SelectItem>
            <SelectItem value="dark">{{ t('appearance.dark') }}</SelectItem>
          </SelectContent>
        </Select>
      </label>
      <label class="field-label"
        >{{ t('settings.summonShortcut') }}
        <Input
          :value="preferences.globalShortcut"
          placeholder="Control+Command+S"
          @keydown="handleShortcutKeydown"
          @change="changeStringPreference('globalShortcut', $event)"
        />
      </label>
    </div>
    <div class="toggle-list">
      <label
        ><Switch
          :model-value="preferences.alwaysOnTop"
          @update:model-value="emit('toggle-always-on-top')"
        />
        {{ t('settings.alwaysOnTop') }}</label
      >
      <label v-if="platform === 'darwin'"
        ><Switch
          :model-value="preferences.showDockIcon"
          @update:model-value="emit('change-preference', 'showDockIcon', $event)"
        />
        {{ t('settings.showDockIcon') }}</label
      >
    </div>
    <div class="settings-actions">
      <Button @click="emit('reset-window')">{{ t('settings.resetWindowPosition') }}</Button>
      <Button @click="emit('reset-preferences')">{{ t('settings.restoreDefaults') }}</Button>
    </div>
  </section>
</template>
