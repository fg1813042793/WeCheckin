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
  { path: 'role', name: 'Role', component: () => import('../views/role/index.vue'), meta: { title: '角色管理' } },
  { path: 'menu', name: 'Menu', component: () => import('../views/menu/index.vue'), meta: { title: '权限管理' } },
  { path: 'event', name: 'Event', component: () => import('../views/event/index.vue'), meta: { title: '赛事活动' } },
  { path: 'setup', name: 'Setup', component: () => import('../views/setup/index.vue'), meta: { title: '系统配置' } },
  { path: 'client-setup', name: 'ClientSetup', component: () => import('../views/client-setup/index.vue'), meta: { title: '系统配置' } }
]

export const fallbackMenuItems: AdminMenuItem[] = [
  { id: 'dashboard', name: '控制台', type: 1, status: 1, path: '/dashboard', icon: 'Odometer' },
  { id: 'user', name: '用户管理', type: 1, status: 1, path: '/user', icon: 'User' },
  { id: 'online', name: '在线用户', type: 1, status: 1, path: '/online', icon: 'Monitor' },
  { id: 'enroll', name: '打卡管理', type: 1, status: 1, path: '/enroll', icon: 'List' },
  { id: 'news', name: '内容管理', type: 1, status: 1, path: '/news', icon: 'Document' },
  { id: 'mgr', name: '管理员管理', type: 1, status: 1, path: '/mgr', icon: 'Setting' },
  { id: 'log', name: '操作日志', type: 1, status: 1, path: '/log', icon: 'Clock' },
  { id: 'dict', name: '字典管理', type: 1, status: 1, path: '/dict', icon: 'Notebook' },
  { id: 'department', name: '部门管理', type: 1, status: 1, path: '/department', icon: 'FolderOpened' },
  { id: 'role', name: '角色管理', type: 1, status: 1, path: '/role', icon: 'UserFilled' },
  { id: 'menu', name: '权限管理', type: 1, status: 1, path: '/menu', icon: 'Key' },
  { id: 'setup', name: '系统配置', type: 1, status: 1, path: '/setup', icon: 'Setting' },
  { id: 'event', name: '赛事活动', type: 1, status: 1, path: '/event', icon: 'TrophyBase' },
  {
    id: 'question-exam',
    name: '问卷考试',
    type: 0,
    status: 1,
    path: '/question-exam',
    icon: 'Collection',
    children: [
      { id: 'question-bank', name: '题库管理', type: 1, status: 1, path: '/question-bank', icon: 'Collection' }
    ]
  },
  {
    id: 'survey',
    name: '问卷调查',
    type: 0,
    status: 1,
    path: '/survey',
    icon: 'List',
    children: [
      { id: 'survey-list', name: '问卷管理', type: 1, status: 1, path: '/survey' },
      { id: 'survey-responses', name: '答卷管理', type: 1, status: 1, path: '/survey/responses' },
      { id: 'survey-statistic', name: '问卷统计', type: 1, status: 1, path: '/survey/statistic' },
      { id: 'survey-stat-report', name: '统计报表', type: 1, status: 1, path: '/survey/stat-report' },
      { id: 'survey-notify', name: '站内通知', type: 1, status: 1, path: '/survey/notify' }
    ]
  },
  {
    id: 'exam',
    name: '在线考试',
    type: 0,
    status: 1,
    path: '/exam',
    icon: 'EditPen',
    children: [
      { id: 'exam-list', name: '考试管理', type: 1, status: 1, path: '/exam/list' },
      { id: 'exam-responses', name: '考试记录', type: 1, status: 1, path: '/exam/responses' },
      { id: 'exam-statistic', name: '考试统计', type: 1, status: 1, path: '/exam/statistic' },
      { id: 'exam-formkit', name: '考试表单设计器', type: 1, status: 1, path: '/exam/formkit' },
      { id: 'exam-formkit-report', name: '考试答题报表', type: 1, status: 1, path: '/exam/formkit/report' }
    ]
  }
]
