<script setup lang="ts">
import type { ApiResult, AppOrderHistoryItem, AppOrderHistoryResp } from '@/api/user'

const props = defineProps<{
  kind: 'recharge' | 'withdraw'
  loader: (data: { currentPage: number, pageSize: number }) => Promise<ApiResult<AppOrderHistoryResp>>
}>()

const { t } = useI18n()

const pageSize = 20
const currentPage = ref(0)
const total = ref(0)
const records = ref<AppOrderHistoryItem[]>([])
const loading = ref(false)
const fetching = ref(false)
const finished = ref(false)
const hasFetched = ref(false)

const showEmpty = computed(() => hasFetched.value && !loading.value && records.value.length === 0)

function formatAmount(item: AppOrderHistoryItem, key: 'amount' | 'netAmount') {
  const symbol = item.currencySymbol || item.currency || ''
  const value = Number(item[key] || 0).toFixed(2)
  return symbol ? `${symbol} ${value}` : value
}

function formatOptionalAmount(item: AppOrderHistoryItem, value?: number) {
  const symbol = item.currencySymbol || item.currency || ''
  const text = Number(value || 0).toFixed(2)
  return symbol ? `${symbol} ${text}` : text
}

function formatTime(raw: string) {
  const d = new Date(raw)
  if (Number.isNaN(d.getTime()))
    return raw || '--'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}`
}

function statusText(status: number) {
  if (props.kind === 'recharge') {
    const map: Record<number, string> = {
      0: t('orderRecordPage.rechargeStatusPending'),
      1: t('orderRecordPage.rechargeStatusSuccess'),
      2: t('orderRecordPage.rechargeStatusFailed'),
    }
    return map[status] || t('orderRecordPage.statusUnknown')
  }

  const map: Record<number, string> = {
    0: t('orderRecordPage.withdrawStatusReview'),
    1: t('orderRecordPage.withdrawStatusPending'),
    2: t('orderRecordPage.withdrawStatusPaying'),
    3: t('orderRecordPage.withdrawStatusSuccess'),
    4: t('orderRecordPage.withdrawStatusFailed'),
    5: t('orderRecordPage.withdrawStatusCanceled'),
    6: t('orderRecordPage.withdrawStatusReturned'),
  }
  return map[status] || t('orderRecordPage.statusUnknown')
}

function statusTone(status: number) {
  if (props.kind === 'recharge')
    return status === 1 ? 'success' : status === 0 ? 'pending' : 'failed'
  return status === 3 ? 'success' : [0, 1, 2].includes(status) ? 'pending' : 'failed'
}

async function loadRecords(reset = false) {
  if (fetching.value)
    return
  if (!reset && finished.value)
    return

  const page = reset ? 0 : currentPage.value
  fetching.value = true
  loading.value = true
  try {
    const res = await props.loader({ currentPage: page, pageSize })
    const list = res?.data?.list || []
    records.value = reset ? list : [...records.value, ...list]
    total.value = Number(res?.data?.total || 0)
    currentPage.value = page + 1
    finished.value = records.value.length >= total.value
  }
  catch {
    if (reset)
      records.value = []
    finished.value = true
  }
  finally {
    fetching.value = false
    loading.value = false
    hasFetched.value = true
  }
}

function onLoadMore() {
  void loadRecords(false)
}

onMounted(() => {
  void loadRecords(true)
})
</script>

<template>
  <van-list
    v-model:loading="loading"
    :finished="finished"
    :immediate-check="false"
    :finished-text="t('orderRecordPage.noMore')"
    @load="onLoadMore"
  >
    <section class="record-list">
      <article v-for="item in records" :key="item.orderNo" class="record-card">
        <div class="record-head">
          <div class="order-meta">
            <span class="order-label">{{ t('orderRecordPage.orderNo') }}</span>
            <strong class="order-no">{{ item.orderNo }}</strong>
          </div>
          <span class="status-badge" :class="statusTone(item.status)">
            {{ statusText(item.status) }}
          </span>
        </div>

        <div class="amount-grid">
          <div class="amount-box">
            <span>{{ t('orderRecordPage.originalAmount') }}</span>
            <strong>{{ formatAmount(item, 'amount') }}</strong>
          </div>
          <div class="amount-box converted">
            <span>{{ t('orderRecordPage.convertedAmount') }}</span>
            <strong>{{ formatAmount(item, 'netAmount') }}</strong>
          </div>
        </div>

        <div class="extra-grid">
          <div v-if="kind === 'recharge'" class="extra-row">
            <span>{{ t('orderRecordPage.bonusAmount') }}</span>
            <strong>{{ formatOptionalAmount(item, item.bonusAmount) }}</strong>
          </div>
          <div v-if="kind === 'withdraw'" class="extra-row">
            <span>{{ t('orderRecordPage.fee') }}</span>
            <strong>{{ formatOptionalAmount(item, item.fee) }}</strong>
          </div>
          <div v-if="kind === 'withdraw' && item.rejectReason" class="extra-row reject">
            <span>{{ t('orderRecordPage.rejectReason') }}</span>
            <strong>{{ item.rejectReason }}</strong>
          </div>
        </div>

        <div class="record-foot">
          <span>{{ t('orderRecordPage.currency') }}: {{ item.currencySymbol || item.currency || '--' }}</span>
          <span>{{ formatTime(item.time) }}</span>
        </div>
      </article>
    </section>
  </van-list>

  <section v-if="showEmpty" class="empty-card">
    <van-icon name="description" />
    <p>{{ t('orderRecordPage.empty') }}</p>
  </section>
</template>

<style scoped>
.record-list {
  display: grid;
  gap: 10px;
}

.record-card {
  position: relative;
  overflow: hidden;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.96), rgba(54, 0, 0, 0.97));
  padding: 14px;
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.record-card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.record-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.order-meta {
  min-width: 0;
  display: grid;
  gap: 5px;
}

.order-label {
  color: rgba(255, 229, 186, 0.58);
  font-size: 11px;
}

.order-no {
  color: #fff0c9;
  font-size: 13px;
  line-height: 1.25;
  word-break: break-all;
}

.status-badge {
  flex: 0 0 auto;
  min-width: 62px;
  border-radius: 999px;
  padding: 5px 9px;
  text-align: center;
  font-size: 11px;
  font-weight: 800;
}

.status-badge.success {
  background: rgba(212, 175, 55, 0.16);
  color: #ffd87f;
}

.status-badge.pending {
  background: rgba(255, 248, 214, 0.1);
  color: #ffe7bf;
}

.status-badge.failed {
  background: rgba(255, 128, 104, 0.14);
  color: #ffb7a7;
}

.amount-grid {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.amount-box {
  min-height: 64px;
  border-radius: 12px;
  border: 1px solid rgba(212, 175, 55, 0.16);
  background: rgba(255, 248, 214, 0.05);
  padding: 10px;
  display: grid;
  align-content: center;
  gap: 6px;
}

.amount-box.converted {
  border-color: rgba(212, 175, 55, 0.3);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.14), rgba(116, 24, 0, 0.2));
}

.amount-box span {
  color: rgba(255, 229, 186, 0.58);
  font-size: 11px;
}

.amount-box strong {
  color: #ffd87f;
  font-size: 16px;
  line-height: 1;
  word-break: break-word;
}

.record-foot {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: rgba(255, 229, 186, 0.58);
  font-size: 11px;
}

.extra-grid {
  margin-top: 10px;
  display: grid;
  gap: 6px;
}

.extra-row {
  min-height: 34px;
  border-radius: 10px;
  background: rgba(255, 248, 214, 0.04);
  border: 1px solid rgba(212, 175, 55, 0.12);
  padding: 7px 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.extra-row span {
  flex: 0 0 auto;
  color: rgba(255, 229, 186, 0.58);
  font-size: 11px;
}

.extra-row strong {
  min-width: 0;
  color: #fff0c9;
  font-size: 12px;
  line-height: 1.3;
  text-align: right;
  word-break: break-word;
}

.extra-row.reject {
  align-items: flex-start;
  background: rgba(255, 128, 104, 0.08);
  border-color: rgba(255, 128, 104, 0.16);
}

.extra-row.reject strong {
  color: #ffb7a7;
}

.empty-card {
  min-height: 180px;
  margin-top: 10px;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.24);
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.8), rgba(54, 0, 0, 0.86));
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 8px;
  color: rgba(255, 229, 186, 0.56);
}

.empty-card :deep(.van-icon) {
  font-size: 18px;
  color: rgba(255, 229, 186, 0.34);
}

.empty-card p {
  margin: 0;
  font-size: 14px;
}
</style>
