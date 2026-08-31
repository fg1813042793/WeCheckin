import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const read = relativePath => fs.readFileSync(path.join(root, relativePath), 'utf8')

const typeSource = read('src/views/workflow/types.ts')
const designerSource = read('src/views/workflow/designer/index.vue')
const formDesignerSource = read('src/views/workflow/designer/components/WorkflowFormDesigner.vue')
const permissionSource = read('src/views/workflow/designer/components/WorkflowFieldPermissions.vue')

const requirements = [
  [typeSource, "export type WorkflowFormFieldType", '缺少流程表单字段类型'],
  [typeSource, 'form: WorkflowFormField[]', '流程草稿缺少 form 字段'],
  [typeSource, 'formPermissions?: WorkflowFieldPermission[]', '流程节点缺少字段权限'],
  [designerSource, 'WorkflowFormDesigner', '流程设计器未接入表单设计'],
  [designerSource, 'WorkflowFieldPermissions', '流程设计器未接入字段权限'],
  [designerSource, '表单设计', '流程设计器缺少表单设计页签'],
  [designerSource, '流程设计', '流程设计器缺少流程设计页签'],
  [designerSource, '字段权限', '流程设计器缺少字段权限页签'],
  [formDesignerSource, "type: 'textarea'", '表单设计器缺少多行文本字段'],
  [formDesignerSource, "type: 'user'", '表单设计器缺少人员字段'],
  [formDesignerSource, "type: 'department'", '表单设计器缺少部门字段'],
  [formDesignerSource, "type: 'attachment'", '表单设计器缺少附件字段'],
  [permissionSource, "value=\"hidden\"", '字段权限缺少隐藏选项'],
  [permissionSource, "value=\"read\"", '字段权限缺少只读选项'],
  [permissionSource, "value=\"write\"", '字段权限缺少可编辑选项'],
]

for (const [source, snippet, message] of requirements) {
  if (!source.includes(snippet)) throw new Error(message)
}

console.log('workflow form designer structure checks passed')
