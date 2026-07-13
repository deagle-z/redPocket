# Vue3 + Vant + pnpm H5 项目初始化分步文档

本文档用于从零初始化一个移动端 H5 前端项目，技术栈包含 Vue 3、Vite、TypeScript、pnpm、Vant、File Based Routing、组件自动导入、Composition API 自动导入、Pinia、PWA、Tailwind CSS、i18n、KeepAlive、axios 通用请求封装、带时效的 localStorage。

示例项目名使用 `RedPocketH5`，如果你要创建其他项目，替换命令里的目录名即可。

## 1. 环境准备

建议环境：

```bash
node -v
corepack enable
corepack prepare pnpm@latest --activate
pnpm -v
```

说明：

- Node.js 建议使用当前 LTS 版本。
- 使用 `corepack` 固定团队 pnpm 行为，减少本机全局 pnpm 版本差异。
- 如果公司内网或服务器需要代理，可配置：

```bash
pnpm config set registry https://registry.npmmirror.com
```

## 2. 创建 Vue3 + TypeScript 项目

```bash
pnpm create vite RedPocketH5 --template vue-ts
cd RedPocketH5
pnpm install
```

创建后先确认基础项目能跑：

```bash
pnpm dev
```

## 3. 安装业务依赖

```bash
pnpm add vue-router pinia vant axios vue-i18n @vueuse/core workbox-window
```

## 4. 安装构建与自动导入依赖

```bash
pnpm add -D @vant/auto-import-resolver unplugin-vue-components unplugin-auto-import vite-plugin-pwa tailwindcss @tailwindcss/vite
```

依赖用途：

- `vue-router`：路由与官方 file-based routing 插件。
- `pinia`：状态管理。
- `vant`：移动端组件库。
- `axios`：HTTP 请求。
- `vue-i18n`：国际化。
- `@vueuse/core`：常用 Composition API 工具。
- `workbox-window`：`vite-plugin-pwa` 的客户端注册逻辑依赖。
- `unplugin-auto-import`：自动导入 `ref`、`computed`、`useRouter`、`defineStore` 等 API。
- `unplugin-vue-components`：自动导入本地组件和 Vant 组件。
- `@vant/auto-import-resolver`：Vant 官方自动导入解析器，并按需导入样式。
- `vite-plugin-pwa`：生成 manifest、service worker 和 PWA 注册逻辑。
- `tailwindcss`、`@tailwindcss/vite`：Tailwind CSS v4 Vite 集成。

## 5. 推荐目录结构

```text
RedPocketH5/
  public/
    pwa-192x192.png
    pwa-512x512.png
  src/
    api/
      user.ts
    assets/
      styles/
        main.css
    components/
      AppTabbar.vue
    composables/
      useTtlLocalStorage.ts
    locales/
      en-US.ts
      zh-CN.ts
      index.ts
    pages/
      index.vue
      login.vue
      mine/
        index.vue
      red-packet/
        [id].vue
    plugins/
      i18n.ts
    router/
      index.ts
    stores/
      app.ts
      user.ts
    types/
      auto-imports.d.ts
      components.d.ts
      vite-plugin-pwa.d.ts
    utils/
      http.ts
      storage.ts
    App.vue
    main.ts
    pwa.ts
  .env.development
  .env.production
  index.html
  package.json
  tsconfig.app.json
  vite.config.ts
```

## 6. 配置 Vite

修改 `vite.config.ts`：

