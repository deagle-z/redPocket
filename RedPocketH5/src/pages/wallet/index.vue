<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getAppCashHistoryList, getCurrentTgUserInfo } from '@/api/user'
import { CURRENCY_CODE, formatCurrency, truncate2 } from '@/utils/currency'

const { t } = useI18n()

interface WalletTx {
  id: string
  title: string
  time: string
  amount: number
}

const router = useRouter()

const wallet = reactive({
  balance: 0,
  commission: 0,
})

const txList = ref<WalletTx[]>([])
const txLoading = ref(false)
const txFetching = ref(false)
const txFinished = ref(false)
const txPage = ref(0)
const txTotal = ref(0)
const txPageSize = 20

const totalAsset = computed(() => wallet.balance + wallet.commission)

const assetRows = computed(() => [
  { key: 'balance', label: t('walletPage.assetBalance'), unit: CURRENCY_CODE, value: wallet.balance, icon: 'gold-coin-o', tone: 'gold' },
  { key: 'commission', label: t('walletPage.assetCommission'), unit: CURRENCY_CODE, value: wallet.commission, icon: 'balance-list-o', tone: 'green' },
])

function goRecharge() {
  router.push('/recharge')
}

function goWithdraw() {
  router.push('/withdraw')
}

function goRechargeRecord() {
  router.push('/rechargeRecord')
}

function goWithdrawRecord() {
  router.push('/withdrawRecord')
}

function goTransform() {
  router.push('/transform')
}

function formatPlain(value: number) {
  return truncate2(Number(value || 0)).toFixed(2)
}

function formatTxAmount(value: number) {
  return formatCurrency(value, { signed: true, spaceBetweenSymbolAndAmount: true })
}

