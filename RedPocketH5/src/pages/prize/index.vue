<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { LuckyWheel } from '@lucky-canvas/vue'
import { showToast } from 'vant'
import AppPageHeader from '@/components/AppPageHeader.vue'
import imgCoin from '@/assets/svg/coin.svg'

interface PageData {
  childCount: number
  normalCount: number
  advancedCount: number
  superCount: number
  rewardList: string[]
  recordList: RecordItem[]
}

interface RecordItem {
  uid: string
  type: 1 | 2 | 3
  reward: string
}

interface DrawResult {
  index: number
  reward: string
}

const { t } = useI18n()
const router = useRouter()

function svgDataUri(svg: string) {
  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`
}

const soundEffectUrl = 'https://pic.bofapic.com/static/_template_/maroon/media/turntable_sound.mp3'
const prizeBg = 'https://pub-93b0b439f98b49c4ba1db81844583907.r2.dev/static/_template_/maroon/img/activity/turntable/prize.png'
const lightBg1 = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 600 600">
  <defs>
    <radialGradient id="g">
      <stop offset="0%" stop-color="rgba(255,233,94,0.58)"/>
      <stop offset="48%" stop-color="rgba(255,181,0,0.28)"/>
      <stop offset="100%" stop-color="rgba(255,181,0,0)"/>
    </radialGradient>
  </defs>
  <circle cx="300" cy="300" r="280" fill="url(#g)"/>
</svg>
`)
const lightBg2 = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 700 700">
  <defs>
    <radialGradient id="g">
      <stop offset="0%" stop-color="rgba(255,245,180,0.92)"/>
      <stop offset="26%" stop-color="rgba(255,213,83,0.62)"/>
      <stop offset="100%" stop-color="rgba(255,213,83,0)"/>
    </radialGradient>
  </defs>
  <circle cx="350" cy="350" r="300" fill="url(#g)"/>
</svg>
`)
const lightBg3 = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 900 900">
  <g fill="rgba(255,220,105,0.25)">
    <path d="M450 40 L500 320 L450 860 L400 320 Z"/>
    <path d="M40 450 L320 400 L860 450 L320 500 Z"/>
    <path d="M144 144 L366 366 L756 756 L534 534 Z"/>
    <path d="M756 144 L534 366 L144 756 L366 534 Z"/>
  </g>
</svg>
`)
const closeWhite = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
  <circle cx="16" cy="16" r="15" fill="rgba(255,255,255,0.1)" stroke="#fff" stroke-width="2"/>
  <path d="M10 10 L22 22 M22 10 L10 22" stroke="#fff" stroke-width="3" stroke-linecap="round"/>
</svg>
`)
const winPrize = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 180 180">
  <defs>
    <linearGradient id="g" x1="0%" x2="100%">
      <stop offset="0%" stop-color="#fff1ad"/>
      <stop offset="55%" stop-color="#ffcb43"/>
      <stop offset="100%" stop-color="#b97206"/>
    </linearGradient>
  </defs>
  <path d="M58 34 H122 V56 C122 80 108 100 90 108 C72 100 58 80 58 56 Z" fill="url(#g)" stroke="#fff0bc" stroke-width="5"/>
  <path d="M58 40 H30 C32 70 46 84 68 88" fill="none" stroke="#e7ba43" stroke-width="8" stroke-linecap="round"/>
  <path d="M122 40 H150 C148 70 134 84 112 88" fill="none" stroke="#e7ba43" stroke-width="8" stroke-linecap="round"/>
  <rect x="76" y="108" width="28" height="24" rx="6" fill="#d69a18"/>
  <rect x="54" y="132" width="72" height="16" rx="8" fill="#8e220f"/>
</svg>
`)

