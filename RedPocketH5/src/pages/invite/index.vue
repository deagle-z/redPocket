<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getCurrentTgInviteRuleConfig, getCurrentTgInviteStats } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'

const router = useRouter()
const { t } = useI18n()

const state = reactive({
  inviteCode: '',
  inviteCount: 0,
  todayCount: 0,
  totalCommission: 0,
  luckySendCommission: 5,
  luckyGrabbingCommission: 5,
  inviteFirstRechargeReward: 10,
  inviteLuckyRebateRate: 40,
  inviteThunderRebateRate: 40,
})
// https://t.me/goodLuckEveryOne66Bot/?start=597811
const tgBotName = (import.meta.env.VITE_TG_BOT_NAME || 'goodLuckEveryOne66Bot').trim()
const tgAppName = (import.meta.env.VITE_TG_APP_NAME || 'luckyapp').trim()

const webInviteLink = computed(() => {
  const code = encodeURIComponent(state.inviteCode || '')
  return `https://red.ai3-mountain.com/register?c=${code}`
})

const telegramInviteLink = computed(() => {
  const startapp = encodeURIComponent(state.inviteCode || '')
  return `https://t.me/${tgBotName}/${tgAppName}?startapp=${startapp}`
})

const qrCodeUrl = computed(() => {
  const text = encodeURIComponent(webInviteLink.value)
  return `https://quickchart.io/qr?size=360&margin=0&text=${text}`
})

function goBack() {
  router.back()
}

function formatAmount(value: number) {
  return Number(value || 0).toFixed(2)
}

async function loadInviteData() {
  try {
    const { data } = await getCurrentTgInviteStats()
    state.inviteCode = String(data?.inviteCode || '')
    state.inviteCount = Number(data?.inviteCount || 0)
    state.todayCount = Number(data?.todayInviteCount || 0)
    state.totalCommission = Number(data?.totalCommission || 0)
  }
  catch {
    state.inviteCode = ''
    state.inviteCount = 0
    state.todayCount = 0
    state.totalCommission = 0
  }
}

async function loadInviteRuleConfig() {
  try {
    const { data } = await getCurrentTgInviteRuleConfig()
    state.luckySendCommission = Number(data?.luckySendCommission || 5)
    state.luckyGrabbingCommission = Number(data?.luckyGrabbingCommission || 5)
    state.inviteFirstRechargeReward = Number(data?.inviteFirstRechargeReward || 10)
    state.inviteLuckyRebateRate = Number(data?.inviteLuckyRebateRate || 40)
    state.inviteThunderRebateRate = Number(data?.inviteThunderRebateRate || 40)
  }
  catch {
    state.luckySendCommission = 5
    state.luckyGrabbingCommission = 5
    state.inviteFirstRechargeReward = 10
    state.inviteLuckyRebateRate = 40
    state.inviteThunderRebateRate = 40
  }
}

async function copyText(text: string) {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    }
    else {
      const input = document.createElement('input')
      input.value = text
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
    }
    showToast(t('invitePage.toastCopySuccess'))
  }
  catch {
    showToast(t('invitePage.toastCopyFailed'))
  }
}

function shareToTelegram() {
  const text = encodeURIComponent(t('invitePage.shareText'))
  const url = encodeURIComponent(webInviteLink.value)
  window.open(`https://t.me/share/url?url=${url}&text=${text}`, '_blank')
}

function saveQrCode() {
  const a = document.createElement('a')
  a.href = qrCodeUrl.value
  a.download = 'invite-qrcode.png'
  a.target = '_blank'
  a.rel = 'noopener'
  a.click()
}

onMounted(() => {
  loadInviteData()
  loadInviteRuleConfig()
})
</script>

