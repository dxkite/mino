import { createRouter, createWebHashHistory } from 'vue-router'
import Home from '../pages/Home.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/error',
    name: 'Error',
    component: () => import(/* webpackChunkName: "error" */ '../pages/Error.vue')
  }
]

const router = createRouter({
  history: createWebHashHistory(process.env.BASE_URL),
  routes
})

export default router
