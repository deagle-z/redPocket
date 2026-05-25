<script setup lang="ts">
import { showToast } from 'vant'
import type { AppHomeGameData, AppHomeGameItem, BannerItem, LuckyHistoryUserFlowItem } from '@/api/user'
import { getAppGameHome, getBanners, getLuckyHistoryUserFlow } from '@/api/user'
import { formatCurrency } from '@/utils/currency'
import { getTokenUserId, isLogin } from '@/utils/auth'
import imgAvatarPlaceholder from '@/assets/images/avatar-placeholder.png'
import imgRedpacketGif from '@/assets/images/redpacket.gif'
import imgRedpacketJpg from '@/assets/images/redpacket.jpg'
import slotsTabIcon from '@/assets/images/game_tabs/slots.webp'
import casinoTabIcon from '@/assets/images/game_tabs/vivo.webp'
import blockchainTabIcon from '@/assets/images/game_tabs/blockchain.webp'
import fishingTabIcon from '@/assets/images/game_tabs/fishing.webp'
import sportsTabIcon from '@/assets/images/game_tabs/sports.webp'
import miniTabIcon from '@/assets/images/game_tabs/mini.webp'
import lotteryTabIcon from '@/assets/images/game_tabs/lottery.webp'

const { t } = useI18n()
const router = useRouter()

const ANNOUNCEMENT_SEEN_PREFIX = 'announcement_popup_seen'
const PACKET_GAME_PAGE_SIZE = 20

const DEFAULT_AVATAR = imgAvatarPlaceholder
const activeIndex = ref(0)
const packetSectionVisibleCounts = ref<Record<string, number>>({})
const activePacketSectionKey = ref('')
const packetGamesLoading = ref(false)

type PacketGameAction = 'game' | 'packet'

interface PacketGameCard {
  title: string
  subtitle: string
  brand: string
  cover?: string
  variant: 'game' | 'packet'
  action: PacketGameAction
  packetMode?: 0 | 1
  rawGame?: AppHomeGameItem
}

interface PacketGameSection {
  label: string
  icon?: string
  iconText?: string
  moreAction: 'packet' | 'lottery'
  games: PacketGameCard[]
}

const packetGameHomeData = ref<AppHomeGameData>({})

const packetSectionMeta = [
  { label: 'Popular', code: 'hot', iconText: '🔥', moreAction: 'packet' as const },
  { label: 'Slots', code: 'slots', icon: slotsTabIcon, moreAction: 'packet' as const },
  { label: 'Casino', code: 'casino', icon: casinoTabIcon, moreAction: 'packet' as const },
  { label: 'Blockchain', code: 'blockchain', icon: blockchainTabIcon, moreAction: 'packet' as const },
  { label: 'Fishing', code: 'fishing', icon: fishingTabIcon, moreAction: 'packet' as const },
  { label: 'Rummy', code: 'rummy', iconText: '🎲', moreAction: 'packet' as const },
  { label: 'Sports', code: 'sports', icon: sportsTabIcon, moreAction: 'packet' as const },
  { label: 'mini', code: 'mini', icon: miniTabIcon, moreAction: 'packet' as const },
  { label: 'Lottery', code: 'lottery', icon: lotteryTabIcon, moreAction: 'lottery' as const },
]

const defaultHotPacketGames = computed<PacketGameCard[]>(() => [
  {
    title: t('homeLucky.playTypeThunder'),
    subtitle: t('homeLucky.game'),
    brand: t('sendPacketPage.playTypeThunderEyebrow'),
    cover: imgRedpacketGif,
    variant: 'packet',
    action: 'packet',
    packetMode: 0,
  },
  {
    title: t('homeLucky.playTypeParity'),
    subtitle: t('homeLucky.game'),
    brand: t('sendPacketPage.playTypeParityEyebrow'),
    cover: imgRedpacketJpg,
    variant: 'packet',
    action: 'packet',
    packetMode: 1,
  },
])

function mapHomeGame(game: AppHomeGameItem): PacketGameCard {
  return {
    title: game.gameName,
    subtitle: game.manufacturer?.toUpperCase() || game.categoryCode,
    brand: game.manufacturer?.toUpperCase() || game.categoryCode?.toUpperCase() || 'GAME',
    cover: game.gameIcon || game.horizontalImage,
    variant: 'game',
    action: 'game',
    rawGame: game,
  }
}

