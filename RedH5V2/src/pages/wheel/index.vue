<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showToast } from 'vant'
import {
  drawLottery,
  getLotteryChances,
  getPrizePoolBalance,
} from '@/api/user'
import { ppmxWheelRules } from '../ppmx-home/pageData'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'

definePage({
  name: 'wheel',
  meta: {
    titleKey: 'ppmx.wheel.title',
    shell: 'ppmx',
    tabbar: false,
  },
})

interface WheelPrize {
  amount: number
  color: string
  gold: boolean
}

const PRIZE_COLORS = ['#c8102e', '#138a52']
const PRIZE_GOLD_COLOR = '#f0bc3d'
const STATIC_WHEEL_AMOUNTS = [0, 8, 18, 38, 88, 188, 288, 588]
const SPIN_DURATION = 5400

const { t } = useI18n()
const userStore = useUserStore()
const prizePoolStore = usePrizePoolStore()
const { pickText } = usePpmxLocale()
const pageChromeRef = ref<InstanceType<typeof PpmxPageChrome> | null>(null)

const pageLoading = ref(false)
const pageError = ref('')
const jackpotError = ref(false)
const spins = ref(0)
const spinning = ref(false)
const rotation = ref(0)
const winAmount = ref<number | null>(null)
const showWinFx = ref(false)
const jackpotDisplay = ref<number | null>(null)
const wheelPrizes = ref<WheelPrize[]>([])
let spinTimer: number | undefined
let fxTimer: number | undefined

const jackpotRealtimeBalance = computed(() => prizePoolStore.luckyBalance)
const jackpotVisibleBalance = computed(() => jackpotRealtimeBalance.value ?? jackpotDisplay.value)
const showWheelJackpot = computed(() => userStore.isLogin)
const segmentSize = computed(() => (wheelPrizes.value.length > 0 ? 360 / wheelPrizes.value.length : 360))
const wheelEmpty = computed(() => !pageLoading.value && !pageError.value && wheelPrizes.value.length === 0)
const wheelBackground = computed(() => {
  if (wheelPrizes.value.length === 0) {
    return 'conic-gradient(from 0deg, rgba(255,255,255,.08), rgba(255,255,255,.02))'
  }

  const stops = wheelPrizes.value
    .map((prize, index) => `${prize.color} ${index * segmentSize.value}deg ${(index + 1) * segmentSize.value}deg`)
    .join(',')

  return `conic-gradient(from ${-segmentSize.value / 2}deg, ${stops})`
})
const ledDots = computed(() =>
  Array.from({ length: 24 }, (_, index) => {
    const angle = (index / 24) * Math.PI * 2

    return {
      id: index,
      x: 50 + 45.5 * Math.cos(angle),
      y: 50 + 45.5 * Math.sin(angle),
      fill: index % 2 ? '#fff3c4' : '#ffd86b',
      delay: `${(index * 0.08).toFixed(2)}s`,
    }
  }),
)
const spokes = computed(() =>
  wheelPrizes.value.map((_, index) => {
    const angle = (index * segmentSize.value - segmentSize.value / 2 - 90) * Math.PI / 180

    return {
      id: index,
      x: 50 + 50 * Math.cos(angle),
      y: 50 + 50 * Math.sin(angle),
    }
  }),
)

function formatPrize(amount: number) {
  return amount.toLocaleString('en-US', {
    maximumFractionDigits: 2,
  })
}

function normalizeAmount(value: number) {
  return Number(Number(value || 0).toFixed(2))
}

function buildWheelPrizes(amounts: number[]) {
  wheelPrizes.value = amounts
    .map(normalizeAmount)
    .filter(amount => Number.isFinite(amount) && amount >= 0)
    .map((amount, index) => ({
      amount,
      color: index === amounts.length - 1 ? PRIZE_GOLD_COLOR : PRIZE_COLORS[index % PRIZE_COLORS.length],
      gold: index === amounts.length - 1,
    }))
}

function resolvePrizeIndex(amount: number) {
  const normalized = normalizeAmount(amount)
  const exactIndex = wheelPrizes.value.findIndex(prize => normalizeAmount(prize.amount) === normalized)
  if (exactIndex >= 0) return exactIndex

  const zeroIndex = wheelPrizes.value.findIndex(prize => normalizeAmount(prize.amount) === 0)

  return zeroIndex >= 0 ? zeroIndex : 0
}

function applyStaticWheelData() {
  pageLoading.value = false
  pageError.value = ''
  jackpotError.value = false
  spins.value = 0
  jackpotDisplay.value = null
  buildWheelPrizes(STATIC_WHEEL_AMOUNTS)
}

