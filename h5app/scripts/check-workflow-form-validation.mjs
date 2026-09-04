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
// eslint-disable-next-line no-new-func
new Function('module', 'exports', 'require', output)(module, module.exports, require)

const { validateWorkflowFormData } = module.exports
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

console.log('workflow form validation checks passed')
