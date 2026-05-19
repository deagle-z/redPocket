<script setup lang="ts">
import { showToast } from 'vant'
import AppPageHeader from '@/components/AppPageHeader.vue'
import type { CheckInRecordItem, CheckInStatusResp } from '@/api/user'
import { doCheckIn, getCheckInRecords, getCheckInStatus } from '@/api/user'
import { CURRENCY_CODE, truncate2 } from '@/utils/currency'
import { safeBack } from '@/utils/navigation'

const router = useRouter()
const { t } = useI18n()

const status = ref<CheckInStatusResp | null>(null)
const records = ref<CheckInRecordItem[]>([])
const loading = ref(false)
const signing = ref(false)
const loadFailed = ref(false)

const totalDays = computed(() => Number(status.value?.totalCheckInDays || 0))
const todayChecked = computed(() => !!status.value?.todayChecked)
const completed = computed(() => !!status.value?.completed)
const rewards = computed(() => status.value?.rewards || [])
const nextRewardAmount = computed(() => Number(status.value?.nextRewardAmount || 0))
const nextSeq = computed(() => Number(status.value?.nextSeq || 1))

const buttonText = computed(() => {
  if (todayChecked.value)
    return t('checkInPage.checkedToday')
  if (completed.value)
    return t('checkInPage.completed')
  return t('checkInPage.signNow')
})

function formatAmount(value: number) {
  return `${truncate2(Number(value || 0)).toFixed(2)} ${CURRENCY_CODE}`
}

function rewardState(index: number) {
  const seq = index + 1
  if (seq <= totalDays.value)
    return 'claimed'
  if (!todayChecked.value && seq === nextSeq.value && !completed.value)
    return 'current'
  return 'locked'
}

