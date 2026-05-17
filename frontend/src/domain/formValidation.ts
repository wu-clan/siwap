import { z } from 'zod'

type Translate = (key: string) => string

const gitBranchForbiddenPattern = /[\s~^:?*[\]\\]|(\.\.)|(@\{)|(^\/)|(\/$)|(\.lock$)|(\.$)/

function isOptionalAbsolutePath(value: string) {
  const path = value.trim()
  if (!path) return true
  return path.startsWith('/') || path.startsWith('~/') || /^[A-Za-z]:[\\/]/.test(path) || path.startsWith('\\\\')
}

export function createWorktreeFormSchema(t: Translate) {
  return z.object({
    branch: z.string()
      .trim()
      .max(120, t('validation.branchTooLong'))
      .refine((value) => !value || !gitBranchForbiddenPattern.test(value), t('validation.branchInvalid')),
    baseBranch: z.string().trim(),
    path: z.string()
      .trim()
      .refine(isOptionalAbsolutePath, t('validation.pathMustBeAbsolute')),
  })
}

export function createTerminalProfileFormSchema(t: Translate) {
  return z.object({
    label: z.string().trim().min(1, t('validation.terminalNameRequired')).max(80, t('validation.terminalNameTooLong')),
    executablePath: z.string().trim().min(1, t('validation.terminalPathRequired')),
  })
}

export type WorktreeFormValues = z.infer<ReturnType<typeof createWorktreeFormSchema>>
export type TerminalProfileFormValues = z.infer<ReturnType<typeof createTerminalProfileFormSchema>>
