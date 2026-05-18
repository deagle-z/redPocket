<script setup lang="ts">
import { showToast } from 'vant'
import { useRouter } from 'vue-router'
import { getTrialLuckyFlowLotteryReward } from '@/api/trial'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { truncate2 } from '@/utils/currency'
import { safeBack } from '@/utils/navigation'

const { t } = useI18n()
const router = useRouter()

const loading = ref(false)
const progress = reactive({
  enabled: false,
  thresholdAmount: 0,
  rewardCount: 0,
  totalFlow: 0,
  remainingFlow: 0,
  progressPercent: 0,
  rewarded: false,
  canReward: false,
  availableRewardCount: 0,
  drawn: false,
  freeLotteryCount: 0,
})

const progressValue = computed(() => Math.max(0, Math.min(100, Number(progress.progressPercent || 0))))
const statusText = computed(() => {
  if (!progress.enabled)
    return t('trialFlowLotteryPage.statusDisabled')
  if (progress.rewarded)
    return t('trialFlowLotteryPage.statusRewarded')
  if (progress.canReward)
    return t('trialFlowLotteryPage.statusReady')
  return t('trialFlowLotteryPage.statusDoing')
})
const drawnText = computed(() => progress.drawn ? t('trialFlowLotteryPage.drawnYes') : t('trialFlowLotteryPage.drawnNo'))
const rewardText = computed(() => progress.rewarded ? t('trialFlowLotteryPage.rewardedYes') : t('trialFlowLotteryPage.rewardedNo'))

function formatAmount(value: number) {
  return truncate2(Number(value || 0)).toFixed(2)
}

function goBack() {
  safeBack(router)
}

function goDemo() {
  router.push('/demo')
}

function goPrize() {
  router.push('/prize')
}

async function loadProgress() {
  if (loading.value)
    return
  loading.value = true
  try {
    const { data } = await getTrialLuckyFlowLotteryReward()
    Object.assign(progress, {
      enabled: Boolean(data?.enabled),
      thresholdAmount: Number(data?.thresholdAmount || 0),
      rewardCount: Number(data?.rewardCount || 0),
      totalFlow: Number(data?.totalFlow || 0),
      remainingFlow: Number(data?.remainingFlow || 0),
      progressPercent: Number(data?.progressPercent || 0),
      rewarded: Boolean(data?.rewarded),
      canReward: Boolean(data?.canReward),
      availableRewardCount: Number(data?.availableRewardCount || 0),
      drawn: Boolean(data?.drawn),
      freeLotteryCount: Number(data?.freeLotteryCount || 0),
    })
  }
  catch {
    showToast(t('trialFlowLotteryPage.loadFailed'))
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  loadProgress()
})
</script>

<template>
  <main class="trial-flow-page">
    <AppPageHeader :title="t('trialFlowLotteryPage.title')" @back="goBack" />

    <section class="hero-panel">
      <div class="hero-copy">
        <span class="hero-eyebrow">{{ t('trialFlowLotteryPage.eyebrow') }}</span>
        <h1>{{ t('trialFlowLotteryPage.heroTitle') }}</h1>
        <p>{{ t('trialFlowLotteryPage.heroDesc', { count: progress.rewardCount || '-' }) }}</p>
      </div>
      <div class="hero-medal">
        <span>{{ progress.rewardCount || 0 }}</span>
        <small>{{ t('trialFlowLotteryPage.chancesUnit') }}</small>
      </div>
    </section>

    <section class="progress-panel">
      <div class="panel-head">
        <div>
          <span class="panel-kicker">{{ t('trialFlowLotteryPage.progressTitle') }}</span>
          <strong>{{ statusText }}</strong>
        </div>
        <button type="button" class="refresh-btn" :disabled="loading" @click="loadProgress">
          <van-icon name="replay" />
        </button>
      </div>

      <van-progress
        class="flow-progress"
        :percentage="progressValue"
        :show-pivot="false"
        color="linear-gradient(90deg, #ffcf62 0%, #f7a923 100%)"
        track-color="rgba(255, 231, 191, 0.16)"
        stroke-width="10"
      />
      <div class="progress-meta">
        <span>{{ formatAmount(progress.totalFlow) }}</span>
        <span>{{ formatAmount(progress.thresholdAmount) }}</span>
      </div>

      <div class="stat-grid">
        <div class="stat-cell">
          <span>{{ t('trialFlowLotteryPage.totalFlow') }}</span>
          <strong>{{ formatAmount(progress.totalFlow) }}</strong>
        </div>
        <div class="stat-cell">
          <span>{{ t('trialFlowLotteryPage.targetFlow') }}</span>
          <strong>{{ formatAmount(progress.thresholdAmount) }}</strong>
        </div>
        <div class="stat-cell">
          <span>{{ t('trialFlowLotteryPage.remainingFlow') }}</span>
          <strong>{{ formatAmount(progress.remainingFlow) }}</strong>
        </div>
        <div class="stat-cell">
          <span>{{ t('trialFlowLotteryPage.freeChances') }}</span>
          <strong>{{ progress.freeLotteryCount }}</strong>
        </div>
      </div>
    </section>

    <section class="status-panel">
      <div class="status-row">
        <span>{{ t('trialFlowLotteryPage.rewardStatus') }}</span>
        <strong :class="{ done: progress.rewarded }">{{ rewardText }}</strong>
      </div>
      <div class="status-row">
        <span>{{ t('trialFlowLotteryPage.drawStatus') }}</span>
        <strong :class="{ done: progress.drawn }">{{ drawnText }}</strong>
      </div>
    </section>

    <div class="action-row">
      <button type="button" class="secondary-action" @click="goDemo">
        {{ t('trialFlowLotteryPage.goTrial') }}
      </button>
      <button type="button" class="primary-action" @click="goPrize">
        {{ t('trialFlowLotteryPage.goDraw') }}
      </button>
    </div>
  </main>
