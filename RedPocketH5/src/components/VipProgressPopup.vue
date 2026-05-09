<script setup lang="ts">
import { showToast } from 'vant'
import { claimVipReward, getClaimableVipRewards, getVipProgress } from '@/api/user'
import type { VipProgressInfo, VipRewardLog } from '@/api/user'
import { truncate2 } from '@/utils/currency'

const emit = defineEmits<{
  claimed: []
}>()
const show = defineModel<boolean>('show', { default: false })

const { t } = useI18n()

const vipLoading = ref(false)
const claimingId = ref<number | null>(null)
const vipProgress = ref<VipProgressInfo | null>(null)
const vipRewards = ref<VipRewardLog[]>([])

function formatAmount3(value: number) {
  return truncate2(Number(value || 0)).toFixed(2)
}

async function loadVipProgress() {
  if (vipLoading.value)
    return
  vipLoading.value = true
  try {
    const [progressRes, rewardsRes] = await Promise.all([
      getVipProgress(),
      getClaimableVipRewards(),
    ])
    vipProgress.value = progressRes.data ?? null
    vipRewards.value = rewardsRes.data ?? []
  }
  catch {
    // toast shown by interceptor
  }
  finally {
    vipLoading.value = false
  }
}

async function handleClaimReward(id: number) {
  if (claimingId.value !== null)
    return
  claimingId.value = id
  try {
    await claimVipReward(id)
    showToast(t('profilePage.toastClaimSuccess'))
    vipRewards.value = vipRewards.value.filter(r => r.id !== id)
    emit('claimed')
  }
  catch {
    // toast shown by interceptor
  }
  finally {
    claimingId.value = null
  }
}

async function handleClaimAll() {
  if (claimingId.value !== null)
    return
  claimingId.value = 0
  try {
    await claimVipReward(0)
    showToast(t('profilePage.toastClaimAllSuccess'))
    vipRewards.value = []
    emit('claimed')
  }
  catch {
    // toast shown by interceptor
  }
  finally {
    claimingId.value = null
  }
}

watch(show, (visible) => {
  if (visible)
    void loadVipProgress()
})
</script>

<template>
  <van-popup v-model:show="show" round position="bottom" class="vip-popup">
    <div class="vip-popup-header">
      <span class="vip-popup-title">{{ t('profilePage.vipTitle') }}</span>
      <button type="button" class="vip-popup-close" @click="show = false">
        ×
      </button>
    </div>

    <div v-if="vipLoading" class="vip-loading">
      <van-loading color="#ffd87f" />
    </div>

    <template v-else-if="vipProgress">
      <div class="vip-levels-row">
        <div class="vip-level-badge">
          <span class="vip-badge-name">{{ vipProgress.prevLevel?.levelName || '—' }}</span>
          <span class="vip-badge-label">{{ t('profilePage.vipPrevLevel') }}</span>
        </div>
        <div class="vip-level-badge current">
          <span class="vip-badge-name">{{ vipProgress.currentLevel?.levelName || t('profilePage.vipDefaultLevel') }}</span>
          <span class="vip-badge-label">{{ t('profilePage.vipCurrentLevel') }}</span>
        </div>
        <div class="vip-level-badge">
          <span class="vip-badge-name">{{ vipProgress.nextLevel?.levelName || '—' }}</span>
          <span class="vip-badge-label">{{ t('profilePage.vipNextLevel') }}</span>
        </div>
      </div>

      <div class="vip-progress-wrap">
        <div class="vip-progress-labels">
          <span class="vip-progress-cur">{{ formatAmount3(vipProgress.currentValue) }}</span>
          <span class="vip-progress-pct">{{ vipProgress.progress.toFixed(0) }}%</span>
          <span class="vip-progress-target">{{ vipProgress.targetValue > 0 ? formatAmount3(vipProgress.targetValue) : '—' }}</span>
        </div>
        <div class="vip-progress-bar">
          <div class="vip-progress-fill" :style="{ width: `${vipProgress.progress}%` }" />
        </div>
        <p v-if="vipProgress.nextLevel" class="vip-progress-hint">
          {{ t('profilePage.vipProgressHint', {
            level: vipProgress.nextLevel.levelName,
            amount: formatAmount3(Math.max(0, vipProgress.targetValue - vipProgress.currentValue)),
          }) }}
        </p>
        <p v-else class="vip-progress-hint">
          {{ t('profilePage.vipTopLevelReached') }}
        </p>
      </div>

      <div v-if="vipProgress.nextLevel && vipProgress.nextBonusAmount > 0" class="vip-next-bonus">
        <span class="vip-next-bonus-label">{{ t('profilePage.vipUpgradeReward') }}</span>
        <span class="vip-next-bonus-amount">+{{ formatAmount3(vipProgress.nextBonusAmount) }}</span>
      </div>

      <div v-if="vipRewards.length > 0" class="vip-rewards-section">
        <div class="vip-rewards-header">
          <span class="vip-rewards-title">{{ t('profilePage.vipPendingRewards') }}</span>
          <button
            type="button"
            class="vip-claim-all-btn"
            :disabled="claimingId !== null"
            @click="handleClaimAll"
          >
            {{ claimingId === 0 ? t('profilePage.vipClaiming') : t('profilePage.vipClaimAll') }}
          </button>
        </div>
        <div v-for="reward in vipRewards" :key="reward.id" class="vip-reward-item">
          <div class="vip-reward-info">
            <span class="vip-reward-name">{{ t('profilePage.vipRewardName', { level: reward.levelName }) }}</span>
            <span class="vip-reward-amount">+{{ formatAmount3(reward.bonusAmount) }}</span>
          </div>
          <button
            type="button"
            class="vip-claim-btn"
            :disabled="claimingId !== null"
            @click="handleClaimReward(reward.id)"
          >
            {{ claimingId === reward.id ? t('profilePage.vipClaiming') : t('profilePage.vipClaim') }}
          </button>
        </div>
      </div>
    </template>
  </van-popup>
