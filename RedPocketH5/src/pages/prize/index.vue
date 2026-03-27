<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { LuckyWheel } from '@lucky-canvas/vue'
import { showToast } from 'vant'
import AppPageHeader from '@/components/AppPageHeader.vue'

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
const spinBtn = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 220 220">
  <defs>
    <radialGradient id="g" cx="50%" cy="35%">
      <stop offset="0%" stop-color="#fff7c8"/>
      <stop offset="48%" stop-color="#f3c84f"/>
      <stop offset="100%" stop-color="#b86705"/>
    </radialGradient>
  </defs>
  <circle cx="110" cy="110" r="105" fill="url(#g)"/>
  <circle cx="110" cy="110" r="74" fill="#7f150f" stroke="#fff2b8" stroke-width="8"/>
  <path d="M110 8 L132 58 L88 58 Z" fill="#ffef9e"/>
  <circle cx="110" cy="110" r="12" fill="#ffe59a"/>
</svg>
`)
const disabledBtn = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 220 220">
  <defs>
    <radialGradient id="g" cx="50%" cy="35%">
      <stop offset="0%" stop-color="#e5dcc0"/>
      <stop offset="48%" stop-color="#b8a77b"/>
      <stop offset="100%" stop-color="#71583b"/>
    </radialGradient>
  </defs>
  <circle cx="110" cy="110" r="105" fill="url(#g)"/>
  <circle cx="110" cy="110" r="74" fill="#5d4333" stroke="#eadfb7" stroke-width="8"/>
  <circle cx="110" cy="110" r="12" fill="#f5edcf"/>
</svg>
`)
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
    <text x="60" y="72" text-anchor="middle" font-size="34" font-weight="800" fill="#7a2708">₹</text>
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

const rewardCatalog = ['₹8', '₹18', '₹28', '₹38', '₹58', '₹88', '₹128', '₹188', '₹588', 'Random']

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
      { uid: 'UID*321', type: 1 as const, reward: '₹18' },
      { uid: 'UID*873', type: 2 as const, reward: '₹88' },
      { uid: 'UID*552', type: 3 as const, reward: '₹188' },
      { uid: 'UID*119', type: 1 as const, reward: '₹8' },
      { uid: 'UID*694', type: 2 as const, reward: '₹58' },
      { uid: 'UID*205', type: 3 as const, reward: '₹588' },
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
    return 320
  return window.innerWidth > 480 ? 360 : Math.floor(window.innerWidth * 0.78)
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

function randomInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function randomDelay(minSeconds: number, maxSeconds: number) {
  return randomInt(minSeconds * 1000, maxSeconds * 1000)
}

function createMockPageData(): PageData {
  const childCount = randomInt(0, 15)
  const rewardList = [...rewardCatalog]
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
    rewardList,
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
  const pools = {
    1: [0, 1, 2, 3, 4, 5, 9],
    2: [1, 2, 3, 4, 5, 6, 7, 9],
    3: [3, 4, 5, 6, 7, 8, 9],
  }
  const pool = pools[type as 1 | 2 | 3] || pools[1]
  return pool[randomInt(0, pool.length - 1)]
}

async function mockLuckDrawDo({ key }: { key: number }): Promise<DrawResult> {
  await new Promise(resolve => setTimeout(resolve, randomInt(450, 900)))
  const index = pickWeightedIndex(key)
  const reward = state.pageData.rewardList[index] || '₹8'
  return { index, reward }
}

