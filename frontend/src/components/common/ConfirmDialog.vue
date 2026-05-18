<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../ui/alert-dialog'

const props = defineProps<{
  open: boolean
  title?: string
  description: string
  confirmLabel?: string
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const { t } = useI18n({ useScope: 'global' })
const pendingChoice = ref<'confirm' | 'cancel' | ''>('')

function mark(choice: 'confirm' | 'cancel') {
  pendingChoice.value = choice
}

function onOpenUpdate(value: boolean) {
  if (value) return
  const choice = pendingChoice.value
  pendingChoice.value = ''
  if (choice === 'confirm') {
    emit('confirm')
    return
  }
  emit('cancel')
}
</script>

<template>
  <AlertDialog :open="props.open" @update:open="onOpenUpdate">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ props.title || t('confirm.title') }}</AlertDialogTitle>
        <AlertDialogDescription>{{ props.description }}</AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel
          @pointerdown="mark('cancel')"
          @keydown.enter="mark('cancel')"
          @keydown.space="mark('cancel')"
        >
          {{ t('common.cancel') }}
        </AlertDialogCancel>
        <AlertDialogAction
          class="bg-destructive text-white hover:bg-destructive/90"
          @pointerdown="mark('confirm')"
          @keydown.enter="mark('confirm')"
          @keydown.space="mark('confirm')"
        >
          {{ props.confirmLabel || t('common.confirm') }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
