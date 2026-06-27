import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/login/index.vue')
    },
    {
      path: '/',
      component: () => import('../views/layout/index.vue'),
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('../views/dashboard/index.vue'), meta: { title: '控制台' } },
        { path: 'user', name: 'User', component: () => import('../views/user/index.vue'), meta: { title: '用户管理' } },
        { path: 'online', name: 'Online', component: () => import('../views/online/OnlineUsers.vue'), meta: { title: '在线用户' } },
        { path: 'enroll', name: 'Enroll', component: () => import('../views/enroll/index.vue'), meta: { title: '打卡管理' } },
        { path: 'survey', name: 'SurveyList', component: () => import('../views/survey/SurveyList.vue'), meta: { title: '问卷管理' } },
        { path: 'survey/designer', name: 'SurveyDesigner', component: () => import('../views/survey/SurveyDesigner.vue'), meta: { title: '问卷设计器' } },
        { path: 'survey/responses', name: 'SurveyResponses', component: () => import('../views/survey/SurveyResponses.vue'), meta: { title: '答卷管理' } },
        { path: 'survey/statistic', name: 'SurveyStatistic', component: () => import('../views/survey/SurveyStatistic.vue'), meta: { title: '问卷统计' } },
        { path: 'survey/formkit', name: 'Formkit', component: () => import('../views/survey/formkit/FormDesigner.vue'), meta: { title: '表单设计器' } },
        { path: 'survey/formkit/report', name: 'FormkitReport', component: () => import('../views/survey/formkit/FormReport.vue'), meta: { title: '答题报表' } },
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
        { path: 'menu', name: 'Menu', component: () => import('../views/menu/index.vue'), meta: { title: '菜单权限' } },
        { path: 'event', name: 'Event', component: () => import('../views/event/index.vue'), meta: { title: '赛事活动' } },
        { path: 'setup', name: 'Setup', component: () => import('../views/setup/index.vue'), meta: { title: '系统配置' } }
      ]
    },
    {
      path: '/sf/:id',
      name: 'SurveyFillPC',
      component: () => import('../views/survey/SurveyFillPC.vue')
    },
    {
      path: '/sf1/:id',
      name: 'SurveyFillPC1',
      meta: { title: '问卷填写' },
      component: () => import('../views/survey/SurveyFillPC1.vue')
    },
    {
      path: '/ef/:id',
      name: 'ExamFillPC',
      component: () => import('../views/exam/ExamFillPC.vue')
    },
    {
      path: '/ef1/:id',
      name: 'ExamFillPC1',
      meta: { title: '考试填写' },
      component: () => import('../views/exam/ExamFillPC1.vue')
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('admin_token')
  if (to.path !== '/login' && !to.path.startsWith('/sf/') && !to.path.startsWith('/sf1/') && !to.path.startsWith('/ef/') && !to.path.startsWith('/ef1/') && !token) {
    next('/login')
  } else {
    next()
  }
})

router.afterEach((to) => {
  if (to.meta?.title) document.title = to.meta.title as string
})

export default router
