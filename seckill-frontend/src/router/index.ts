import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore, type UserRole } from '../store/user'
import { toast } from '../utils/toast'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    requiresRole?: UserRole
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'mall',
    component: () => import('../views/Mall.vue'),
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/Login.vue'),
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('../views/Profile.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/merchant',
    name: 'merchant',
    component: () => import('../views/Merchant.vue'),
    meta: { requiresAuth: true, requiresRole: 1 },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.token) {
    toast('请先完成身份接入', 'warning')
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresRole !== undefined && userStore.role !== to.meta.requiresRole) {
    toast('商家控制台仅允许商家身份访问', 'error')
    return { name: 'mall' }
  }

  return true
})

export default router