function buildPrizeConfig() {
  const texts = state.pageData.rewardList
  prizes.value = [
    { background: '#f7d4ca', imgs: [prizeImg1], fonts: [{ text: texts[0], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#fca89f', imgs: [prizeImg2], fonts: [{ text: texts[1], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#f7d4ca', imgs: [prizeImg3], fonts: [{ text: texts[2], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#fca89f', imgs: [prizeImg2], fonts: [{ text: texts[3], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#f7d4ca', imgs: [prizeImg3], fonts: [{ text: texts[4], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#fca89f', imgs: [prizeImg4], fonts: [{ text: texts[5], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#f7d4ca', imgs: [prizeImg1], fonts: [{ text: texts[6], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#fca89f', imgs: [prizeImg2], fonts: [{ text: texts[7], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#f7d4ca', imgs: [prizeImg1], fonts: [{ text: texts[8], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
    { background: '#fca89f', imgs: [prizeImg2], fonts: [{ text: texts[9], top: '20%', fontColor: '#9f3800', fontSize: '12px' }] },
  ]

  buttons.value = [{
    radius: '45%',
    fonts: [{
      text: state.type > 0 ? currentSpinText.value : t('prizePage.spinAction'),
      fontColor: '#561919',
      top: '-30%',
      fontSize: '12',
    }, {
      text: 'Spin',
      fontColor: '#561919',
      top: '0%',
      fontSize: '12',
    }],
    imgs: [{
      src: state.type > 0 ? spinBtn : disabledBtn,
      width: '80%',
      top: '-115%',
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
  state.pageData.recordList.unshift({
    uid: 'ME***',
    type: (state.type || 1) as 1 | 2 | 3,
    reward,
  })
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
  if (type === 3)
    return t('prizePage.superSpin')
  if (type === 2)
    return t('prizePage.advancedSpin')
  return t('prizePage.normalSpin')
}

onMounted(async () => {
  await refreshPageState()
  startScrolling()
})

onBeforeUnmount(() => {
  if (animationFrame.value)
    cancelAnimationFrame(animationFrame.value)
  pauseSound()
})
</script>

<template>
  <div class="prize-page">
    <AppPageHeader :title="t('prizePage.title')" @back="goBack" />

    <div class="prize-scroll">
      <audio ref="audioRef" :src="soundEffectUrl" preload="auto" />

      <div class="turntable">
        <div class="content">
          <div class="wheel-container">
            <div class="turntable-prize">
              <div class="wheel">
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
            </div>
          </div>

          <button type="button" class="invite-btn" @click="goInvite">
            Play
          </button>

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
                  <span class="winning-item__cell winning-item__reward">{{ item.reward }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <van-overlay :show="state.winningShow" class="winning-overlay">
        <div class="light-bg2">
          <img :src="lightBg3">
        </div>
        <div class="light-bg">
          <img :src="lightBg2">
        </div>
        <div class="spins-tips">
          <img class="close" :src="closeWhite" @click="closeWinning">
          <button type="button" class="spins-btn" @click="closeWinning">
            {{ t('prizePage.claimAction') }}
          </button>
          <div class="tip">
            <h1>{{ t('prizePage.resultTitle') }}</h1>
            <img class="tip__prize" :src="winPrize">
            <div class="tip__amount">
              +{{ state.reward }}
            </div>
          </div>
          <div class="light-bg1" :style="{ backgroundImage: `url(${lightBg1})` }" />
        </div>
      </van-overlay>
    </div>
  </div>
</template>

<style scoped>
.prize-page {
  min-height: 100vh;
  background: #190709;
}

.prize-scroll {
  position: relative;
  min-height: calc(100vh - 46px);
  padding-bottom: calc(104px + env(safe-area-inset-bottom));
  overflow: hidden;
}

.turntable {
  display: flex;
  min-height: 100%;
  color: #fff;
  background-color: #1a0313;
  flex-direction: column;
}

.content {
  flex: 1;
  padding-top: 8px;
  background:
    radial-gradient(circle at top, rgba(255, 211, 88, 0.18), transparent 28%),
    linear-gradient(180deg, rgba(93, 14, 11, 0.86) 0%, rgba(47, 2, 2, 0.96) 38%, rgba(25, 3, 19, 1) 100%);
  overflow-x: hidden;
}

.winning-report,
.winning-item {
  display: flex;
  align-items: center;
}

.wheel-container {
  display: flex;
  justify-content: center;
  width: 100%;
  height: 71vw;
  margin-top: 4px;
  padding: 0 42px;
}

.turntable-prize {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.wheel {
  position: relative;
  display: flex;
  justify-content: center;
  width: 100%;
}

.wheel-canvas {
  z-index: 1;
  max-width: 360px;
  max-height: 360px;
}

.invite-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: calc(100% - 36px);
  height: 48px;
  margin: 28px auto 12px;
  border: 1px solid rgba(255, 236, 157, 0.4);
  border-radius: 999px;
  background: linear-gradient(
    180deg,
    rgba(255, 249, 208, 0.98) 0%,
    rgba(255, 211, 88, 0.96) 48%,
    rgba(191, 120, 12, 1) 100%
  );
  box-shadow:
    0 10px 18px rgba(77, 17, 19, 0.45),
    inset 0 1px 0 rgba(255, 255, 255, 0.35);
  color: #7c2200;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 0.03em;
}

.winning {
  margin: 30px 16px 0;
  border: 1px solid #fdf07e;
  border-radius: 8px;
  background: linear-gradient(180deg, #270000 0%, #560007 100%);
}

.winning-report {
  height: 36px;
  margin: 6px 6px 0;
  justify-content: center;
  background-color: #5d1519;
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
  padding: 10px 12px;
  color: #fff;
  font-size: 12px;
}

.winning-item__reward {
  color: #ffea00;
  font-weight: 700;
}

:deep(.winning-overlay.van-overlay) {
  position: absolute;
  z-index: 12;
}

.light-bg2,
.light-bg,
.spins-tips {
  position: fixed;
  left: 50%;
}

.light-bg2 {
  top: 40%;
  z-index: 2002;
  width: 100%;
  max-width: 500px;
  transform: translate(-50%, -60%);
}

.light-bg2 img {
  width: 100%;
  animation: scale-element 2s linear infinite;
}

.light-bg {
  top: 40%;
  z-index: 2002;
  width: 282px;
  transform: translate(-41%, -63%);
}

.light-bg img {
  width: 100%;
}

.spins-tips {
  top: 40%;
  z-index: 9999;
  width: 80%;
  max-width: 360px;
  padding: 96px 0 112px;
  transform: translate(-50%, -60%);
  text-align: center;
  font-size: 14px;
}

.light-bg1 {
  position: absolute;
  inset: 0;
  z-index: -1;
  background-repeat: no-repeat;
  background-position: center;
  background-size: contain;
}

.close {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 20px;
  cursor: pointer;
}

.spins-btn {
  position: absolute;
  bottom: 28px;
  left: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 148px;
  height: 44px;
  padding: 0 22px;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffd601 0%, #ffa800 100%);
  color: #921111;
  font-weight: 700;
  transform: translateX(-50%);
}

.tip {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0 40px;
}

.tip h1 {
  margin: 0;
  color: #fff;
  font-size: 22px;
}

.tip__prize {
  width: 96px;
  margin: 12px 0;
}

.tip__amount {
  color: #ffe65c;
  font-size: 36px;
  font-weight: 800;
}

@keyframes scale-element {
  0% {
    transform: scale(0.96);
  }

  50% {
    transform: scale(1.04);
  }

  100% {
    transform: scale(0.96);
  }
}
</style>

<route lang="json5">
{
  name: 'Prize',
}
</route>
