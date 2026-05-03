<script setup lang="ts">
import { showToast } from 'vant'
import type { BannerItem, LuckyHistoryUserFlowItem } from '@/api/user'
import { getBanners, getLuckyHistoryUserFlow } from '@/api/user'
import { formatCurrency } from '@/utils/currency'
import imgAvatarPlaceholder from '@/assets/images/avatar-placeholder.png'
import coinSvgUrl from '@/assets/svg/coin.svg'

const { t } = useI18n()
const router = useRouter()

const DISMISSED_KEY = 'dismissed_popup_banner_ids'

const DEFAULT_AVATAR = imgAvatarPlaceholder
const activeIndex = ref(0)

const homeBanners = ref<BannerItem[]>([])
const popupQueue = ref<BannerItem[]>([])
const popupVisible = ref(false)
const popupIndex = ref(0)

const currentPopup = computed(() => popupQueue.value[popupIndex.value] ?? null)
const recentWinnersLoading = ref(false)
const recentWinners = ref<any[]>([])
const thunderCanvasRef = ref<HTMLCanvasElement | null>(null)
const parityCanvasRef = ref<HTMLCanvasElement | null>(null)

interface CoinAnimationController {
  stop: () => void
}

let coinAnimationControllers: CoinAnimationController[] = []
let coinImagePromise: Promise<HTMLImageElement> | null = null

const visibleWinners = computed(() => recentWinners.value)
const showWinnerLoading = computed(() => recentWinnersLoading.value && visibleWinners.value.length === 0)
const showWinnerEmpty = computed(() => visibleWinners.value.length === 0)
const marqueeText = computed(() => homeBanners.value[activeIndex.value]?.bannerName || '')

function onSwipeChange(index: number) {
  activeIndex.value = index
}

function goPacketList(mode: 0 | 1) {
  router.push({
    path: '/packetList',
    query: { mode: String(mode) },
  })
}

function ensureCoinImage() {
  if (coinImagePromise)
    return coinImagePromise

  coinImagePromise = new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = reject
    image.src = coinSvgUrl
  })

  return coinImagePromise
}

class CoinParticle {
  ctx: CanvasRenderingContext2D
  image: HTMLImageElement
  width: number
  height: number
  radius = 3
  x = 0
  y = 0
  vx = 0
  vy = 0
  gravity = 0.12
  bounce = 0.4
  rotation = 0
  rotationSpeed = 0

  constructor(context: CanvasRenderingContext2D, image: HTMLImageElement, width: number, height: number) {
    this.ctx = context
    this.image = image
    this.width = width
    this.height = height
    this.init()
  }

  init() {
    this.radius = Math.random() > 0.72 ? 3 : 2
    this.x = Math.random() * (this.width - this.radius * 2) + this.radius
    this.y = -24 - Math.random() * 80
    this.vx = (Math.random() - 0.5) * 2.2
    this.vy = Math.random() * 1 + 0.45
    this.gravity = 0.065
    this.bounce = 0.5
    this.rotation = Math.random() * 360
    this.rotationSpeed = Math.random() * 4 - 2
  }

  update() {
    this.vy += this.gravity
    this.x += this.vx
    this.y += this.vy

    if (this.y + this.radius > this.height - 18) {
      this.y = this.height - 18 - this.radius
      this.vy *= -this.bounce
      this.vx *= 0.82
    }

    if (this.x + this.radius > this.width || this.x - this.radius < 0)
      this.vx *= -1

    this.rotation += this.rotationSpeed
  }

  draw() {
    const ctx = this.ctx
    const size = this.radius * 2.6
    ctx.save()
    ctx.translate(this.x, this.y)
    ctx.rotate((this.rotation * Math.PI) / 180)
    ctx.drawImage(this.image, -size / 2, -size / 2, size, size)
    ctx.restore()
  }
}

function setupCoinCanvas(canvas: HTMLCanvasElement, coinImage: HTMLImageElement) {
  const reducedMotion = typeof window !== 'undefined'
    ? window.matchMedia('(prefers-reduced-motion: reduce)').matches
    : false
  const ratio = typeof window !== 'undefined' ? Math.min(window.devicePixelRatio || 1, 2) : 1
  const rect = canvas.getBoundingClientRect()
  const width = Math.max(1, Math.floor(rect.width))
  const height = Math.max(1, Math.floor(rect.height))
  const context = canvas.getContext('2d')
  if (!context)
    return { stop: () => {} }

  canvas.width = Math.floor(width * ratio)
  canvas.height = Math.floor(height * ratio)
  context.setTransform(ratio, 0, 0, ratio, 0, 0)

  if (reducedMotion) {
    context.clearRect(0, 0, width, height)
    return { stop: () => {} }
  }

  const particles: CoinParticle[] = []
  const maxParticles = 10
  let frameId = 0

  const render = () => {
    context.clearRect(0, 0, width, height)

    if (particles.length < maxParticles && Math.random() < 0.035)
      particles.push(new CoinParticle(context, coinImage, width, height))

    for (const particle of particles) {
      particle.update()
      particle.draw()
      if (Math.abs(particle.vy) < 0.08 && particle.y > height - 36)
        particle.init()
    }

    frameId = window.requestAnimationFrame(render)
  }

  frameId = window.requestAnimationFrame(render)

  return {
    stop: () => window.cancelAnimationFrame(frameId),
  }
}

