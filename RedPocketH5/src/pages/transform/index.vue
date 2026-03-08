<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { showToast } from 'vant'
import { useRouter } from 'vue-router'
import { getAppCashHistoryList, getCurrentTgUserInfo, transferRebateToBalance } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import gameIcon from '@/assets/my/game.svg'
import { CURRENCY_SYMBOL, formatCurrency } from '@/utils/currency'

const { t } = useI18n()

interface TransformTx {
  id: string
  title: string
  time: string
  amount: number
}

const router = useRouter()
const loading = ref(false)
const confirming = ref(false)
const amountInput = ref('')
const recentList = ref<TransformTx[]>([])

const wallet = reactive({
  commission: 0,
  game: 0,
})

const canTransferAll = computed(() => wallet.commission > 0)
const commissionText = computed(() => formatCurrency(wallet.commission))
const gameText = computed(() => formatCurrency(wallet.game))

function goBack() {
  router.back()
}

function isCommissionTx(item: any) {
  const type = Number(item?.type || 0)
  if (type === 11)
    return true
  const mark = String(item?.cashMark || item?.cashDesc || '').toLowerCase()
  return mark.includes('commission')
}

function formatTime(raw: string) {
  const d = new Date(raw)
  if (Number.isNaN(d.getTime()))
    return raw || ''
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}/${m}/${day} ${hh}:${mm}`
}

function formatAmount(value: number) {
  return formatCurrency(Number(value || 0))
}

function fillAll() {
  if (!canTransferAll.value)
    return
  amountInput.value = wallet.commission.toFixed(2)
}

async function loadWalletInfo() {
  try {
    const { data } = await getCurrentTgUserInfo()
    wallet.commission = Number(data?.rebate_amount || 0)
    wallet.game = Number(data?.balance || 0)
  }
  catch {
    wallet.commission = 0
    wallet.game = 0
  }
}

async function loadRecentList() {
  loading.value = true
  try {
    const { data } = await getAppCashHistoryList({
      currentPage: 0,
      pageSize: 20,
    })
    const list = (data?.list || [])
      .filter((item: any) => isCommissionTx(item))
      .map((item: any) => ({
        id: String(`${item?.createdAt || ''}_${item?.amount || 0}`),
        title: item?.cashDesc || item?.cashMark || t('transformPage.txTitleDefault'),
        time: formatTime(item?.createdAt || ''),
        amount: Number(item?.amount || 0),
      }))
    recentList.value = list
  }
  finally {
    loading.value = false
  }
}

function goHistory() {
  router.push('/history')
}

async function handleConfirm() {
  if (confirming.value)
    return
  const amount = Number(amountInput.value)
  if (!amount || Number.isNaN(amount)) {
    showToast(t('transformPage.toastEnterAmount'))
    return
  }
  if (amount < 1) {
    showToast(t('transformPage.toastMinAmount', { min: `${CURRENCY_SYMBOL}1` }))
    return
  }
  if (amount > wallet.commission) {
    showToast(t('transformPage.toastExceedBalance'))
    return
  }
  confirming.value = true
  try {
    const { data } = await transferRebateToBalance()
    const transferAmount = Number(data?.transferAmount || 0)
    wallet.commission = Number(data?.rebateAmount || 0)
    wallet.game = Number(data?.balance || 0)
    amountInput.value = ''
    showToast(t('transformPage.toastSuccess', { amount: formatCurrency(transferAmount) }))
    void loadRecentList()
  }
  finally {
    confirming.value = false
  }
}

onMounted(() => {
  void loadWalletInfo()
  void loadRecentList()
})
</script>

<template>
  <div class="transform-page">
    <AppPageHeader :title="t('transformPage.title')" @back="goBack" />

    <section class="flow-row card">
      <div class="flow-item">
        <span class="flow-icon green">
          <van-icon name="gold-coin-o" />
        </span>
        <p>{{ t('transformPage.walletCommission') }}</p>
      </div>
      <van-icon name="arrow" class="flow-arrow" />
      <div class="flow-item">
        <span class="flow-icon orange">
          <img :src="gameIcon" alt="" class="flow-icon-img">
        </span>
        <p>{{ t('transformPage.walletGame') }}</p>
      </div>
      <van-icon name="arrow" class="flow-arrow" />
      <div class="flow-item">
        <span class="flow-icon blue">
          <van-icon name="balance-list-o" />
        </span>
        <p>{{ t('transformPage.walletWithdraw') }}</p>
      </div>
    </section>

    <section class="cards-wrap">
      <article class="wallet-card commission">
        <div class="wallet-row">
          <h3>{{ t('transformPage.walletCommission') }}</h3>
          <span>{{ t('transformPage.availableBalance') }}</span>
        </div>
        <p class="wallet-value">
          {{ commissionText }}
        </p>
      </article>

      <article class="wallet-card game">
        <div class="wallet-row">
          <h3>{{ t('transformPage.walletGame') }}</h3>
          <span>{{ t('transformPage.availableBalance') }}</span>
        </div>
        <p class="wallet-value">
          {{ gameText }}
        </p>
      </article>
    </section>

    <section class="input-section card">
      <div class="input-row">
        <p class="input-label">
          {{ t('transformPage.transferAmount') }}
        </p>
        <input v-model="amountInput" class="amount-input" type="number" min="1" :placeholder="t('transformPage.amountPlaceholder', { min: `${CURRENCY_SYMBOL}1` })">
        <button type="button" class="fill-btn" :disabled="!canTransferAll" @click="fillAll">
          {{ t('transformPage.fillAll') }}
        </button>
      </div>
      <p class="input-hint">
        {{ t('transformPage.minTip', { min: `${CURRENCY_SYMBOL}1` }) }}
      </p>
    </section>

    <section class="confirm-wrap">
      <button type="button" class="confirm-btn" :disabled="confirming" @click="handleConfirm">
        {{ t('transformPage.confirmTransfer') }}
      </button>
    </section>

    <section class="recent-wrap card">
      <header class="recent-header" @click="goHistory">
        <h3>{{ t('transformPage.recentTx') }}</h3>
        <van-icon name="arrow" />
      </header>

      <div v-if="loading" class="state-wrap">
        <van-loading size="24px" color="var(--color-primary)" vertical>
          {{ t('transformPage.loading') }}
        </van-loading>
      </div>

      <template v-else-if="recentList.length > 0">
        <article v-for="item in recentList" :key="item.id" class="recent-item">
          <div>
            <p class="recent-title">
              {{ item.title }}
            </p>
            <p class="recent-time">
              {{ item.time }}
            </p>
          </div>
          <p class="recent-amount">
            {{ formatAmount(item.amount) }}
          </p>
        </article>
      </template>

      <div v-else class="empty-state">
        <div class="empty-icon">
          <van-icon name="description" />
        </div>
        <p>{{ t('transformPage.noMore') }}</p>
      </div>
    </section>
  </div>
</template>

<style scoped>
.transform-page {
  min-height: 100vh;
  background: #f5f6fa;
  padding-bottom: calc(16px + env(safe-area-inset-bottom));
}

.card {
  background: #fff;
}

.flow-row {
  margin-top: 10px;
  padding: 18px 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.flow-item {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.flow-item p {
  margin: 0;
  font-size: 11px;
  color: #6b7280;
}

.flow-icon {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #fff;
}

.flow-icon-img {
  width: 24px;
  height: 24px;
  object-fit: contain;
  filter: brightness(0) invert(1);
}

.flow-icon.green {
  background: var(--color-primary);
}

.flow-icon.orange {
  background: #ff9500;
}

.flow-icon.blue {
  background: #4a5cff;
}

.flow-arrow {
  color: #d1d5db;
  font-size: 14px;
}

.cards-wrap {
  padding: 10px 16px 0;
  display: grid;
  gap: 10px;
}

.wallet-card {
  border-radius: 12px;
  padding: 12px 14px;
}

.wallet-card.commission {
  background: var(--color-primary-soft);
  border: 1px solid var(--color-primary);
}

.wallet-card.game {
  background: #eef0ff;
  border: 1px solid #4a5cff;
}

.wallet-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.wallet-row h3 {
  margin: 0;
  font-size: 13px;
  color: #1a1a2e;
  font-weight: 600;
}

.wallet-row span {
  font-size: 11px;
  color: #6b7280;
}

.wallet-value {
  margin: 6px 0 0;
  color: var(--color-primary);
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
}

.input-section {
  margin-top: 14px;
  padding: 14px 16px 8px;
}

.input-row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 10px;
}

.input-label {
  margin: 0;
  color: #1a1a2e;
  font-size: 14px;
  font-weight: 700;
}

.amount-input {
  min-width: 0;
  height: 32px;
  border: none;
  background: transparent;
  color: #1a1a2e;
  font-size: 13px;
  outline: none;
}

.amount-input::placeholder {
  color: #d1d5db;
}

.fill-btn {
  border: 1px solid var(--color-primary);
  color: var(--color-primary);
  background: #fff;
  border-radius: 16px;
  height: 30px;
  padding: 0 14px;
  font-size: 12px;
  font-weight: 500;
}

.fill-btn:disabled {
  border-color: #d1d5db;
  color: #9ca3af;
}

.input-hint {
  margin: 6px 0 0;
  text-align: right;
  color: #9ca3af;
  font-size: 11px;
}

.confirm-wrap {
  padding: 16px;
}

.confirm-btn {
  width: 100%;
  height: 48px;
  border: none;
  border-radius: 24px;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  background: linear-gradient(90deg, var(--color-primary) 0%, var(--color-primary-gradient-end) 100%);
}

.confirm-btn:disabled {
  opacity: 0.7;
}

.recent-wrap {
  margin-top: 6px;
}

.recent-header {
  padding: 16px 16px 8px;
  display: flex;
  align-items: center;
  cursor: pointer;
}

.recent-header h3 {
  margin: 0;
  font-size: 14px;
  color: #1a1a2e;
  font-weight: 700;
}

.recent-header .van-icon {
  margin-left: auto;
  color: #d1d5db;
}

.state-wrap {
  min-height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.recent-item {
  padding: 10px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid #f0f0f5;
}

.recent-title {
  margin: 0;
  color: #1a1a2e;
  font-size: 13px;
}

.recent-time {
  margin: 3px 0 0;
  color: #9ca3af;
  font-size: 11px;
}

.recent-amount {
  margin: 0;
  color: var(--color-primary);
  font-size: 13px;
  font-weight: 600;
}

.empty-state {
  padding: 36px 16px 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.empty-icon {
  width: 100px;
  height: 100px;
  border-radius: 12px;
  background: #f5f6fa;
  color: #d1d5db;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.empty-state p {
  margin: 0;
  color: #9ca3af;
  font-size: 14px;
}
</style>

<route lang="json5">
{
  name: 'Transform',
}
</route>
