import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import AdminLogin from './views/AdminLogin.vue'
import AdminLayout from './views/AdminLayout.vue'
import AdminVendors from './views/admin/Vendors.vue'
import AdminModels from './views/admin/Models.vue'
import AdminLinks from './views/admin/Links.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Home },
    { path: '/admin/login', component: AdminLogin },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { auth: true },
      children: [
        { path: '', redirect: '/admin/vendors' },
        { path: 'vendors', component: AdminVendors },
        { path: 'models',  component: AdminModels },
        { path: 'links',   component: AdminLinks },
      ]
    }
  ]
})

router.beforeEach((to) => {
  if (to.meta?.auth) {
    const tok = localStorage.getItem('mm_token')
    if (!tok) return '/admin/login'
  }
})

export default router