function formatTxTime(raw: string) {
  const d = new Date(raw)
  if (Number.isNaN(d.getTime()))
    return raw || ''
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${month}/${day} ${hh}:${mm}`
}

function mapCashTypeTitle(type: number) {
  const map: Record<number, string> = {
    1: t('walletPage.typeSend'),
    2: t('walletPage.typeGrabNoThunderWin'),
    3: t('walletPage.typeGrabThunderLoss'),
    4: t('walletPage.typeSendThunderWin'),
    5: t('walletPage.typeLuckyCommission'),
    6: t('walletPage.typeRecharge'),
    7: t('walletPage.typeManualAdd'),
    8: t('walletPage.typeManualDeduct'),
    9: t('walletPage.typeWithdrawApply'),
    10: t('walletPage.typeWithdrawReturn'),
    11: t('walletPage.typeRebateTransfer'),
    12: t('walletPage.typeLuckyExpiredRefund'),
    13: t('walletPage.typeLotteryConsume'),
    14: t('walletPage.typeLotteryWin'),
    15: t('walletPage.typeVipUpgradeReward'),
    16: t('walletPage.typeFirstRechargeGift'),
    17: t('walletPage.typeTodayFirstRechargeGift'),
    18: t('walletPage.typeLuckyNumsGift'),
    20: t('walletPage.typeCheckInGift'),
  }
  return map[Number(type) || 0] || t('walletPage.typeAccountChange')
}

function mapTxItem(item: any): WalletTx {
  return {
    id: String(`${item?.userId || 0}_${item?.createdAt || ''}_${item?.type || 0}_${item?.amount || 0}`),
    title: mapCashTypeTitle(Number(item?.type || 0)),
    time: formatTxTime(item?.createdAt || ''),
    amount: Number(item?.amount || 0),
  }
}

async function loadWallet() {
  try {
    const { data } = await getCurrentTgUserInfo()
    wallet.balance = Number(data?.balance || 0)
    wallet.commission = Number(data?.rebate_amount || 0)
  }
  catch {
    wallet.balance = 0
    wallet.commission = 0
  }
}

async function loadTxList(reset = false) {
  if (txFetching.value)
    return

  if (!reset && txFinished.value)
    return

  const page = reset ? 0 : txPage.value
  txFetching.value = true
  txLoading.value = true
  try {
    const { data } = await getAppCashHistoryList({
      currentPage: page,
      pageSize: txPageSize,
    })
    const list = (data?.list || []).map((item: any) => mapTxItem(item))
    if (reset)
      txList.value = list
    else
      txList.value = [...txList.value, ...list]

    txTotal.value = Number(data?.total || 0)
    txPage.value = page + 1
    txFinished.value = txList.value.length >= txTotal.value
  }
  catch {
    if (reset)
      txList.value = []
    txFinished.value = true
  }
  finally {
    txFetching.value = false
    txLoading.value = false
  }
}

function onLoadTxMore() {
  void loadTxList(false)
}

onMounted(() => {
  loadWallet()
  void loadTxList(true)
})
</script>

<template>
  <div class="wallet-page">
    <section class="asset-card card">
      <p class="asset-label">
        {{ t('walletPage.totalAsset') }}
      </p>
      <p class="asset-value">
        <CoinAmount :text="formatPlain(totalAsset)" />
      </p>

      <div class="action-row">
        <button type="button" class="asset-action recharge" @click="goRecharge">
          <span class="action-icon">
            <van-icon name="gold-coin-o" />
          </span>
          <span>{{ t('walletPage.recharge') }}</span>
        </button>
        <button type="button" class="asset-action withdraw" @click="goWithdraw">
          <span class="action-icon">
            <van-icon name="balance-pay" />
          </span>
          <span>{{ t('walletPage.withdraw') }}</span>
        </button>
        <button type="button" class="asset-action transfer" @click="goTransform">
          <span class="action-icon">
            <van-icon name="exchange" />
          </span>
          <span>{{ t('walletPage.transfer') }}</span>
        </button>
      </div>

      <div class="record-row">
        <button type="button" class="record-action" @click="goRechargeRecord">
          <van-icon name="orders-o" />
          <span>{{ t('profilePage.accountRechargeRecord') }}</span>
        </button>
        <button type="button" class="record-action" @click="goWithdrawRecord">
          <van-icon name="records-o" />
          <span>{{ t('profilePage.accountWithdrawRecord') }}</span>
        </button>
      </div>
    </section>

    <section class="card list-card">
      <div class="list-header">
        <p>{{ t('walletPage.assetDetails') }}</p>
        <van-icon name="arrow" />
      </div>
      <article v-for="item in assetRows" :key="item.key" class="list-row">
        <div class="row-left">
          <span class="row-icon" :class="item.tone">
            <van-icon :name="item.icon" />
          </span>
          <div class="row-meta">
            <p>{{ item.label }}</p>
            <span>{{ item.unit }}</span>
          </div>
        </div>
        <p class="row-value">
          <CoinAmount :text="formatPlain(item.value)" />
        </p>
      </article>
    </section>

    <section class="card list-card">
      <div class="list-header">
        <p>{{ t('walletPage.recentTx') }}</p>
        <van-icon name="arrow" />
      </div>
      <van-list
        v-model:loading="txLoading" :finished="txFinished" :immediate-check="false"
        :finished-text="t('walletPage.noMore')" @load="onLoadTxMore"
      >
        <article v-for="item in txList" :key="item.id" class="list-row">
          <div class="row-left">
            <span class="row-icon tx">
              <van-icon name="records-o" />
            </span>
            <div class="row-meta">
              <p>{{ item.title }}</p>
              <span>{{ item.time }}</span>
            </div>
          </div>
          <p class="row-value" :class="{ income: item.amount > 0, expense: item.amount < 0 }">
            <CoinAmount :text="formatTxAmount(item.amount)" />
          </p>
        </article>
      </van-list>
      <div v-if="!txLoading && txList.length === 0" class="wallet-empty">
        {{ t('walletPage.emptyTx') }}
      </div>
    </section>
  </div>
</template>

<style scoped>
.wallet-page {
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
}

.card {
  position: relative;
  overflow: hidden;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.96), rgba(54, 0, 0, 0.97));
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.38);
  box-shadow:
    0 14px 28px rgba(0, 0, 0, 0.32),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.asset-card {
  margin-top: 10px;
  padding: 24px 16px 18px;
  text-align: center;
}

.asset-label {
  margin: 0;
  color: rgba(255, 229, 186, 0.66);
  font-size: 13px;
  font-weight: 600;
}

.asset-value {
  margin: 8px 0 0;
  color: #ffd87f;
  font-size: 28px;
  line-height: 1;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.action-row {
  margin-top: 20px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.asset-action {
  min-height: 78px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 16px;
  background: rgba(255, 248, 214, 0.05);
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 700;
}

.asset-action.recharge {
  color: #ffd87f;
}

.asset-action.withdraw {
  color: #ffd3a0;
}

.asset-action.transfer {
  color: #ffe7b4;
}

.action-icon {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  border: 1px solid rgba(255, 248, 214, 0.16);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.asset-action.recharge .action-icon {
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.2), rgba(94, 10, 0, 0.28));
}

.asset-action.withdraw .action-icon {
  background: linear-gradient(180deg, rgba(255, 207, 138, 0.2), rgba(94, 10, 0, 0.28));
}

.asset-action.transfer .action-icon {
  background: linear-gradient(180deg, rgba(255, 248, 214, 0.18), rgba(94, 10, 0, 0.28));
}

.record-row {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.record-action {
  min-height: 42px;
  padding: 0 12px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  border-radius: 14px;
  background: rgba(255, 248, 214, 0.06);
  color: #ffe7b4;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 800;
}

.record-action:active {
  transform: translateY(1px);
}

.record-action :deep(.van-icon) {
  flex: 0 0 auto;
  color: #ffd87f;
  font-size: 16px;
}

.record-action span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-card {
  margin-top: 10px;
}

.list-header {
  height: 52px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.list-header p {
  margin: 0;
  color: #fff0c9;
  font-size: 15px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.list-header :deep(.van-icon) {
  color: rgba(255, 229, 186, 0.4);
  font-size: 18px;
}

.list-row {
  min-height: 58px;
  padding: 12px 16px;
  border-top: 1px solid rgba(212, 175, 55, 0.12);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.row-left {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.row-icon {
  width: 32px;
  height: 32px;
  border-radius: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
}

.row-icon.gold,
.row-icon.tx {
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.18), rgba(94, 10, 0, 0.24));
}

.row-icon.green {
  background: linear-gradient(180deg, rgba(255, 248, 214, 0.18), rgba(94, 10, 0, 0.24));
}

.row-meta {
  min-width: 0;
}

.row-meta p {
  margin: 0;
  color: #fff0c9;
  font-size: 14px;
  font-weight: 600;
}

.row-meta span {
  display: block;
  margin-top: 2px;
  color: rgba(255, 229, 186, 0.58);
  font-size: 11px;
}

.row-value {
  margin: 0;
  color: #fff0c9;
  font-size: 14px;
  line-height: 1;
  font-weight: 700;
  white-space: nowrap;
}

.row-value.income {
  color: #ffd87f;
}

.row-value.expense {
  color: #ffb7a7;
}

.wallet-empty {
  padding: 18px 16px 22px;
  text-align: center;
  color: rgba(255, 229, 186, 0.56);
  font-size: 13px;
}
</style>

<route lang="json5">
{
  name: 'Wallet'
}
</route>
