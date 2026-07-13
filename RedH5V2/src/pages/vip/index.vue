<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showToast } from 'vant'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'
import { ppmxAsset } from '../ppmx-home/assets'
import { VIP_SVGS } from '@/utils/vipBadge'
import { claimVipWeeklySalary, getVipLevels, getVipProgress, getVipWeeklySalary } from '@/api/vip'
import type { VipLevelItem, VipProgressData, VipWeeklySalaryStatus } from '@/api/vip'
import { HttpError } from '@/utils/http'

definePage({
  name: 'vip',
  meta: {
    titleKey: 'vip.title',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

const { pickText } = usePpmxLocale()
const { t } = useI18n()
const userStore = useUserStore()

// ---- i18n strings ----
const tx = {
  eyebrow:      { es: 'Programa VIP',                    en: 'VIP Program',                    zh: 'VIP 计划'        } as LocalizedText,
  title:        { es: 'Nivel VIP',                       en: 'VIP Level',                      zh: 'VIP 等级'        } as LocalizedText,
  sub:          { es: 'Apuesta para subir de nivel y desbloquea mejores recompensas.', en: 'Wager to level up and unlock better rewards.', zh: '投注升级，解锁更高奖励。' } as LocalizedText,
  current:      { es: 'Nivel actual',                    en: 'Current level',                  zh: '当前等级'        } as LocalizedText,
  points:       { es: 'puntos',                          en: 'points',                         zh: '积分'            } as LocalizedText,
  progress:     { es: 'de avance',                       en: 'progress',                       zh: '进度'            } as LocalizedText,
  tonext:       { es: 'Te faltan',                       en: 'You need',                       zh: '还差'            } as LocalizedText,
  tonext2:      { es: 'para el siguiente nivel.',        en: 'to reach the next level.',       zh: '即可升至下一级。' } as LocalizedText,
  maxed:        { es: '¡Nivel máximo alcanzado!',        en: 'Maximum level reached!',         zh: '已达最高等级！'  } as LocalizedText,
  nextgoal:     { es: 'Próxima meta',                    en: 'Next goal',                      zh: '下一目标'        } as LocalizedText,
  ladder:       { es: 'Ruta de niveles',                 en: 'Level path',                     zh: '等级路径'        } as LocalizedText,
  ladderHint:   { es: 'Desliza',                         en: 'Slide',                          zh: '左右滑动'        } as LocalizedText,
  reqTitle:     { es: 'Requisitos de nivel',             en: 'Level requirements',             zh: '等级要求'        } as LocalizedText,
  reqHint:      { es: 'Toca un nivel para ver sus requisitos.', en: 'Tap a level to see its requirements.', zh: '点击等级查看详情。' } as LocalizedText,
  reqAttain:    { es: 'Requisitos para alcanzar',        en: 'Requirements to reach',          zh: '达到条件'        } as LocalizedText,
  reqPoints:    { es: 'Apuesta válida acumulada',        en: 'Cumulative valid bet',           zh: '累计有效投注'    } as LocalizedText,
  reqDeposit:   { es: 'Depósito acumulado',              en: 'Cumulative deposit',             zh: '累计存款'        } as LocalizedText,
  reqBenefits:  { es: 'Beneficios desbloqueados',        en: 'Unlocked benefits',              zh: '解锁权益'        } as LocalizedText,
  reqGift:      { es: 'Bono de ascenso',                 en: 'Upgrade bonus',                  zh: '晋级奖励'        } as LocalizedText,
  reqHost:      { es: 'Anfitrión VIP',                   en: 'VIP host',                       zh: '专属客服'        } as LocalizedText,
  reqYes:       { es: 'Incluido',                        en: 'Included',                       zh: '包含'            } as LocalizedText,
  reqNo:        { es: 'No incluido',                     en: 'Not included',                   zh: '不包含'          } as LocalizedText,
  rwCashback:   { es: 'Salario semanal',                 en: 'Weekly cashback',                zh: '每周薪资'        } as LocalizedText,
  rwWdCount:    { es: 'Retiros por día',                 en: 'Withdrawals per day',            zh: '每日提现次数'    } as LocalizedText,
  rwWdMax:      { es: 'Retiro máx. por operación',      en: 'Max withdrawal per op.',         zh: '单次提现上限'    } as LocalizedText,
  rwTimes:      { es: 'veces',                           en: 'times',                          zh: '次'              } as LocalizedText,
  youarehere:   { es: 'Actual',                          en: 'Current',                        zh: '当前'            } as LocalizedText,
  statusDone:   { es: 'Logrado',                         en: 'Achieved',                       zh: '已达成'          } as LocalizedText,
  statusLocked: { es: 'Bloqueado',                       en: 'Locked',                         zh: '未解锁'          } as LocalizedText,
  rewards:      { es: 'Beneficios por nivel',            en: 'Benefits by level',              zh: '各等级权益'      } as LocalizedText,
  rewardsSub:   { es: 'Cada nivel desbloquea estos beneficios.', en: 'Each level unlocks these benefits.', zh: '每个等级解锁以下权益。' } as LocalizedText,
  salEyebrow:   { es: 'Cobro semanal',                   en: 'Weekly payout',                  zh: '每周薪资'        } as LocalizedText,
  salTitle:     { es: 'Salario semanal VIP',             en: 'VIP weekly salary',              zh: 'VIP 每周薪资'    } as LocalizedText,
  salNote:      { es: 'Reclámalo una vez por semana.',   en: 'Claim it once a week.',          zh: '每周可领取一次。' } as LocalizedText,
  salClaim:     { es: 'Reclamar',                        en: 'Claim',                          zh: '立即领取'        } as LocalizedText,
  salClaimed:   { es: 'Reclamado',                       en: 'Claimed',                        zh: '已领取'          } as LocalizedText,
  salNext:      { es: 'Ya cobrado esta semana. Vuelve el próximo lunes.', en: 'Already collected this week. Come back next Monday.', zh: '本周已领取，下周一可再次领取。' } as LocalizedText,
  salOk:        { es: '¡Salario semanal cobrado!',       en: 'Weekly salary claimed!',         zh: '每周薪资已到账！' } as LocalizedText,
  salUnavailable: { es: 'Tu nivel actual no tiene salario semanal.', en: 'Your current level has no weekly salary.', zh: '当前等级暂无每周薪资。' } as LocalizedText,
}

// ---- types ----
interface VipLevel { key: string; pts: number; name: LocalizedText }
interface VipReq   {
  dep: number; wager: number; gift: number
  wd: number; wdmax: number
  cashback: number; host: boolean
}

// ---- static rewards grid ----
const vipRewards = [
  { id: 'bonus',    icon: 'fa-gift',       label: { es: 'Bono de recarga',        en: 'Reload bonus',          zh: '充值奖金'     } as LocalizedText },
  { id: 'cashback', icon: 'fa-percent',    label: { es: 'Salario semanal',        en: 'Weekly cashback',       zh: '每周薪资'     } as LocalizedText },
  { id: 'withdraw', icon: 'fa-bolt',       label: { es: 'Retiros prioritarios',   en: 'Priority withdrawals',  zh: '优先提现'     } as LocalizedText },
  { id: 'host',     icon: 'fa-user-tie',   label: { es: 'Anfitrión VIP',          en: 'VIP host',              zh: '专属客服经理' } as LocalizedText },
  { id: 'wdcount',  icon: 'fa-repeat',     label: { es: 'Retiros por día',        en: 'Daily withdrawals',     zh: '每日提现次数' } as LocalizedText },
  { id: 'wdmax',    icon: 'fa-arrow-up',   label: { es: 'Retiro máx. por op.',   en: 'Max per withdrawal',    zh: '单次提现上限' } as LocalizedText },
]

// ---- state ----
const vipLoading = ref(false)
const vipError   = ref('')
const vipLoaded  = ref(false)

// ---- weekly salary ----
const chestImage = ppmxAsset('vip-chest.png')
const salary = ref<VipWeeklySalaryStatus | null>(null)
const claimingSalary = ref(false)
const showSalaryCard = computed(() => salary.value !== null)
const salaryClaimed = computed(() => Boolean(salary.value?.claimed))
const salaryClaimable = computed(() => Boolean(salary.value?.claimable))
const salaryAmount = computed(() => Number(salary.value?.weeklySalary ?? 0))
const salaryAmountText = computed(() =>
  `$${salaryAmount.value.toLocaleString('en-US', { maximumFractionDigits: 2 })}`,
)
const salaryNote = computed(() => {
  if (salaryClaimed.value) return tx.salNext
  if (salaryAmount.value <= 0) return tx.salUnavailable
  return tx.salNote
})

// ---- ladder nav ----
const ladderScrollRef = ref<HTMLElement | null>(null)
const ladderAtStart = ref(true)
const ladderAtEnd = ref(false)

function updateLadderNav() {
  const el = ladderScrollRef.value
  if (!el) return
  const max = el.scrollWidth - el.clientWidth - 1
  ladderAtStart.value = el.scrollLeft <= 2
  ladderAtEnd.value = el.scrollLeft >= max || max <= 0
}

function scrollLadder(dir: number) {
  const el = ladderScrollRef.value
  if (!el) return
  el.scrollBy({ left: dir * Math.max(190, el.clientWidth * 0.7), behavior: 'smooth' })
}

const apiLevels   = ref<VipLevelItem[]>([])
const apiProgress = ref<VipProgressData | null>(null)

// ---- derived display data ----
function levelNameToIndex(name: string | null | undefined): number {
  if (!name) return 0
  const n = parseInt(name.replace(/[^0-9]/g, ''), 10)
  return Number.isFinite(n) ? n : 0
}

const vipLevels = computed<VipLevel[]>(() => {
  if (!apiLevels.value.length) return []
  return [...apiLevels.value]
    .filter(l => l.status === 1)
    .sort((a, b) => a.level - b.level)
    .slice(0, 11)
    .map(l => ({
      key:  l.levelName.toLowerCase(),
      pts:  l.totalValidBet ?? 0,
      name: { es: l.levelName, en: l.levelName, zh: l.levelName } as LocalizedText,
    }))
})

const vipReqs = computed<VipReq[]>(() => {
  if (!apiLevels.value.length) return []
  return [...apiLevels.value]
    .filter(l => l.status === 1)
    .sort((a, b) => a.level - b.level)
    .slice(0, 11)
    .map(l => ({
      dep:      l.totalRechargeAmount ?? 0,
      wager:    l.totalValidBet ?? 0,
      gift:     l.upgradeBonusAmount ?? 0,
      wd:       l.totalWithdrawCount ?? 0,
      wdmax:    l.maxWithdrawAmount ?? 0,
      cashback: l.weeklySalary ?? 0,
      host:     Boolean(l.exclusiveService),
    }))
})

const currentLevelIndex = computed(() => {
  const level = apiProgress.value?.currentLevel?.level
  if (typeof level === 'number' && level > 0) return level - 1
  return levelNameToIndex(apiProgress.value?.currentLevel?.levelName)
})

const currentPoints = computed(() => Number(apiProgress.value?.currentValue ?? 0))

const selectedIndex = ref(0)

// keep selectedIndex pointing at current level after load
watch(currentLevelIndex, (idx) => { selectedIndex.value = idx }, { immediate: true })

// ---- computed display values ----
const currentLevel  = computed(() => vipLevels.value[currentLevelIndex.value] ?? null)
const nextLevel     = computed(() => vipLevels.value[currentLevelIndex.value + 1] ?? null)
const selectedLevel = computed(() => vipLevels.value[selectedIndex.value] ?? null)
const selectedReq   = computed(() => vipReqs.value[selectedIndex.value] ?? null)

const progressPercent = computed(() => {
  const serverProgress = Number(apiProgress.value?.progress)
  if (Number.isFinite(serverProgress)) return Math.min(100, Math.max(0, Math.round(serverProgress)))

  const next = nextLevel.value
  const cur  = currentLevel.value
  if (!next || !cur) return 100
  const span = next.pts - cur.pts
  if (span <= 0) return 100
  return Math.min(100, Math.round(((currentPoints.value - cur.pts) / span) * 100))
})

const remainingPoints = computed(() => {
  const next = nextLevel.value
  if (!next) return 0
  return Math.max(0, next.pts - currentPoints.value)
})

function isUnlocked(index: number) { return index <= currentLevelIndex.value }

function fmt(n: number) { return n < 0 ? '♾️' : n.toLocaleString('en-US') }
function fmtUsd(n: number) { return n < 0 ? '♾️' : n === 0 ? '$0' : '$' + n.toLocaleString('en-US') }

// ---- data loading ----
async function loadVipData() {
  if (!userStore.isLogin) {
    vipLoading.value = false
    vipError.value   = ''
    vipLoaded.value  = false
    apiLevels.value   = []
    apiProgress.value = null
    salary.value      = null
    return
  }

  vipLoading.value = true
  vipError.value   = ''

  try {
    const [levelsData, progressData] = await Promise.all([
      getVipLevels(),
      getVipProgress(),
    ])
    apiLevels.value   = levelsData
    apiProgress.value = progressData
    vipLoaded.value   = true
    void loadSalary()
  } catch (err) {
    apiLevels.value   = []
    apiProgress.value = null
    vipLoaded.value   = false
    vipError.value    = err instanceof Error ? err.message : t('error.requestFailed')
  } finally {
    vipLoading.value = false
  }
}

async function loadSalary() {
  try {
    salary.value = await getVipWeeklySalary()
  } catch {
    salary.value = null
  }
}

async function claimSalary() {
  if (claimingSalary.value || !salary.value?.claimable) return

  claimingSalary.value = true
  try {
    await claimVipWeeklySalary()
    showToast({ type: 'success', message: pickText(tx.salOk) })
    await Promise.allSettled([loadSalary(), userStore.loadUserInfo()])
  } catch (err) {
    const code = err instanceof HttpError ? err.message : ''
    if (code === 'vip_weekly_salary_already_claimed') {
      showToast(pickText(tx.salNext))
      void loadSalary()
    } else if (code === 'vip_weekly_salary_unavailable') {
      showToast(pickText(tx.salUnavailable))
      void loadSalary()
    } else {
      showToast(err instanceof Error ? err.message : t('error.requestFailed'))
    }
  } finally {
    claimingSalary.value = false
  }
}

onMounted(() => { void loadVipData() })
watch(() => userStore.isLogin, () => { void loadVipData() })
watch(vipLevels, () => { void nextTick(updateLadderNav) })
</script>

<template>
  <main class="ppmx-page ppmx-subpage ppmx-vip-page">
    <PpmxPageChrome>
      <section class="ppmx-subpage__compact ppmx-vip-stack">

        <!-- Header (always visible) -->
        <header class="ppmx-subpage-head ppmx-vip-head reveal">
          <div class="ppmx-eyebrow">{{ pickText(tx.eyebrow) }}</div>
          <h1 class="ppmx-display">{{ pickText(tx.title) }}</h1>
          <p>{{ pickText(tx.sub) }}</p>
        </header>

        <!-- Loading state -->
        <AppState
          v-if="vipLoading"
          status="loading"
          :title="t('state.loading.title')"
          :message="t('state.loading.message')"
        />

        <!-- Error state -->
        <AppState
          v-else-if="vipError"
          status="error"
          :title="t('state.error.title')"
          :message="vipError"
          :action-text="t('common.retry')"
          @retry="loadVipData"
        />

        <template v-else-if="vipLoaded && currentLevel">

          <!-- Hero: current level + progress -->
          <div class="vip2-hero reveal">
            <span class="vip2-hero__glow" />
            <div class="vip2-hero__head">
              <span class="vip2-hero__badge" v-html="VIP_SVGS[currentLevelIndex]" />
              <div class="vip2-hero__identity">
                <p class="vip2-hero__eyebrow">{{ pickText(tx.current) }}</p>
                <p class="vip2-hero__name">{{ pickText(currentLevel.name) }}</p>
              </div>
              <div class="vip2-hero__pts-block">
                <p class="vip2-hero__pts-num">{{ fmt(currentPoints) }}</p>
                <p class="vip2-hero__pts-lbl">{{ pickText(tx.points) }}</p>
              </div>
            </div>

            <div class="vip2-hero__prog">
              <div class="vip2-hero__prog-meta">
                <span><b>{{ progressPercent }}%</b> {{ pickText(tx.progress) }}</span>
                <span v-if="nextLevel">
                  <i class="fa-solid fa-arrow-right-long" style="font-size:10px;opacity:.7" />
                  {{ pickText(nextLevel.name) }} · {{ fmt(nextLevel.pts) }} pts
                </span>
              </div>
              <div class="vip2-hero__track" role="progressbar" :aria-valuenow="progressPercent" aria-valuemin="0" aria-valuemax="100">
                <div class="vip2-hero__bar" :style="{ width: `${progressPercent}%` }">
                  <span class="vip2-hero__bar-shine" />
                </div>
              </div>
              <p v-if="nextLevel" class="vip2-hero__foot">
                <i class="fa-solid fa-bolt vip2-hero__foot-ic" />
                {{ pickText(tx.tonext) }}
                <b>{{ fmt(remainingPoints) }} pts</b>
                {{ pickText(tx.tonext2) }}
              </p>
              <p v-else class="vip2-hero__foot">{{ pickText(tx.maxed) }}</p>

              <div v-if="nextLevel" class="vip2-hero__next">
                <span class="vip2-hero__next-tile" v-html="VIP_SVGS[currentLevelIndex + 1]" />
                <div>
                  <p class="vip2-hero__next-lbl">{{ pickText(tx.nextgoal) }}</p>
                  <p class="vip2-hero__next-name">{{ pickText(nextLevel.name) }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Weekly salary -->
          <div
            v-if="showSalaryCard"
            class="vip-salary reveal"
            :class="{ 'is-claimed': salaryClaimed }"
          >
            <span class="vip-salary-glow" />
            <div class="vip-salary-chest">
              <span class="vip-salary-coin c1" />
              <span class="vip-salary-coin c2" />
              <span class="vip-salary-coin c3" />
              <img class="vip-chest-svg" :src="chestImage" alt="" aria-hidden="true" />
            </div>
            <div class="vip-salary-info">
              <p class="vip-salary-eyebrow">
                <i class="fa-solid fa-calendar-week" />
                {{ pickText(tx.salEyebrow) }}
              </p>
              <p class="vip-salary-title">{{ pickText(tx.salTitle) }}</p>
              <p class="vip-salary-amount">{{ salaryAmountText }}</p>
              <p class="vip-salary-note">{{ pickText(salaryNote) }}</p>
            </div>
            <button
              type="button"
              class="vip-salary-btn"
              :disabled="!salaryClaimable || claimingSalary"
              @click="claimSalary"
            >
              <i v-if="claimingSalary" class="fa-solid fa-spinner fa-spin" />
              <span>{{ pickText(salaryClaimed ? tx.salClaimed : tx.salClaim) }}</span>
            </button>
          </div>

          <!-- Tier ladder -->
          <div class="vip2-ladder-wrap reveal">
            <div class="vip2-ladder-header">
              <h3 class="vip2-ladder-title">
                <i class="fa-solid fa-route" />
                {{ pickText(tx.ladder) }}
              </h3>
              <span class="vip2-ladder-hint">
                <i class="fa-solid fa-arrows-left-right" />
                {{ pickText(tx.ladderHint) }}
              </span>
            </div>
            <div class="vip2-ladder-scroller">
              <button
                type="button"
                class="vip2-ladder-nav vip2-ladder-prev"
                :disabled="ladderAtStart"
                aria-label="Previous"
                @click="scrollLadder(-1)"
              >
                <i class="fa-solid fa-chevron-left" />
              </button>
              <div ref="ladderScrollRef" class="vip2-ladder-scroll" @scroll="updateLadderNav">
                <div class="vip2-ladder">
                  <button
                    v-for="level, i in vipLevels"
                    :key="level.key"
                    type="button"
                    class="vip2-ladder-item"
                    :class="{
                      'is-done': isUnlocked(i) && i !== currentLevelIndex,
                      'is-now': i === currentLevelIndex,
                      'is-locked': !isUnlocked(i),
                      'is-view': i === selectedIndex,
                    }"
                    @click="selectedIndex = i"
                  >
                    <span class="vip2-ladder-tile" v-html="VIP_SVGS[i]" />
                    <span class="vip2-ladder-name">{{ pickText(level.name) }}</span>
                    <span class="vip2-ladder-pts">{{ fmt(level.pts) }} pts</span>
                  </button>
                </div>
              </div>
              <button
                type="button"
                class="vip2-ladder-nav vip2-ladder-next"
                :disabled="ladderAtEnd"
                aria-label="Next"
                @click="scrollLadder(1)"
              >
                <i class="fa-solid fa-chevron-right" />
              </button>
            </div>
          </div>

          <!-- Requirements panel -->
          <div v-if="selectedLevel && selectedReq" class="vip2-req-wrap reveal">
            <div class="vip2-req-head">
              <h3 class="vip2-req-title">
                <i class="fa-solid fa-clipboard-check" />
                {{ pickText(tx.reqTitle) }}
              </h3>
              <p class="vip2-req-hint">
                <i class="fa-solid fa-hand-pointer" style="font-size:10px" />
                {{ pickText(tx.reqHint) }}
              </p>
            </div>

            <div
              class="vip2-req-card"
              :class="{ 'is-now': selectedIndex === currentLevelIndex }"
            >
              <div class="vip2-req-card__head">
                <span class="vip2-req-badge" v-html="VIP_SVGS[selectedIndex]" />
                <div class="vip2-req-card__meta">
                  <p class="vip2-req-card__name">{{ pickText(selectedLevel.name) }}</p>
                  <p class="vip2-req-card__pts">{{ fmt(selectedLevel.pts) }} pts</p>
                </div>
                <span
                  class="vip2-status"
                  :class="selectedIndex === currentLevelIndex ? 's-now' : isUnlocked(selectedIndex) ? 's-done' : 's-lock'"
                >
                  <i class="fa-solid" :class="selectedIndex === currentLevelIndex ? 'fa-location-dot' : isUnlocked(selectedIndex) ? 'fa-circle-check' : 'fa-lock'" />
                  {{ pickText(selectedIndex === currentLevelIndex ? tx.youarehere : isUnlocked(selectedIndex) ? tx.statusDone : tx.statusLocked) }}
                </span>
              </div>

              <!-- Attainment requirements -->
              <div class="vip2-req-group">
                <p class="vip2-req-gtitle">
                  <i class="fa-solid fa-arrow-up-right-dots" />
                  {{ pickText(tx.reqAttain) }}
                </p>
                <div class="vip2-req-row">
                  <span class="vip2-req-ic"><i class="fa-solid fa-dice" /></span>
                  <span class="vip2-req-label">{{ pickText(tx.reqPoints) }}</span>
                  <span class="vip2-req-val">{{ fmtUsd(selectedReq.wager) }}</span>
                </div>
                <div class="vip2-req-row">
                  <span class="vip2-req-ic"><i class="fa-solid fa-wallet" /></span>
                  <span class="vip2-req-label">{{ pickText(tx.reqDeposit) }}</span>
                  <span class="vip2-req-val">{{ fmtUsd(selectedReq.dep) }}</span>
                </div>
              </div>

              <!-- Benefits -->
              <div class="vip2-req-group">
                <p class="vip2-req-gtitle">
                  <i class="fa-solid fa-gift" />
                  {{ pickText(tx.reqBenefits) }}
                </p>
                <div class="vip2-req-row">
                  <span class="vip2-req-ic"><i class="fa-solid fa-percent" /></span>
                  <span class="vip2-req-label">{{ pickText(tx.rwCashback) }}</span>
                  <span class="vip2-req-val">{{ fmtUsd(selectedReq.cashback) }}</span>
                </div>
                <div class="vip2-req-row">
                  <span class="vip2-req-ic"><i class="fa-solid fa-sack-dollar" /></span>
                  <span class="vip2-req-label">{{ pickText(tx.reqGift) }}</span>
                  <span class="vip2-req-val">{{ fmtUsd(selectedReq.gift) }}</span>
                </div>
                <div class="vip2-req-row">
                  <span class="vip2-req-ic"><i class="fa-solid fa-money-bill-transfer" /></span>
                  <span class="vip2-req-label">{{ pickText(tx.rwWdCount) }}</span>
                  <span class="vip2-req-val">{{ fmt(selectedReq.wd) }} {{ pickText(tx.rwTimes) }}</span>
                </div>
                <div class="vip2-req-row">
                  <span class="vip2-req-ic"><i class="fa-solid fa-money-bill-wave" /></span>
                  <span class="vip2-req-label">{{ pickText(tx.rwWdMax) }}</span>
                  <span class="vip2-req-val">{{ fmtUsd(selectedReq.wdmax) }}</span>
                </div>
                <div class="vip2-req-row">
                  <span class="vip2-req-ic"><i class="fa-solid fa-user-tie" /></span>
                  <span class="vip2-req-label">{{ pickText(tx.reqHost) }}</span>
                  <span
                    class="vip2-req-val"
                    :class="selectedReq.host ? 'vip2-req-yes' : 'vip2-req-no'"
                  >
                    <i class="fa-solid" :class="selectedReq.host ? 'fa-check' : 'fa-minus'" />
                    {{ pickText(selectedReq.host ? tx.reqYes : tx.reqNo) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Rewards overview -->
          <section class="ppmx-vip-rewards reveal">
            <div class="ppmx-vip-rewards__head">
              <h2>{{ pickText(tx.rewards) }}</h2>
              <p>{{ pickText(tx.rewardsSub) }}</p>
            </div>
            <div class="ppmx-vip-reward-grid">
              <article v-for="reward in vipRewards" :key="reward.id" class="ppmx-vip-reward-card">
                <span><i class="fa-solid" :class="reward.icon" /></span>
                <p>{{ pickText(reward.label) }}</p>
              </article>
            </div>
          </section>

        </template>

      </section>
    </PpmxPageChrome>
  </main>
</template>