const prizeImg1 = {
  src: svgDataUri(`
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120">
    <defs>
      <linearGradient id="g" x1="0%" x2="100%">
        <stop offset="0%" stop-color="#fff6b2"/>
        <stop offset="55%" stop-color="#ffd347"/>
        <stop offset="100%" stop-color="#b97608"/>
      </linearGradient>
    </defs>
    <circle cx="60" cy="60" r="42" fill="url(#g)" stroke="#fff0bf" stroke-width="6"/>
    <text x="60" y="72" text-anchor="middle" font-size="34" font-weight="800" fill="#7a2708">$</text>
  </svg>
  `),
  width: '40%',
  top: '45%',
}
const prizeImg2 = {
  src: svgDataUri(`
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120">
    <defs>
      <linearGradient id="g" x1="0%" x2="100%">
        <stop offset="0%" stop-color="#fff1aa"/>
        <stop offset="55%" stop-color="#ffcb3a"/>
        <stop offset="100%" stop-color="#bb7200"/>
      </linearGradient>
    </defs>
    <path d="M38 36 H82 L90 50 L78 92 H42 L30 50 Z" fill="url(#g)" stroke="#fff1c5" stroke-width="5"/>
    <path d="M50 36 C50 24 70 24 70 36" fill="none" stroke="#8f2d10" stroke-width="6" stroke-linecap="round"/>
  </svg>
  `),
  width: '40%',
  top: '45%',
}
const prizeImg3 = {
  src: svgDataUri(`
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120">
    <defs>
      <linearGradient id="g" x1="0%" x2="100%">
        <stop offset="0%" stop-color="#fff7be"/>
        <stop offset="55%" stop-color="#ffd44b"/>
        <stop offset="100%" stop-color="#b87406"/>
      </linearGradient>
    </defs>
    <path d="M26 40 L44 28 L60 44 L76 28 L94 40 L84 78 H36 Z" fill="url(#g)" stroke="#fff3c7" stroke-width="5"/>
    <circle cx="60" cy="58" r="10" fill="#a31d14"/>
  </svg>
  `),
  width: '40%',
  top: '45%',
}
const prizeImg4 = {
  src: svgDataUri(`
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120">
    <defs>
      <linearGradient id="g" x1="0%" x2="100%">
        <stop offset="0%" stop-color="#fff2b5"/>
        <stop offset="55%" stop-color="#ffd041"/>
        <stop offset="100%" stop-color="#ba7607"/>
      </linearGradient>
    </defs>
    <path d="M60 24 L74 44 L98 48 L80 66 L84 92 L60 80 L36 92 L40 66 L22 48 L46 44 Z" fill="url(#g)" stroke="#fff2c0" stroke-width="5"/>
  </svg>
  `),
  width: '40%',
  top: '45%',
}

const rewardCatalog = ['8', '18', '28', '38', '58', '88', '128', '188', '588', 'Random']

const state = reactive({
  winningShow: false,
  type: 0,
  drawType: 2,
  reward: '',
  pageData: {
    childCount: 0,
    normalCount: 3,
    advancedCount: 8,
    superCount: 15,
    rewardList: rewardCatalog,
    recordList: [
      { uid: 'UID*321', type: 1 as const, reward: '18' },
      { uid: 'UID*873', type: 2 as const, reward: '88' },
      { uid: 'UID*552', type: 3 as const, reward: '188' },
      { uid: 'UID*119', type: 1 as const, reward: '8' },
      { uid: 'UID*694', type: 2 as const, reward: '58' },
      { uid: 'UID*205', type: 3 as const, reward: '588' },
    ],
  } as PageData,
})

const blocks = ref([
  {
    padding: '20px',
    imgs: [{
      src: prizeBg,
      width: '100%',
      height: '100%',
    }],
  },
])

const buttons = ref<any[]>([])
const prizes = ref<any[]>([])
const wheelCanvas = ref<any>(null)
const listContainer = ref<HTMLElement | null>(null)
const listWrapper = ref<HTMLElement | null>(null)
const audioRef = ref<HTMLAudioElement | null>(null)
const animationFrame = ref<number>()
const spinning = ref(false)

