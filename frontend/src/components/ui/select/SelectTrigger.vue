<script setup lang="ts">
import type { SelectTriggerProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { ChevronDownIcon } from '@radix-icons/vue'
import { reactiveOmit } from '@vueuse/core'
import { SelectIcon, SelectTrigger } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps<SelectTriggerProps & { class?: HTMLAttributes['class'] }>()
const delegatedProps = reactiveOmit(props, 'class')
</script>

<template>
  <SelectTrigger
    data-slot="select-trigger"
    v-bind="delegatedProps"
    :class="
      cn(
        'border-input bg-background dark:bg-input/30 dark:hover:bg-input/50 flex h-9 w-full min-w-0 items-center justify-between gap-2 rounded-md border px-3 py-2 text-sm whitespace-nowrap shadow-xs transition-[color,box-shadow] outline-none disabled:cursor-not-allowed disabled:opacity-50 data-[placeholder]:text-muted-foreground',
        'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
        '[&>span]:min-w-0 [&>span]:truncate',
        props.class,
      )
    "
  >
    <slot />
    <SelectIcon as-child>
      <ChevronDownIcon class="text-muted-foreground size-4 shrink-0 opacity-50" />
    </SelectIcon>
  </SelectTrigger>
</template>
