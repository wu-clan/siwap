export const ALL_PROJECTS_SCOPE_ID = '__all_projects'

export function isAllProjectsScope(id?: string) {
  return id === ALL_PROJECTS_SCOPE_ID
}