async function loadWheelData() {
  if (!userStore.isLogin) {
    applyStaticWheelData()
    return
  }

  pageLoading.value = true
  pageError.value = ''
  jackpotError.value = false

  try {
    const [chanceData, jackpotResult] = await Promise.all([
      getLotteryChances(),
      getPrizePoolBalance('lucky').catch(() => {
        jackpotError.value = true
        return null
      }),
    ])

    buildWheelPrizes(Array.isArray(chanceData.amounts) ? chanceData.amounts : [])
    spins.value = Math.max(0, Number(chanceData.availableCount ?? 0))
    jackpotDisplay.value = jackpotResult ? Number(jackpotResult.balance ?? 0) : null
    if (jackpotResult) {
      prizePoolStore.updatePrizePoolBalance({
        poolCode: jackpotResult.poolCode || 'lucky',
        balance: Number(jackpotResult.balance ?? 0),
      })
    }
  } catch (error) {
    wheelPrizes.value = []
    spins.value = 0
    jackpotDisplay.value = null
    pageError.value = error instanceof Error ? error.message : t('ppmx.wheel.loadFailed')
  } finally {
    pageLoading.value = false
  }
}

function requireRegisterAuth() {
  return pageChromeRef.value?.requireRegisterAuth() ?? false
}

function clearTimers() {
  if (spinTimer) window.clearTimeout(spinTimer)
  if (fxTimer) window.clearTimeout(fxTimer)
  spinTimer = undefined
  fxTimer = undefined
}

async function spinWheel() {
  if (spinning.value) return
  if (!requireRegisterAuth()) return

  if (spins.value <= 0) {
    showToast(t('ppmx.wheel.noSpins'))
    return
  }

  if (wheelPrizes.value.length === 0) {
    showToast(t('ppmx.wheel.emptyPrizeConfig'))
    return
  }

  spinning.value = true
  showWinFx.value = false

  try {
    const result = await drawLottery()
    const awardAmount = normalizeAmount(Number(result.awardAmount ?? 0))
    const prizeIndex = resolvePrizeIndex(awardAmount)
    const jitter = (((Date.now() % 1000) / 1000) * 2 - 1) * (segmentSize.value * 0.22)
    const base = rotation.value - (rotation.value % 360)
    rotation.value = base + 360 * 7 - prizeIndex * segmentSize.value - jitter

    spinTimer = window.setTimeout(() => {
      spinning.value = false
      winAmount.value = awardAmount
      showWinFx.value = awardAmount > 0
      spins.value = Math.max(0, spins.value - 1)

      if (awardAmount > 0) {
        showToast(t('ppmx.wheel.winToast', { amount: formatPrize(awardAmount) }))
      } else {
        showToast(t('ppmx.wheel.drawMiss'))
      }

      fxTimer = window.setTimeout(() => {
        showWinFx.value = false
      }, 1800)

      void userStore.loadUserInfo()
      void loadWheelData()
    }, SPIN_DURATION)
  } catch (error) {
    spinning.value = false
    showToast(error instanceof Error ? error.message : t('ppmx.wheel.noSpins'))
    void loadWheelData()
  }
}

onMounted(() => {
  void loadWheelData()
})

watch(
  () => userStore.isLogin,
  () => {
    clearTimers()
    spinning.value = false
    showWinFx.value = false
    void loadWheelData()
  },
)

onBeforeUnmount(() => {
  clearTimers()
})
</script>

