import { readFileSync } from 'node:fs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')

const api = read('src/api/index.ts')
for (const snippet of [
  'workflowDefinitionVersionChanges',
  'workflowDefinitionVersionDelete',
  'workflowDefinitionVersionRollback',
]) {
  if (!api.includes(snippet)) throw new Error(`workflow version API missing ${snippet}`)
}

const component = read('src/views/workflow/components/WorkflowVersionDrawer.vue')
for (const snippet of [
  'append-to-body',
  'changeHeadline',
  'deleteBlockedReason',
  '查看变更',
  '回滚',
  '删除版本',
]) {
  if (!component.includes(snippet)) throw new Error(`workflow version drawer missing ${snippet}`)
}

for (const path of ['src/views/workflow/index.vue', 'src/views/workflow/designer/index.vue']) {
  const source = read(path)
  if (!source.includes('WorkflowVersionDrawer')) {
    throw new Error(`${path} must reuse WorkflowVersionDrawer`)
  }
}
