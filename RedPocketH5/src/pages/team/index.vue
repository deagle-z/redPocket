<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { getCurrentTgInviteStats } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { CURRENCY_SYMBOL, formatCurrency } from '@/utils/currency'
import imgRedpacketJpg from '@/assets/images/redpacket.jpg'

const { t } = useI18n()

const router = useRouter()
const teamRuleBannerImage = imgRedpacketJpg

const stats = reactive({
  inviteUsers: 0,
  rechargeUsers: 0,
  todayInviteUsers: 0,
  todayRechargeUsers: 0,
  totalCommission: 0,
  availableCommission: 0,
  todayCommission: 0,
})

const overviewCards = computed(() => [
  { key: 'inviteUsers', icon: 'friends-o', label: t('teamPage.statInviteUsers'), value: stats.inviteUsers },
  { key: 'rechargeUsers', icon: 'cash-back-record', label: t('teamPage.statRechargeUsers'), value: stats.rechargeUsers },
  { key: 'todayInviteUsers', icon: 'calendar-o', label: t('teamPage.statTodayInviteUsers'), value: stats.todayInviteUsers },
  { key: 'todayRechargeUsers', icon: 'fire-o', label: t('teamPage.statTodayRechargeUsers'), value: stats.todayRechargeUsers },
])

const commissionCards = computed(() => [
  { key: 'total', label: t('teamPage.commissionTotal'), value: stats.totalCommission, tone: 'warm' },
  { key: 'available', label: t('teamPage.commissionAvailable'), value: stats.availableCommission, tone: 'success' },
  { key: 'today', label: t('teamPage.commissionToday'), value: stats.todayCommission, tone: 'info' },
])

function goBack() {
  router.back()
}

function formatAmount(value: number) {
  return formatCurrency(Number(value || 0))
}

async function loadTeamData() {
  try {
    const { data } = await getCurrentTgInviteStats()
    stats.inviteUsers = Number(data?.inviteCount || 0)
    stats.rechargeUsers = Number(data?.rechargeUsers || 0)
    stats.todayInviteUsers = Number(data?.todayInviteCount || 0)
    stats.todayRechargeUsers = Number(data?.todayRechargeUsers || 0)
    stats.totalCommission = Number(data?.totalCommission || 0)
    stats.availableCommission = Number(data?.availableCommission || 0)
    stats.todayCommission = Number(data?.todayCommission || 0)
  }
  catch {
    // Keep zero values as fallback.
  }
}

onMounted(() => {
  loadTeamData()
})
</script>

<template>
  <div class="team-page">
    <AppPageHeader class="team-header" :title="t('teamPage.title')" @back="goBack" />

    <section class="section-card">
      <div class="section-title">
        <span class="dot green" />
        <span>{{ t('teamPage.overviewTitle') }}</span>
      </div>
      <div class="overview-grid">
        <article v-for="item in overviewCards" :key="item.key" class="overview-item">
          <div class="overview-icon">
            <van-icon :name="item.icon" />
          </div>
          <p class="overview-value">
            {{ item.value }}
          </p>
          <p class="overview-label">
            {{ item.label }}
          </p>
        </article>
      </div>
    </section>

    <section class="section-card">
      <div class="section-title">
        <span class="dot yellow" />
        <span>{{ t('teamPage.commissionTitle') }}</span>
      </div>
      <div class="commission-grid">
        <article v-for="item in commissionCards" :key="item.key" class="commission-item" :class="item.tone">
          <p class="commission-label">
            {{ item.label }}
          </p>
          <p class="commission-value">
            {{ formatAmount(item.value) }}
          </p>
        </article>
      </div>
    </section>

    <section class="section-card">
      <div class="section-title">
        <span class="dot orange" />
        <span>{{ t('teamPage.ruleTitle') }}</span>
      </div>

      <div class="rule-banner">
        <img :src="teamRuleBannerImage" alt="commission banner">
      </div>

      <div class="rule-content">
        <h3>{{ t('teamPage.ruleIntroTitle') }}</h3>
        <ul class="rule-list">
          <li>
            <van-icon name="friends-o" />
            {{ t('teamPage.ruleItem1', { symbol: CURRENCY_SYMBOL }) }}
          </li>
          <li>
            <van-icon name="gift-o" />
            {{ t('teamPage.ruleItem2') }}
          </li>
          <li>
            <van-icon name="flash" />
            {{ t('teamPage.ruleItem3') }}
          </li>
        </ul>

        <p class="rule-text">
          {{ t('teamPage.ruleText1') }}
        </p>

        <h3>{{ t('teamPage.exampleTitle') }}</h3>
        <p class="rule-text">
          {{ t('teamPage.exampleText', { symbol: CURRENCY_SYMBOL }) }}
        </p>

        <div class="rule-highlight">
          {{ t('teamPage.highlightText') }}
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.team-page {
  min-height: 100vh;
  background: var(--bg-secondary);
  padding: 0 var(--space-md) calc(18px + env(safe-area-inset-bottom));
}

