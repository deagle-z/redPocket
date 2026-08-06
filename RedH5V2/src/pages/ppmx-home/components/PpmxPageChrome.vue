<script setup lang="ts">
import { showToast } from 'vant'
import {
  ppmxDrawerItems,
  ppmxNavItems,
  ppmxNotices,
} from '../data'
import { formatMoney, toDisplayCents } from '@/utils/money'
import { APP_CURRENCY_SYMBOL } from '@/config/market'
import { isPublicTabbarRoute } from '@/utils/authGate'
import { usePpmxLocale } from '../composables/usePpmxLocale'
import type { LocaleCode, PpmxAuthMode, PpmxRouteTarget } from '../types'
import PpmxAuthModal from './PpmxAuthModal.vue'
import PpmxHeader from './PpmxHeader.vue'
import PpmxLogo from './PpmxLogo.vue'
import PpmxMobileDrawer from './PpmxMobileDrawer.vue'
import PpmxNotificationPopover from './PpmxNotificationPopover.vue'
import PpmxSymbols from './PpmxSymbols.vue'
import PpmxTopStrip from './PpmxTopStrip.vue'
import PpmxThemeToggle from '@/components/PpmxThemeToggle.vue'
import PpmxSupportModal from '@/components/PpmxSupportModal.vue'

defineOptions({ name: 'PpmxPageChrome' })

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const drawerOpen = ref(false)
const notificationOpen = ref(false)
const authModalOpen = ref(false)
const authModalMode = ref<PpmxAuthMode>('register')
const languageMenuOpen = ref(false)
const drawerLanguageOpen = ref(false)
const supportModalOpen = ref(false)
const notificationButtonRef = ref<HTMLElement | null>(null)
const notificationPopoverRef = ref<{ $el: HTMLElement } | null>(null)
const languageSwitcherRef = ref<HTMLElement | null>(null)
const { t } = useI18n()
const { homeLocale, pickText, setHomeLocale } = usePpmxLocale()
const ppmxLanguageOptions: Array<{
  locale: LocaleCode
  region: string
  label: string
  triggerLabel: string
  country: string
}> = [
  { locale: 'es', region: 'PE', label: 'Español', triggerLabel: 'Español', country: 'Perú' },
  { locale: 'en', region: 'PE', label: 'English', triggerLabel: 'English', country: 'Peru' },
  { locale: 'zh', region: 'CN', label: '简体中文', triggerLabel: '中文', country: 'China' },
]
const headerBalance = computed(() => formatMoney(toDisplayCents(userStore.userInfo?.balance), {
  currency: APP_CURRENCY_SYMBOL,
}))
const activeLanguage = computed(() =>
  ppmxLanguageOptions.find((item) => item.locale === homeLocale.value) || ppmxLanguageOptions[0],
)
const headerAccountName = computed(() => (
  userStore.userInfo?.nickname
  || userStore.userInfo?.username
  || t('ppmx.chrome.accountFallback')
))
const isTabbarRoute = computed(() => isPublicTabbarRoute(route.path))
const headerMenuAriaLabel = computed(() =>
  isTabbarRoute.value ? t('ppmx.chrome.openMenu') : t('ppmx.chrome.back'),
)
const headerMenuIcon = computed(() =>
  isTabbarRoute.value ? 'fa-bars' : 'fa-arrow-left',
)

function isTargetActive(target: PpmxRouteTarget | string) {
  const routePath = resolveRouteTarget(target)

  if (!routePath) return false
  if (routePath === '/') return route.path === '/'

  return route.path === routePath || route.path.startsWith(`${routePath}/`)
}

function openAuth(mode: PpmxAuthMode) {
  drawerOpen.value = false
  notificationOpen.value = false
  authModalMode.value = mode
  authModalOpen.value = true
}

function requireRegisterAuth() {
  if (userStore.isLogin) return true

  openAuth('register')
  return false
}

function chooseHomeLocale(locale: LocaleCode) {
  setHomeLocale(locale)
  drawerOpen.value = false
  languageMenuOpen.value = false
  drawerLanguageOpen.value = false
}

