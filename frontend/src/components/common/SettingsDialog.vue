<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '../ui/button'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from '../ui/sidebar'
import type { SettingsSection, SettingsSectionItem } from '../../domain/settings'

defineProps<{
  visible: boolean
  isSettingsWindow: boolean
  settingsSection: SettingsSection
  settingsSections: readonly SettingsSectionItem[]
}>()

const emit = defineEmits<{
  'update:settings-section': [section: SettingsSection]
  close: []
}>()

const { t } = useI18n({ useScope: 'global' })
</script>

<template>
  <section
    v-if="visible"
    class="settings-dialog"
    :role="isSettingsWindow ? undefined : 'dialog'"
    :aria-modal="isSettingsWindow ? undefined : 'true'"
    :aria-label="t('settings.title')"
  >
    <SidebarProvider class="settings-sidebar-provider">
      <Sidebar collapsible="none" class="settings-sidebar">
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem v-for="item in settingsSections" :key="item.id">
                  <SidebarMenuButton
                    size="lg"
                    :is-active="settingsSection === item.id"
                    @click="emit('update:settings-section', item.id)"
                  >
                    {{ t(item.label) }}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>
      <SidebarInset class="settings-inset">
        <section class="settings-pane settings-pane-layout">
          <slot />
        </section>
        <footer class="settings-footer settings-footer-layout">
          <Button variant="default" @click="emit('close')">{{ t('common.done') }}</Button>
        </footer>
      </SidebarInset>
    </SidebarProvider>
  </section>
</template>