<template>
  <div class="invite-page">
    <AppPageHeader :title="t('invitePage.title')" @back="goBack" />

    <section class="card qr-card">
      <h2>{{ t('invitePage.heroTitle') }}</h2>
      <p class="sub-title">
        {{ t('invitePage.heroSubTitle') }}
      </p>
      <img class="qr-image" :src="qrCodeUrl" alt="invite qrcode">
      <div class="qr-actions">
        <button type="button" class="main-btn" @click="saveQrCode">
          {{ t('invitePage.saveQr') }}
        </button>
        <button type="button" class="main-btn" @click="shareToTelegram">
          {{ t('invitePage.shareToTelegram') }}
        </button>
      </div>
    </section>

    <section class="card link-card">
      <div class="link-row">
        <p class="link-title">
          {{ t('invitePage.linkTelegram') }}
        </p>
        <div class="link-content">
          <p>{{ telegramInviteLink }}</p>
          <button type="button" class="copy-btn" @click="copyText(telegramInviteLink)">
            {{ t('invitePage.copy') }}
          </button>
        </div>
      </div>
      <div class="link-row">
        <p class="link-title">
          {{ t('invitePage.linkH5') }}
        </p>
        <div class="link-content">
          <p>{{ webInviteLink }}</p>
          <button type="button" class="copy-btn" @click="copyText(webInviteLink)">
            {{ t('invitePage.copy') }}
          </button>
        </div>
      </div>
    </section>

    <section class="card">
      <h3 class="section-title">
        {{ t('invitePage.benefitTitle') }}
      </h3>
      <div class="benefit-grid">
        <article class="benefit-item">
          <div class="benefit-icon blue">
            👥
          </div>
          <p class="benefit-name">
            {{ t('invitePage.benefitGameShare') }}
          </p>
          <p class="benefit-desc">
            {{ t('invitePage.benefitGameShareDesc', { rate: state.inviteLuckyRebateRate }) }}
          </p>
        </article>
        <article class="benefit-item">
          <div class="benefit-icon orange">
            <img src="@/assets/svg/coin.svg" alt="" style="width:1.1em;height:1.1em;display:block;">
          </div>
          <p class="benefit-name">
            {{ t('invitePage.benefitWithdraw') }}
          </p>
          <p class="benefit-desc">
            {{ t('invitePage.benefitWithdrawDesc') }}
          </p>
        </article>
        <article class="benefit-item">
          <div class="benefit-icon green">
            📈
          </div>
          <p class="benefit-name">
            {{ t('invitePage.benefitLifetime') }}
          </p>
          <p class="benefit-desc">
            {{ t('invitePage.benefitLifetimeDesc') }}
          </p>
        </article>
        <article class="benefit-item">
          <div class="benefit-icon red">
            🎁
          </div>
          <p class="benefit-name">
            {{ t('invitePage.benefitExtra') }}
          </p>
          <p class="benefit-desc">
            {{ t('invitePage.benefitExtraDesc') }}
          </p>
        </article>
      </div>
    </section>

    <section class="card">
      <h3 class="section-title">
        {{ t('invitePage.ruleTitle') }}
      </h3>
      <ul class="rule-list">
        <li>{{ t('invitePage.rule1', { reward: state.inviteFirstRechargeReward }) }}</li>
        <li>{{ t('invitePage.rule2', { commission: state.luckyGrabbingCommission, rebate: state.inviteLuckyRebateRate }) }}</li>
        <li>{{ t('invitePage.rule3', { commission: state.luckySendCommission, rebate: state.inviteThunderRebateRate }) }}</li>
      </ul>
    </section>

    <section class="card">
      <h3 class="section-title">
        {{ t('invitePage.stepTitle') }}
      </h3>
      <ol class="steps">
        <li>
          <h4>{{ t('invitePage.step1Title') }}</h4>
          <p>{{ t('invitePage.step1Desc') }}</p>
        </li>
        <li>
          <h4>{{ t('invitePage.step2Title') }}</h4>
          <p>{{ t('invitePage.step2Desc') }}</p>
        </li>
        <li>
          <h4>{{ t('invitePage.step3Title') }}</h4>
          <p>{{ t('invitePage.step3Desc') }}</p>
        </li>
        <li>
          <h4>{{ t('invitePage.step4Title') }}</h4>
          <p>{{ t('invitePage.step4Desc') }}</p>
        </li>
      </ol>
    </section>

    <section class="card data-card">
      <div class="data-head">
        <h3 class="section-title">
          {{ t('invitePage.dataTitle') }}
        </h3>
        <van-icon name="arrow" />
      </div>
      <div class="data-grid">
        <div>
          <p class="data-value">
            {{ state.inviteCount }}
          </p>
          <p class="data-label">
            {{ t('invitePage.dataInviteCount') }}
          </p>
        </div>
        <div>
          <p class="data-value">
            {{ state.todayCount }}
          </p>
          <p class="data-label">
            {{ t('invitePage.dataTodayNew') }}
          </p>
        </div>
        <div>
          <p class="data-value">
            {{ formatAmount(state.totalCommission) }}
          </p>
          <p class="data-label">
            {{ t('invitePage.dataTotalCommission') }}
          </p>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.invite-page {
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

.card {
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

.card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.qr-card {
  text-align: center;
}

.qr-card h2 {
  margin: 0;
  font-size: 20px;
  color: #ffd98b;
  letter-spacing: 0.06em;
}

.sub-title {
  margin: 8px 0 12px;
  color: rgba(255, 229, 186, 0.62);
  font-size: 12px;
}

.qr-image {
  width: 210px;
  height: 210px;
  display: block;
  margin: 0 auto;
  border-radius: 4px;
}

.qr-actions {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.main-btn {
  height: 44px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 12px;
  font-weight: 800;
}

.link-card {
  padding: 0;
  overflow: hidden;
}

.link-row {
  padding: 14px;
}

.link-row + .link-row {
  border-top: 1px solid rgba(212, 175, 55, 0.12);
}

.link-title {
  margin: 0;
  color: #ffd98b;
  font-size: 14px;
  font-weight: 700;
}

.link-content {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.link-content p {
  margin: 0;
  flex: 1;
  color: rgba(255, 229, 186, 0.78);
  font-size: 12px;
  word-break: break-all;
}

.copy-btn {
  flex-shrink: 0;
  min-width: 76px;
  height: 40px;
  border-radius: 999px;
  border: 1px solid rgba(255, 248, 214, 0.3);
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 12px;
  font-weight: 800;
}

.section-title {
  margin: 0;
  font-size: 14px;
  font-weight: 800;
  color: #ffd98b;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.rule-list {
  margin: 10px 0 0;
  padding-left: 18px;
  color: rgba(255, 229, 186, 0.78);
  font-size: 13px;
  line-height: 1.7;
}

.rule-list li + li {
  margin-top: 6px;
}

.benefit-grid {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 10px;
}

.benefit-item {
  text-align: center;
}

.benefit-icon {
  width: 54px;
  height: 54px;
  margin: 0 auto 10px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  border: 2px solid currentcolor;
}

.benefit-icon.blue {
  color: #2ea5e6;
}

.benefit-icon.orange {
  color: #f2a100;
}

.benefit-icon.green {
  color: var(--color-primary-link);
}

.benefit-icon.red {
  color: #f05b63;
}

.benefit-name {
  margin: 0 0 8px;
  font-size: 15px;
  font-weight: 700;
  color: #fff0c9;
}

.benefit-desc {
  margin: 0;
  color: rgba(255, 229, 186, 0.62);
  font-size: 13px;
  line-height: 1.35;
}

.steps {
  list-style: none;
  margin: 16px 0 0;
  padding: 0 0 0 18px;
  position: relative;
}

.steps::before {
  content: '';
  position: absolute;
  left: 6px;
  top: 8px;
  bottom: 12px;
  width: 2px;
  background: rgba(212, 175, 55, 0.4);
}

.steps li {
  position: relative;
  padding: 0 0 18px 18px;
}

.steps li::before {
  content: '';
  position: absolute;
  left: -1px;
  top: 6px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ffd87f;
}

.steps li:last-child::before {
  width: 18px;
  height: 18px;
  left: -6px;
  top: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffd87f;
  content: '✓';
  color: #5a1b00;
  font-size: 12px;
  font-weight: 700;
}

.steps h4 {
  margin: 0;
  font-size: 15px;
  color: #fff0c9;
}

.steps p {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.62);
  font-size: 13px;
  line-height: 1.35;
}

.data-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.data-head :deep(.van-icon) {
  color: rgba(255, 229, 186, 0.32);
  font-size: 18px;
}

.data-grid {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  text-align: center;
}

.data-value {
  margin: 0;
  font-size: 18px;
  line-height: 1;
  color: #ffd87f;
  font-weight: 800;
}

.data-label {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.54);
  font-size: 13px;
}
</style>

<route lang="json5">
{
  name: 'Invite'
}
</route>
