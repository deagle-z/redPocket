<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showToast } from 'vant'
import type { AppPopupAnnouncement } from '@/api/banner'
import { getPopupAnnouncements } from '@/api/banner'
import type { CheckInRecordItem, CheckInStatusResp } from '@/api/user'
import { doCheckIn, getCheckInRecords, getCheckInStatus } from '@/api/user'
import { filterPopupAnnouncements } from '@/utils/announcementPopup'
import { APP_CURRENCY, APP_CURRENCY_SYMBOL } from '@/config/market'
import {
  ppmxBanners,
  ppmxCheckinText,
  ppmxFooterColumns,
  ppmxFloatingWidgets,
  ppmxNotices,
  ppmxSocialItems,
  ppmxSocialText,
  ppmxTrustItems,
} from './data'
import { getPpmxFallbackGradient, getPpmxGameTileBackground } from './assets'
import PpmxSkeleton from '@/components/PpmxSkeleton.vue'
import { useCategoryScrollSpy } from './composables/useCategoryScrollSpy'
import { useFloatingWidgets } from './composables/useFloatingWidgets'
import { useHeroCarousel } from './composables/useHeroCarousel'
import { useNoticeCarousel } from './composables/useNoticeCarousel'
import { usePpmxGameHome } from './composables/usePpmxGameHome'
import { ppmxLocaleToAppLocale, usePpmxLocale } from './composables/usePpmxLocale'
import type { PpmxAuthMode, PpmxFloatingWidget, PpmxGame, PpmxHomeCategory, PpmxSocialItem } from './types'
import PpmxFloatingWidgets from './components/PpmxFloatingWidgets.vue'
import PpmxAnnouncementPopup from './components/PpmxAnnouncementPopup.vue'
import PpmxGameNav from './components/PpmxGameNav.vue'
import PpmxGameTile from './components/PpmxGameTile.vue'
import PpmxHeroCarousel from './components/PpmxHeroCarousel.vue'
import PpmxLobbySection from './components/PpmxLobbySection.vue'
import PpmxLogo from './components/PpmxLogo.vue'
import PpmxNoticeBar from './components/PpmxNoticeBar.vue'
import PpmxPageChrome from './components/PpmxPageChrome.vue'
import PpmxTrustSection from './components/PpmxTrustSection.vue'
import PpmxUnlockModal from '@/components/PpmxUnlockModal.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()
const { unlockOpen, ensureUnlocked } = useGameUnlockGate()

const pageChromeRef = ref<InstanceType<typeof PpmxPageChrome> | null>(null)
const activeCategoryId = ref('')
const touchStartX = ref<number | null>(null)
// Download-app promo activities are hidden from the home/promo/guide surfaces.
const HIDDEN_ACTIVITY_IDS = new Set(['download'])
const visibleBanners = ppmxBanners.filter((banner) => !HIDDEN_ACTIVITY_IDS.has(banner.id))
const hero = useHeroCarousel({ count: visibleBanners.length })
const notices = useNoticeCarousel({ count: ppmxNotices.length })
const floating = useFloatingWidgets()
const gameHome = usePpmxGameHome()
const { homeLocale, pickText } = usePpmxLocale()
useCategoryScrollSpy(gameHome.categoryIds, activeCategoryId)

const casinoCategoryCodes = new Set(['slots', 'casino', 'blockchain', 'fishing', 'rummy', 'sports', 'mini', 'lottery'])

const activeNotice = computed(() => ppmxNotices[notices.activeIndex.value])
const visibleFloatingWidgets = computed(() =>
  ppmxFloatingWidgets.filter((widget) => floating.isWidgetVisible(widget.id)),
)
const homeCategories = computed(() => gameHome.categories.value)

let heroTimer: number | undefined
let noticeTimer: number | undefined
const socialText = ppmxSocialText
const checkinText = ppmxCheckinText

const socialModalOpen = ref(false)
const checkinModalOpen = ref(false)
const checkinStatus = ref<CheckInStatusResp | null>(null)
const checkinRecords = ref<CheckInRecordItem[]>([])
const checkinLoading = ref(false)
const checkinFailed = ref(false)
const checkinSubmitting = ref(false)
const popupQueue = ref<AppPopupAnnouncement[]>([])
const popupVisible = ref(false)
const currentPopup = computed(() => popupQueue.value[0])
let popupRequestVersion = 0