const canvasWidth = computed(() => {
  if (typeof window === 'undefined')
    return 300
  return window.innerWidth > 480 ? 340 : Math.floor(window.innerWidth * 0.74)
})

const wheelDefaultConfig = ref({
  speed: 15,
  decelerationTime: 2500,
})

const currentSpinText = computed(() => {
  if (state.type === 3)
    return t('prizePage.superSpin')
  if (state.type === 2)
    return t('prizePage.advancedSpin')
  return t('prizePage.normalSpin')
})

// ── Jackpot counter ───────────────────────────────────────────────
const jackpotValue = ref(5290463)
const jackpotDisplay = ref(5290463)
const isJpFlashing = ref(false)
let jpRafId: number | undefined
let jpIntervalId: number | undefined

function rollJackpot(target: number) {
  cancelAnimationFrame(jpRafId!)
  const step = () => {
    const diff = target - jackpotDisplay.value
    if (Math.abs(diff) > 0.5) {
      jackpotDisplay.value += diff * 0.1
      jpRafId = requestAnimationFrame(step)
    }
    else {
      jackpotDisplay.value = target
    }
  }
  jpRafId = requestAnimationFrame(step)
}

function startJackpot() {
  jpIntervalId = window.setInterval(() => {
    jackpotValue.value += Math.floor(Math.random() * 888 + 200)
    isJpFlashing.value = false
    nextTick(() => {
      isJpFlashing.value = true
    })
    setTimeout(() => {
      isJpFlashing.value = false
    }, 600)
    rollJackpot(jackpotValue.value)
  }, 2000)
}

// ── LED ring ──────────────────────────────────────────────────────
const LED_COUNT = 24
const ledBulbs = Array.from({ length: LED_COUNT }, (_, i) => ({
  angle: i * (360 / LED_COUNT),
  delay: `${((i * 2) / LED_COUNT).toFixed(2)}s`,
}))

// ── Existing wheel logic ──────────────────────────────────────────
function randomInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function randomDelay(minSeconds: number, maxSeconds: number) {
  return randomInt(minSeconds * 1000, maxSeconds * 1000)
}

function createMockPageData(): PageData {
  const childCount = randomInt(0, 15)
  const recordList = Array.from({ length: 10 }, (_, index) => ({
    uid: `UID*${randomInt(100, 999)}`,
    type: (index % 3 + 1) as 1 | 2 | 3,
    reward: rewardCatalog[randomInt(0, rewardCatalog.length - 2)],
  }))
  return {
    childCount,
    normalCount: 3,
    advancedCount: 8,
    superCount: 15,
    rewardList: [...rewardCatalog],
    recordList,
  }
}

async function mockPhoneCanDraw() {
  await new Promise(resolve => setTimeout(resolve, 120))
  const childCount = state.pageData.childCount
  if (childCount >= state.pageData.superCount)
    return 4
  if (childCount >= state.pageData.advancedCount)
    return 3
  if (childCount >= state.pageData.normalCount)
    return 2
  return 6
}

async function mockLuckDrawInit() {
  await new Promise(resolve => setTimeout(resolve, 160))
  return createMockPageData()
}

function pickWeightedIndex(type: number) {
  const pools: Record<number, number[]> = {
    1: [0, 1, 2, 3, 4, 5, 9],
    2: [1, 2, 3, 4, 5, 6, 7, 9],
    3: [3, 4, 5, 6, 7, 8, 9],
  }
  const pool = pools[type] || pools[1]
  return pool[randomInt(0, pool.length - 1)]
}

async function mockLuckDrawDo({ key }: { key: number }): Promise<DrawResult> {
  await new Promise(resolve => setTimeout(resolve, randomInt(450, 900)))
  const index = pickWeightedIndex(key)
  const reward = state.pageData.rewardList[index] || '8'
  return { index, reward }
}

