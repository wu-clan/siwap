<script setup lang="ts">
import type { SelectContentEmits, SelectContentProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { ChevronDownIcon, ChevronUpIcon } from '@radix-icons/vue'
import { reactiveOmit } from '@vueuse/core'
import {
  SelectContent,
  SelectPortal,
  SelectScrollDownButton,
  SelectScrollUpButton,
  SelectViewport,
  useForwardPropsEmits,
} from 'reka-ui'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<SelectContentProps & { class?: HTMLAttributes['class'] }>(),
  {
    align: 'start',
    position: 'popper',
    sideOffset: 6,
    bodyLock: false,
  },
)
const emits = defineEmits<SelectContentEmits>()

const delegatedProps = reactiveOmit(props, 'class')
const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <SelectPortal>
    <SelectContent
      data-slot="select-content"
      v-bind="{ ...forwarded, ...$attrs }"
      :class="
        cn(
          'bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 relative z-[100] max-h-72 min-w-[8rem] overflow-hidden rounded-md border shadow-md',
          'data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1',
          props.class,
        )
      "
    >
      <SelectScrollUpButton class="flex cursor-default items-center justify-center py-1">
        <ChevronUpIcon class="size-4" />
      </SelectScrollUpButton>
      <SelectViewport class="p-1" :style="{ minWidth: 'var(--reka-select-trigger-width)' }">
        <slot />
      </SelectViewport>
      <SelectScrollDownButton class="flex cursor-default items-center justify-center py-1">
        <ChevronDownIcon class="size-4" />
      </SelectScrollDownButton>
    </SelectContent>
  </SelectPortal>
</template>
