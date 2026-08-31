import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const read = relativePath => fs.readFileSync(path.join(root, relativePath), 'utf8')

const apiTypesSource = read('src/api/types.ts')
const apiSource = read('src/api/index.ts')
const userPageSource = read('src/views/user/index.vue')
const workflowTypesSource = read('src/views/workflow/types.ts')
const workflowDesignerSource = read('src/views/workflow/designer/index.vue')
const nodeInspectorSource = read('src/views/workflow/designer/components/NodeInspector.vue')
const nodeCardSource = read('src/views/workflow/designer/components/WorkflowNodeCard.vue')
const routesSource = read('src/router/adminRoutes.ts')
const orgApproverPageSource = read('src/views/workflow/org-approvers/index.vue')

const requirements = [
  [apiTypesSource, 'managerUserId', '用户类型缺少直属上级 ID'],
  [apiTypesSource, 'managerUserName', '用户类型缺少直属上级名称'],
  [apiSource, 'workflowOrgApproverIdentities', 'API 缺少组织审批身份列表接口'],
  [apiSource, 'workflowOrgApproverAssignments', 'API 缺少组织审批身份配置接口'],
  [apiSource, 'workflowOrgApproverAssignmentsSave', 'API 缺少组织审批身份保存接口'],
  [userPageSource, '直属上级', '用户页缺少直属上级展示或表单项'],
  [userPageSource, 'managerUserTreeData', '用户页缺少直属上级用户树'],
  [userPageSource, 'managerUserId', '用户页未提交直属上级 ID'],
  [routesSource, 'workflow/org-approvers', '后台路由缺少组织审批身份设置入口'],
  [routesSource, 'WorkflowOrgApprovers', '后台路由缺少组织审批身份设置页面名称'],
  [orgApproverPageSource, '组织审批身份设置', '组织审批身份设置页缺少标题'],
  [orgApproverPageSource, 'workflowOrgApproverIdentities', '组织审批身份设置页缺少身份列表加载'],
  [orgApproverPageSource, 'workflowOrgApproverAssignments', '组织审批身份设置页缺少配置读取'],
  [orgApproverPageSource, 'workflowOrgApproverAssignmentsSave', '组织审批身份设置页缺少配置保存'],
  [orgApproverPageSource, 'orgApproverUserTreeData', '组织审批身份设置页缺少用户树数据'],
  [orgApproverPageSource, '@check="onOrgApproverUserCheck"', '组织审批身份设置页缺少用户树选择事件'],
  [workflowTypesSource, "'org_identity'", '流程审批人类型缺少组织审批身份'],
  [workflowDesignerSource, 'workflowOrgApproverIdentities', '流程设计器未加载组织审批身份'],
  [workflowDesignerSource, 'workflowAssigneeUsers', '流程设计器未加载审批用户树数据'],
  [nodeInspectorSource, 'assignee-user-tree', '审批节点缺少用户树选择'],
  [nodeInspectorSource, 'orgIdentityOptions', '审批节点缺少组织审批身份选项'],
  [nodeInspectorSource, 'starter_department:', '组织审批身份缺少发起人部门作用域编码'],
  [nodeInspectorSource, '任一人审批', '单人审批缺少多人组织身份说明'],
  [nodeCardSource, '组织审批身份', '流程节点卡片缺少组织审批身份描述'],
]

for (const [source, snippet, message] of requirements) {
  if (!source.includes(snippet)) throw new Error(message)
}

console.log('workflow organization approver structure checks passed')