function buildPrizeConfig() {
  const texts = state.pageData.rewardList
  prizes.value = [
    { background: '#8b0000', imgs: [prizeImg1], fonts: [{ text: texts[0], top: '20%', fontColor: '#ffe59a', fontSize: '13px' }] },
    { background: '#1a1a1a', imgs: [prizeImg2], fonts: [{ text: texts[1], top: '20%', fontColor: '#ffbb00', fontSize: '13px' }] },
    { background: '#8b0000', imgs: [prizeImg3], fonts: [{ text: texts[2], top: '20%', fontColor: '#ffe59a', fontSize: '13px' }] },
    { background: '#1a1a1a', imgs: [prizeImg2], fonts: [{ text: texts[3], top: '20%', fontColor: '#ffbb00', fontSize: '13px' }] },
    { background: '#8b0000', imgs: [prizeImg3], fonts: [{ text: texts[4], top: '20%', fontColor: '#ffe59a', fontSize: '13px' }] },
    { background: '#1a1a1a', imgs: [prizeImg4], fonts: [{ text: texts[5], top: '20%', fontColor: '#ffbb00', fontSize: '13px' }] },
    { background: '#8b0000', imgs: [prizeImg1], fonts: [{ text: texts[6], top: '20%', fontColor: '#ffe59a', fontSize: '13px' }] },
    { background: '#1a1a1a', imgs: [prizeImg2], fonts: [{ text: texts[7], top: '20%', fontColor: '#ffbb00', fontSize: '13px' }] },
    { background: '#8b0000', imgs: [prizeImg1], fonts: [{ text: texts[8], top: '20%', fontColor: '#ffe59a', fontSize: '13px' }] },
    { background: '#1a1a1a', imgs: [prizeImg2], fonts: [{ text: texts[9], top: '20%', fontColor: '#ffbb00', fontSize: '13px' }] },
  ]

  const isActive = state.type > 0
  buttons.value = [{
    radius: '30%',
    background: isActive ? '#d4a017' : '#8a7355',
    pointer: false,
    fonts: [{
      text: isActive ? currentSpinText.value : t('prizePage.spinAction'),
      fontColor: isActive ? '#5a1206' : '#c4b89a',
      fontSize: '12',
      fontWeight: 'bold',
      top: '20%',
    }],
  }]
}

async function refreshPageState() {
  state.pageData = await mockLuckDrawInit()
  const status = await mockPhoneCanDraw()

  if (status === 4)
    state.type = 3
  else if (status === 3)
    state.type = 2
  else if (status === 2)
    state.type = 1
  else
    state.type = 0

  buildPrizeConfig()
  window.setTimeout(() => {
    wheelCanvas.value?.init?.()
  }, 80)
}

function addLatestRecord(reward: string) {
  state.pageData.recordList.unshift({ uid: 'ME***', type: (state.type || 1) as 1 | 2 | 3, reward })
  state.pageData.recordList = state.pageData.recordList.slice(0, 12)
}

function pauseSound() {
  if (!audioRef.value)
    return
  audioRef.value.pause()
  audioRef.value.currentTime = 0
}

function startScrolling() {
  const step = () => {
    const container = listContainer.value
    const wrapper = listWrapper.value
    if (container && wrapper) {
      container.scrollTop += 0.5
      if (container.scrollTop >= wrapper.scrollHeight / 2) {
        container.scrollTop = 0
        const first = state.pageData.recordList.shift()
        if (first)
          state.pageData.recordList.push(first)
      }
    }
    animationFrame.value = window.requestAnimationFrame(step)
  }
  animationFrame.value = window.requestAnimationFrame(step)
}

async function startCallback() {
  if (!state.type || spinning.value) {
    showToast(t('prizePage.noChance'))
    return
  }
  spinning.value = true
  state.reward = ''
  wheelCanvas.value?.play?.()
  if (audioRef.value) {
    audioRef.value.currentTime = 0
    audioRef.value.play().catch(() => {})
  }
  try {
    const { index, reward } = await mockLuckDrawDo({ key: state.type })
    state.reward = reward
    const stopDelay = randomDelay(3, 5)
    window.setTimeout(() => {
      wheelCanvas.value?.stop?.(index)
    }, stopDelay)
  }
  catch (error: any) {
    wheelCanvas.value?.init?.()
    pauseSound()
    spinning.value = false
    showToast(error?.message || t('prizePage.noChance'))
  }
}