```ts
import path from 'node:path'
import { defineConfig } from 'vite'
import Vue from '@vitejs/plugin-vue'
import VueRouter from 'vue-router/vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { VantResolver } from '@vant/auto-import-resolver'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    VueRouter({
      routesFolder: 'src/pages',
      dts: 'src/types/typed-router.d.ts',
    }),
    Vue(),
    tailwindcss(),
    AutoImport({
      imports: [
        'vue',
        'vue-router',
        'pinia',
        '@vueuse/core',
        {
          'vue-i18n': ['useI18n'],
        },
      ],
      dirs: ['src/composables', 'src/stores'],
      vueTemplate: true,
      resolvers: [VantResolver()],
      dts: 'src/types/auto-imports.d.ts',
    }),
    Components({
      dirs: ['src/components'],
      resolvers: [VantResolver()],
      dts: 'src/types/components.d.ts',
      types: [
        {
          from: 'vue-router',
          names: ['RouterLink', 'RouterView'],
        },
      ],
    }),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico'],
      manifest: {
        name: 'RedPocketH5',
        short_name: 'RedPocket',
        description: 'Mobile H5 app',
        theme_color: '#dc2626',
        background_color: '#ffffff',
        display: 'standalone',
        start_url: '/',
        scope: '/',
        icons: [
          {
            src: '/pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: '/pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: '/pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable',
          },
        ],
      },
      workbox: {
        cleanupOutdatedCaches: true,
      },
      devOptions: {
        enabled: false,
      },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
})
```

注意：`VueRouter()` 必须放在 `Vue()` 前面，否则官方 file-based routing 生成类型和路由时可能异常。

## 7. 配置 TypeScript include

修改 `tsconfig.app.json`，确保包含自动生成的类型文件：

```json
{
  "extends": "@vue/tsconfig/tsconfig.dom.json",
  "include": [
    "env.d.ts",
    "src/**/*",
    "src/**/*.vue",
    "src/types/**/*.d.ts"
  ],
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "moduleResolution": "Bundler",
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  }
}
```

新增 `src/types/vite-plugin-pwa.d.ts`：

```ts
/// <reference types="vite-plugin-pwa/client" />
```

## 8. 配置 Tailwind CSS

新增或修改 `src/assets/styles/main.css`：

```css
@import "tailwindcss";

:root {
  color: #111827;
  background: #f7f7f8;
  font-family:
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    sans-serif;
}

html,
body,
#app {
  min-height: 100%;
  margin: 0;
}

body {
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;
}

button,
input,
textarea {
  font: inherit;
}
```

如果 Tailwind base 样式与 Vant 组件出现冲突，优先在业务 CSS 中覆盖具体元素，不要直接修改 `node_modules/vant`。

## 9. 配置入口文件

修改 `src/main.ts`：

```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import { i18n } from './plugins/i18n'
import './assets/styles/main.css'
import './pwa'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.mount('#app')
```

## 10. File Based Routing

新增 `src/router/index.ts`：

```ts
import { createRouter, createWebHashHistory } from 'vue-router'
import { handleHotUpdate, routes } from 'vue-router/auto-routes'

export const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

if (import.meta.hot) {
  handleHotUpdate(router)
}
```

说明：

- H5 项目如果部署在 Nginx、OSS、CDN 或后端静态目录中，`createWebHashHistory` 对刷新 404 更稳。
- 如果服务器已配置 SPA fallback，可改为 `createWebHistory(import.meta.env.BASE_URL)`。

页面文件与路由关系：

```text
src/pages/index.vue              -> /
src/pages/login.vue              -> /login
src/pages/mine/index.vue         -> /mine
src/pages/red-packet/[id].vue    -> /red-packet/:id
```

页面内可通过 `definePage()` 增加 meta：

```vue
<script setup lang="ts">
definePage({
  name: 'mine',
  meta: {
    title: '我的',
    requiresAuth: true,
    keepAlive: true,
  },
})
</script>

<template>
  <main class="min-h-screen bg-gray-50 px-4 py-3">
    <h1 class="text-lg font-semibold">{{ $t('mine.title') }}</h1>
  </main>
</template>
```

## 11. App.vue 与 KeepAlive

修改 `src/App.vue`：

