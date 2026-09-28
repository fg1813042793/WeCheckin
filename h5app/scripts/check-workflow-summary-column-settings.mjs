import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowSummarySection.vue', import.meta.url),
  'utf8',
)

for (const snippet of [
  'const summaryColumnStorageKey = \'workflow_summary_column_settings_v2\'',
  'const legacySummaryColumnStorageKey = \'workflow_summary_column_settings_v1\'',
  'activeDefinitionId: string',
  'formSelectedKeysByDefinition: Record<string, string[]>',
  'const formColumnKeyword = ref(\'\')',
  'const formFieldDropdownOpen = ref(false)',
  'const selectedColumnDefinitionId = computed<string>',
  'const activeFormFieldGroup = computed',
  'const filteredActiveFormColumns = computed',
  'const formFieldSelectionText = computed',
  'formSelectedKeysByDefinition',
  'function toggleFormFieldDropdown()',
  'formFieldDropdownOpen.value = false',
  'class="workflow-summary__column-settings-form-trigger"',
  ':aria-expanded="formFieldDropdownOpen"',
  'aria-label="选择表单字段"',
  'v-if="formFieldDropdownOpen"',
  'class="workflow-summary__column-settings-form-dropdown"',
  'v-if="filteredActiveFormColumns.length > 0"',
  'class="workflow-summary__column-settings-list workflow-summary__column-settings-list--form"',
  '请选择表单字段',
  'placeholder="搜索字段名称"',
  'aria-label="选择流程字段所属流程"',
  '请选择流程后配置表单字段',
  '未找到匹配字段',
  'v-for="column in filteredActiveFormColumns"',
]) {
  assert.ok(source.includes(snippet), `WorkflowSummarySection.vue missing ${snippet}`)
}

assert.equal(
  source.match(/v-for="group in formFieldGroups"/g)?.length || 0,
  1,
  'column settings must use workflow groups only for the process selector',
)
assert.match(source, /<option v-for="group in formFieldGroups"/)
assert.doesNotMatch(
  source,
  /<template v-if="activeFormFieldGroup">\s*<view class="workflow-summary__column-settings-field-head">/,
  'form fields must stay collapsed behind the form-field selector',
)
assert.doesNotMatch(
  source,
  /<u-checkbox-group v-model="selectedFormColumnKeys">\s*<view[^>]+workflow-summary__column-settings-list--form/,
  'the scrollable form field list must be the checkbox group itself so rows keep their height',
)
for (const snippet of [
  'display: grid !important;',
  'min-height: 28px;',
  'line-height: 20px !important;',
  'overflow-wrap: anywhere;',
]) {
  assert.ok(source.includes(snippet), `WorkflowSummarySection.vue missing readable checkbox rule ${snippet}`)
}

console.log('Workflow summary column settings checks passed')