const checkinRewards = computed(() => checkinStatus.value?.rewards ?? [])
const checkinTotalDays = computed(() => Number(checkinStatus.value?.totalCheckInDays ?? 0))
const checkinNextSeq = computed(() => Number(checkinStatus.value?.nextSeq ?? 1))
const checkinTodayChecked = computed(() => Boolean(checkinStatus.value?.todayChecked))
const checkinCompleted = computed(() => Boolean(checkinStatus.value?.completed))
const checkinRecentRecords = computed(() => checkinRecords.value.slice(0, 3))
const checkinTotalRewardDays = computed(() => checkinRewards.value.length)
const checkinClaimedDays = computed(() => Math.min(checkinTotalDays.value, checkinTotalRewardDays.value))
const checkinFocusSeq = computed(() => {
  if (checkinTotalRewardDays.value === 0) return 0
  if (checkinCompleted.value) return checkinTotalRewardDays.value
  if (checkinTodayChecked.value) return Math.max(1, checkinClaimedDays.value)

  return Math.min(Math.max(checkinNextSeq.value, 1), checkinTotalRewardDays.value)
})
const checkinSummaryAmount = computed(() => {
  const focusIndex = Math.max(checkinFocusSeq.value - 1, 0)
  const focusAmount = checkinRewards.value[focusIndex]

  if (checkinTodayChecked.value || checkinCompleted.value) return formatCheckinAmount(focusAmount, true)

  return formatCheckinMysteryAmount(true)
})
const checkinSummaryLabel = computed(() =>
  checkinTodayChecked.value || checkinCompleted.value
    ? pickText(checkinText.revealed)
    : pickText(checkinText.mystery),
)
const checkinSummaryState = computed(() => {
  if (checkinCompleted.value) return pickText(checkinText.completed)
  if (checkinTodayChecked.value) return pickText(checkinText.claimed)

  return pickText(checkinText.pending)
})
const checkinProgressStyle = computed(() => {
  const total = Math.max(checkinTotalRewardDays.value, 1)
  const progress = Math.min(100, Math.max(0, (checkinClaimedDays.value / total) * 100))

  return { '--ppmx-checkin-progress': `${progress}%` }
})
const checkinButtonDisabled = computed(
  () =>
    checkinLoading.value ||
    checkinFailed.value ||
    checkinSubmitting.value ||
    !checkinStatus.value ||
    checkinTodayChecked.value ||
    checkinCompleted.value ||
    checkinRewards.value.length === 0,
)
const checkinButtonText = computed(() => {
  if (checkinSubmitting.value) return pickText(checkinText.submitting)
  if (checkinTodayChecked.value) return pickText(checkinText.claimed)
  if (checkinCompleted.value) return pickText(checkinText.completed)

  return pickText(checkinText.claim)
})

function goToTarget(target: string) {
  if (target === 'home') {
    window.scrollTo({ top: 0, behavior: 'smooth' })
    return
  }

  if (target === 'registro' || target === 'register') {
    if (userStore.isLogin) {
      void router.push('/promo/registro')
    } else {
      openAuth('register')
    }
    return
  }

  if (target === 'responsible') {
    void router.push('/responsible')
    return
  }

  if (!requireRegisterAuth()) return

  if (target === 'recarga') {
    showToast(t('ppmx.common.comingSoon'))
    return
  }

  if (target === 'invita') {
    void router.push('/team')
    return
  }

  if (target === 'account' || target === 'profile') {
    void router.push('/profile')
    return
  }

  if (target === 'casino') {
    void router.push('/casino')
    return
  }

  if (target === 'download') {
    void router.push('/download')
    return
  }

  if (target === 'wheel' || target === 'team' || target === 'vip') {
    void router.push(`/${target}`)
    return
  }

  const category = homeCategories.value.find((item) => item.id === target || item.target === target)
  if (category) {
    scrollToCategory(category.id)
  }
}

function scrollToCategory(categoryId: PpmxHomeCategory['id']) {
  activeCategoryId.value = categoryId
  document.getElementById(`ppmx-cat-${categoryId}`)?.scrollIntoView({
    behavior: 'smooth',
    block: 'start',
  })
}

function openCategoryList(category: PpmxHomeCategory) {
  if (!requireRegisterAuth()) return

  if (category.target === 'casino') {
    void router.push({
      path: '/casino',
      query: casinoCategoryCodes.has(category.id)
        ? { categoryCode: category.id }
        : undefined,
    })
    return
  }

  goToTarget(category.target || category.id)
}

