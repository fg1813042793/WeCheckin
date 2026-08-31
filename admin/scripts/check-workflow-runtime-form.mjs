import fs from 'node:fs'
import path from 'node:path'
import { createRequire } from 'node:module'
import ts from 'typescript'

const root = process.cwd()
const require = createRequire(import.meta.url)

function loadTypeScriptModule(relativePath) {
  const filename = path.join(root, relativePath)
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
  const runner = new Function('module', 'exports', 'require', output)
  runner(module, module.exports, require)
  return module.exports
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

const {
  initialWorkflowFormData,
  visibleWorkflowFormFields,
  workflowFieldAccessMap,
  writableWorkflowFormData,
} = loadTypeScriptModule('src/views/workflow/runtimeForm.ts')

const fields = [
  { key: 'reason', label: '申请原因', type: 'textarea', required: true, default: '出差' },
  { key: 'amount', label: '报销金额', type: 'amount', default: 100 },
  { key: 'attachments', label: '附件', type: 'attachment' },
  { key: 'internalNote', label: '内部备注', type: 'text' },
]
const permissions = {
  start: [
    { field: 'internalNote', access: 'hidden' },
    { field: 'amount', access: 'write' },
  ],
  manager: [
    { field: 'reason', access: 'read' },
    { field: 'amount', access: 'write' },
    { field: 'attachments', access: 'write' },
    { field: 'internalNote', access: 'hidden' },
  ],
}

const initial = initialWorkflowFormData(fields, { reason: '客户拜访' })
assert(initial.reason === '客户拜访', 'existing form data should win over field defaults')
assert(initial.amount === 100, 'field defaults should seed missing form values')
assert(Array.isArray(initial.attachments), 'attachment fields should default to an empty array')

const startAccess = workflowFieldAccessMap(fields, permissions, 'start', 'write')
assert(startAccess.reason === 'write', 'start node should default to writable fields')
assert(startAccess.internalNote === 'hidden', 'explicit hidden permission should be preserved')
assert(visibleWorkflowFormFields(fields, startAccess).length === 3, 'hidden fields should not render')

const managerAccess = workflowFieldAccessMap(fields, permissions, 'manager', 'read')
const payload = writableWorkflowFormData(fields, {
  reason: '篡改原因',
  amount: 120,
  attachments: ['receipt.pdf'],
  internalNote: 'secret',
}, managerAccess)
assert(!('reason' in payload), 'read-only fields must not be submitted as a task form patch')
assert(payload.amount === 120, 'writable scalar field should be submitted')
assert(Array.isArray(payload.attachments) && payload.attachments[0] === 'receipt.pdf', 'writable array field should be submitted')
assert(!('internalNote' in payload), 'hidden fields must not be submitted')

console.log('workflow runtime form checks passed')
