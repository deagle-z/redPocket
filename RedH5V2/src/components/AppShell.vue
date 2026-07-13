<script setup lang="ts">
import { useResponsiveLayout } from '@/composables/useResponsiveLayout'
import { APP_BRAND } from '@/config/brand'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { deviceKind, isMobile, isPc } = useResponsiveLayout()

const navItems = computed(() => [
  {
    icon: 'wap-home-o',
    label: t('nav.home'),
    path: '/',
  },
  {
    icon: 'user-o',
    label: t('nav.account'),
    path: '/profile',
  },
])

const pageTitle = computed(() =>
  route.meta.titleKey ? t(route.meta.titleKey) : String(route.meta.title || APP_BRAND.name),
)
const showMobileTabbar = computed(() => route.meta.tabbar === true && isMobile.value)

function goTo(path: string) {
  if (route.path !== path) {
    void router.push(path)
  }
}
</script>

<template>
  <div class="app-shell" :class="`app-shell--${deviceKind}`">
    <header v-if="isPc" class="pc-topbar">
      <button class="pc-brand" type="button" aria-label="返回首页" @click="goTo('/')">
        <span class="pc-brand__mark">{{ APP_BRAND.mark }}</span>
        <span class="pc-brand__text">
          <strong>{{ APP_BRAND.name }}</strong>
          <small>{{ APP_BRAND.tagline }}</small>
        </span>
      </button>

      <nav class="pc-nav" aria-label="主导航">
        <button
          v-for="item in navItems"
          :key="item.path"
          class="pc-nav__item"
          :class="{ 'is-active': route.path === item.path }"
          type="button"
          @click="goTo(item.path)"
        >
          <van-icon :name="item.icon" />
          <span>{{ item.label }}</span>
        </button>
      </nav>

      <div class="pc-topbar__meta">
        <span class="live-dot" />
        <span>{{ pageTitle }}</span>
      </div>
    </header>

    <div class="app-stage" :class="{ 'app-stage--with-tabbar': showMobileTabbar }">
      <section v-if="isPc" class="pc-context-bar">
        <div>
          <p class="eyebrow">{{ APP_BRAND.contextEyebrow }}</p>
          <h1>{{ pageTitle }}</h1>
        </div>
        <span class="pc-context-bar__badge">{{ APP_BRAND.contextBadge }}</span>
      </section>

      <div class="app-content">
        <slot />
      </div>
    </div>

    <AppTabbar v-if="showMobileTabbar" />
  </div>
</template>
