<script setup lang="ts">
import {
  assistantInitials,
  assistantLogoKind,
  assistantLogoPaths,
  assistantLogoViewBox,
} from '../../domain/assistant'
import type { Harness } from '../../domain/types'

defineProps<{
  harness: Harness
  mini?: boolean
}>()
</script>

<template>
  <span
    class="assistant-logo"
    :class="[{ 'assistant-mini': mini }, `logo-${assistantLogoKind(harness)}`]"
    :style="{ '--assistant-color': harness.tint }"
    aria-hidden="true"
  >
    <svg v-if="assistantLogoPaths(harness).length" :viewBox="assistantLogoViewBox(harness)" role="img">
      <path
        v-for="(path, index) in assistantLogoPaths(harness)"
        :key="index"
        :d="path.d"
        :fill="path.fill || 'currentColor'"
        :fill-rule="path.fillRule"
        :clip-rule="path.clipRule"
      />
    </svg>
    <span v-else>{{ assistantInitials(harness) }}</span>
  </span>
</template>