```vue
<script setup lang="ts">
const keepAliveNames = ['home', 'mine']
</script>

<template>
  <RouterView v-slot="{ Component, route }">
    <KeepAlive :include="keepAliveNames">
      <component
        :is="Component"
        v-if="route.meta.keepAlive"
        :key="String(route.name)"
      />
    </KeepAlive>

    <component
      :is="Component"
      v-if="!route.meta.keepAlive"
      :key="route.fullPath"
    />
  </RouterView>
</template>
```

KeepAlive 要点：

- `include` 匹配的是组件名或路由组件生成名，建议在 `definePage({ name: 'home' })` 中显式声明。
- 列表页、首页、我的页适合缓存。
- 登录页、详情页、支付页通常不缓存，避免状态残留。

## 12. Pinia 状态管理

新增 `src/stores/app.ts`：

```ts
export type Locale = 'zh-CN' | 'en-US'

export const useAppStore = defineStore('app', () => {
  const locale = ref<Locale>('zh-CN')
  const theme = ref<'light' | 'dark'>('light')

  function setLocale(value: Locale) {
    locale.value = value
  }

  function setTheme(value: 'light' | 'dark') {
    theme.value = value
  }

  return {
    locale,
    theme,
    setLocale,
    setTheme,
  }
})
```

新增 `src/stores/user.ts`：

```ts
import { getTtlStorage, removeTtlStorage, setTtlStorage } from '@/utils/storage'

export interface UserInfo {
  id: number
  username: string
  nickname?: string
}

const TOKEN_KEY = 'Authorization'
const TOKEN_TTL = 7 * 24 * 60 * 60 * 1000

export const useUserStore = defineStore('user', () => {
  const token = ref(getTtlStorage<string>(TOKEN_KEY) ?? '')
  const userInfo = ref<UserInfo | null>(null)

  const isLogin = computed(() => Boolean(token.value))

  function setToken(value: string) {
    token.value = value
    setTtlStorage(TOKEN_KEY, value, TOKEN_TTL)
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    removeTtlStorage(TOKEN_KEY)
  }

  return {
    token,
    userInfo,
    isLogin,
    setToken,
    logout,
  }
})
```

因为 `vite.config.ts` 已配置 `AutoImport`，这里可以直接使用 `ref`、`computed`、`defineStore`。

## 13. i18n 国际化

新增 `src/locales/zh-CN.ts`：

```ts
export default {
  common: {
    confirm: '确认',
    cancel: '取消',
    loading: '加载中',
  },
  home: {
    title: '红包大厅',
  },
  mine: {
    title: '我的',
  },
}
```

新增 `src/locales/en-US.ts`：

```ts
export default {
  common: {
    confirm: 'Confirm',
    cancel: 'Cancel',
    loading: 'Loading',
  },
  home: {
    title: 'Red Packets',
  },
  mine: {
    title: 'Mine',
  },
}
```

新增 `src/locales/index.ts`：

```ts
import enUS from './en-US'
import zhCN from './zh-CN'

export const messages = {
  'zh-CN': zhCN,
  'en-US': enUS,
}
```

新增 `src/plugins/i18n.ts`：

```ts
import { createI18n } from 'vue-i18n'
import { getTtlStorage } from '@/utils/storage'
import { messages } from '@/locales'
import type { Locale } from '@/stores/app'

export const i18n = createI18n({
  legacy: false,
  locale: getTtlStorage<Locale>('locale') ?? 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages,
})
```

组件里使用：

```vue
<script setup lang="ts">
const { t, locale } = useI18n()

function switchLocale() {
  locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
}
</script>

<template>
  <van-button type="primary" @click="switchLocale">
    {{ t('common.confirm') }}
  </van-button>
</template>
```

## 14. 带时效 localStorage

新增 `src/utils/storage.ts`：