</template>

<style scoped>
.trial-flow-page {
  min-height: 100vh;
  background-image:
    radial-gradient(circle at 18% 12%, rgba(212, 175, 55, 0.18), transparent 28%),
    radial-gradient(circle at 82% 84%, rgba(255, 215, 0, 0.12), transparent 24%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #240000 60%, #150000 100%);
  padding: 0 12px calc(24px + env(safe-area-inset-bottom));
  color: #fff0c9;
}

.hero-panel,
.progress-panel,
.status-panel {
  border: 1px solid rgba(212, 175, 55, 0.34);
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.28),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.hero-panel {
  display: flex;
  align-items: center;
  gap: 14px;
  min-height: 150px;
  border-radius: 18px;
  padding: 18px 16px;
  overflow: hidden;
}

.hero-copy {
  flex: 1;
  min-width: 0;
}

.hero-eyebrow,
.panel-kicker {
  display: block;
  color: #ffd98b;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.hero-copy h1 {
  margin: 8px 0 8px;
  font-size: 24px;
  line-height: 1.12;
  font-weight: 900;
  color: #fff3d2;
}

.hero-copy p {
  margin: 0;
  color: rgba(255, 229, 186, 0.7);
  font-size: 13px;
  line-height: 1.5;
}

.hero-medal {
  width: 86px;
  height: 86px;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  box-shadow:
    0 12px 22px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.35);
}

.hero-medal span {
  font-size: 28px;
  line-height: 1;
  font-weight: 900;
}

.hero-medal small {
  margin-top: 4px;
  font-size: 11px;
  font-weight: 800;
}

.progress-panel {
  margin-top: 14px;
  border-radius: 16px;
  padding: 16px;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.panel-head strong {
  display: block;
  margin-top: 6px;
  color: #fff3d2;
  font-size: 18px;
}

.refresh-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid rgba(212, 175, 55, 0.34);
  background: rgba(255, 248, 214, 0.08);
  color: #ffd98b;
}

.refresh-btn:disabled {
  opacity: 0.55;
}

.flow-progress {
  margin-top: 18px;
}

.progress-meta {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  color: rgba(255, 229, 186, 0.62);
  font-size: 12px;
}

.stat-grid {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.stat-cell {
  min-height: 72px;
  border-radius: 12px;
  border: 1px solid rgba(212, 175, 55, 0.16);
  background: rgba(255, 248, 214, 0.05);
  padding: 12px;
}

.stat-cell span {
  display: block;
  color: rgba(255, 229, 186, 0.62);
  font-size: 12px;
}

.stat-cell strong {
  display: block;
  margin-top: 8px;
  color: #ffd87f;
  font-size: 18px;
  line-height: 1;
}

.status-panel {
  margin-top: 14px;
  border-radius: 16px;
  overflow: hidden;
}

.status-row {
  min-height: 54px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.status-row + .status-row {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.status-row span {
  color: rgba(255, 229, 186, 0.72);
  font-size: 14px;
}

.status-row strong {
  color: rgba(255, 229, 186, 0.62);
  font-size: 14px;
}

.status-row strong.done {
  color: #ffd87f;
}

.action-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 16px;
}

.primary-action,
.secondary-action {
  height: 46px;
  border-radius: 24px;
  font-size: 15px;
  font-weight: 800;
}

.primary-action {
  border: 1px solid rgba(255, 248, 214, 0.46);
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
}

.secondary-action {
  border: 1px solid rgba(212, 175, 55, 0.42);
  background: rgba(255, 248, 214, 0.08);
  color: #fff0c9;
}

@media (max-width: 360px) {
  .hero-panel {
    align-items: flex-start;
  }

  .hero-medal {
    width: 72px;
    height: 72px;
  }

  .hero-copy h1 {
    font-size: 21px;
  }
}
</style>

<route lang="json5">
{
  name: 'TrialFlowLottery',
}
</route>