async function initEntryCardAnimations() {
  await nextTick()
  coinAnimationControllers.forEach(controller => controller.stop())
  coinAnimationControllers = []

  const coinImage = await ensureCoinImage()

  const canvases = [thunderCanvasRef.value, parityCanvasRef.value].filter(Boolean) as HTMLCanvasElement[]
  canvases.forEach((canvas) => {
    coinAnimationControllers.push(setupCoinCanvas(canvas, coinImage))
  })
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

function getDismissedIds(): number[] {
  try {
    return JSON.parse(localStorage.getItem(DISMISSED_KEY) || '[]')
  }
  catch {
    return []
  }
}

function addDismissedId(id: number) {
  const ids = getDismissedIds()
  if (!ids.includes(id)) {
    ids.push(id)
    localStorage.setItem(DISMISSED_KEY, JSON.stringify(ids))
  }
}

async function loadBanners() {
  try {
    const { data } = await getBanners()
    homeBanners.value = (data?.home ?? []).filter(b => b.status === 1)
    const dismissed = getDismissedIds()
    popupQueue.value = (data?.popup ?? [])
      .filter(b => b.status === 1 && !dismissed.includes(b.id))
      .sort((a, b) => a.sort - b.sort)
    if (popupQueue.value.length > 0) {
      popupIndex.value = 0
      popupVisible.value = true
    }
  }
  catch { /* silent — carousel stays empty */ }
}

function onBannerClick(banner: BannerItem) {
  if (banner.jumpType === 'url' && banner.jumpValue)
    window.open(banner.jumpValue, '_blank')
}

function onPopupOk() {
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
    addDismissedId(item.id)
  popupQueue.value.splice(popupIndex.value, 1)
  if (popupQueue.value.length === 0) {
    popupVisible.value = false
  }
  else {
    popupIndex.value = Math.min(popupIndex.value, popupQueue.value.length - 1)
  }
}

onMounted(async () => {
  void loadBanners()
  void loadRecentWinners()
  await initEntryCardAnimations()
})

onBeforeUnmount(() => {
  coinAnimationControllers.forEach(controller => controller.stop())
  coinAnimationControllers = []
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
            :alt="item.bannerName"
            :style="item.jumpType === 'url' ? 'cursor:pointer' : ''"
            @click="onBannerClick(item)"
          >
        </van-swipe-item>
      </van-swipe>
      <div class="banner-stripe" />
      <van-notice-bar class="home-marquee" :scrollable="true" :text="marqueeText" />
    </section>

    <section class="packet-entry-card">
      <div class="packet-entry-grid">
        <button type="button" class="packet-entry-btn thunder" @click="goPacketList(0)">
          <span class="packet-entry-btn__sunburst" aria-hidden="true" />
          <span class="packet-entry-btn__tag">{{ t('homeLucky.playModeThunderEyebrow') }}</span>
          <svg class="packet-entry-btn__kranok packet-entry-btn__kranok--tl" viewBox="0 0 100 100" aria-hidden="true">
            <path d="M10 90 Q 10 10 90 10 L 70 30 Q 30 30 30 70 Z" />
          </svg>
          <svg class="packet-entry-btn__kranok packet-entry-btn__kranok--br" viewBox="0 0 100 100" aria-hidden="true">
            <path d="M10 90 Q 10 10 90 10 L 70 30 Q 30 30 30 70 Z" />
          </svg>
          <canvas ref="thunderCanvasRef" class="packet-entry-btn__coins" aria-hidden="true" />

          <span class="packet-entry-btn__visual packet-entry-btn__visual--bomb" aria-hidden="true">
            <span class="bomb-wrapper">
              <span class="bomb-fuse" />
              <span class="bomb-spark">✦</span>
              <span class="bomb-main" />
              <span class="bomb-rim" />
            </span>
          </span>

          <span class="packet-entry-btn__content">
            <span class="packet-entry-btn__ticker">{{ t('homeLucky.playModeThunderDesc') }}</span>
            <span class="packet-entry-btn__footer">
              <strong class="packet-entry-btn__title">BOMB</strong>
              <small class="packet-entry-btn__subtitle">{{ t('packetListPage.modeThunder') }}</small>
            </span>
          </span>
        </button>
        <button type="button" class="packet-entry-btn parity" @click="goPacketList(1)">
          <span class="packet-entry-btn__sunburst" aria-hidden="true" />
          <span class="packet-entry-btn__tag">{{ t('homeLucky.playModeParityEyebrow') }}</span>
          <svg class="packet-entry-btn__kranok packet-entry-btn__kranok--tl" viewBox="0 0 100 100" aria-hidden="true">
            <path d="M10 90 Q 10 10 90 10 L 70 30 Q 30 30 30 70 Z" />
          </svg>
          <svg class="packet-entry-btn__kranok packet-entry-btn__kranok--br" viewBox="0 0 100 100" aria-hidden="true">
            <path d="M10 90 Q 10 10 90 10 L 70 30 Q 30 30 30 70 Z" />
          </svg>
          <canvas ref="parityCanvasRef" class="packet-entry-btn__coins" aria-hidden="true" />

          <span class="packet-entry-btn__visual packet-entry-btn__visual--parity" aria-hidden="true">
            <span class="pill-group">
              <span class="pill odd">ODD</span>
              <span class="pill even">EVEN</span>
            </span>
          </span>

          <span class="packet-entry-btn__content">
            <span class="packet-entry-btn__ticker">{{ t('homeLucky.playModeParityDesc') }}</span>
            <span class="packet-entry-btn__footer">
              <strong class="packet-entry-btn__title">ODD/EVEN</strong>
              <small class="packet-entry-btn__subtitle">{{ t('packetListPage.modeParity') }}</small>
            </span>
          </span>
        </button>
      </div>
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
          v-if="currentPopup"
          :src="currentPopup.imageUrl"
          class="banner-popup__img"
          :alt="currentPopup.bannerName"
        >
        <div class="banner-popup__actions">
          <button type="button" class="popup-btn popup-btn--dismiss" @click="onPopupDismiss">
            {{ t('common.noRemind') }}
          </button>
          <button type="button" class="popup-btn popup-btn--ok" @click="onPopupOk">
            {{ t('common.ok') }}
          </button>
        </div>
      </div>
    </van-overlay>
  </div>
</template>

<style scoped>
.home-page {
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
  margin-top: 14px;
}

.packet-entry-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.packet-entry-btn {
  position: relative;
  min-height: 228px;
  padding: 0;
  border-radius: 24px;
  border: 1px solid rgba(255, 221, 149, 0.34);
  color: #fff3de;
  overflow: hidden;
  isolation: isolate;
  background:
    radial-gradient(circle at 50% 18%, rgba(160, 22, 0, 0.42), rgba(160, 22, 0, 0) 35%),
    linear-gradient(180deg, rgba(255, 243, 212, 0.04), transparent 18%),
    linear-gradient(180deg, rgba(39, 2, 2, 0.98), rgba(27, 2, 2, 0.98));
  box-shadow:
    0 18px 32px rgba(0, 0, 0, 0.36),
    inset 0 0 0 1px rgba(255, 248, 214, 0.05),
    inset 0 -30px 48px rgba(0, 0, 0, 0.28);
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    border-color 180ms ease;
}

.packet-entry-btn.parity {
  background:
    radial-gradient(circle at 50% 18%, rgba(123, 80, 18, 0.28), rgba(123, 80, 18, 0) 35%),
    linear-gradient(180deg, rgba(255, 243, 212, 0.04), transparent 18%),
    linear-gradient(180deg, rgba(39, 2, 2, 0.98), rgba(27, 2, 2, 0.98));
}

.packet-entry-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(140deg, rgba(255, 255, 255, 0.1), transparent 24%),
    radial-gradient(circle at center, rgba(212, 175, 55, 0.06), transparent 58%);
  pointer-events: none;
}

.packet-entry-btn::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 24px;
  padding: 2px;
  background: linear-gradient(135deg, #fff5c3 0%, #ffbb00 25%, #8b4513 50%, #ffbb00 75%, #fff5c3 100%);
  -webkit-mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  mask-composite: exclude;
  filter: drop-shadow(0 0 5px rgba(255, 187, 0, 0.6));
  z-index: 0;
  pointer-events: none;
}

.packet-entry-btn:hover,
.packet-entry-btn:active {
  transform: translateY(-2px);
  border-color: rgba(255, 231, 169, 0.52);
  box-shadow:
    0 22px 36px rgba(0, 0, 0, 0.38),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08),
    0 0 18px rgba(212, 175, 55, 0.16);
}

.packet-entry-btn__sunburst {
  position: absolute;
  top: -58%;
  left: -58%;
  width: 220%;
  height: 220%;
  background: conic-gradient(from 0deg, transparent 0deg, rgba(255, 215, 0, 0.05) 15deg, transparent 30deg);
  animation: rotateSun 20s linear infinite;
  pointer-events: none;
  z-index: 0;
}

.packet-entry-btn__tag {
  position: relative;
  position: absolute;
  top: 18px;
  right: -28px;
  z-index: 3;
  background: linear-gradient(90deg, #ff8a00, #e52d27);
  padding: 5px 36px;
  transform: rotate(45deg);
  font-size: 11px;
  line-height: 1;
  font-weight: 800;
  color: #fff6eb;
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.28);
}

.packet-entry-btn__kranok {
  position: absolute;
  width: 34px;
  height: 34px;
  fill: rgba(255, 187, 0, 0.5);
  z-index: 1;
}

.packet-entry-btn__kranok--tl {
  top: 4px;
  left: 4px;
}

.packet-entry-btn__kranok--br {
  right: 8px;
  bottom: 8px;
  transform: rotate(180deg);
}

.packet-entry-btn__coins {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 2;
}

.packet-entry-btn__visual {
  position: absolute;
  left: 0;
  right: 0;
  top: 62px;
  display: flex;
  justify-content: center;
  z-index: 3;
}

.packet-entry-btn__visual--bomb {
  top: 64px;
}

.packet-entry-btn__visual--parity {
  top: 52px;
}

.bomb-wrapper {
  position: relative;
  width: 80px;
  height: 80px;
  animation: bombPulse 2s ease-in-out infinite;
}

.bomb-main {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, #555 0%, #141414 42%, #050505 100%);
  box-shadow:
    0 20px 40px rgba(0, 0, 0, 0.56),
    inset -5px -5px 15px rgba(255, 255, 255, 0.08);
}

.bomb-rim {
  position: absolute;
  inset: 4px;
  border-radius: 50%;
  border: 2px solid #ffd700;
  box-shadow: 0 0 10px rgba(255, 215, 0, 0.34);
  transform: translate(-1px, -2px);
}

.bomb-fuse {
  position: absolute;
  top: -16px;
  left: 50%;
  width: 5px;
  height: 22px;
  background: linear-gradient(180deg, #55331f 0%, #3d2b1f 100%);
  border-radius: 999px;
  transform: translateX(-50%) rotate(15deg);
}

.bomb-spark {
  position: absolute;
  top: -24px;
  left: 55%;
  font-size: 14px;
  color: #ffea33;
  filter: drop-shadow(0 0 10px rgba(255, 187, 0, 0.9));
  animation: sparkFlicker 0.1s infinite alternate;
}

.pill-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
}

.pill {
  min-width: 92px;
  padding: 7px 0;
  border-radius: 999px;
  font-size: 19px;
  line-height: 1;
  font-weight: 900;
  letter-spacing: 0.06em;
  border: 2px solid #ffbb00;
  text-align: center;
  box-shadow:
    0 0 14px rgba(255, 184, 0, 0.24),
    inset 0 0 8px rgba(255, 184, 0, 0.2);
}

.pill.odd {
  background: #ffbb00;
  color: #1a0000;
  transform: rotate(-3deg);
}

.pill.even {
  background: #1a0000;
  color: #ffbb00;
  transform: rotate(3deg);
}

.packet-entry-btn__content {
  position: absolute;
  inset: auto 16px 10px;
  z-index: 3;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.packet-entry-btn__ticker {
  margin-top: 8px;
  width: 100%;
  min-height: 32px;
  padding: 5px 10px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.48);
  color: rgba(255, 229, 186, 0.72);
  font-size: 9px;
  line-height: 1.25;
  text-align: center;
}

.packet-entry-btn__footer {
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.packet-entry-btn__title {
  color: #fff;
  font-size: 16px;
  line-height: 1;
  font-weight: 900;
  letter-spacing: 0.02em;
}

.packet-entry-btn__subtitle {
  margin-top: 4px;
  color: #ffbb00;
  font-size: 12px;
  line-height: 1;
  letter-spacing: 0.18em;
}

@keyframes rotateSun {
  to {
    transform: rotate(360deg);
  }
}

@keyframes bombPulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.06);
  }
}

@keyframes sparkFlicker {
  from {
    opacity: 0.8;
    transform: scale(0.8);
  }
  to {
    opacity: 1;
    transform: scale(1.2);
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

@media (prefers-reduced-motion: reduce) {
  .packet-entry-btn,
  .packet-entry-btn__sunburst,
  .bomb-wrapper,
  .bomb-spark {
    animation: none !important;
    transition: none !important;
    transform: none !important;
  }
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
  object-fit: cover;
  display: block;
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
