export interface WorkflowRecordColumn {
  key: string
  label: string
  width?: string
  mobileHidden?: boolean
  copyable?: boolean
  configurable?: boolean
}

export function configurableWorkflowRecordColumns(columns: WorkflowRecordColumn[]) {
  return columns.filter(column => column.configurable === true)
}

export function visibleWorkflowRecordColumns(columns: WorkflowRecordColumn[], hiddenKeys: string[]) {
  const hidden = new Set(hiddenKeys)
  return columns.filter(column => !column.configurable || !hidden.has(column.key))
}

export function normalizeWorkflowRecordHiddenKeys(columns: WorkflowRecordColumn[], values: unknown) {
  if (!Array.isArray(values))
    return []
  const configurableKeys = new Set(configurableWorkflowRecordColumns(columns).map(column => column.key))
  return [...new Set(values.map(value => String(value)).filter(key => configurableKeys.has(key)))]
}
