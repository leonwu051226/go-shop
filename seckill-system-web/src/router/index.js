import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import SecKill from '../views/SecKill.vue'

const routes = [
  { path: '/', redirect: '/seckill' },
  { path: '/login', component: Login, meta: { public: true } },
  { path: '/seckill', component: SecKill }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (!to.meta.public && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