<template>
  <main class="ppmx-page ppmx-subpage ppmx-wheel-page">
    <PpmxPageChrome ref="pageChromeRef">
      <section class="ppmx-subpage__narrow">
        <header class="ppmx-subpage-head ppmx-subpage-head--center">
          <div class="ppmx-eyebrow">{{ t('ppmx.wheel.eyebrow') }}</div>
          <h1 class="ppmx-display">{{ t('ppmx.wheel.title') }}</h1>
          <p>{{ t('ppmx.wheel.description') }}</p>
        </header>

        <AppState
          v-if="pageLoading"
          class="ppmx-wheel-state"
          status="loading"
          :title="t('state.loading.title')"
          :message="t('state.loading.message')"
          skeleton-variant="wheel-page"
        />

        <AppState
          v-else-if="pageError"
          class="ppmx-wheel-state"
          status="error"
          :title="t('ppmx.wheel.loadFailed')"
          :message="pageError"
          :action-text="t('common.retry')"
          @retry="loadWheelData"
        />

        <AppState
          v-else-if="wheelEmpty"
          class="ppmx-wheel-state"
          status="empty"
          :title="t('ppmx.wheel.emptyPrizeConfig')"
          :message="t('ppmx.wheel.emptyMessage')"
          :action-text="t('common.retry')"
          @retry="loadWheelData"
        />

        <template v-else>
          <!-- <section v-if="showWheelJackpot" class="ppmx-wheel-jackpot">
            <span>{{ t('ppmx.wheel.jackpot') }}</span>
            <strong v-if="jackpotVisibleBalance !== null">${{ formatPrize(jackpotVisibleBalance) }}</strong>
            <strong v-else>{{ jackpotError ? t('ppmx.wheel.jackpotUnavailable') : '--' }}</strong>
          </section> -->

          <div class="ppmx-wheel-spins">
            <i class="fa-solid fa-ticket" />
            <span>{{ t('ppmx.wheel.availableSpins') }}</span>
            <b>{{ spins }}</b>
          </div>

          <section class="ppmx-wheel-stage" :aria-label="t('ppmx.wheel.stageAria')">
            <div class="ppmx-wheel-glow" />
            <div class="ppmx-wheel-pointer">
              <span class="ppmx-wheel-pointer-gem" />
            </div>
            <div class="ppmx-wheel-ring">
              <div
                class="ppmx-wheel"
                :class="{ 'is-spinning': spinning }"
                :style="{ background: wheelBackground, transform: `rotate(${rotation}deg)` }"
              >
                <svg class="ppmx-wheel-spokes" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
                  <line
                    v-for="spoke in spokes"
                    :key="spoke.id"
                    x1="50"
                    y1="50"
                    :x2="spoke.x.toFixed(2)"
                    :y2="spoke.y.toFixed(2)"
                    stroke="rgba(255,255,255,.22)"
                    stroke-width="0.5"
                  />
                </svg>

                <div
                  v-for="(prize, index) in wheelPrizes"
                  :key="`${prize.amount}-${index}`"
                  class="ppmx-wheel-seg-label"
                  :class="{ 'is-gold': prize.gold }"
                  :style="{ transform: `rotate(${index * segmentSize}deg)` }"
                >
                  <span class="ppmx-wheel-seg-text">
                    <svg class="ppmx-wheel-seg-coin" viewBox="0 0 100 100"><use href="#ppCoin" /></svg>
                    <span>{{ formatPrize(prize.amount) }}</span>
                  </span>
                </div>
              </div>

              <svg class="ppmx-wheel-leds" viewBox="0 0 100 100" aria-hidden="true">
                <circle
                  v-for="dot in ledDots"
                  :key="dot.id"
                  class="ppmx-wheel-led"
                  :cx="dot.x.toFixed(2)"
                  :cy="dot.y.toFixed(2)"
                  r="1.7"
                  :fill="dot.fill"
                  :style="{ animationDelay: dot.delay }"
                />
              </svg>

              <button
                class="ppmx-wheel-hub"
                :class="{ 'is-spinning': spinning }"
                type="button"
                :disabled="spinning"
                :aria-label="t('ppmx.wheel.hubAria')"
                @click="spinWheel"
              >
                <span class="ppmx-wheel-hub-in">
                  <svg class="ppmx-wheel-hub-coin" viewBox="0 0 100 100"><use href="#ppCoin" /></svg>
                  <span>{{ t('ppmx.wheel.hubText') }}</span>
                </span>
              </button>
            </div>
          </section>

          <button
            class="ppmx-btn-gold ppmx-wheel-spin-button"
            type="button"
            :disabled="spinning"
            @click="spinWheel"
          >
            <i class="fa-solid fa-arrows-spin" />
            <span>{{ t('ppmx.wheel.spinButton') }}</span>
          </button>

          <section class="ppmx-panel ppmx-wheel-rules">
            <h2>
              <i class="fa-solid fa-circle-info" />
              <span>{{ t('ppmx.wheel.rulesTitle') }}</span>
            </h2>
            <ul>
              <li v-for="rule in ppmxWheelRules" :key="pickText(rule)">
                <i class="fa-solid fa-check" />
                <span>{{ pickText(rule) }}</span>
              </li>
            </ul>
          </section>
        </template>
      </section>

      <div class="ppmx-wheel-winfx" :class="{ 'is-visible': showWinFx }" aria-live="polite">
        <div class="ppmx-wheel-winfx__bg" />
        <div class="ppmx-wheel-winfx__coins" aria-hidden="true">
          <svg v-for="coin in 12" :key="coin" viewBox="0 0 100 100"><use href="#ppCoin" /></svg>
        </div>
        <div v-if="winAmount !== null" class="ppmx-wheel-winfx__amount">
          +${{ formatPrize(winAmount) }}
        </div>
      </div>
    </PpmxPageChrome>
  </main>
</template>
