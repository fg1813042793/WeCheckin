import type { RouteRecordRaw } from 'vue-router'

export interface AdminMenuItem {
  id: string
  name: string
  type: 0 | 1 | 2
  status: number
  path?: string
  icon?: string
  children?: AdminMenuItem[]
}

export const adminChildRoutes: RouteRecordRaw[] = [
  { path: 'dashboard', name: 'Dashboard', component: () => import('../views/dashboard/index.vue'), meta: { title: '控制台' } },
  { path: 'user', name: 'User', component: () => import('../views/user/index.vue'), meta: { title: '用户管理' } },
  { path: 'online', name: 'Online', component: () => import('../views/online/OnlineUsers.vue'), meta: { title: '在线用户' } },
  { path: 'enroll', name: 'Enroll', component: () => import('../views/enroll/index.vue'), meta: { title: '打卡管理' } },
  { path: 'survey', name: 'SurveyList', component: () => import('../views/survey/SurveyList.vue'), meta: { title: '问卷管理' } },
  { path: 'survey/designer', name: 'SurveyDesigner', component: () => import('../views/survey/SurveyDesigner.vue'), meta: { title: '问卷设计器' } },
  { path: 'survey/responses', name: 'SurveyResponses', component: () => import('../views/survey/SurveyResponses.vue'), meta: { title: '答卷管理' } },
  { path: 'survey/statistic', name: 'SurveyStatistic', component: () => import('../views/survey/SurveyStatistic.vue'), meta: { title: '问卷统计' } },
  { path: 'survey/stat-report', name: 'SurveyStatReport', component: () => import('../views/survey/SurveyStatReport.vue'), meta: { title: '统计报表' } },
  { path: 'survey/notify', name: 'SurveyNotify', component: () => import('../views/survey/SurveyNotify.vue'), meta: { title: '站内通知' } },
  { path: 'survey/formkit', name: 'Formkit', component: () => import('../views/survey/formkit/FormDesigner.vue'), meta: { title: '表单设计器' } },
  { path: 'survey/formkit/report', name: 'FormkitReport', component: () => import('../views/survey/formkit/FormReport.vue'), meta: { title: '答题报表' } },
  { path: 'question-bank', name: 'QuestionBank', component: () => import('../views/question-bank/QuestionBank.vue'), meta: { title: '题库管理' } },
  { path: 'exam/list', name: 'ExamList', component: () => import('../views/exam/ExamList.vue'), meta: { title: '考试管理' } },
  { path: 'exam/designer', name: 'ExamDesigner', component: () => import('../views/exam/ExamDesigner.vue'), meta: { title: '考试设计器' } },
  { path: 'exam/responses', name: 'ExamResponses', component: () => import('../views/exam/ExamResponses.vue'), meta: { title: '考试记录' } },
  { path: 'exam/statistic', name: 'ExamStatistic', component: () => import('../views/exam/ExamStatistic.vue'), meta: { title: '考试统计' } },
  { path: 'exam/formkit', name: 'ExamFormkit', component: () => import('../views/exam/formkit/FormDesigner.vue'), meta: { title: '考试表单设计器' } },
  { path: 'exam/formkit/report', name: 'ExamFormkitReport', component: () => import('../views/exam/formkit/FormReport.vue'), meta: { title: '考试答题报表' } },
  { path: 'news', name: 'News', component: () => import('../views/news/index.vue'), meta: { title: '内容管理' } },
  { path: 'mgr', name: 'Mgr', component: () => import('../views/mgr/index.vue'), meta: { title: '管理员管理' } },
  { path: 'log', name: 'Log', component: () => import('../views/log/index.vue'), meta: { title: '操作日志' } },
  { path: 'dict', name: 'Dict', component: () => import('../views/dict/index.vue'), meta: { title: '字典管理' } },
  { path: 'department', name: 'Department', component: () => import('../views/department/index.vue'), meta: { title: '部门管理' } },
  { path: 'position', name: 'Position', component: () => import('../views/position/index.vue'), meta: { title: '岗位管理' } },
  { path: 'workflow/definitions', name: 'WorkflowDefinitions', component: () => import('../views/workflow/index.vue'), meta: { title: '流程定义' } },
  { path: 'workflow/definitions/:id/designer', name: 'WorkflowDesigner', component: () => import('../views/workflow/designer/index.vue'), meta: { title: '流程设计器' } },
  { path: 'workflow/instances', name: 'WorkflowInstances', component: () => import('../views/workflow/instances/index.vue'), meta: { title: '流程实例' } },
  { path: 'workflow/tasks', name: 'WorkflowTasks', component: () => import('../views/workflow/tasks/index.vue'), meta: { title: '流程任务' } },
  { path: 'workflow/org-approvers', name: 'WorkflowOrgApprovers', component: () => import('../views/workflow/org-approvers/index.vue'), meta: { title: '组织审批身份设置' } },
  { path: 'scheduled-task/tasks', name: 'ScheduledTaskTasks', component: () => import('../views/scheduled-task/tasks/index.vue'), meta: { title: '定时任务' } },
  { path: 'scheduled-task/runs', name: 'ScheduledTaskRuns', component: () => import('../views/scheduled-task/runs/index.vue'), meta: { title: '运行记录' } },
  { path: 'scheduled-task/workers', name: 'ScheduledTaskWorkers', component: () => import('../views/scheduled-task/workers/index.vue'), meta: { title: '执行节点' } },
  { path: 'role', name: 'Role', component: () => import('../views/role/index.vue'), meta: { title: '角色管理' } },
  { path: 'menu', name: 'Menu', component: () => import('../views/menu/index.vue'), meta: { title: '权限管理' } },
  { path: 'event', name: 'Event', component: () => import('../views/event/index.vue'), meta: { title: '赛事活动' } },
  { path: 'dingtalk/config', name: 'DingTalkConfig', component: () => import('../views/dingtalk-setup/index.vue'), meta: { title: '钉钉配置' } },
  { path: 'dingtalk/bindings', name: 'DingTalkBindings', component: () => import('../views/dingtalk-bindings/index.vue'), meta: { title: '钉钉用户绑定' } },
  { path: 'dingtalk/perf-reviews', name: 'DingTalkPerfReviews', component: () => import('../views/dingtalk-perf-reviews/index.vue'), meta: { title: '绩效考评单' } },
  { path: 'dingtalk/perf-histories', name: 'DingTalkPerfHistories', component: () => import('../views/dingtalk-perf-histories/index.vue'), meta: { title: '绩效流转记录' } },
  { path: 'setup', name: 'Setup', component: () => import('../views/setup/index.vue'), meta: { title: '系统配置' } },
  { path: 'client-setup', name: 'ClientSetup', component: () => import('../views/client-setup/index.vue'), meta: { title: '系统配置' } }
]