function getGameTileStyle(game: PpmxGame, index = 0) {
  const imageId = game.imageId || index

  if (game.cover) {
    return {
      backgroundImage: `url("${game.cover.replace(/"/g, '%22')}"), ${getPpmxFallbackGradient(imageId)}`,
    }
  }

  return getPpmxGameTileBackground(imageId)
}

function openPromo(bannerId: string) {
  if (bannerId === 'registro') {
    if (userStore.isLogin) {
      void router.push('/promo/registro')
    } else {
      openAuth('register')
    }
    return
  }

  void router.push(`/promo/${bannerId}`)
}

function openAuth(mode: PpmxAuthMode) {
  pageChromeRef.value?.openAuth(mode)
}

function requireRegisterAuth() {
  return pageChromeRef.value?.requireRegisterAuth() ?? false
}

function showGuardedToast(message: string) {
  if (!requireRegisterAuth()) return

  showToast(message)
}

function handleFloatingWidgetTap(widget: PpmxFloatingWidget) {
  if (floating.isDragging.value) return

  if (widget.id === 'checkin') {
    openCheckinModal()
    return
  }

  if (widget.id === 'social') {
    openSocialModal()
    return
  }

  if (widget.id === 'download') {
    goToTarget('download')
    return
  }

  if (!requireRegisterAuth()) return

  showToast(pickText(widget.label))
}

function openSocialModal() {
  socialModalOpen.value = true
}

function closeSocialModal() {
  socialModalOpen.value = false
}

function openSocialChannel(social: PpmxSocialItem) {
  window.open(social.url, '_blank', 'noopener,noreferrer')
}

function checkinDayLabel(day: number) {
  return pickText({
    es: `Día ${day}`,
    en: `Day ${day}`,
    zh: `第${day}天`,
  })
}

function checkinRewardState(index: number) {
  const seq = index + 1

  if (seq <= checkinTotalDays.value) return 'claimed'
  if (!checkinTodayChecked.value && !checkinCompleted.value && seq === checkinNextSeq.value) return 'current'

  return 'locked'
}

function checkinAmountRevealed(index: number) {
  return checkinRewardState(index) === 'claimed'
}

function formatCheckinAmount(value?: number, showCurrency = false) {
  const amount = Number(value ?? 0)
  const formatted = amount.toLocaleString('en-US', {
    maximumFractionDigits: 2,
  })

  return showCurrency ? `+${APP_CURRENCY_SYMBOL}${formatted} ${APP_CURRENCY}` : `+${APP_CURRENCY_SYMBOL}${formatted}`
}

function formatCheckinMysteryAmount(showCurrency = false) {
  return showCurrency ? `+${APP_CURRENCY_SYMBOL}? ${APP_CURRENCY}` : `+${APP_CURRENCY_SYMBOL}?`
}

function formatCheckinTileAmount(value: number, index: number) {
  const isGrand = index === checkinRewards.value.length - 1

  if (!checkinAmountRevealed(index)) return formatCheckinMysteryAmount(isGrand)

  return formatCheckinAmount(value, isGrand)
}