async function endCallback() {
  pauseSound()
  spinning.value = false
  if (state.reward) {
    addLatestRecord(state.reward)
    state.winningShow = true
  }
  await refreshPageState()
}

function closeWinning() {
  state.winningShow = false
  state.reward = ''
}

function goBack() {
  router.back()
}
function goInvite() {
  router.push('/invite')
}

function spinTypeLabel(type: 1 | 2 | 3) {
  if (type === 3) return t('prizePage.superSpin')
  if (type === 2) return t('prizePage.advancedSpin')
  return t('prizePage.normalSpin')
}

onMounted(async () => {
  await refreshPageState()
  startScrolling()
  startJackpot()
})

onBeforeUnmount(() => {
  if (animationFrame.value)
    cancelAnimationFrame(animationFrame.value)
  cancelAnimationFrame(jpRafId!)
  clearInterval(jpIntervalId)
  pauseSound()
})
</script>

<template>
  <div class="prize-page">
    <!-- 环境光 -->
    <div class="ambient-bg" />
    <div class="velvet-texture" />

    <AppPageHeader :title="t('prizePage.title')" @back="goBack" />

    <div class="prize-scroll">
      <audio ref="audioRef" :src="soundEffectUrl" preload="auto" />

      <!-- 顶部奖池 -->
      <div class="jp-container" :class="{ 'jp-flash': isJpFlashing }">
        <p class="jp-super-label">SUPER JACKPOT</p>
        <div class="jp-amount-row">
          <img :src="imgCoin" class="jp-coin" alt="">
          <span class="jp-amount">{{ Math.floor(jackpotDisplay).toLocaleString('en-US') }}</span>
        </div>
      </div>

      <!-- 转盘系统 -->
      <div class="wheel-system" :style="{ '--ws': `${canvasWidth + 64}px`, '--wr': `${(canvasWidth + 64) / 2}px` }">
        <!-- 3D 机械外框 -->
        <div class="outer-frame" />

        <!-- LED 灯珠轨道 -->
        <div class="led-track">
          <div
            v-for="(bulb, i) in ledBulbs"
            :key="i"
            class="led-bulb"
            :style="{ transform: `rotate(${bulb.angle}deg)`, animationDelay: bulb.delay }"
          />
        </div>

        <!-- 霓虹指针 -->
        <svg class="modern-pointer" viewBox="0 0 100 120">
          <defs>
            <linearGradient id="ptrGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#00f2ff" />
              <stop offset="100%" stop-color="#0066ff" />
            </linearGradient>
          </defs>
          <path d="M50 120 L10 20 L90 20 Z" fill="url(#ptrGrad)" />
          <circle cx="50" cy="35" r="10" fill="#fff" opacity="0.5" />
        </svg>

        <!-- LuckyCanvas 转盘 -->
        <LuckyWheel
          ref="wheelCanvas"
          :width="canvasWidth"
          :height="canvasWidth"
          :prizes="prizes"
          :blocks="blocks"
          :buttons="buttons"
          :default-config="wheelDefaultConfig"
          class="wheel-canvas"
          @start="startCallback"
          @end="endCallback"
        />
      </div>

      <!-- 邀请按钮 -->
      <button type="button" class="invite-btn" @click="goInvite">
        {{ t('prizePage.spinAction') }}
      </button>

      <!-- 中奖记录 -->
      <div class="winning">
        <div class="winning-report">
          <span class="active">{{ t('prizePage.recordUser') }}</span>
          <span>{{ t('prizePage.recordType') }}</span>
          <span>{{ t('prizePage.recordReward') }}</span>
        </div>
        <div ref="listContainer" class="winning-list">
          <div ref="listWrapper" class="scroll-up">
            <div
              v-for="(item, index) in state.pageData.recordList"
              :key="`${item.uid}-${item.reward}-${index}`"
              class="winning-item"
            >
              <span class="winning-item__cell">{{ item.uid }}</span>
              <span class="winning-item__cell">{{ spinTypeLabel(item.type) }}</span>
              <span class="winning-item__cell winning-item__reward">
                <CoinAmount :text="item.reward" />
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 中奖弹窗 -->
      <van-overlay :show="state.winningShow" class="winning-overlay" @click.self="closeWinning">
        <div class="win-modal">
          <!-- 光晕背景层 -->
          <img class="win-glow-outer" :src="lightBg3" alt="">
          <img class="win-glow-inner" :src="lightBg2" alt="">
          <!-- 内容卡片 -->
          <div class="win-card" :style="{ backgroundImage: `url(${lightBg1})` }">
            <img class="win-close" :src="closeWhite" @click="closeWinning">
            <div class="win-body">
              <h1 class="win-title">{{ t('prizePage.resultTitle') }}</h1>
              <img class="win-prize-img" :src="winPrize">
              <div class="win-amount">
                <CoinAmount :text="state.reward" class="win-coin-amount" />
              </div>
            </div>
            <button type="button" class="win-btn" @click="closeWinning">
              {{ t('prizePage.claimAction') }}
            </button>
          </div>
        </div>
      </van-overlay>
    </div>
  </div>