.team-header {
  margin-bottom: 10px;
}

.section-card {
  background: #fff;
  border-radius: 12px;
  padding: 12px;
  margin-bottom: 12px;
  border: 1px solid #edf1f5;
}

.section-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 16px;
  color: #0f172a;
  font-weight: 700;
  margin-bottom: 12px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot.green {
  background: var(--color-primary);
}

.dot.yellow {
  background: #f59e0b;
}

.dot.orange {
  background: #fb923c;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.overview-item {
  border-radius: 12px;
  border: 1px solid #edf1f5;
  height: 100px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.overview-icon {
  width: 24px;
  height: 24px;
  border-radius: 8px;
  background: var(--color-primary-soft);
  color: var(--color-primary-medium);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.overview-value {
  margin: 10px 0 6px;
  font-size: 18px;
  line-height: 1;
  color: #f59e0b;
  font-weight: 700;
}

.overview-label {
  margin: 0;
  font-size: 12px;
  line-height: 1.2;
  color: #6b7280;
}

.commission-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.commission-item {
  height: 80px;
  border-radius: 12px;
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.commission-item.warm {
  background: #fff4e5;
}

.commission-item.success {
  background: var(--color-primary-soft);
}

.commission-item.info {
  background: #e6f3ff;
}

.commission-label {
  margin: 0 0 8px;
  font-size: 12px;
  line-height: 1;
  font-weight: 600;
}

.commission-item.warm .commission-label {
  color: #b58a30;
}

.commission-item.success .commission-label {
  color: var(--color-primary-link);
}

.commission-item.info .commission-label {
  color: #4389d8;
}

.commission-value {
  margin: 0;
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
}

.commission-item.warm .commission-value {
  color: #f59e0b;
}

.commission-item.success .commission-value {
  color: var(--color-primary-medium);
}

.commission-item.info .commission-value {
  color: #2563eb;
}

.rule-banner {
  width: 100%;
  height: 180px;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 12px;
}

.rule-banner img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.rule-content {
  background: #fff;
  border: 1px solid #edf1f5;
  border-radius: 12px;
  padding: 12px;
}

.rule-content h3 {
  margin: 0 0 10px;
  font-size: 15px;
  line-height: 1.2;
  color: #1f2937;
  font-weight: 700;
}

.rule-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.rule-list li {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  color: #374151;
  font-size: 13px;
  line-height: 1.4;
}

.rule-list :deep(.van-icon) {
  margin-top: 1px;
  color: #f59e0b;
  font-size: 14px;
}

.rule-text {
  margin: 10px 0;
  color: #4b5563;
  font-size: 13px;
  line-height: 1.5;
}

.rule-highlight {
  margin-top: 8px;
  background: #fffbe6;
  border-radius: 8px;
  padding: 12px;
  color: #92400e;
  font-size: 13px;
  line-height: 1.5;
  font-weight: 700;
}
</style>

<route lang="json5">
{
  name: 'Team'
}
</route>
