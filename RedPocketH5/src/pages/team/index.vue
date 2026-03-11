<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { getCurrentTgInviteStats } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { CURRENCY_SYMBOL, formatCurrency } from '@/utils/currency'
import imgTeamJpg from '@/assets/images/team.jpg'

const { t } = useI18n()

const router = useRouter()
const teamRuleBannerImage = imgTeamJpg

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
  padding: 0 12px calc(90px + env(safe-area-inset-bottom));
  color: #fff0c9;
}

.team-header {
  margin-bottom: 10px;
}

.section-card {
  position: relative;
  overflow: hidden;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  padding: 14px;
  margin-bottom: 12px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.section-card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.section-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #ffd98b;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
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
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.18);
  background: rgba(255, 248, 214, 0.05);
  min-height: 100px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.overview-icon {
  width: 30px;
  height: 30px;
  border-radius: 10px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.overview-value {
  margin: 10px 0 6px;
  font-size: 20px;
  line-height: 1;
  color: #ffd87f;
  font-weight: 800;
}

.overview-label {
  margin: 0;
  font-size: 12px;
  line-height: 1.2;
  color: rgba(255, 229, 186, 0.6);
}

.commission-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.commission-item {
  min-height: 80px;
  border-radius: 16px;
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.commission-item.warm,
.commission-item.success,
.commission-item.info {
  background: rgba(255, 248, 214, 0.06);
  border: 1px solid rgba(212, 175, 55, 0.16);
}

.commission-label {
  margin: 0 0 8px;
  font-size: 12px;
  line-height: 1;
  font-weight: 600;
}

.commission-item .commission-label {
  color: rgba(255, 229, 186, 0.62);
}

.commission-value {
  margin: 0;
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
}

.commission-item .commission-value {
  color: #ffd87f;
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
  background: rgba(255, 248, 214, 0.04);
  border: 1px solid rgba(212, 175, 55, 0.16);
  border-radius: 16px;
  padding: 12px;
}

.rule-content h3 {
  margin: 0 0 10px;
  font-size: 15px;
  line-height: 1.2;
  color: #fff0c9;
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
  color: rgba(255, 229, 186, 0.78);
  font-size: 13px;
  line-height: 1.4;
}

.rule-list :deep(.van-icon) {
  margin-top: 1px;
  color: #ffd87f;
  font-size: 14px;
}

.rule-text {
  margin: 10px 0;
  color: rgba(255, 229, 186, 0.64);
  font-size: 13px;
  line-height: 1.5;
}

.rule-highlight {
  margin-top: 8px;
  background: rgba(212, 175, 55, 0.12);
  border-radius: 12px;
  padding: 12px;
  color: #ffd87f;
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