```ts
interface TtlStoragePayload<T> {
  value: T
  expiresAt: number | null
}

function now() {
  return Date.now()
}

function getStorage() {
  if (typeof window === 'undefined') return null
  return window.localStorage
}

export function setTtlStorage<T>(key: string, value: T, ttlMs?: number) {
  const storage = getStorage()
  if (!storage) return

  const payload: TtlStoragePayload<T> = {
    value,
    expiresAt: ttlMs ? now() + ttlMs : null,
  }

  storage.setItem(key, JSON.stringify(payload))
}

export function getTtlStorage<T>(key: string): T | null {
  const storage = getStorage()
  if (!storage) return null

  const raw = storage.getItem(key)
  if (!raw) return null

  try {
    const payload = JSON.parse(raw) as TtlStoragePayload<T>

    if (payload.expiresAt && payload.expiresAt <= now()) {
      storage.removeItem(key)
      return null
    }

    return payload.value
  } catch {
    storage.removeItem(key)
    return null
  }
}

export function removeTtlStorage(key: string) {
  getStorage()?.removeItem(key)
}

export function clearExpiredTtlStorage(prefix?: string) {
  const storage = getStorage()
  if (!storage) return

  for (let index = storage.length - 1; index >= 0; index -= 1) {
    const key = storage.key(index)
    if (!key) continue
    if (prefix && !key.startsWith(prefix)) continue

    const raw = storage.getItem(key)
    if (!raw) continue

    try {
      const payload = JSON.parse(raw) as TtlStoragePayload<unknown>

      if (payload.expiresAt && payload.expiresAt <= now()) {
        storage.removeItem(key)
      }
    } catch {
      // 普通 localStorage 项不是 TTL 结构，不处理。
    }
  }
}
```

新增响应式组合式函数 `src/composables/useTtlLocalStorage.ts`：

```ts
import {
  getTtlStorage,
  removeTtlStorage,
  setTtlStorage,
} from '@/utils/storage'

export function useTtlLocalStorage<T>(
  key: string,
  initialValue: T,
  ttlMs?: number,
) {
  const stored = getTtlStorage<T>(key)
  const state = ref<T>(stored ?? initialValue)

  watch(
    state,
    value => {
      setTtlStorage(key, value, ttlMs)
    },
    { deep: true },
  )

  function remove() {
    removeTtlStorage(key)
    state.value = initialValue
  }

  return {
    state,
    remove,
  }
}
```

使用示例：

```ts
const { state: inviteCode } = useTtlLocalStorage('invite_code', '', 30 * 60 * 1000)
```

## 15. axios 通用 HTTP 请求封装

新增 `.env.development`：

```bash
VITE_API_BASE_URL=/api
```

新增 `.env.production`：

```bash
VITE_API_BASE_URL=/api
```

新增 `src/utils/http.ts`：

```ts
import axios from 'axios'
import type {
  AxiosError,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
} from 'axios'
import { getTtlStorage } from '@/utils/storage'

export interface ApiEnvelope<T = unknown> {
  code: number
  msg?: string
  message?: string
  data: T
}

const TOKEN_KEY = 'Authorization'

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
})

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getTtlStorage<string>(TOKEN_KEY)

  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  return config
})

http.interceptors.response.use(
  response => {
    const body = response.data as ApiEnvelope

    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0 || body.code === 200) {
        return body.data
      }

      showFailToast(body.msg || body.message || '请求失败')
      return Promise.reject(body)
    }

    return body
  },
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      showFailToast('登录已过期')
      return Promise.reject(error)
    }

    showFailToast(error.message || '网络异常')
    return Promise.reject(error)
  },
)

export function request<T = unknown>(config: AxiosRequestConfig) {
  return http.request<ApiEnvelope<T>, T>(config)
}

export const get = <T = unknown>(url: string, config?: AxiosRequestConfig) =>
  request<T>({ ...config, url, method: 'GET' })

export const post = <T = unknown>(
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig,
) => request<T>({ ...config, url, data, method: 'POST' })

export const put = <T = unknown>(
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig,
) => request<T>({ ...config, url, data, method: 'PUT' })

export const del = <T = unknown>(url: string, config?: AxiosRequestConfig) =>
  request<T>({ ...config, url, method: 'DELETE' })

export { http }
```

