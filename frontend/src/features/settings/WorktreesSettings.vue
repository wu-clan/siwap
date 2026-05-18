<script setup lang="ts">
import { computed, watch } from 'vue'
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
import { createWorktreeFormSchema, type WorktreeFormValues } from '../../domain/formValidation'
import type { Preferences, Project, Worktree } from '../../domain/types'

const props = defineProps<{
  projects: Project[]
  allWorktrees: Worktree[]
  settingsWorktreeProjectId: string
  worktreeBranches: string[]
  canCreateWorktree: boolean
  selectedWorktreePath: string
  worktreeCreateOpen: boolean
  preferences: Preferences
  branchDraft: string
  baseBranchDraft: string
  worktreePathDraft: string
  projectName: (project: Project) => string
}>()

const emit = defineEmits<{
  'open-create': []
  'close-create': []
  create: []
  delete: [item: Worktree]
  'update-settings-worktree-project': [id: string]
  'set-default-worktree': [path: string]
  'update-branch-draft': [value: string]
  'update-base-branch-draft': [value: string]
  'update-worktree-path-draft': [value: string]
  'change-preference': [key: keyof Preferences, value: Preferences[keyof Preferences]]
}>()

const { t } = useI18n({ useScope: 'global' })

const worktreeSchema = computed(() => toTypedSchema(createWorktreeFormSchema(t)))
const baseBranchOptions = computed(() => props.worktreeBranches)
const defaultBaseBranch = computed(
  () =>
    props.worktreeBranches.find((branch) => branch === 'main') ??
    props.worktreeBranches.find((branch) => branch.endsWith('/main')) ??
    props.worktreeBranches.find((branch) => branch === 'master') ??
    props.worktreeBranches.find((branch) => branch.endsWith('/master')) ??
    props.worktreeBranches[0] ??
    '',
)
const worktreeForm = useForm<WorktreeFormValues>({
  validationSchema: worktreeSchema,
  initialValues: {
    branch: props.branchDraft,
    baseBranch: props.baseBranchDraft || defaultBaseBranch.value,
    path: props.worktreePathDraft,
  },
})

watch(
  () => props.worktreeCreateOpen,
  (open) => {
    if (!open) return
    worktreeForm.resetForm({
      values: {
        branch: props.branchDraft,
        baseBranch: props.baseBranchDraft || defaultBaseBranch.value,
        path: props.worktreePathDraft,
      },
    })
  },
)

const filteredWorktrees = computed(() => {
  if (!props.settingsWorktreeProjectId) return props.allWorktrees
  return props.allWorktrees.filter((w) => w.projectId === props.settingsWorktreeProjectId)
})

watch(defaultBaseBranch, (branch) => {
  if (!props.worktreeCreateOpen || !branch) return
  worktreeForm.setFieldValue('baseBranch', branch)
  emit('update-base-branch-draft', branch)
})

function isProtectedBranch(branch: string) {
  return branch === 'main' || branch === 'master'
}

function canRemoveWorktree(item: Worktree) {
  return !item.isMain && !isProtectedBranch(item.branch)
}

function isDefaultWorktree(item: Worktree) {
  if (item.projectId !== props.preferences.selectedProjectId) return false
  return (
    props.selectedWorktreePath === item.path || (item.isMain && props.selectedWorktreePath === '')
  )
}

function canSetDefaultWorktree(item: Worktree) {
  return item.projectId === props.preferences.selectedProjectId && !isDefaultWorktree(item)
}

function setDefaultWorktree(item: Worktree) {
  emit('set-default-worktree', item.isMain ? '' : item.path)
}

const submitWorktree = worktreeForm.handleSubmit((values) => {
  emit('update-branch-draft', values.branch)
  emit('update-base-branch-draft', values.baseBranch || defaultBaseBranch.value)
  emit('update-worktree-path-draft', values.path)
  emit('create')
})
</script>

<template>
  <section class="settings-page settings-page-layout">
    <template v-if="!worktreeCreateOpen">
      <div class="form-grid">
        <div class="field-label">
          {{ t('nav.projects') }}
          <Select
            :model-value="settingsWorktreeProjectId || '__all'"
            @update:model-value="
              emit(
                'update-settings-worktree-project',
                String($event) === '__all' ? '' : String($event),
              )
            "
          >
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__all">{{ t('worktree.allProjects') }}</SelectItem>
              <SelectItem v-for="project in projects" :key="project.id" :value="project.id">
                {{ projectName(project) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
      <div class="settings-actions">
        <Button variant="default" :disabled="!canCreateWorktree" @click="emit('open-create')">{{
          t('worktree.create')
        }}</Button>
      </div>
      <div class="native-list">
        <article
          v-for="item in filteredWorktrees"
          :key="item.id"
          :class="[
            'native-row transition-all',
            isDefaultWorktree(item) ? 'border-primary/50 bg-primary/5' : '',
          ]"
        >
          <div>
            <strong>{{ item.branch || 'detached' }}</strong>
            <span v-if="isDefaultWorktree(item)" class="status-pill">{{
              t('common.default')
            }}</span>
            <small>{{ item.path }}</small>
            <small v-if="item.dirty" class="text-destructive text-xs">{{
              t('worktree.modified')
            }}</small>
          </div>
          <div class="row-actions row-actions-layout">
            <Button v-if="canSetDefaultWorktree(item)" @click="setDefaultWorktree(item)">{{
              t('worktree.setDefault')
            }}</Button>
            <Button
              v-if="canRemoveWorktree(item)"
              variant="destructive"
              @click="emit('delete', item)"
              >{{ t('common.delete') }}</Button
            >
          </div>
        </article>
        <p v-if="filteredWorktrees.length === 0" class="settings-empty">
          {{ t('worktree.empty') }}
        </p>
      </div>
    </template>
    <template v-else>
      <form class="settings-form" @submit.prevent="submitWorktree">
        <div class="form-grid worktree-create-form">
          <FormField v-slot="{ value, handleChange, handleBlur }" name="branch">
            <FormItem>
              <FormLabel>{{ t('worktree.newBranchName') }}</FormLabel>
              <FormControl>
                <Input
                  :model-value="value"
                  placeholder="feature/example"
                  @update:model-value="
                    (next) => {
                      handleChange(next)
                      emit('update-branch-draft', String(next))
                    }
                  "
                  @blur="handleBlur"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
          <FormField v-slot="{ value, handleChange }" name="baseBranch">
            <FormItem>
              <FormLabel>{{ t('worktree.baseBranch') }}</FormLabel>
              <Select
                :model-value="String(value || defaultBaseBranch)"
                @update:model-value="
                  (next) => {
                    const value = String(next)
                    handleChange(value)
                    emit('update-base-branch-draft', value)
                  }
                "
              >
                <FormControl>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem v-for="branch in baseBranchOptions" :key="branch" :value="branch">
                    {{ branch }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>
          <FormField v-slot="{ value, handleChange, handleBlur }" name="path">
            <FormItem>
              <FormLabel>{{ t('worktree.customPath') }}</FormLabel>
              <FormControl>
                <Input
                  :model-value="value"
                  :placeholder="t('worktree.defaultPathHint')"
                  @update:model-value="
                    (next) => {
                      handleChange(next)
                      emit('update-worktree-path-draft', String(next))
                    }
                  "
                  @blur="handleBlur"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
        <div class="settings-actions">
          <Button variant="default" :disabled="!canCreateWorktree" type="submit">{{
            t('common.create')
          }}</Button>
          <Button type="button" @click="emit('close-create')">{{ t('common.cancel') }}</Button>
        </div>
      </form>
    </template>
  </section>
</template>
