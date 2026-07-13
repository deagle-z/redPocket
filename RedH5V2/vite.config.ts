import path from 'node:path'
import { defineConfig, loadEnv } from 'vite'
import Vue from '@vitejs/plugin-vue'
import VueRouter from 'vue-router/vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { VantResolver } from '@vant/auto-import-resolver'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'
import { APP_BRAND } from './src/config/brand'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const appBase = env.VITE_CDN_BASE_URL || env.VITE_APP_BASE || '/'
  const apiProxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080'

  return {
    base: appBase,
    plugins: [
      VueRouter({
        routesFolder: 'src/pages',
        exclude: [
          'src/pages/ppmx-home/**',
          'src/pages/**/components/**',
          'src/pages/**/composables/**',
          'src/pages/**/*.test.ts',
        ],
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
        dirs: [
          'src/composables',
          'src/stores',
          'src/stores/modules',
          'src/stores/modules/**',
        ],
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
        includeAssets: ['images/ppmx/emblem.svg', 'images/ppmx/logo.svg', 'offline.html'],
        manifest: {
          name: APP_BRAND.name,
          short_name: APP_BRAND.shortName,
          description: APP_BRAND.description,
          theme_color: APP_BRAND.themeColor,
          background_color: APP_BRAND.themeColor,
          display: 'standalone',
          start_url: appBase,
          scope: appBase,
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
          navigateFallback: '/offline.html',
          navigateFallbackDenylist: [/^\/api\//],
          runtimeCaching: [
            {
              urlPattern: ({ request }) =>
                ['style', 'script', 'worker', 'font', 'image'].includes(
                  request.destination,
                ),
              handler: 'CacheFirst',
              options: {
                cacheName: 'ppmx-static-assets',
                expiration: {
                  maxEntries: 80,
                  maxAgeSeconds: 7 * 24 * 60 * 60,
                },
              },
            },
          ],
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
          target: apiProxyTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
