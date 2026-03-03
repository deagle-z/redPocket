<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getAppCashHistoryList, getCurrentTgUserInfo } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'

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
  { key: 'balance', label: '余额', unit: 'PHP', value: wallet.balance, emoji: '💰', tone: 'gold' },
  { key: 'commission', label: '佣金', unit: 'PHP', value: wallet.commission, emoji: '💵', tone: 'green' },
])

function goBack() {
  router.back()
}

function goRecharge() {
  router.push('/recharge')
}

function goWithdraw() {
  router.push('/withdraw')
}

function goTransform() {
  router.push('/transform')
}

function formatPlain(value: number) {
  return Number(value || 0).toFixed(2)
}

function formatTxAmount(value: number) {
  const sign = value >= 0 ? '+' : '-'
  return `${sign}₱ ${Math.abs(value).toFixed(2)}`
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

function mapCashTypeTitle(type: number, cashMark?: string, cashDesc?: string) {
  if (cashDesc)
    return cashDesc
  if (cashMark)
    return cashMark
  const map: Record<number, string> = {
    1: '发送红包',
    2: '抢红包收益',
    3: '抢红包中雷',
    4: '发包雷收益',
    6: '充值到账',
    7: '手工加款',
    8: '手工扣款',
    9: '提现申请',
    10: '提现退回',
    11: '佣金转余额',
    12: '红包过期退回',
  }
  return map[Number(type) || 0] || '账户变动'
}

function mapTxItem(item: any): WalletTx {
  return {
    id: String(`${item?.userId || 0}_${item?.createdAt || ''}_${item?.type || 0}_${item?.amount || 0}`),
    title: mapCashTypeTitle(Number(item?.type || 0), item?.cashMark, item?.cashDesc),
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
    <AppPageHeader title="我的钱包" @back="goBack">
      <template #right>
        <van-icon name="ellipsis" />
      </template>
    </AppPageHeader>

    <section class="asset-card card">
      <p class="asset-label">
        总资产 (₱)
      </p>
      <p class="asset-value">
        {{ formatPlain(totalAsset) }}
      </p>

      <div class="action-row">
        <button type="button" class="asset-action recharge" @click="goRecharge">
          <span class="action-icon">
            <van-icon name="gold-coin-o" />
          </span>
          <span>充值</span>
        </button>
        <button type="button" class="asset-action withdraw" @click="goWithdraw">
          <span class="action-icon">
            <van-icon name="balance-pay" />
          </span>
          <span>提现</span>
        </button>
        <button type="button" class="asset-action transfer" @click="goTransform">
          <span class="action-icon">
            <van-icon name="exchange" />
          </span>
          <span>转账</span>
        </button>
      </div>
    </section>

    <section class="card list-card">
      <div class="list-header">
        <p>资产明细</p>
        <van-icon name="arrow" />
      </div>
      <article
        v-for="item in assetRows"
        :key="item.key"
        class="list-row"
      >
        <div class="row-left">
          <span class="row-icon" :class="item.tone">{{ item.emoji }}</span>
          <div class="row-meta">
            <p>{{ item.label }}</p>
            <span>{{ item.unit }}</span>
          </div>
        </div>
        <p class="row-value">
          {{ formatPlain(item.value) }}
        </p>
      </article>
    </section>

    <section class="card list-card">
      <div class="list-header">
        <p>最近交易</p>
        <van-icon name="arrow" />
      </div>
      <van-list
        v-model:loading="txLoading"
        :finished="txFinished"
        :immediate-check="false"
        finished-text="没有更多了"
        @load="onLoadTxMore"
      >
        <article
          v-for="item in txList"
          :key="item.id"
          class="list-row"
        >
          <div class="row-left">
            <span class="row-icon tx">💬</span>
            <div class="row-meta">
              <p>{{ item.title }}</p>
              <span>{{ item.time }}</span>
            </div>
          </div>
          <p class="row-value" :class="{ income: item.amount > 0, expense: item.amount < 0 }">
            {{ formatTxAmount(item.amount) }}
          </p>
        </article>
      </van-list>
      <div v-if="!txLoading && txList.length === 0" class="wallet-empty">
        暂无交易流水
      </div>
    </section>
  </div>
</template>

<style scoped>
.wallet-page {
  min-height: 100vh;
  background: #f4f6fa;
  padding: 0 0 calc(14px + env(safe-area-inset-bottom));
}

.card {
  background: #fff;
}

.asset-card {
  margin-top: 10px;
  padding: 24px 16px 20px;
  text-align: center;
}

.asset-label {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
  font-weight: 500;
}

.asset-value {
  margin: 8px 0 0;
  color: var(--color-primary);
  font-size: 18px;
  line-height: 1;
  font-weight: 700;
}

.action-row {
  margin-top: 20px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.asset-action {
  border: 0;
  background: transparent;
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
}

.asset-action.recharge {
  color: var(--color-primary);
}

.asset-action.withdraw {
  color: #ff9500;
}

.asset-action.transfer {
  color: #4a5cff;
}

.action-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.asset-action.recharge .action-icon {
  background: #f0f7ff;
}

.asset-action.withdraw .action-icon {
  background: #fff4e5;
}

.asset-action.transfer .action-icon {
  background: #f0f7ff;
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
  color: #1a1a2e;
  font-size: 15px;
  font-weight: 700;
}

.list-header :deep(.van-icon) {
  color: #d1d5db;
  font-size: 18px;
}

.list-row {
  min-height: 58px;
  padding: 12px 16px;
  border-top: 1px solid #f5f5f5;
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
}

.row-icon.gold,
.row-icon.tx {
  background: #fff4e5;
}

.row-icon.green {
  background: var(--color-primary-soft);
}

.row-meta {
  min-width: 0;
}

.row-meta p {
  margin: 0;
  color: #1a1a2e;
  font-size: 14px;
  font-weight: 600;
}

.row-meta span {
  display: block;
  margin-top: 2px;
  color: #9ca3af;
  font-size: 11px;
}

.row-value {
  margin: 0;
  color: #1a1a2e;
  font-size: 14px;
  line-height: 1;
  font-weight: 700;
  white-space: nowrap;
}

.row-value.income {
  color: var(--color-primary);
}

.row-value.expense {
  color: #ff4d4f;
}

.wallet-empty {
  padding: 18px 16px 22px;
  text-align: center;
  color: #9ca3af;
  font-size: 13px;
}
</style>

<route lang="json5">
{
  name: 'Wallet'
}
</route>