const packetGameSections = computed<PacketGameSection[]>(() => [
  ...packetSectionMeta.map(section => ({
    label: section.label,
    icon: section.icon,
    iconText: section.iconText,
    moreAction: section.moreAction,
    games: [
      ...(section.code === 'hot' ? defaultHotPacketGames.value : []),
      ...(packetGameHomeData.value[section.code] || []).map(mapHomeGame),
    ],
  })),
].filter(section => section.games.length > 0))

const displayPacketGameSections = computed<PacketGameSection[]>(() => {
  return packetGameSections.value
})

function getPacketSectionKey(section: PacketGameSection) {
  return section.label
}

function getPacketSectionDomId(section: PacketGameSection) {
  return `packet-game-section-${getPacketSectionKey(section).replace(/\s+/g, '-').toLowerCase()}`
}

function getPacketSectionVisibleCount(section: PacketGameSection) {
  const key = getPacketSectionKey(section)
  return packetSectionVisibleCounts.value[key] ?? PACKET_GAME_PAGE_SIZE
}

function visiblePacketGames(section: PacketGameSection) {
  return section.games.slice(0, getPacketSectionVisibleCount(section))
}

function hasMorePacketGames(section: PacketGameSection) {
  return getPacketSectionVisibleCount(section) < section.games.length
}

function loadMorePacketGames(section: PacketGameSection) {
  const key = getPacketSectionKey(section)
  const nextCount = Math.min(getPacketSectionVisibleCount(section) + PACKET_GAME_PAGE_SIZE, section.games.length)
  packetSectionVisibleCounts.value = {
    ...packetSectionVisibleCounts.value,
    [key]: nextCount,
  }
}

const homeBanners = ref<BannerItem[]>([])
const popupQueue = ref<BannerItem[]>([])
const popupVisible = ref(false)
const popupIndex = ref(0)

const currentPopup = computed(() => popupQueue.value[popupIndex.value] ?? null)
const recentWinnersLoading = ref(false)
const recentWinners = ref<any[]>([])

const visibleWinners = computed(() => recentWinners.value)
const showWinnerLoading = computed(() => recentWinnersLoading.value && visibleWinners.value.length === 0)
const showWinnerEmpty = computed(() => visibleWinners.value.length === 0)
const marqueeText = computed(() => getBannerTitle(homeBanners.value[activeIndex.value]))

function onSwipeChange(index: number) {
  activeIndex.value = index
}

function getBannerTitle(banner?: BannerItem | null) {
  return banner?.title || banner?.bannerName || ''
}

function goPacketListDirectly(mode: 0 | 1) {
  router.push({
    path: '/packetList',
    query: { mode: String(mode) },
  })
}

function goPacketList(mode: 0 | 1) {
  goPacketListDirectly(mode)
}

function goPrize() {
  router.push('/prize')
}

function goPacketSection(section: PacketGameSection) {
  if (section.moreAction === 'lottery') {
    goPrize()
    return
  }
  goPacketList(0)
}

function scrollToPacketSection(section: PacketGameSection) {
  const key = getPacketSectionKey(section)
  activePacketSectionKey.value = key
  const target = document.getElementById(getPacketSectionDomId(section))
  if (!target)
    return
  const top = target.getBoundingClientRect().top + window.scrollY - 72
  window.scrollTo({
    top,
    behavior: 'smooth',
  })
}

function playPacketGame(game: PacketGameCard) {
  if (game.action === 'packet') {
    goPacketList(game.packetMode ?? 0)
    return
  }

  const gameId = game.rawGame?.gameId
  if (!gameId)
    return
  router.push({
    path: '/gamePlay',
    query: {
      gameId: String(gameId),
      title: game.title,
    },
  })
}

async function loadAppGames() {
  try {
    packetGamesLoading.value = true
    const { data } = await getAppGameHome()
    packetGameHomeData.value = data || {}
    const firstSection = displayPacketGameSections.value[0]
    if (firstSection && !displayPacketGameSections.value.some(section => getPacketSectionKey(section) === activePacketSectionKey.value))
      activePacketSectionKey.value = getPacketSectionKey(firstSection)
  }
  catch {
    packetGameHomeData.value = {}
  }
  finally {
    packetGamesLoading.value = false
  }
}

