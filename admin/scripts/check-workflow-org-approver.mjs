import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const read = relativePath => fs.readFileSync(path.join(root, relativePath), 'utf8')

const apiTypesSource = read('src/api/types.ts')
const apiSource = read('src/api/index.ts')
const userPageSource = read('src/views/user/index.vue')
const workflowTypesSource = read('src/types/workflow.ts')
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
  [orgApproverPageSource, 'org-panel--departments', '组织审批身份设置页缺少部门区域布局标识'],
  [orgApproverPageSource, 'org-panel--identities', '组织审批身份设置页缺少身份区域布局标识'],
  [orgApproverPageSource, 'panel-title__count', '组织审批身份设置页缺少区域数量反馈'],
  [orgApproverPageSource, 'candidate-users__header', '组织审批身份设置页缺少候选处理人标题'],
  [orgApproverPageSource, 'activeDepartmentPath', '组织审批身份设置页缺少当前部门路径反馈'],
  [orgApproverPageSource, 'activeSubjectType', '组织审批身份设置页缺少适用对象类型'],
  [orgApproverPageSource, '部门内员工', '组织审批身份设置页缺少部门员工范围'],
  [orgApproverPageSource, '指定员工', '组织审批身份设置页缺少指定员工范围'],
  [orgApproverPageSource, '由谁担任该身份', '组织审批身份设置页缺少身份担任人步骤'],
  [orgApproverPageSource, '优先于部门规则', '组织审批身份设置页缺少指定员工规则优先级提示'],
  [orgApproverPageSource, 'assignmentSummary', '组织审批身份设置页缺少人员身份关系摘要'],
  [orgApproverPageSource, 'subjectType', '组织审批身份设置页未提交适用对象类型'],
  [orgApproverPageSource, 'subjectId', '组织审批身份设置页未提交适用对象 ID'],
  [orgApproverPageSource, 'defaultExpandedUserTreeKeys', '组织审批身份设置页缺少当前部门树展开状态'],
  [orgApproverPageSource, 'empty-text="尚未指定担任人员"', '组织审批身份设置页缺少身份担任人空状态'],
  [workflowTypesSource, "'org_identity'", '流程审批人类型缺少组织审批身份'],
  [workflowTypesSource, 'WorkflowDepartmentApprovalChain', '流程类型缺少分层审批链配置'],
  [workflowDesignerSource, 'workflowOrgApproverIdentities', '流程设计器未加载组织审批身份'],
  [workflowDesignerSource, 'workflowAssigneeUsers', '流程设计器未加载审批用户树数据'],
  [nodeInspectorSource, 'assignee-user-tree', '审批节点缺少用户树选择'],
  [nodeInspectorSource, 'orgIdentityOptions', '审批节点缺少组织审批身份选项'],
  [nodeInspectorSource, 'starter_department:', '组织审批身份缺少发起人部门作用域编码'],
  [nodeInspectorSource, '逐级向上审批', '组织审批身份缺少逐级审批开关'],
  [nodeInspectorSource, 'departmentApprovalChainStopMode', '逐级审批缺少终止范围配置'],
  [nodeInspectorSource, 'missingAssigneePolicy', '逐级审批缺少负责人缺失策略'],
  [nodeInspectorSource, "code: 'supervisor', name: '主管'", '审批身份默认选项缺少主管'],
  [nodeInspectorSource, '审批方式应用于同一部门', '组织身份缺少同层审批方式说明'],
  [nodeCardSource, '组织审批身份', '流程节点卡片缺少组织审批身份描述'],
]

for (const [source, snippet, message] of requirements) {
  if (!source.includes(snippet)) throw new Error(message)
}

const chainSwitchIndex = nodeInspectorSource.indexOf('逐级向上审批')
const assigneeSourceIndex = nodeInspectorSource.indexOf('处理人来源')
if (chainSwitchIndex < 0 || assigneeSourceIndex < 0 || chainSwitchIndex > assigneeSourceIndex) {
  throw new Error('逐级审批开关必须在审批节点中直接可见，不能隐藏在处理人来源配置内')
}

console.log('workflow organization approver structure checks passed')
