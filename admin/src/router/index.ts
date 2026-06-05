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
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('admin_token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