function formatRecordDate(value: string) {
  if (!value)
    return ''
  if (/^\d{4}-\d{2}-\d{2}$/.test(value))
    return value
  const date = new Date(value)
  if (Number.isNaN(date.getTime()))
    return value
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

async function loadCheckInData() {
  loading.value = true
  loadFailed.value = false
  try {
    const [statusResult, recordsResult] = await Promise.all([
      getCheckInStatus(),
      getCheckInRecords(30),
    ])
    status.value = statusResult.data
    records.value = recordsResult.data || []
  }
  catch {
    loadFailed.value = true
    status.value = null
    records.value = []
  }
  finally {
    loading.value = false
  }
}

async function handleCheckIn() {
  if (signing.value || todayChecked.value || completed.value)
    return

  signing.value = true
  try {
    const { data } = await doCheckIn()
    showToast(t('checkInPage.successToast', { amount: formatAmount(Number(data?.rewardAmount || 0)) }))
    await loadCheckInData()
  }
  finally {
    signing.value = false
  }
}

function goBack() {
  safeBack(router)
}

onMounted(() => {
  void loadCheckInData()
})
</script>

<template>
  <main class="checkin-page">
    <AppPageHeader :title="t('checkInPage.title')" @back="goBack" />

    <van-loading v-if="loading && !status" class="page-loading" color="#ffd98b" />

    <template v-else>
      <section class="hero-panel">
        <div class="hero-copy">
          <p class="eyebrow">
            {{ t('checkInPage.eyebrow') }}
          </p>
          <h1>{{ t('checkInPage.heroTitle') }}</h1>
          <p class="hero-subtitle">
            {{ t('checkInPage.heroSubtitle') }}
          </p>
        </div>

        <div class="coin-stack" aria-hidden="true">
          <span class="coin coin-a" />
          <span class="coin coin-b" />
          <span class="coin coin-c" />
        </div>
      </section>

      <section v-if="loadFailed" class="state-panel">
        <van-empty :description="t('checkInPage.loadFailed')" image="error">
          <van-button round type="primary" size="small" @click="loadCheckInData">
            {{ t('checkInPage.retry') }}
          </van-button>
        </van-empty>
      </section>

      <template v-else>
        <section class="status-panel">
          <div class="status-item">
            <span>{{ t('checkInPage.totalDays') }}</span>
            <strong>{{ totalDays }}</strong>
          </div>
          <div class="status-item">
            <span>{{ t('checkInPage.nextReward') }}</span>
            <strong>{{ formatAmount(nextRewardAmount) }}</strong>
          </div>
          <div class="status-item wide">
            <span>{{ t('checkInPage.timezone') }}</span>
            <strong>{{ status?.timezone || 'America/New_York' }}</strong>
          </div>
        </section>

        <button
          type="button"
          class="checkin-button"
          :class="{ disabled: todayChecked || completed }"
          :disabled="todayChecked || completed || signing"
          @click="handleCheckIn"
        >
          <van-loading v-if="signing" size="18" color="#650400" />
          <van-icon v-else :name="todayChecked ? 'success' : 'gold-coin-o'" />
          <span>{{ buttonText }}</span>
        </button>

        <section class="reward-panel">
          <div class="section-heading">
            <h2>{{ t('checkInPage.rewardTitle') }}</h2>
            <span>{{ t('checkInPage.rewardHint') }}</span>
          </div>
          <div class="reward-grid">
            <article
              v-for="(amount, index) in rewards"
              :key="`${index}-${amount}`"
              class="reward-card"
              :class="rewardState(index)"
            >
              <span class="day-label">{{ t('checkInPage.dayLabel', { day: index + 1 }) }}</span>
              <strong>{{ formatAmount(amount) }}</strong>
              <small>{{ t(`checkInPage.state.${rewardState(index)}`) }}</small>
            </article>
          </div>
        </section>

        <section class="record-panel">
          <div class="section-heading">
            <h2>{{ t('checkInPage.recordTitle') }}</h2>
            <span>{{ t('checkInPage.recordHint') }}</span>
          </div>

          <van-empty v-if="records.length === 0" :description="t('checkInPage.emptyRecords')" />

          <div v-else class="record-list">
            <article v-for="item in records" :key="item.id" class="record-row">
              <div>
                <strong>{{ t('checkInPage.recordSeq', { seq: item.checkInSeq }) }}</strong>
                <span>{{ formatRecordDate(item.checkInDate) }}</span>
              </div>
              <em>+{{ formatAmount(item.rewardAmount) }}</em>
            </article>
          </div>
        </section>
      </template>
    </template>
  </main>
</template>

<style scoped>
.checkin-page {
  min-height: 100vh;
  padding: 0 12px calc(24px + env(safe-area-inset-bottom));
  background-image:
    radial-gradient(circle at 20% 8%, rgba(255, 224, 128, 0.18), transparent 26%),
    radial-gradient(circle at 86% 32%, rgba(192, 57, 43, 0.28), transparent 24%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.05) 18px,
      rgba(212, 175, 55, 0.05) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #260000 58%, #160000 100%);
  color: #fff7d6;
}

.page-loading {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.hero-panel,
.status-panel,
.reward-panel,
.record-panel,
.state-panel {
  border: 1px solid rgba(255, 224, 128, 0.16);
  background: rgba(76, 7, 2, 0.66);
  box-shadow:
    0 16px 32px rgba(0, 0, 0, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.hero-panel {
  position: relative;
  overflow: hidden;
  min-height: 156px;
  border-radius: 18px;
  padding: 22px 18px;
  display: flex;
  align-items: center;
}

.hero-panel::after {
  content: '';
  position: absolute;
  inset: auto -24px -52px auto;
  width: 160px;
  height: 160px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 216, 92, 0.24), transparent 66%);
}

.hero-copy {
  position: relative;
  z-index: 1;
  width: 68%;
}

.eyebrow {
  margin: 0 0 8px;
  font-size: 12px;
  font-weight: 800;
  color: #ffd98b;
}

.hero-copy h1 {
  margin: 0;
  font-size: 26px;
  line-height: 1.12;
  letter-spacing: 0;
  color: #fff4c5;
}

.hero-subtitle {
  margin: 8px 0 0;
  font-size: 13px;
  line-height: 1.45;
  color: rgba(255, 247, 214, 0.78);
}

.coin-stack {
  position: absolute;
  right: 20px;
  top: 36px;
  width: 78px;
  height: 78px;
}

.coin {
  position: absolute;
  display: block;
  width: 54px;
  height: 54px;
  border-radius: 50%;
  background:
    radial-gradient(circle at 34% 28%, #fff8ba 0 16%, transparent 17%),
    linear-gradient(145deg, #ffe98d, #d79325 52%, #924100);
  box-shadow:
    0 8px 18px rgba(0, 0, 0, 0.28),
    inset 0 0 0 3px rgba(144, 61, 0, 0.18);
}

.coin-a {
  right: 2px;
  top: 2px;
}

.coin-b {
  right: 28px;
  top: 24px;
  transform: scale(0.76) rotate(-12deg);
}

.coin-c {
  right: 0;
  top: 44px;
  transform: scale(0.62) rotate(14deg);
}

.status-panel {
  margin-top: 12px;
  border-radius: 16px;
  padding: 14px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.status-item {
  min-width: 0;
  padding: 12px;
  border-radius: 12px;
  background: rgba(255, 237, 177, 0.08);
}

.status-item.wide {
  grid-column: 1 / -1;
}

.status-item span {
  display: block;
  margin-bottom: 6px;
  color: rgba(255, 247, 214, 0.7);
  font-size: 12px;
}

.status-item strong {
  display: block;
  min-width: 0;
  color: #ffe18a;
  font-size: 18px;
  line-height: 1.2;
  word-break: break-word;
}

.checkin-button {
  width: 100%;
  height: 52px;
  margin-top: 12px;
  border: 0;
  border-radius: 16px;
  background: linear-gradient(135deg, #ffe18a, #d99627 46%, #b51f16);
  color: #650400;
  font-size: 16px;
  font-weight: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.26);
}

.checkin-button.disabled {
  color: rgba(255, 247, 214, 0.72);
  background: linear-gradient(135deg, rgba(108, 56, 19, 0.92), rgba(85, 20, 14, 0.92));
}

.reward-panel,
.record-panel,
.state-panel {
  margin-top: 12px;
  border-radius: 16px;
  padding: 14px;
}

.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 12px;
}

.section-heading h2 {
  margin: 0;
  font-size: 17px;
  line-height: 1.2;
  color: #fff4c5;
}

.section-heading span {
  flex: 0 0 auto;
  font-size: 12px;
  color: rgba(255, 247, 214, 0.62);
}

.reward-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 9px;
}

.reward-card {
  min-height: 86px;
  padding: 10px 8px;
  border-radius: 12px;
  border: 1px solid rgba(255, 224, 128, 0.12);
  background: rgba(255, 247, 214, 0.07);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 5px;
}

.reward-card.claimed {
  background: rgba(40, 123, 70, 0.24);
  border-color: rgba(128, 224, 153, 0.28);
}

.reward-card.current {
  background: linear-gradient(160deg, rgba(255, 219, 107, 0.24), rgba(181, 31, 22, 0.24));
  border-color: rgba(255, 224, 128, 0.55);
}

.reward-card.locked {
  opacity: 0.76;
}

.day-label,
.reward-card small {
  font-size: 11px;
  color: rgba(255, 247, 214, 0.66);
}

.reward-card strong {
  color: #ffe18a;
  font-size: 13px;
  line-height: 1.2;
  word-break: break-word;
}

.record-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.record-row {
  min-height: 54px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 247, 214, 0.07);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.record-row div {
  min-width: 0;
}

.record-row strong,
.record-row span {
  display: block;
}

.record-row strong {
  color: #fff4c5;
  font-size: 14px;
}

.record-row span {
  margin-top: 3px;
  color: rgba(255, 247, 214, 0.62);
  font-size: 12px;
}

.record-row em {
  flex: 0 0 auto;
  color: #ffe18a;
  font-style: normal;
  font-weight: 900;
  font-size: 14px;
}

:deep(.van-empty__description) {
  color: rgba(255, 247, 214, 0.66);
}

@media (max-width: 360px) {
  .hero-copy {
    width: 72%;
  }

  .hero-copy h1 {
    font-size: 23px;
  }

  .reward-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>

<route lang="json5">
{
  name: 'CheckIn',
}
</route>