新增 `src/api/user.ts`：

```ts
import { get, post } from '@/utils/http'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
}

export interface Profile {
  id: number
  username: string
  nickname?: string
}

export function loginApi(data: LoginParams) {
  return post<LoginResult>('/login', data)
}

export function profileApi() {
  return get<Profile>('/user/profile')
}
```

如果后端响应结构不是 `{ code, msg, data }`，只需要调整 `http.interceptors.response` 中的成功码和返回字段。

## 16. PWA 注册

新增 `src/pwa.ts`：

```ts
import { registerSW } from 'virtual:pwa-register'

export const updateServiceWorker = registerSW({
  immediate: true,
  onNeedRefresh() {
    showConfirmDialog({
      title: '发现新版本',
      message: '是否立即刷新并使用最新版本？',
    }).then(() => {
      updateServiceWorker(true)
    })
  },
  onOfflineReady() {
    showToast('已支持离线访问')
  },
})
```

PWA 图标要求：

- `public/pwa-192x192.png`
- `public/pwa-512x512.png`
- 建议提供 maskable 图标，避免 Android 安装后图标被裁切。

生产验证：

```bash
pnpm build
pnpm preview
```

然后在 Chrome DevTools 的 Application 面板检查 Manifest 和 Service Worker。

## 17. 示例页面

新增 `src/pages/index.vue`：

```vue
<script setup lang="ts">
definePage({
  name: 'home',
  meta: {
    title: '红包大厅',
    keepAlive: true,
  },
})

const { t } = useI18n()
const router = useRouter()
</script>

<template>
  <main class="min-h-screen bg-gray-50 px-4 py-4">
    <header class="mb-4 flex items-center justify-between">
      <h1 class="text-lg font-semibold text-gray-950">{{ t('home.title') }}</h1>
      <van-button size="small" type="primary" @click="router.push('/mine')">
        {{ t('mine.title') }}
      </van-button>
    </header>

    <van-cell-group inset>
      <van-cell title="普通红包" value="进入" is-link to="/red-packet/1" />
      <van-cell title="扫雷红包" value="进入" is-link to="/red-packet/2" />
    </van-cell-group>
  </main>
</template>
```

新增 `src/pages/mine/index.vue`：

```vue
<script setup lang="ts">
definePage({
  name: 'mine',
  meta: {
    title: '我的',
    requiresAuth: true,
    keepAlive: true,
  },
})

const userStore = useUserStore()
const { t } = useI18n()
</script>

<template>
  <main class="min-h-screen bg-gray-50 px-4 py-4">
    <h1 class="mb-4 text-lg font-semibold text-gray-950">{{ t('mine.title') }}</h1>

    <van-cell-group inset>
      <van-cell title="登录状态" :value="userStore.isLogin ? '已登录' : '未登录'" />
    </van-cell-group>
  </main>
</template>
```

新增 `src/pages/red-packet/[id].vue`：

```vue
<script setup lang="ts">
definePage({
  name: 'red-packet-detail',
  meta: {
    title: '红包详情',
  },
})

const route = useRoute()
const packetId = computed(() => {
  const params = route.params as { id?: string | string[] }
  const id = params.id

  return Array.isArray(id) ? id[0] : (id ?? '')
})
</script>

<template>
  <main class="min-h-screen bg-red-50 px-4 py-4">
    <h1 class="text-lg font-semibold text-red-700">红包 {{ packetId }}</h1>
  </main>
</template>
```

## 18. 路由鉴权

在 `src/router/index.ts` 中追加：

```ts
router.beforeEach(to => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.isLogin) {
    return {
      path: '/login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  return true
})
```

新增 `src/pages/login.vue`：