function formatCheckinRecordDate(value: string) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) return value

  return new Intl.DateTimeFormat(undefined, {
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

async function loadCheckinData() {
  checkinLoading.value = true
  checkinFailed.value = false

  try {
    const [status, records] = await Promise.all([
      getCheckInStatus(),
      getCheckInRecords(7),
    ])

    checkinStatus.value = status
    checkinRecords.value = records ?? []
  } catch {
    checkinFailed.value = true
    checkinStatus.value = null
    checkinRecords.value = []
  } finally {
    checkinLoading.value = false
  }
}

function openCheckinModal() {
  if (!requireRegisterAuth()) return

  checkinModalOpen.value = true
  void loadCheckinData()
}

function closeCheckinModal() {
  checkinModalOpen.value = false
}

async function claimCheckin() {
  if (checkinButtonDisabled.value) return

  checkinSubmitting.value = true

  try {
    const result = await doCheckIn()
    showToast(`${pickText(checkinText.success)} ${formatCheckinAmount(result.rewardAmount, true)}`)
    await Promise.allSettled([
      loadCheckinData(),
      userStore.loadUserInfo(),
    ])
  } finally {
    checkinSubmitting.value = false
  }
}

async function openGame(game: PpmxGame) {
  if (!requireRegisterAuth()) return

  if (!game.gameId) {
    showToast(pickText(game.name))
    return
  }

  if (!(await ensureUnlocked())) return

  void router.push({
    path: '/play',
    query: {
      gameId: String(game.gameId),
      title: pickText(game.name),
    },
  })
}

function onUnlockActivate() {
  void router.push('/recharge')
}

function stopHeroAuto() {
  if (heroTimer) window.clearInterval(heroTimer)
}

function startHeroAuto() {
  stopHeroAuto()
  heroTimer = window.setInterval(hero.next, 6000)
}

function stopNoticeAuto() {
  if (noticeTimer) window.clearInterval(noticeTimer)
}

function startNoticeAuto() {
  stopNoticeAuto()
  noticeTimer = window.setInterval(notices.next, 4500)
}

function onHeroTouchStart(event: TouchEvent) {
  touchStartX.value = event.touches[0]?.clientX ?? null
}

function onHeroTouchEnd(event: TouchEvent) {
  const startX = touchStartX.value
  const endX = event.changedTouches[0]?.clientX
  touchStartX.value = null

  if (startX === null || endX === undefined) return

  const delta = endX - startX
  if (Math.abs(delta) < 36) return
  if (delta < 0) hero.next()
  else hero.previous()
}

function onNoticeTouchStart(event: TouchEvent) {
  touchStartX.value = event.touches[0]?.clientX ?? null
}

function onNoticeTouchEnd(event: TouchEvent) {
  const startX = touchStartX.value
  const endX = event.changedTouches[0]?.clientX
  touchStartX.value = null

  if (startX === null || endX === undefined || Math.abs(endX - startX) < 36) return
  notices.next()
}

function closeCurrentPopup() {
  const announcement = currentPopup.value

  if (!announcement) return

  popupQueue.value.shift()
  popupVisible.value = popupQueue.value.length > 0
}

function openPopupTarget(announcement: AppPopupAnnouncement) {
  const target = announcement.jumpValue?.trim()

  if (!target) return

  if (announcement.jumpType === 'internal') {
    if (announcement.openMode === 'new_window') {
      window.open(router.resolve(target).href, '_blank', 'noopener,noreferrer')
      return
    }

    if (announcement.openMode === 'current') {
      void router.push(target)
    }
    return
  }

  if (announcement.jumpType === 'url' && /^https?:\/\//i.test(target)) {
    if (announcement.openMode === 'current') {
      window.location.assign(target)
      return
    }

    if (announcement.openMode === 'new_window') {
      window.open(target, '_blank', 'noopener,noreferrer')
    }
  }
}

function confirmCurrentPopup() {
  const announcement = currentPopup.value

  if (!announcement) return

  closeCurrentPopup()
  openPopupTarget(announcement)
}

async function loadPopupAnnouncements(
  requestVersion: number,
  lang: string,
) {
  try {
    const groups = await getPopupAnnouncements(lang)

    if (requestVersion !== popupRequestVersion) return

    popupQueue.value = filterPopupAnnouncements(groups.popup ?? [])
    popupVisible.value = popupQueue.value.length > 0
  } catch {
    if (requestVersion !== popupRequestVersion) return

    popupQueue.value = []
    popupVisible.value = false
  }
}

onMounted(() => {
  void gameHome.loadGameHome()
  startHeroAuto()
  startNoticeAuto()
  document.addEventListener('pointermove', floating.dragWidget)
  document.addEventListener('pointerup', floating.endWidgetDrag)
  window.addEventListener('resize', floating.resetWidgetsToSafeArea)
  window.addEventListener('scroll', floating.handlePageScroll, { passive: true })
})

watch(
  gameHome.categoryIds,
  (ids) => {
    if (ids.length > 0 && !ids.includes(activeCategoryId.value)) {
      activeCategoryId.value = ids[0]
    }
  },
  { immediate: true },
)

watch(
  () => route.query.action,
  (action) => {
    if (action === 'checkin') {
      void router.replace({ query: { ...route.query, action: undefined } })
      void nextTick(() => openCheckinModal())
    }
  },
  { immediate: true },
)

watch(
  [
    () => userStore.isLogin,
    () => userStore.userInfo?.id,
    () => ppmxLocaleToAppLocale(homeLocale.value),
  ],
  ([isLogin, userId, lang]) => {
    const requestVersion = ++popupRequestVersion

    if (!isLogin || userId === undefined || userId === '') {
      popupQueue.value = []
      popupVisible.value = false
      return
    }

    void loadPopupAnnouncements(requestVersion, lang)
  },
  { immediate: true },
)

onDeactivated(() => {
  popupRequestVersion += 1
  popupQueue.value = []
  popupVisible.value = false
})

onBeforeUnmount(() => {
  popupRequestVersion += 1
  stopHeroAuto()
  stopNoticeAuto()
  document.removeEventListener('pointermove', floating.dragWidget)
  document.removeEventListener('pointerup', floating.endWidgetDrag)
  window.removeEventListener('resize', floating.resetWidgetsToSafeArea)
  window.removeEventListener('scroll', floating.handlePageScroll)
  floating.stopScrollWatcher()
})
</script>

<template>
  <main class="ppmx-page ppmx-home">
    <PpmxPageChrome ref="pageChromeRef">

    <section class="ppmx-container ppmx-home-main">
      <PpmxNoticeBar
        :class="{ 'is-clickable': !!activeNotice?.target }"
        @click="activeNotice?.target && goToTarget(activeNotice.target)"
        @mouseenter="stopNoticeAuto"
        @mouseleave="startNoticeAuto"
        @touchstart.passive="onNoticeTouchStart"
        @touchend.passive="onNoticeTouchEnd"
      >
        <div class="ppmx-notif-ico">
          <i class="fa-solid" :class="activeNotice?.icon" />
        </div>
        <p>{{ activeNotice ? pickText(activeNotice.text) : '' }}</p>
        <span class="ppmx-btn-gold ppmx-notice-cta">
          {{ t('ppmx.home.noticeCta') }}
        </span>
      </PpmxNoticeBar>

      <PpmxHeroCarousel
        @mouseenter="stopHeroAuto"
        @mouseleave="startHeroAuto"
        @touchstart.passive="onHeroTouchStart"
        @touchend.passive="onHeroTouchEnd"
      >
        <article
          v-for="(banner, index) in visibleBanners"
          :key="banner.id"
          class="ppmx-hero-slide"
          :class="{ 'is-active': index === hero.activeIndex.value }"
        >
          <div class="ppmx-hero-bg" :style="{ background: banner.background }" />
          <div class="ppmx-hero-scrim" />
          <div class="ppmx-hero-content">
            <span class="ppmx-hero-eyebrow">
              <i class="fa-solid fa-bolt" />
              {{ pickText(banner.tag) }}
            </span>
            <div v-if="banner.amountHtml" class="ppmx-hero-amount" v-html="pickText(banner.amountHtml)" />
            <p v-if="banner.benefitHtml" class="ppmx-hero-benefit" v-html="pickText(banner.benefitHtml)" />
            <h1 v-if="!banner.amountHtml" class="ppmx-display">{{ pickText(banner.title) }}</h1>
            <p class="ppmx-hero-sub">{{ pickText(banner.sub) }}</p>
            <div class="ppmx-hero-actions">
              <button class="ppmx-btn-gold ppmx-hero-cta" type="button" @click="openPromo(banner.id)">
                {{ pickText(banner.cta) }}
                <i class="fa-solid fa-arrow-right" />
              </button>
              <span v-if="banner.valid">{{ pickText(banner.valid) }}</span>
            </div>
          </div>
          <div v-if="!banner.image" class="ppmx-hero-coins" aria-hidden="true">
            <svg v-for="coin in 3" :key="coin" viewBox="0 0 100 100">
              <use href="#ppCoin" />
            </svg>
          </div>
          <div v-if="!banner.image" class="ppmx-hero-thumbs" aria-hidden="true">
            <span
              v-for="thumb in banner.thumbs"
              :key="thumb"
              class="ppmx-hero-thumb"
              :style="getPpmxGameTileBackground(thumb)"
            />
          </div>
        </article>

        <button class="ppmx-hero-arrow ppmx-hero-arrow--prev" type="button" :aria-label="t('ppmx.home.heroPrevious')" @click="hero.previous()">
          <i class="fa-solid fa-chevron-left" />
        </button>
        <button class="ppmx-hero-arrow ppmx-hero-arrow--next" type="button" :aria-label="t('ppmx.home.heroNext')" @click="hero.next()">
          <i class="fa-solid fa-chevron-right" />
        </button>
        <div class="ppmx-hero-dots" :aria-label="t('ppmx.home.heroSlidesAria')">
          <button
            v-for="banner, index in visibleBanners"
            :key="banner.id"
            class="ppmx-hero-dot"
            :class="{ 'is-active': index === hero.activeIndex.value }"
            type="button"
            :aria-label="t('ppmx.home.heroSlideAria', { number: index + 1 })"
            @click="hero.goTo(index)"
          />
        </div>
      </PpmxHeroCarousel>
    </section>

    <PpmxGameNav>
      <div class="ppmx-container ppmx-game-nav__inner">
        <button
          v-for="category in homeCategories"
          :key="category.id"
          class="ppmx-game-nav__item"
          :class="{ 'is-active': category.id === activeCategoryId }"
          type="button"
          @click="scrollToCategory(category.id)"
        >
          <i class="fa-solid" :class="category.icon" />
          <span>{{ pickText(category.name) }}</span>
        </button>
      </div>
    </PpmxGameNav>

    <section class="ppmx-container ppmx-lobby">
      <PpmxSkeleton
        v-if="gameHome.loading.value"
        variant="home-games"
        :items="12"
        :aria-label="t('ppmx.home.loadingGames')"
      />

      <AppState
        v-else-if="gameHome.empty.value"
        class="ppmx-game-empty"
        status="empty"
        :title="t('ppmx.home.emptyTitle')"
        :message="t('ppmx.home.emptyMessage')"
        :action-text="t('common.retry')"
        @retry="gameHome.loadGameHome"
      />

      <template v-else>
        <PpmxLobbySection
          v-for="category in homeCategories"
          :id="`ppmx-cat-${category.id}`"
          :key="category.id"
          class="ppmx-hcat"
          :data-ppmx-category-id="category.id"
        >
          <div class="ppmx-hcat-head">
            <h2 class="ppmx-hcat-title">
              <i class="fa-solid" :class="category.icon" />
              {{ pickText(category.name) }}
            </h2>
            <button class="ppmx-btn-ghost ppmx-more-button" type="button" @click="openCategoryList(category)">
              {{ t('ppmx.home.more') }}
            </button>
          </div>

          <div class="ppmx-hcat-grid">
            <PpmxGameTile
              v-for="game, gameIndex in category.games"
              :key="game.id"
              @click="openGame(game)"
            >
              <span class="ppmx-htile-thumb" :style="getGameTileStyle(game, gameIndex)">
                <span v-if="game.hot" class="ppmx-htile-badge ppmx-htile-badge--hot">HOT</span>
                <span v-if="game.isNew" class="ppmx-htile-badge ppmx-htile-badge--new">NEW</span>
              </span>
              <span class="ppmx-htile-name">{{ pickText(game.name) }}</span>
              <span v-if="game.subtitle" class="ppmx-htile-provider">{{ game.subtitle }}</span>
            </PpmxGameTile>
          </div>
        </PpmxLobbySection>
      </template>
    </section>

    <PpmxTrustSection>
      <div class="ppmx-trust-grid">
        <article v-for="item in ppmxTrustItems" :key="item.id" class="ppmx-trust-card">
          <i class="fa-solid" :class="item.icon" />
          <strong>{{ pickText(item.title) }}</strong>
          <p>{{ pickText(item.description) }}</p>
        </article>
      </div>

      <RouterLink to="/responsible" class="ppmx-responsible-pledge reveal">
        <i class="fa-solid fa-shield-heart ppmx-responsible-pledge__icon" />
        <div class="ppmx-responsible-pledge__body">
          <strong>{{ t('ppmx.home.pledgeTitle') }}</strong>
          <p>{{ t('ppmx.home.pledgeSub') }}</p>
        </div>
        <span class="ppmx-responsible-hotline">
          <i class="fa-solid fa-chevron-right" />
        </span>
      </RouterLink>

      <footer id="ppmx-footer" class="ppmx-footer">
        <div class="ppmx-footer__inner">
          <div class="ppmx-footer-brand">
            <PpmxLogo />
            <p>{{ t('ppmx.home.footerDescription') }}</p>
            <div class="ppmx-footer-socials" :aria-label="t('ppmx.home.footerSocialAria')">
              <button
                v-for="social in ppmxSocialItems"
                :key="social.id"
                type="button"
                :aria-label="social.label"
                @click="openSocialChannel(social)"
              >
                <i class="fa-brands" :class="social.icon" />
              </button>
            </div>
          </div>

          <nav class="ppmx-footer-columns" :aria-label="t('ppmx.home.footerAria')">
            <section v-for="column in ppmxFooterColumns" :key="column.id">
              <h4>{{ pickText(column.title) }}</h4>
              <ul>
                <li v-for="link in column.links" :key="link.id">
                  <button
                    type="button"
                    @click="goToTarget(link.target)"
                  >
                    {{ pickText(link.label) }}
                  </button>
                </li>
              </ul>
            </section>
          </nav>
        </div>

        <div class="ppmx-footer-legal">
          <div class="ppmx-footer-legal__inner">
            <p>{{ t('ppmx.home.footerLegal') }}</p>
          </div>
        </div>

        <div class="ppmx-footer-bottom">
          <div class="ppmx-footer-bottom__inner">
            <div>© 2026 PP.PE · <span>{{ t('ppmx.home.footerRights') }}</span></div>
            <div class="ppmx-footer-badges">
              <span><strong>+18</strong></span>
              <span>Responsible Gaming</span>
              <span>PP.PE SUPPORT</span>
            </div>
          </div>
        </div>
      </footer>
    </PpmxTrustSection>

    <PpmxFloatingWidgets :hiding="floating.isScrollHiding.value">
      <div
        v-for="(widget, index) in visibleFloatingWidgets"
        :key="widget.id"
        class="ppmx-float-widget"
        :class="`ppmx-float-widget--${widget.id}`"
        :style="floating.getWidgetStyle(widget.id, index)"
      >
        <button class="ppmx-float-close" type="button" :aria-label="t('ppmx.home.closeShortcut')" @click="floating.hideWidget(widget.id)">
          <i class="fa-solid fa-xmark" />
        </button>
        <button
          class="ppmx-float-inner"
          type="button"
          @click="handleFloatingWidgetTap(widget)"
          @pointerdown="floating.startWidgetDrag(widget.id, $event, index)"
        >
          <i class="fa-solid" :class="widget.icon" />
        </button>
        <span>{{ pickText(widget.label) }}</span>
      </div>
    </PpmxFloatingWidgets>

    <Transition name="ppmx-social-fade">
      <div
        v-if="socialModalOpen"
        id="socialModal"
        class="ppmx-social-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="socialModalTitle"
      >
        <button
          class="ppmx-social-modal__backdrop"
          type="button"
          :aria-label="pickText(socialText.close)"
          @click="closeSocialModal"
        />

        <section class="ppmx-social-card-wrap" role="document">
          <button
            class="ppmx-social-close"
            type="button"
            :aria-label="pickText(socialText.close)"
            @click="closeSocialModal"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <div class="ppmx-social-icon" aria-hidden="true">
            <i class="fa-solid fa-share-nodes" />
          </div>
          <h2 id="socialModalTitle">{{ pickText(socialText.title) }}</h2>
          <p>{{ pickText(socialText.sub) }}</p>

          <div class="ppmx-social-grid">
            <button
              v-for="social in ppmxSocialItems"
              :key="social.id"
              class="ppmx-social-card"
              :class="`is-${social.tone}`"
              type="button"
              :aria-label="`${pickText(socialText.open)} ${social.label}`"
              @click="openSocialChannel(social)"
            >
              <span class="ppmx-social-mark" aria-hidden="true">
                <i class="fa-brands" :class="social.icon" />
              </span>
              <strong>{{ social.label }}</strong>
              <small>{{ social.handle }}</small>
            </button>
          </div>
        </section>
      </div>
    </Transition>

    <Transition name="ppmx-checkin-fade">
      <div
        v-if="checkinModalOpen"
        id="checkinModal"
        class="ppmx-checkin-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="checkinModalTitle"
      >
        <button
          class="ppmx-checkin-modal__backdrop"
          type="button"
          :aria-label="pickText(checkinText.close)"
          @click="closeCheckinModal"
        />

        <section class="ppmx-checkin-card" role="document">
          <span class="ppmx-checkin-glow" aria-hidden="true" />
          <button
            class="ppmx-checkin-close"
            type="button"
            :aria-label="pickText(checkinText.close)"
            @click="closeCheckinModal"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <div class="ppmx-checkin-body">
            <div class="ppmx-checkin-icon" aria-hidden="true">
              <i class="fa-solid fa-calendar-check" />
            </div>
            <h2 id="checkinModalTitle">{{ pickText(checkinText.title) }}</h2>
            <p>{{ pickText(checkinText.sub) }}</p>

            <div
              v-if="checkinStatus && checkinRewards.length"
              class="ppmx-checkin-summary"
              :style="checkinProgressStyle"
            >
              <div class="ppmx-checkin-summary__reward">
                <span>{{ checkinSummaryLabel }}</span>
                <strong>{{ checkinSummaryAmount }}</strong>
              </div>
              <div class="ppmx-checkin-summary__meta">
                <span>{{ checkinDayLabel(checkinFocusSeq || 1) }}</span>
                <b>{{ checkinSummaryState }}</b>
              </div>
              <div class="ppmx-checkin-progress" aria-hidden="true">
                <span />
              </div>
            </div>

            <PpmxSkeleton
              v-if="checkinLoading && !checkinStatus"
              variant="checkin-modal"
              :items="7"
              compact
              :aria-label="pickText(checkinText.loading)"
            />

            <div v-else-if="checkinFailed" class="ppmx-checkin-state is-error">
              <i class="fa-solid fa-triangle-exclamation" />
              <span>{{ pickText(checkinText.failed) }}</span>
              <button class="ppmx-btn-gold" type="button" @click="loadCheckinData">
                {{ pickText(checkinText.retry) }}
              </button>
            </div>

            <div v-else-if="checkinRewards.length === 0" class="ppmx-checkin-state">
              <i class="fa-regular fa-calendar-xmark" />
              <span>{{ pickText(checkinText.empty) }}</span>
            </div>

            <template v-else>
              <div id="checkinGrid" class="ppmx-checkin-grid">
                <article
                  v-for="(amount, index) in checkinRewards"
                  :key="`${index}-${amount}`"
                  class="ppmx-checkin-tile"
                  :class="[
                    `is-${checkinRewardState(index)}`,
                    { 'is-grand': index === checkinRewards.length - 1 },
                  ]"
                >
                  <span class="ppmx-checkin-day">
                    {{ checkinDayLabel(index + 1) }}
                    <small v-if="index === checkinRewards.length - 1">{{ pickText(checkinText.grand) }}</small>
                  </span>
                  <span class="ppmx-checkin-coin" aria-hidden="true">
                    <i class="fa-solid" :class="index === checkinRewards.length - 1 ? 'fa-trophy' : 'fa-coins'" />
                  </span>
                  <strong :class="{ 'is-mystery': !checkinAmountRevealed(index) }">
                    {{ formatCheckinTileAmount(amount, index) }}
                  </strong>
                  <i v-if="checkinRewardState(index) === 'claimed'" class="fa-solid fa-check ppmx-checkin-mark" />
                  <i v-else-if="checkinRewardState(index) === 'locked'" class="fa-solid fa-lock ppmx-checkin-mark" />
                </article>
              </div>

              <button
                id="checkinBtn"
                class="ppmx-btn-gold ppmx-checkin-submit"
                type="button"
                :disabled="checkinButtonDisabled"
                @click="claimCheckin"
              >
                <i v-if="checkinSubmitting" class="fa-solid fa-circle-notch fa-spin" />
                <span>{{ checkinButtonText }}</span>
              </button>

              <div v-if="checkinRecentRecords.length" class="ppmx-checkin-records">
                <strong>{{ pickText(checkinText.records) }}</strong>
                <span
                  v-for="record in checkinRecentRecords"
                  :key="record.id"
                >
                  {{ formatCheckinRecordDate(record.checkInDate || record.createdAt) }}
                  <b>{{ formatCheckinAmount(record.rewardAmount, true) }}</b>
                </span>
              </div>
            </template>
          </div>
        </section>
      </div>
    </Transition>

    </PpmxPageChrome>

    <PpmxAnnouncementPopup
      v-if="currentPopup"
      :announcement="currentPopup"
      :close-text="t('common.close')"
      :confirm-text="t('common.confirm')"
      :model-value="popupVisible"
      @close="closeCurrentPopup"
      @confirm="confirmCurrentPopup"
    />

    <PpmxUnlockModal v-model="unlockOpen" @activate="onUnlockActivate" />
  </main>
</template>
