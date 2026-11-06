import fs from 'node:fs'
import { createRequire } from 'node:module'
import path from 'node:path'
import process from 'node:process'
import ts from 'typescript'

const root = process.cwd()
const require = createRequire(import.meta.url)
const filename = path.join(root, 'src/pages/workflow/workflow-form.ts')
const source = fs.readFileSync(filename, 'utf8')
const output = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
  fileName: filename,
}).outputText
const module = { exports: {} }
function moduleRequire(id) {
  if (id === './workflow-calculation')
    return { calculateWorkflowFormData: (_fields, values) => values }
  return require(id)
}
// eslint-disable-next-line no-new-func
new Function('module', 'exports', 'require', output)(module, module.exports, moduleRequire)

const {
  validateWorkflowFormData,
  validateWorkflowFormDataDetailed,
  workflowDetailCellValidationKey,
  workflowFieldValidationTrigger,
} = module.exports
const fields = [
  { key: 'grade', label: '上级分档', type: 'radio', options: [{ label: '01', value: '01' }, { label: '1', value: '1' }] },
  {
    key: 'result',
    label: '复核结果',
    type: 'select',
    options: [{ label: '01', value: '01' }, { label: '1', value: '1' }],
    rules: [{ id: 'same_grade', type: 'compare_field', field: 'grade', operator: 'eq' }],
  },
]

if (!validateWorkflowFormData(fields, { grade: '1', result: '01' }).result) {
  throw new Error('选项值 01 与 1 不应按数值判定为相等')
}
if (Object.keys(validateWorkflowFormData(fields, { grade: '01', result: '01' })).length !== 0) {
  throw new Error('完全相同的选项值应通过字段比较')
}

if (workflowFieldValidationTrigger({ type: 'text' }) !== 'blur')
  throw new Error('文本字段应在失焦时提前校验')
if (workflowFieldValidationTrigger({ type: 'number' }) !== 'blur')
  throw new Error('数字字段应在失焦时提前校验')
if (workflowFieldValidationTrigger({ type: 'select' }) !== 'change')
  throw new Error('选择字段应在值变化时提前校验')
if (workflowFieldValidationTrigger({ type: 'attachment' }) !== 'change')
  throw new Error('附件字段应在值变化时提前校验')
if (workflowFieldValidationTrigger({ type: 'detail_list' }) !== 'change')
  throw new Error('明细表应在结构或内容变化时提前校验')

const detailFields = [{
  key: 'items',
  label: '费用明细',
  type: 'detail_list',
  rowKey: 'id',
  columns: [
    { key: 'name', label: '项目', type: 'textarea', required: true },
    { key: 'amount', label: '金额', type: 'number', min: 0 },
  ],
}]
const detailValues = { items: [{ id: 'row-a', name: '', amount: -1 }] }
const detailed = validateWorkflowFormDataDetailed(detailFields, detailValues)
const nameKey = workflowDetailCellValidationKey('items', 'row-a', 'name')
const amountKey = workflowDetailCellValidationKey('items', 'row-a', 'amount')
if (detailed.detailCellErrors[nameKey] !== '项目不能为空')
  throw new Error('明细必填列应生成稳定行 ID 对应的单元格错误')
if (detailed.detailCellErrors[amountKey] !== '金额不能小于0')
  throw new Error('明细数字列应独立校验范围')
if (!String(detailed.errors.items || '').includes('费用明细第1行：项目不能为空'))
  throw new Error('明细表提交错误应保留首个行列汇总')

const runtimeFormSource = fs.readFileSync(path.join(root, 'src/pages/workflow/components/WorkflowRuntimeForm.vue'), 'utf8')
const fieldControlSource = fs.readFileSync(path.join(root, 'src/pages/workflow/components/WorkflowFieldControl.vue'), 'utf8')
const textareaSource = fs.readFileSync(path.join(root, 'src/pages/workflow/components/WorkflowTextarea.vue'), 'utf8')
for (const pattern of [
  'validateWorkflowFormDataDetailed(props.fields, props.modelValue, accessMap.value)',
  'touchedFields',
  'touchedDetailCells',
  'workflowDetailCellValidationKey',
  '@blur="touchField(field, \'blur\')"',
  '@change="touchField(field, \'change\')"',
  'detailCellError(field, row, rowIndex, column)',
  'workflow-detail__column--error',
]) {
  if (!runtimeFormSource.includes(pattern))
    throw new Error(`WorkflowRuntimeForm missing validation interaction contract: ${pattern}`)
}
for (const pattern of ['\'blur\': []', '\'change\': []', '@blur="emit(\'blur\')"', '@change="emit(\'change\')"']) {
  if (!fieldControlSource.includes(pattern))
    throw new Error(`WorkflowFieldControl missing validation event contract: ${pattern}`)
}
if (!textareaSource.includes('\'blur\': []') || !textareaSource.includes('emit(\'blur\')'))
  throw new Error('WorkflowTextarea must report blur after IME-safe value commit')

console.log('workflow form validation checks passed')
