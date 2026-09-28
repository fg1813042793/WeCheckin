import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')

const sourceURL = new URL('../src/pages/workflow/workflow-summary-columns.ts', import.meta.url)
const source = readFileSync(sourceURL, 'utf8')
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2022,
    esModuleInterop: true,
  },
  fileName: sourceURL.pathname,
}).outputText
const module = { exports: {} }
const execute = runInThisContext(`(function (module, exports, require) {${compiled}\n})`)
execute(module, module.exports, request => request === './workflow-form'
  ? { flattenWorkflowOptions: options => options || [], workflowDataFields: fields => fields || [] }
  : require(request))

const definition = {
  id: 7,
  key: 'review',
  name: '绩效评语单',
  form: [
    {
      key: 'monthly',
      label: '本月目标',
      type: 'group',
      fields: [
        { key: 'goal', label: '目标', type: 'text' },
        {
          key: 'scores',
          label: '分数',
          type: 'group',
          fields: [{ key: 'result', label: '结果', type: 'textarea' }],
        },
      ],
    },
  ],
}
const groups = module.exports.workflowSummaryFormFieldGroups([definition])
assert.deepEqual(
  groups[0].columns.map(column => column.label),
  ['本月目标.目标', '本月目标.分数.结果'],
)
assert.deepEqual(
  groups[0].columns.map(column => column.key),
  ['form:7:goal', 'form:7:result'],
)

console.log('Workflow summary column label checks passed')