async function loadRecentWinners() {
  if (recentWinnersLoading.value)
    return
  try {
    recentWinnersLoading.value = true
    const { data } = await getLuckyHistoryUserFlow({
      currentPage: 0,
      pageSize: 20,
    })
    recentWinners.value = (data?.list || []).map((item: LuckyHistoryUserFlowItem, index: number) => ({
      id: Number(item.userId || index),
      avatar: item.avatar || DEFAULT_AVATAR,
      amount: formatCurrency(Number(item.flowAmount || 0)),
      name: item.firstName || 'User',
      time: t('homeLucky.timeJustNow'),
    }))
  }
  catch {
    showToast(t('homeLucky.winnerLoadFailed'))
  }
  finally {
    recentWinnersLoading.value = false
  }
}

function getTodayKey() {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

function getSeenStorageKey() {
  return `${ANNOUNCEMENT_SEEN_PREFIX}:${getTokenUserId() || 0}:${getTodayKey()}`
}

function getSeenIds(): number[] {
  try {
    return JSON.parse(localStorage.getItem(getSeenStorageKey()) || '[]')
  }
  catch {
    return []
  }
}

function markPopupSeen(id: number) {
  const ids = getSeenIds()
  if (!ids.includes(id)) {
    ids.push(id)
    localStorage.setItem(getSeenStorageKey(), JSON.stringify(ids))
  }
}

async function loadBanners() {
  try {
    const { data } = await getBanners()
    homeBanners.value = (data?.home ?? []).filter(b => b.status === 1 && b.imageUrl)
    if (!isLogin()) {
      popupQueue.value = []
      popupVisible.value = false
      return
    }
    const seen = getSeenIds()
    popupQueue.value = (data?.popup ?? [])
      .filter(b => b.status === 1 && !seen.includes(b.id))
      .sort((a, b) => a.sort - b.sort)
    if (popupQueue.value.length > 0) {
      popupIndex.value = 0
      popupVisible.value = true
    }
  }
  catch { /* silent — carousel stays empty */ }
}

function onBannerClick(banner: BannerItem) {
  if (!banner.jumpValue)
    return

  if (banner.jumpType === 'internal') {
    router.push(banner.jumpValue)
    return
  }

  if (banner.jumpType === 'url')
    window.open(banner.jumpValue, '_blank')
}

function onPopupOk() {
  const item = currentPopup.value
  if (item)
    markPopupSeen(item.id)
  if (popupIndex.value + 1 < popupQueue.value.length) {
    popupIndex.value++
  }
  else {
    popupVisible.value = false
  }
}

function onPopupDismiss() {
  const item = currentPopup.value
  if (item)
    markPopupSeen(item.id)
  popupQueue.value.splice(popupIndex.value, 1)
  if (popupQueue.value.length === 0) {
    popupVisible.value = false
  }
  else {
    popupIndex.value = Math.min(popupIndex.value, popupQueue.value.length - 1)
  }
}

onMounted(async () => {
  void loadAppGames()
  void loadBanners()
  void loadRecentWinners()
})
</script>

<template>
  <div class="home-page">
    <section class="home-carousel-card">
      <van-swipe class="home-swipe" :autoplay="3200" lazy-render indicator-color="#d4af37" @change="onSwipeChange">
        <van-swipe-item v-for="item in homeBanners" :key="item.id">
          <img
            :src="item.imageUrl"
            class="banner-image"
            :alt="getBannerTitle(item)"
            :style="item.jumpValue ? 'cursor:pointer' : ''"
            @click="onBannerClick(item)"
          >
        </van-swipe-item>
      </van-swipe>
      <div class="banner-stripe" />
      <van-notice-bar class="home-marquee" :scrollable="true" :text="marqueeText" />
    </section>

    <section class="packet-entry-card">
      <div v-if="packetGamesLoading" class="packet-game-skeleton">
        <div class="packet-game-skeleton__tabs">
          <van-skeleton
            v-for="item in 5"
            :key="`packet-tab-skeleton-${item}`"
            class="packet-game-skeleton__tab"
            title
            :row="0"
          />
        </div>
        <van-skeleton class="packet-game-skeleton__title" title :row="0" />
        <div class="packet-game-skeleton__grid">
          <van-skeleton
            v-for="item in 8"
            :key="`packet-game-skeleton-${item}`"
            class="packet-game-skeleton__card"
            title
            :row="2"
          />
        </div>
      </div>

      <template v-else-if="displayPacketGameSections.length > 0">
        <div class="packet-section-tabs" aria-label="Game categories">
          <button
            v-for="section in displayPacketGameSections"
            :key="`tab-${section.label}`"
            type="button"
            class="packet-section-tab"
            :class="{ 'packet-section-tab--active': activePacketSectionKey === getPacketSectionKey(section) }"
            @click="scrollToPacketSection(section)"
          >
            <img
              v-if="section.icon"
              :src="section.icon"
              class="packet-section-tab__img"
              alt=""
            >
            <span v-else class="packet-section-tab__emoji">{{ section.iconText }}</span>
            <span>{{ section.label }}</span>
          </button>
        </div>

        <div
          v-for="section in displayPacketGameSections"
          :id="getPacketSectionDomId(section)"
          :key="section.label"
          class="packet-game-section"
        >
          <div class="packet-section-heading">
            <div class="packet-section-heading__title">
              <img
                v-if="section.icon"
                :src="section.icon"
                class="packet-section-heading__img"
                alt=""
              >
              <span v-else class="packet-section-heading__emoji">{{ section.iconText }}</span>
              <p>{{ section.label }}</p>
            </div>
            <button type="button" class="packet-section-heading__more" @click="goPacketSection(section)">
              Todo
              <van-icon name="arrow" />
            </button>
          </div>

          <div class="packet-game-grid">
            <button
              v-for="(game, gameIndex) in visiblePacketGames(section)"
              :key="`${section.label}-${game.rawGame?.gameId || gameIndex}`"
              type="button"
              class="packet-game-card"
              :class="`packet-game-card--${game.variant}`"
              @click="playPacketGame(game)"
            >
              <span class="packet-game-card__brand">{{ game.brand }}</span>
              <img
                v-if="game.cover"
                :src="game.cover"
                class="packet-game-card__cover"
                :alt="game.title"
              >
              <strong>{{ game.title }}</strong>
              <small>{{ game.subtitle }}</small>
            </button>
          </div>

          <div v-if="hasMorePacketGames(section)" class="packet-game-pager">
            <p>{{ getPacketSectionVisibleCount(section) }}/{{ section.games.length }} {{ section.label }} games</p>
            <button type="button" class="packet-game-pager__more" @click="loadMorePacketGames(section)">
              Carregue mais
              <van-icon name="arrow-down" />
            </button>
          </div>
        </div>
      </template>
    </section>

    <section class="winner-section">
      <header class="packet-header">
        <div class="section-divider" />
        <div class="packet-title-wrap">
          <van-icon name="trophy-o" />
          <span>{{ t('homeLucky.latestWinners') }}</span>
        </div>
      </header>

      <div v-if="showWinnerLoading" class="winner-skeleton-card">
        <article v-for="idx in 4" :key="`winner-skeleton-${idx}`" class="winner-item">
          <van-skeleton-avatar avatar-size="44px" />
          <div class="winner-main">
            <van-skeleton title :row="1" />
          </div>
        </article>
      </div>

      <AppEmpty v-else-if="showWinnerEmpty" :text="t('homeLucky.emptyWinners')" :min-height="120" />

      <template v-else>
        <div class="winner-card">
          <article v-for="item in visibleWinners" :key="item.id" class="winner-item">
            <img :src="item.avatar" alt="" class="winner-avatar">
            <div class="winner-main">
              <p class="winner-amount">
                {{ item.name }} <strong><CoinAmount :text="item.amount" class="coin-amount--winner" /></strong>
              </p>
              <p class="winner-name">
                {{ t('homeLucky.gotPrefix') }}
              </p>
            </div>
          </article>
        </div>
      </template>
    </section>

    <!-- 弹窗广告 -->
    <van-overlay :show="popupVisible" class="banner-popup-overlay" @click.self="onPopupOk">
      <div class="banner-popup">
        <img
          v-if="currentPopup?.imageUrl"
          :src="currentPopup.imageUrl"
          class="banner-popup__img"
          :alt="getBannerTitle(currentPopup)"
        >
        <div v-if="currentPopup" class="banner-popup__content">
          <h3 v-if="getBannerTitle(currentPopup)" class="banner-popup__title">
            {{ getBannerTitle(currentPopup) }}
          </h3>
          <p v-if="currentPopup.description" class="banner-popup__desc">
            {{ currentPopup.description }}
          </p>
        </div>
        <div class="banner-popup__actions">
          <button type="button" class="popup-btn popup-btn--dismiss" @click="onPopupDismiss">
            {{ t('common.noRemind') }}
          </button>
          <button type="button" class="popup-btn popup-btn--ok" @click="onPopupOk">
            {{ currentPopup?.buttonText || t('common.ok') }}
          </button>
        </div>
      </div>
    </van-overlay>
  </div>
</template>

<style scoped>
.home-page {
  box-sizing: border-box;
  min-height: 100vh;
  background-image:
    radial-gradient(circle at 20% 10%, rgba(212, 175, 55, 0.18), transparent 30%),
    radial-gradient(circle at 80% 90%, rgba(255, 215, 0, 0.12), transparent 28%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #230000 62%, #160000 100%);
  padding: 12px 12px 90px;
  width: 100%;
  max-width: 100%;
  overflow: auto;
  color: #f9e8c6;
}

.home-carousel-card,
.packet-entry-card,
.winner-card,
.winner-skeleton-card {
  position: relative;
  overflow: hidden;
}

.home-carousel-card {
  border-radius: 14px;
  background: #5a0000;
  aspect-ratio: 16 / 6;
  min-height: 132px;
}

.home-swipe {
  width: 100%;
  height: 100%;
  background:
    linear-gradient(90deg, rgba(255, 248, 214, 0.04), rgba(255, 248, 214, 0.1), rgba(255, 248, 214, 0.04)), #5a0000;
}

:deep(.home-swipe .van-swipe__track),
:deep(.home-swipe .van-swipe-item) {
  height: 100%;
}

.banner-image {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.banner-stripe {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 16px,
    rgba(212, 175, 55, 0.05) 16px,
    rgba(212, 175, 55, 0.05) 18px
  );
  z-index: 1;
}

:deep(.home-marquee.van-notice-bar) {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 26px;
  min-height: 26px;
  background: linear-gradient(180deg, rgba(133, 0, 0, 0) 0%, rgba(80, 0, 0, 0.96) 100%) !important;
  color: #ffd98b;
  padding: 0 10px;
  font-size: 11px;
  line-height: 26px;
  z-index: 2;
}

.packet-entry-card {
  box-sizing: border-box;
  width: 100%;
  max-width: 100%;
  margin-top: 14px;
  padding: 14px 0 2px;
}

.packet-section-tabs {
  position: sticky;
  top: 0;
  z-index: 8;
  display: flex;
  gap: 8px;
  width: 100%;
  margin: -2px 0 14px;
  padding: 0 0 8px;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  scrollbar-width: none;
  background: transparent;
}

.packet-section-tabs::-webkit-scrollbar {
  display: none;
}

.packet-section-tab {
  flex: 0 0 72px;
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  min-width: 0;
  height: 56px;
  padding: 6px 6px 5px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(117, 22, 14, 0.72), rgba(64, 0, 0, 0.82));
  color: rgba(255, 232, 194, 0.78);
  box-shadow:
    inset 0 1px 0 rgba(255, 237, 180, 0.08),
    0 4px 10px rgba(0, 0, 0, 0.16);
}

.packet-section-tab--active {
  border-color: rgba(255, 237, 172, 0.78);
  color: #3a0800;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.34), transparent 48%),
    linear-gradient(135deg, #ffe78a 0%, #ffb300 54%, #d57200 100%);
  box-shadow:
    0 6px 14px rgba(122, 30, 0, 0.32),
    inset 0 1px 0 rgba(255, 255, 255, 0.54);
}

.packet-section-tab__img,
.packet-section-tab__emoji {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
}

.packet-section-tab__img {
  object-fit: contain;
  filter: drop-shadow(0 3px 6px rgba(0, 0, 0, 0.28));
}

.packet-section-tab__emoji {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 23px;
  line-height: 1;
}

.packet-section-tab span:last-child {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  font-size: 12px;
  line-height: 14px;
  font-weight: 700;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.packet-game-skeleton {
  padding-bottom: 18px;
}

.packet-game-skeleton__tabs {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 18px;
}

.packet-game-skeleton__tab,
.packet-game-skeleton__card,
.packet-game-skeleton__title {
  --van-skeleton-paragraph-background: rgba(255, 232, 194, 0.12);
  --van-skeleton-title-background: rgba(255, 232, 194, 0.16);
}

.packet-game-skeleton__tab {
  height: 56px;
  padding: 8px 6px;
  border: 1px solid rgba(212, 175, 55, 0.14);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(117, 22, 14, 0.5), rgba(64, 0, 0, 0.62));
}

.packet-game-skeleton__title {
  width: 42%;
  margin-bottom: 12px;
}

.packet-game-skeleton__grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px 8px;
}

.packet-game-skeleton__card {
  min-height: 132px;
  padding: 12px 8px;
  border: 1px solid rgba(255, 218, 127, 0.16);
  border-radius: 12px;
  background: linear-gradient(165deg, rgba(127, 27, 16, 0.56) 0%, rgba(50, 0, 0, 0.72) 100%);
}

.packet-game-section {
  position: relative;
  width: 100%;
  min-width: 0;
  scroll-margin-top: 72px;
  padding: 0 0 18px;
}

.packet-game-section + .packet-game-section {
  padding-top: 16px;
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.packet-section-heading {
  min-height: 38px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.packet-section-heading__title {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 9px;
}

.packet-section-heading__title p {
  margin: 0;
  overflow: hidden;
  color: #ffe7a3;
  font-size: 20px;
  line-height: 1.1;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-shadow: 0 2px 8px rgba(84, 0, 0, 0.45);
}

.packet-section-heading__img,
.packet-section-heading__emoji {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
}

.packet-section-heading__img {
  object-fit: contain;
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.32));
}

.packet-section-heading__emoji {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 25px;
  line-height: 1;
}

.packet-section-heading__more {
  flex-shrink: 0;
  min-width: 72px;
  height: 36px;
  padding: 0 12px;
  border: 1px solid rgba(255, 237, 172, 0.64);
  border-radius: 10px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.32), transparent 44%),
    linear-gradient(135deg, #ffe78a 0%, #ffb300 52%, #d57200 100%);
  color: #3a0800;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  font-size: 15px;
  line-height: 1;
  font-weight: 800;
  box-shadow:
    0 8px 18px rgba(122, 30, 0, 0.32),
    inset 0 1px 0 rgba(255, 255, 255, 0.58);
}

