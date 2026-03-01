import { createRouter, createWebHistory } from 'vue-router'
import { handleHotUpdate, routes } from 'vue-router/auto-routes'

import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

import type { EnhancedRouteLocation } from './types'
import { useRouteCacheStore, useUserStore } from '@/stores'

import { isLogin } from '@/utils/auth'
import setPageTitle from '@/utils/set-page-title'

NProgress.configure({ showSpinner: true, parent: '#app' })

const publicRouteNames = new Set(['Home', 'Login', 'Register'])
const publicRoutePaths = new Set(['/', '/login', '/register'])

const router = createRouter({
  history: createWebHistory(import.meta.env.VITE_APP_PUBLIC_PATH),
  routes,
  scrollBehavior() {
    return {
      left: 0,
      top: 0,
    }
  },
})

// This will update routes at runtime without reloading the page
if (import.meta.hot)
  handleHotUpdate(router)

router.beforeEach(async (to: EnhancedRouteLocation) => {
  NProgress.start()

  const routeCacheStore = useRouteCacheStore()
  const userStore = useUserStore()

  // Route cache
  routeCacheStore.addRoute(to)

  // Set page title
  setPageTitle(to.name)

  const routeName = String(to.name || '')
  const isPublicRoute = publicRouteNames.has(routeName) || publicRoutePaths.has(to.path)
  if (!isLogin() && !isPublicRoute) {
    return {
      name: 'Login',
      query: { redirect: to.fullPath },
      replace: true,
    }
  }

  if (isLogin() && !userStore.userInfo?.uid)
    await userStore.info()
})

router.afterEach(() => {
  NProgress.done()
})

export default router