```vue
<script setup lang="ts">
definePage({
  name: 'login',
  meta: {
    title: '登录',
  },
})

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

async function login() {
  userStore.setToken('mock-token')
  await router.replace(String(route.query.redirect || '/'))
}
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-gray-50 px-4">
    <van-button block type="primary" @click="login">登录</van-button>
  </main>
</template>
```

## 19. package.json 脚本

建议保留或调整为：

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "preview": "vite preview --host 0.0.0.0",
    "typecheck": "vue-tsc -b --noEmit"
  }
}
```

常用命令：

```bash
pnpm dev
pnpm typecheck
pnpm build
pnpm preview
```

## 20. 初始化完成检查清单

- `pnpm dev` 可以启动项目。
- 页面中可直接使用 `ref`、`computed`、`useRouter`、`useRoute`、`useI18n`、`defineStore`。
- 页面中可直接使用 `<van-button>`、`<van-cell>` 等 Vant 组件，无需手动 import。
- `src/pages/index.vue` 自动映射到 `/`。
- `src/pages/red-packet/[id].vue` 自动映射到 `/red-packet/:id`。
- `definePage({ meta: { keepAlive: true } })` 的页面可以被缓存。
- `pnpm build` 会同时执行 TypeScript 检查和 Vite 构建。
- `pnpm preview` 后 Application 面板可以看到 Manifest。
- axios 请求会自动携带 `Authorization`。
- 过期 localStorage 会在读取时自动清理。

## 21. 常见问题

### 21.1 自动导入类型不生效

先启动一次开发服务生成类型文件：

```bash
pnpm dev
```

然后确认这些文件已在 `tsconfig.app.json` 的 `include` 中：

```text
src/types/auto-imports.d.ts
src/types/components.d.ts
src/types/typed-router.d.ts
```

### 21.2 Vant 样式没有自动引入

确认 `vite.config.ts` 中 `Components` 使用的是：

```ts
import { VantResolver } from '@vant/auto-import-resolver'

Components({
  resolvers: [VantResolver()],
})
```

### 21.3 i18n 的 `useI18n` 报错

确认 `createI18n` 设置：

```ts
createI18n({
  legacy: false,
})
```

Composition API 模式必须关闭 legacy。

### 21.4 PWA 本地开发看不到 Service Worker

本文档默认：

```ts
devOptions: {
  enabled: false,
}
```

如果需要本地调试 PWA，可临时改为：

```ts
devOptions: {
  enabled: true,
}
```

调试完建议改回 `false`，避免开发缓存干扰页面调试。

### 21.5 部署后刷新页面 404

如果使用 `createWebHistory`，服务器必须把所有前端路由 fallback 到 `index.html`。如果不方便配置服务器，H5 项目建议使用本文档默认的 `createWebHashHistory`。

## 22. 官方参考

- Vite 创建项目：https://vite.dev/guide/
- Vue TypeScript：https://vuejs.org/guide/typescript/overview
- Vue Router File Based Routing：https://router.vuejs.org/file-based-routing/
- Vue Router Extending Routes：https://router.vuejs.org/file-based-routing/extending-routes
- Vant：https://vant-ui.github.io/
- Vant 自动导入解析器：https://www.npmjs.com/package/@vant/auto-import-resolver
- unplugin-auto-import：https://unplugin.unjs.io/showcase/unplugin-auto-import.html
- unplugin-vue-components：https://unplugin.unjs.io/showcase/unplugin-vue-components
- Pinia：https://pinia.vuejs.org/
- Vue I18n Composition API：https://vue-i18n.intlify.dev/guide/advanced/composition
- Tailwind CSS Vite 安装：https://tailwindcss.com/docs/installation/using-vite
- Vite PWA：https://vite-pwa-org.netlify.app/guide/
- Axios Instance：https://axios-http.com/docs/instance
- VueUse useStorage：https://vueuse.org/core/useStorage/