.packet-game-grid {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px 8px;
  align-items: stretch;
  overflow: hidden;
}

.packet-game-card {
  box-sizing: border-box;
  width: 100%;
  position: relative;
  min-width: 0;
  aspect-ratio: 0.72;
  min-height: 132px;
  padding: 7px 5px 9px;
  border: 1px solid rgba(255, 218, 127, 0.34);
  border-radius: 12px;
  overflow: hidden;
  color: #fff7e8;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-end;
  text-align: left;
  box-shadow:
    0 10px 18px rgba(0, 0, 0, 0.26),
    inset 0 1px 0 rgba(255, 248, 214, 0.18);
}

.packet-game-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 50% 24%, rgba(255, 248, 214, 0.3), transparent 34%),
    linear-gradient(180deg, rgba(0, 0, 0, 0) 44%, rgba(45, 0, 0, 0.45) 100%);
  z-index: 1;
  pointer-events: none;
}

.packet-game-card--game {
  background: linear-gradient(165deg, #7f1b10 0%, #320000 100%);
}

.packet-game-card--packet {
  border-color: rgba(255, 225, 128, 0.58);
  background:
    radial-gradient(circle at 50% 20%, rgba(255, 231, 150, 0.22), transparent 34%),
    linear-gradient(165deg, #9a1e10 0%, #4a0000 100%);
}

.packet-game-card--game::before {
  background:
    linear-gradient(180deg, rgba(0, 0, 0, 0) 38%, rgba(49, 0, 0, 0.72) 100%),
    radial-gradient(circle at 50% 8%, rgba(255, 226, 151, 0.2), transparent 36%);
}

.packet-game-card--packet::before {
  background:
    radial-gradient(circle at 50% 22%, rgba(255, 244, 186, 0.38), transparent 32%),
    linear-gradient(180deg, rgba(0, 0, 0, 0) 42%, rgba(67, 0, 0, 0.78) 100%);
}

.packet-game-card__brand {
  position: absolute;
  top: 6px;
  left: 6px;
  z-index: 2;
  padding: 1px 4px;
  border-radius: 4px;
  background: rgba(61, 0, 0, 0.46);
  color: rgba(255, 237, 183, 0.9);
  font-size: 9px;
  line-height: 1.2;
  font-weight: 900;
}

.packet-game-card__cover {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.packet-game-card strong,
.packet-game-card small {
  position: relative;
  z-index: 2;
  width: 100%;
  overflow: hidden;
  display: block;
  text-align: center;
  text-shadow: 0 1px 4px rgba(70, 0, 0, 0.58);
}

.packet-game-card strong {
  font-size: 13px;
  line-height: 1.15;
  font-weight: 900;
}

.packet-game-card small {
  margin-top: 3px;
  color: rgba(255, 236, 191, 0.88);
  font-size: 10px;
  line-height: 1.15;
}

.packet-game-card:active {
  transform: translateY(1px) scale(0.98);
}

.packet-game-pager {
  margin-top: 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7px;
  color: rgba(255, 232, 193, 0.78);
  text-align: center;
}

.packet-game-pager p {
  margin: 0;
  font-size: 14px;
  line-height: 1.2;
  font-weight: 700;
}

.packet-game-pager__more {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  color: rgba(255, 232, 193, 0.82);
  background: rgba(84, 0, 0, 0.34);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  font-size: 14px;
  line-height: 1;
}

.packet-game-pager__more :deep(.van-icon) {
  font-size: 13px;
}

@media (max-width: 360px) {
  .packet-section-heading__title p {
    font-size: 18px;
  }

  .packet-game-grid {
    gap: 8px 6px;
  }

  .packet-game-card {
    min-height: 118px;
    border-radius: 10px;
  }

  .packet-game-card strong {
    font-size: 11px;
  }

  .packet-game-card small {
    font-size: 9px;
  }
}

.winner-section {
  margin-top: 14px;
}

.packet-header {
  margin-bottom: 10px;
}

.section-divider {
  height: 1px;
  margin-bottom: 10px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(212, 175, 55, 0.6) 20%,
    rgba(212, 175, 55, 0.9) 50%,
    rgba(212, 175, 55, 0.6) 80%,
    transparent 100%
  );
}

.packet-title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #ffd98b;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.winner-card {
  border-radius: 16px;
  background: linear-gradient(170deg, rgba(125, 0, 0, 0.97), rgba(60, 0, 0, 0.97));
}

.winner-skeleton-card {
  border-radius: 14px;
  background: linear-gradient(170deg, rgba(116, 0, 0, 0.95), rgba(68, 0, 0, 0.95));
}

.winner-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 13px 14px;
}