</template>

<style scoped>
.vip-popup {
  padding: 0 0 32px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  background:
    radial-gradient(circle at top, rgba(212, 175, 55, 0.14), transparent 26%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
}

.vip-popup-header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 56px;
  border-bottom: 1px solid rgba(212, 175, 55, 0.18);
}

.vip-popup-title {
  color: #fff0c9;
  font-size: 18px;
  font-weight: 700;
}

.vip-popup-close {
  position: absolute;
  top: 50%;
  right: 16px;
  border: none;
  background: transparent;
  color: rgba(255, 229, 186, 0.7);
  font-size: 20px;
  line-height: 1;
  transform: translateY(-50%);
}

.vip-loading {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.vip-levels-row {
  display: flex;
  gap: 8px;
  justify-content: space-between;
  padding: 20px 20px 0;
}

.vip-level-badge {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 6px;
  align-items: center;
  padding: 12px 8px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.04);
}

.vip-level-badge.current {
  border-color: rgba(255, 216, 127, 0.6);
  background: linear-gradient(160deg, rgba(140, 30, 0, 0.7), rgba(80, 0, 0, 0.7));
  box-shadow: 0 0 14px rgba(212, 175, 55, 0.18);
}

.vip-badge-name {
  color: #ffd87f;
  font-size: 14px;
  font-weight: 800;
  line-height: 1;
}

.vip-badge-label {
  color: rgba(255, 229, 186, 0.56);
  font-size: 10px;
  line-height: 1;
}

.vip-progress-wrap {
  padding: 20px 20px 0;
}

.vip-progress-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
  color: rgba(255, 229, 186, 0.6);
  font-size: 11px;
}

.vip-progress-pct {
  color: #ffd87f;
  font-size: 12px;
  font-weight: 700;
}

.vip-progress-bar {
  height: 8px;
  overflow: hidden;
  border-radius: 99px;
  background: rgba(255, 255, 255, 0.1);
}

.vip-progress-fill {
  min-width: 4px;
  height: 100%;
  border-radius: 99px;
  background: linear-gradient(90deg, #d4af37 0%, #ffd700 100%);
  transition: width 0.5s ease;
}

.vip-progress-hint {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.6);
  font-size: 12px;
  text-align: center;
}

.vip-next-bonus {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 16px 20px 0;
  padding: 12px 16px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 12px;
  background: rgba(212, 175, 55, 0.08);
}

.vip-next-bonus-label {
  color: rgba(255, 229, 186, 0.7);
  font-size: 13px;
}

.vip-next-bonus-amount {
  color: #ffd87f;
  font-size: 16px;
  font-weight: 800;
}

.vip-rewards-section {
  margin: 16px 20px 0;
}

.vip-rewards-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.vip-rewards-title {
  color: #fff0c9;
  font-size: 13px;
  font-weight: 700;
}

.vip-claim-all-btn {
  padding: 5px 14px;
  border: 1px solid rgba(255, 216, 127, 0.5);
  border-radius: 99px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 12px;
  font-weight: 700;
}

.vip-claim-all-btn:disabled {
  opacity: 0.5;
}

.vip-reward-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 12px 14px;
  border: 1px solid rgba(212, 175, 55, 0.18);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.04);
}

.vip-reward-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.vip-reward-name {
  color: #fff0c9;
  font-size: 13px;
  font-weight: 600;
}

.vip-reward-amount {
  color: #ffd87f;
  font-size: 15px;
  font-weight: 800;
}

.vip-claim-btn {
  padding: 7px 18px;
  border: 1px solid rgba(255, 216, 127, 0.5);
  border-radius: 99px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 13px;
  font-weight: 700;
}

.vip-claim-btn:disabled {
  opacity: 0.5;
}
</style>