</template>

<style scoped>
/* ── 页面基础 ── */
.prize-page {
  min-height: 100vh;
  background: #050000;
  position: relative;
}

.ambient-bg {
  position: fixed;
  inset: 0;
  background: radial-gradient(circle at center, #8b0000 0%, #050000 78%);
  opacity: 0.65;
  z-index: 1;
  pointer-events: none;
}

.velvet-texture {
  position: fixed;
  inset: 0;
  opacity: 0.04;
  pointer-events: none;
  z-index: 2;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
}

.prize-scroll {
  position: relative;
  z-index: 5;
  min-height: calc(100vh - 46px);
  padding: 0 16px calc(80px + env(safe-area-inset-bottom));
  overflow: hidden;
}

/* ── 顶部奖池 ── */
.jp-container {
  margin: 16px auto 0;
  max-width: 360px;
  text-align: center;
}

@keyframes jpFlash {
  0% {
    transform: scale(1);
    filter: brightness(1);
  }

  30% {
    transform: scale(1.03);
    filter: brightness(1.8) drop-shadow(0 0 16px #ffbb00);
  }

  100% {
    transform: scale(1);
    filter: brightness(1);
  }
}

.jp-flash {
  animation: jpFlash 0.6s ease-out forwards;
}

.jp-super-label {
  margin: 0 0 6px;
  font-size: 11px;
  letter-spacing: 6px;
  text-transform: uppercase;
  color: #ffcc00;
}

.jp-amount-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.jp-coin {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
}

.jp-amount {
  font-size: 44px;
  font-weight: 900;
  line-height: 1;
  background: linear-gradient(180deg, #fff5c3 0%, #ffbb00 50%, #8b4513 100%);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.9));
  font-variant-numeric: tabular-nums;
}

/* ── 转盘系统 ── */
.wheel-system {
  position: relative;
  width: var(--ws);
  height: var(--ws);
  margin: 20px auto 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

.outer-frame {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: linear-gradient(135deg, #2a2a2a, #0d0d0d, #2a2a2a);
  border: 4px solid #5d1a1a;
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.85),
    inset 0 0 30px #000;
}

/* LED 灯珠轨道 */
.led-track {
  position: absolute;
  inset: 6%;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  box-shadow: inset 0 0 12px rgba(255, 187, 0, 0.15);
}

.led-bulb {
  position: absolute;
  width: 10px;
  height: 10px;
  background: #fff;
  border-radius: 50%;
  left: calc(50% - 5px);
  top: 6px;
  /* rotate around the center of led-track: radius(44%*ws) minus bulb top offset */
  transform-origin: 5px calc(var(--ws) * 0.44 - 6px);
  box-shadow: 0 0 4px #fff, 0 0 12px #ffbb00;
  animation: ledFlow 2s infinite ease-in-out;
}

@keyframes ledFlow {
  0%,
  100% {
    opacity: 0.25;
    filter: brightness(1);
  }

  50% {
    opacity: 1;
    filter: brightness(1.6) blur(0.5px);
  }
}

/* 霓虹指针 */
.modern-pointer {
  position: absolute;
  top: -44px;
  width: 52px;
  height: 76px;
  z-index: 20;
  filter: drop-shadow(0 0 16px #00f2ff);
}

/* LuckyCanvas */
.wheel-canvas {
  z-index: 10;
  position: relative;
}

/* ── 邀请按钮 ── */
.invite-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  max-width: 360px;
  height: 50px;
  margin: 24px auto 0;
  border: 1px solid rgba(255, 236, 157, 0.4);
  border-radius: 999px;
  background: linear-gradient(180deg, #fff9d0 0%, #ffd358 48%, #bf780c 100%);
  box-shadow:
    0 10px 20px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.35);
  color: #7c2200;
  font-size: 16px;
  font-weight: 900;
  letter-spacing: 0.06em;
}

/* ── 中奖记录 ── */
.winning {
  margin: 24px auto 0;
  max-width: 360px;
  border: 1px solid #fdf07e;
  border-radius: 10px;
  background: linear-gradient(180deg, #200000 0%, #500007 100%);
  overflow: hidden;
}

.winning-report,
.winning-item {
  display: flex;
  align-items: center;
}

.winning-report {
  height: 36px;
  background: #5d1519;
  color: rgba(255, 255, 255, 0.4);
  font-size: 12px;
}

.winning-report span,
.winning-item__cell {
  flex: 1;
  text-align: center;
}

.winning-report .active {
  color: #fff;
  font-weight: 700;
}

.winning-list {
  max-height: 126px;
  overflow: hidden;
}

.winning-item {
  padding: 9px 12px;
  color: #fff;
  font-size: 12px;
}

.winning-item__reward {
  color: #ffea00;
  font-weight: 700;
  display: flex;
  justify-content: center;
}

/* ── 中奖弹窗 ── */
:deep(.winning-overlay.van-overlay) {
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.win-modal {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  max-width: 360px;
  padding: 0 16px;
}

.win-glow-outer {
  position: absolute;
  width: 120%;
  max-width: 500px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  animation: scaleFloat 2s linear infinite;
  z-index: 0;
}

.win-glow-inner {
  position: absolute;
  width: 80%;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 0;
}

.win-card {
  position: relative;
  z-index: 1;
  width: 100%;
  padding: 40px 0 80px;
  text-align: center;
  background-repeat: no-repeat;
  background-position: center;
  background-size: contain;
}

.win-close {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 24px;
  cursor: pointer;
}

.win-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 32px;
}

.win-title {
  margin: 0 0 8px;
  color: #fff;
  font-size: 22px;
  font-weight: 700;
}

.win-prize-img {
  width: 96px;
  margin: 8px 0;
}

.win-amount {
  display: flex;
  justify-content: center;
  margin-top: 4px;
}

.win-btn {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 148px;
  height: 44px;
  padding: 0 24px;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffd601 0%, #ffa800 100%);
  color: #921111;
  font-size: 15px;
  font-weight: 700;
  white-space: nowrap;
}

:deep(.win-coin-amount) {
  font-size: 36px;
  font-weight: 800;
  color: #ffe65c;
}

:deep(.tip-coin-amount .coin-amount-icon) {
  width: 1.1em;
  height: 1.1em;
}

@keyframes scaleFloat {
  0%,
  100% {
    transform: scale(0.96);
  }

  50% {
    transform: scale(1.04);
  }
}
</style>

<route lang="json5">
{
  name: 'Prize',
}
</route>