.winner-item + .winner-item {
  border-top: 1px solid rgba(212, 175, 55, 0.15);
}

.winner-avatar {
  width: 46px;
  height: 46px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(212, 175, 55, 0.7);
}

.winner-main {
  min-width: 0;
  flex: 1;
}

.winner-amount {
  margin: 0;
  color: #ffefca;
  font-size: 16px;
  line-height: 1.3;
  display: flex;
  align-items: center;
  gap: 8px;
}

.winner-amount strong {
  margin-left: auto;
  flex-shrink: 0;
}

.winner-amount :deep(.coin-amount--winner.coin-amount-wrap) {
  justify-content: flex-end;
}

.winner-name {
  margin: 6px 0 0;
  color: rgba(255, 229, 186, 0.68);
  font-size: 14px;
  line-height: 1;
}

/* ── 弹窗广告 ── */
:deep(.banner-popup-overlay.van-overlay) {
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3000;
}

.banner-popup {
  width: 80vw;
  max-width: 360px;
  max-height: 70vh;
  border-radius: 12px;
  overflow: hidden;
  background: #1a0a00;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
}

.banner-popup__img {
  width: 100%;
  flex: 1;
  min-height: 0;
  max-height: 48vh;
  object-fit: cover;
  display: block;
}

.banner-popup__content {
  padding: 22px 20px 20px;
  color: #ffe9bc;
  text-align: center;
}

.banner-popup__title {
  margin: 0;
  color: #fff4d9;
  font-size: 20px;
  font-weight: 700;
  line-height: 1.25;
}

.banner-popup__desc {
  max-height: 30vh;
  margin: 12px 0 0;
  overflow: auto;
  color: rgba(255, 233, 188, 0.78);
  font-size: 15px;
  line-height: 1.55;
  white-space: pre-wrap;
  text-align: left;
}

.banner-popup__actions {
  display: flex;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.popup-btn {
  flex: 1;
  height: 48px;
  font-size: 15px;
  font-weight: 600;
  background: transparent;
}

.popup-btn--dismiss {
  color: rgba(255, 255, 255, 0.4);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
}

.popup-btn--ok {
  color: #f3c84f;
}
</style>

<route lang="json5">
{
  name: 'Home'
}
</route>
