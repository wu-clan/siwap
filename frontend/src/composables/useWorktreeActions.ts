import type { ComputedRef, Ref } from 'vue'
import { CreateWorktree, RemoveWorktree } from '../../bindings/siwap/internal/desktop/app'
import type { Preferences, Project, Worktree, WorktreeActionResult } from '../domain/types'

type Run = <T>(label: string, fn: () => Promise<T>) => Promise<T | undefined>
type Translate = (key: string, named?: Record<string, unknown>) => string
type Confirm = (description: string) => Promise<boolean>

/** useWorktreeActions 封装 worktree 创建、删除和创建面板状态 */
export function useWorktreeActions(options: {
  selectedProject: ComputedRef<Project | undefined>
  selectedWorktreePath: Ref<string>
  branchDraft: Ref<string>
  baseBranchDraft: Ref<string>
  worktreePathDraft: Ref<string>
  worktreeCreateOpen: Ref<boolean>
  preferences: Ref<Preferences>
  actionMessage: Ref<string>
  run: Run
  t: Translate
  refreshWorktreeBranches: () => Promise<void>
  syncWorktrees: (next?: Worktree[]) => Promise<void>
  canCreateWorktree: ComputedRef<boolean>
  confirm: Confirm
}) {
  const {
    selectedProject,
    selectedWorktreePath,
    branchDraft,
    baseBranchDraft,
    worktreePathDraft,
    worktreeCreateOpen,
    preferences,
    actionMessage,
    run,
    t,
    refreshWorktreeBranches,
    syncWorktrees,
    canCreateWorktree,
    confirm,
  } = options

  function removeDirtyWorktreePrompt(name: string) {
    return t('confirm.removeDirtyWorktree', { name })
  }

  function removeWorktreePrompt(name: string) {
    return t('confirm.removeWorktree', { name })
  }

  function isProtectedWorktree(item: Worktree) {
    return item.isMain || item.branch === 'main' || item.branch === 'master'
  }

  async function openWorktreeCreate() {
    if (!selectedProject.value) {
      actionMessage.value = t('project.selectFirst')
      return
    }
    await refreshWorktreeBranches()
    if (!canCreateWorktree.value) {
      actionMessage.value = t('worktree.gitRequired')
      return
    }
    worktreeCreateOpen.value = true
  }

  function closeWorktreeCreate() {
    worktreeCreateOpen.value = false
  }

  async function createWorktree() {
    if (!selectedProject.value) return
    if (!canCreateWorktree.value) {
      actionMessage.value = t('worktree.gitRequired')
      return
    }
    const created = await run(
      'worktree.create',
      () =>
        CreateWorktree({
          projectId: selectedProject.value!.id,
          projectPath: selectedProject.value!.path,
          branch: branchDraft.value,
          baseBranch: baseBranchDraft.value,
          path: worktreePathDraft.value,
          baseDir:
            preferences.value.worktreeLocation === 'custom'
              ? preferences.value.worktreeBaseDir
              : '',
        }) as unknown as Promise<WorktreeActionResult>,
    )
    if (!created) return
    // 创建结果已包含 Go 聚合后的最新列表，前端直接消费而不再额外查询
    await syncWorktrees(created.worktrees ?? [])
    branchDraft.value = ''
    baseBranchDraft.value = ''
    worktreePathDraft.value = ''
    worktreeCreateOpen.value = false
    actionMessage.value = t('worktree.createdSelectFromMain')
  }

  async function deleteWorktree(item: Worktree) {
    if (isProtectedWorktree(item)) {
      actionMessage.value = t('worktree.protected')
      return
    }
    const force = item.dirty
    const name = item.branch || item.path
    if (force && !(await confirm(removeDirtyWorktreePrompt(name)))) return
    if (!force && !(await confirm(removeWorktreePrompt(name)))) return
    const result = await run(
      'action.deleteWorktree',
      () =>
        RemoveWorktree(
          item.projectId,
          item.path,
          force,
        ) as unknown as Promise<WorktreeActionResult>,
    )
    if (selectedWorktreePath.value === item.path) selectedWorktreePath.value = ''
    if (result) {
      actionMessage.value = result.action.message
      // 删除结果同样以后端返回的全量 worktree 列表为准
      await syncWorktrees(result.worktrees ?? [])
    }
  }

  return { openWorktreeCreate, closeWorktreeCreate, createWorktree, deleteWorktree }
}
