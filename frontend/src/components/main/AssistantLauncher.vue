<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AssistantLogo from '../common/AssistantLogo.vue'
import { Button } from '../ui/button'
import type { Harness } from '../../domain/types'

defineProps<{
  enabledHarnesses: Harness[]
}>()

const emit = defineEmits<{
  launch: [harness: Harness]
  'open-settings-ai': []
}>()

const { t } = useI18n({ useScope: 'global' })
</script>

<template>
  <section class="launcher-section launcher-layout" aria-labelledby="ai-title">
    <div class="section-heading section-heading-layout">
      <h2 id="ai-title">AI</h2>
    </div>
    <div class="assistant-grid assistant-grid-layout" role="list">
      <Button
        v-for="harness in enabledHarnesses"
        :key="harness.id"
        variant="ghost"
        class="assistant-tile"
        :title="harness.label"
        :aria-label="harness.label"
        :style="{ '--assistant-color': harness.tint }"
        role="listitem"
        @click="emit('launch', harness)"
      >
        <AssistantLogo :harness="harness" />
      </Button>
    </div>
    <Button
      v-if="enabledHarnesses.length === 0"
      class="empty-action"
      @click="emit('open-settings-ai')"
      >{{ t('ai.enableAssistants') }}</Button
    >
  </section>
</template>
