import { createRouter, createWebHistory } from 'vue-router'
import { adminChildRoutes } from './adminRoutes'

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
      children: adminChildRoutes
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
