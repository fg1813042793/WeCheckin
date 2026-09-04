import { createRouter, createWebHistory } from 'vue-router'
import { adminChildRoutes } from './adminRoutes'
import { canAccessAdminRoute, loadAdminAccessSnapshot } from './adminAccess'

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
      children: adminChildRoutes
    },
    {
      path: '/sf/:id',
      name: 'SurveyFillPC',
      component: () => import('../views/survey/SurveyFillPC.vue'),
      meta: { public: true }
    },
    {
      path: '/sf1/:id',
      name: 'SurveyFillPC1',
      meta: { title: '问卷填写', public: true },
      component: () => import('../views/survey/SurveyFillPC1.vue')
    },
    {
      path: '/ef/:id',
      name: 'ExamFillPC',
      component: () => import('../views/exam/ExamFillPC.vue'),
      meta: { public: true }
    },
    {
      path: '/ef1/:id',
      name: 'ExamFillPC1',
      meta: { title: '考试填写', public: true },
      component: () => import('../views/exam/ExamFillPC1.vue')
    }
  ]
})

router.beforeEach(async (to) => {
  const token = localStorage.getItem('admin_token')
  if (to.path === '/login' || to.meta.public) return true
  if (!token) return { path: '/login', query: { redirect: to.fullPath } }
  if (to.meta.allowWithoutMenu) return true
  try {
    const snapshot = await loadAdminAccessSnapshot()
    if (to.path === '/') {
      const firstRoute = adminChildRoutes.find(route => {
        return typeof route.path === 'string'
          && !route.path.includes(':')
          && canAccessAdminRoute(route.meta || {}, snapshot)
      })
      return firstRoute ? { path: `/${firstRoute.path}` } : { path: '/forbidden' }
    }
    if (canAccessAdminRoute(to.meta, snapshot)) return true
    return { path: '/forbidden', query: { from: to.fullPath } }
  } catch {
    return { path: '/forbidden', query: { reason: 'load', from: to.fullPath } }
  }
})

router.afterEach((to) => {
  if (to.meta?.title) document.title = to.meta.title as string
})

export default router