function toggleLanguageMenu() {
  languageMenuOpen.value = !languageMenuOpen.value
  notificationOpen.value = false
}

function resolveRouteTarget(target: PpmxRouteTarget | string) {
  if (target === 'home') return '/'
  if (target === 'wheel') return '/wheel'
  if (target === 'team') return '/team'
  if (target === 'promo') return '/promo'
  if (target === 'casino') return '/casino'
  if (target === 'download') return '/download'
  if (target === 'guide') return '/guide'
  if (target === 'records') return '/records'
  if (target === 'security') return '/security'
  if (target === 'vip') return '/vip'
  if (target === 'responsible') return '/responsible'
  if (target === 'account' || target === 'profile') return '/profile'

  return ''
}

function goToTarget(target: PpmxRouteTarget | string) {
  drawerOpen.value = false
  drawerLanguageOpen.value = false
  notificationOpen.value = false

  if (target === 'help' || target === 'support') {
    openSupport()
    return
  }

  const routePath = resolveRouteTarget(target)

  if (routePath) {
    void router.push(routePath)
    return
  }

  if (!requireRegisterAuth()) return

  showToast(t('ppmx.common.comingSoon'))
}

function openSupport() {
  drawerOpen.value = false
  drawerLanguageOpen.value = false
  notificationOpen.value = false
  languageMenuOpen.value = false
  supportModalOpen.value = true
}

function goBack() {
  drawerOpen.value = false
  drawerLanguageOpen.value = false
  notificationOpen.value = false
  languageMenuOpen.value = false

  if (window.history.length > 1) {
    router.back()
    return
  }

  void router.push('/profile')
}

function handleHeaderMenuClick() {
  if (isTabbarRoute.value) {
    drawerOpen.value = true
    return
  }

  goBack()
}

function onDocumentPointerDown(event: PointerEvent) {
  if (!notificationOpen.value && !languageMenuOpen.value) return

  const target = event.target
  if (!(target instanceof Node)) return

  const popoverElement = notificationPopoverRef.value?.$el
  if (popoverElement?.contains(target) || notificationButtonRef.value?.contains(target)) {
    return
  }

  if (languageSwitcherRef.value?.contains(target)) {
    return
  }

  notificationOpen.value = false
  languageMenuOpen.value = false
}

onMounted(() => {
  document.addEventListener('pointerdown', onDocumentPointerDown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocumentPointerDown)
})

defineExpose({
  goToTarget,
  openAuth,
  openSupport,
  requireRegisterAuth,
})
</script>

<template>
  <PpmxSymbols />

  <PpmxTopStrip>
    <div class="ppmx-top-strip__inner">
      <div class="ppmx-top-strip__support">
        <button class="ppmx-top-strip__support-button" type="button" @click="openSupport">
          <i class="fa-solid fa-headset" /> {{ t('ppmx.chrome.support247') }}
        </button>
        <span><i class="fa-solid fa-shield-halved" /> {{ t('ppmx.chrome.responsibleGaming') }}</span>
      </div>
      <div class="ppmx-top-strip__mobile">
        <span>{{ t('ppmx.chrome.mobileResponsible') }}</span>
      </div>
      <div class="ppmx-top-strip__actions">
        <button class="ppmx-top-strip__age" type="button" @click="goToTarget('responsible')">
          {{ t('ppmx.chrome.mobileResponsible') }}
        </button>
        <PpmxThemeToggle class="ppmx-theme-toggle--top" />
        <div
          ref="languageSwitcherRef"
          class="ppmx-language-switcher"
          :class="{ 'is-open': languageMenuOpen }"
        >
          <button
            class="ppmx-language-trigger"
            type="button"
            aria-haspopup="listbox"
            :aria-expanded="languageMenuOpen"
            @click="toggleLanguageMenu"
          >
            <span>{{ activeLanguage.region }}</span>
            <strong>{{ activeLanguage.triggerLabel }}</strong>
            <i class="fa-solid fa-chevron-down" />
          </button>

          <div v-if="languageMenuOpen" class="ppmx-language-menu" role="listbox">
            <button
              v-for="option in ppmxLanguageOptions"
              :key="option.locale"
              class="ppmx-language-option"
              :class="{ 'is-active': option.locale === homeLocale }"
              type="button"
              role="option"
              :aria-selected="option.locale === homeLocale"
              @click="chooseHomeLocale(option.locale)"
            >
              <span class="ppmx-language-option__region">{{ option.region }}</span>
              <span class="ppmx-language-option__label">
                <strong>{{ option.label }}</strong>
                <small>{{ option.country }}</small>
              </span>
              <i v-if="option.locale === homeLocale" class="fa-solid fa-check" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </PpmxTopStrip>

  <PpmxHeader>
    <div class="ppmx-header__inner">
      <button
        id="navMenu"
        class="ppmx-icon-button ppmx-header__menu"
        type="button"
        :aria-label="headerMenuAriaLabel"
        @click="handleHeaderMenuClick"
      >
        <i class="fa-solid" :class="headerMenuIcon" />
      </button>

      <button class="ppmx-logo-button" type="button" :aria-label="t('ppmx.chrome.homeAria')" @click="goToTarget('home')">
        <PpmxLogo />
      </button>

      <nav class="ppmx-header__desktop-nav" :aria-label="t('ppmx.chrome.mainNavigation')">
        <button
          v-for="item in ppmxNavItems"
          :key="item.id"
          class="ppmx-header-nav__item"
          :class="{ 'is-active': isTargetActive(item.target) }"
          type="button"
          @click="goToTarget(item.target)"
        >
          <i class="fa-solid" :class="item.icon" />
          <span>{{ pickText(item.label) }}</span>
        </button>
      </nav>

      <div class="ppmx-header__actions">
        <button v-if="!userStore.isLogin" class="ppmx-btn-ghost ppmx-auth-button ppmx-auth-button--login" type="button" @click="openAuth('login')">
          {{ t('ppmx.chrome.login') }}
        </button>
        <button v-if="!userStore.isLogin" class="ppmx-btn-red ppmx-auth-button" type="button" @click="openAuth('register')">
          {{ t('ppmx.chrome.register') }}
        </button>
        <button v-if="userStore.isLogin" class="ppmx-wallet" type="button" @click="goToTarget('profile')">
          <i class="fa-solid fa-wallet" />
          <span>{{ headerBalance }}</span>
          <small>{{ headerAccountName }}</small>
        </button>
        <button
          ref="notificationButtonRef"
          class="ppmx-icon-button ppmx-notification-button"
          type="button"
          :aria-label="t('ppmx.chrome.notifications')"
          @click="notificationOpen = !notificationOpen"
        >
          <i class="fa-solid fa-bell" />
          <span class="ppmx-notification-button__label">{{ t('ppmx.chrome.notifications') }}</span>
          <span class="ppmx-unread-dot" />
        </button>
      </div>

      <PpmxNotificationPopover v-if="notificationOpen" ref="notificationPopoverRef">
        <div class="ppmx-popover-head">
          <strong>{{ t('ppmx.chrome.notifications') }}</strong>
          <button type="button" :aria-label="t('ppmx.chrome.closeNotifications')" @click="notificationOpen = false">
            <i class="fa-solid fa-xmark" />
          </button>
        </div>
        <button
          v-for="notice in ppmxNotices.slice(0, 4)"
          :key="notice.id"
          class="ppmx-popover-item"
          type="button"
          @click="notice.target ? goToTarget(notice.target) : undefined"
        >
          <i class="fa-solid" :class="notice.icon" />
          <span>{{ pickText(notice.text) }}</span>
        </button>
      </PpmxNotificationPopover>
    </div>
  </PpmxHeader>

  <div v-if="drawerOpen" class="ppmx-drawer-backdrop" @click="drawerOpen = false" />
  <PpmxMobileDrawer v-if="drawerOpen" id="drawer">
    <div class="ppmx-drawer-head">
      <button class="ppmx-drawer-logo" type="button" :aria-label="t('ppmx.chrome.homeAria')" @click="goToTarget('home')">
        <PpmxLogo />
      </button>
      <button class="ppmx-drawer-close" type="button" :aria-label="t('ppmx.chrome.closeMenu')" @click="drawerOpen = false">
        <i class="fa-solid fa-xmark" />
      </button>
    </div>

    <section class="ppmx-drawer-account" :class="{ 'is-guest': !userStore.isLogin }">
      <template v-if="userStore.isLogin">
        <span class="ppmx-drawer-account__icon">
          <i class="fa-solid fa-wallet" />
        </span>
        <div class="ppmx-drawer-account__body">
          <small>{{ headerAccountName }}</small>
          <strong>{{ headerBalance }}</strong>
        </div>
        <button type="button" class="ppmx-drawer-account__action" @click="goToTarget('profile')">
          <i class="fa-solid fa-chevron-right" />
        </button>
      </template>
      <template v-else>
        <div class="ppmx-drawer-account__body">
          <small>{{ t('ppmx.chrome.accountFallback') }}</small>
          <strong>{{ t('auth.registerSub1') }}</strong>
        </div>
        <div class="ppmx-drawer-auth-actions">
          <button type="button" class="ppmx-drawer-auth ppmx-drawer-auth--ghost" @click="openAuth('login')">
            {{ t('ppmx.chrome.login') }}
          </button>
          <button type="button" class="ppmx-drawer-auth ppmx-drawer-auth--red" @click="openAuth('register')">
            {{ t('ppmx.chrome.register') }}
          </button>
        </div>
      </template>
    </section>

    <nav class="ppmx-drawer-nav" :aria-label="t('ppmx.chrome.mobileMenu')">
      <button
        v-for="item in ppmxDrawerItems"
        :key="item.id"
        :class="{ 'is-active': isTargetActive(item.target) }"
        type="button"
        @click="goToTarget(item.target)"
      >
        <span class="ppmx-drawer-nav__icon">
          <i class="fa-solid" :class="item.icon" />
        </span>
        <span>{{ pickText(item.label) }}</span>
        <i class="fa-solid fa-chevron-right ppmx-drawer-nav__chevron" />
      </button>
    </nav>
    <div class="ppmx-drawer-language">
      <div class="ppmx-drawer-theme-row">
        <span class="ppmx-drawer-theme-label">
          <i class="fa-solid fa-circle-half-stroke" />
          {{ t('theme.toggle') }}
        </span>
        <PpmxThemeToggle class="ppmx-theme-toggle--drawer" />
      </div>
      <button
        class="ppmx-drawer-lang-trigger"
        type="button"
        :aria-expanded="drawerLanguageOpen"
        aria-controls="drawerLangPanel"
        @click="drawerLanguageOpen = !drawerLanguageOpen"
      >
        <span class="ppmx-drawer-lang-trigger__label">
          <i class="fa-solid fa-globe" />
          {{ t('ppmx.chrome.language') }}
        </span>
        <span class="ppmx-drawer-lang-trigger__state">
          <strong>{{ activeLanguage.region }}</strong>
          <span>{{ activeLanguage.triggerLabel }}</span>
          <i
            id="drawerLangChev"
            class="fa-solid fa-chevron-down"
            :class="{ 'is-open': drawerLanguageOpen }"
          />
        </span>
      </button>
      <div
        id="drawerLangPanel"
        class="ppmx-drawer-lang-panel"
        :class="{ open: drawerLanguageOpen }"
      >
      <button
        v-for="option in ppmxLanguageOptions"
        :key="option.locale"
        class="ppmx-drawer-lang"
        :class="{ 'is-active': option.locale === homeLocale }"
        type="button"
        @click="chooseHomeLocale(option.locale)"
      >
        <span class="dl-flag">{{ option.region }}</span>
        <span>{{ option.triggerLabel }}</span>
      </button>
      </div>
    </div>
  </PpmxMobileDrawer>

  <slot />

  <PpmxAuthModal v-model="authModalOpen" v-model:mode="authModalMode" />
  <PpmxSupportModal v-model="supportModalOpen" />
</template>
