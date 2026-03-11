<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { createRechargeOrder, getCurrentTgUserInfo } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { CURRENCY_CODE, CURRENCY_SYMBOL, formatCurrency } from '@/utils/currency'
const { t } = useI18n()

const router = useRouter()

const balance = ref(0)

const channels = [
  { id: 'gtpay', name: 'GTPAY(maya)' },
  { id: 'hipay', name: 'HIPAY' },
]

const amountOptions = [
  100,
  200,
  500,
  1000,
  5000,
  10000,
  20000,
  50000,
  'custom',
]

const payMethods = [
  {
    id: 'gcash',
    name: 'GCash QR',
    subKey: 'rechargePage.paySubQr',
    logo: 'G',
  },
  {
    id: 'maya',
    name: 'Maya',
    subKey: 'rechargePage.paySubWallet',
    logo: 'M',
  },
]

const selectedChannel = ref(channels[0].id)
const selectedAmount = ref<number | 'custom'>(amountOptions[0] as number)
const customAmount = ref('')
const selectedPay = ref(payMethods[0].id)
const submitLoading = ref(false)

const displayAmount = computed(() => {
  if (selectedAmount.value === 'custom') {
    return customAmount.value ? Number(customAmount.value) : 0
  }
  return selectedAmount.value
})

const canSubmit = computed(() => Number(displayAmount.value) > 0 && !submitLoading.value)

function chooseAmount(value: number | 'custom') {
  selectedAmount.value = value
  if (value !== 'custom') {
    customAmount.value = ''
  }
}

function goBack() {
  router.back()
}

function showCenterToast(message: string) {
  showToast({
    message,
    position: 'middle',
    teleport: '#app',
    wordBreak: 'break-word',
  })
}

function showHelpTip() {
  showCenterToast(t('rechargePage.helpTip'))
}

async function loadBalance() {
  try {
    const { data } = await getCurrentTgUserInfo()
    balance.value = Number(data?.balance ?? 0)
  }
  catch {
    balance.value = 0
  }
}

async function handleSubmitRecharge() {
  if (!canSubmit.value)
    return
  const amount = Number(displayAmount.value)
  if (!amount || amount <= 0) {
    showCenterToast(t('rechargePage.invalidAmount'))
    return
  }

  submitLoading.value = true
  try {
    const { data } = await createRechargeOrder({
      amount,
      channel: selectedChannel.value,
      payMethod: selectedPay.value,
      currency: CURRENCY_CODE,
    })

    if (data?.payUrl) {
      showCenterToast(t('rechargePage.orderToPay'))
      window.location.href = data.payUrl
      return
    }

    if (data?.devCallback) {
      showCenterToast(t('rechargePage.orderRechargeSuccess', { orderNo: data.orderNo }))
      await loadBalance()
      return
    }

    showCenterToast(t('rechargePage.orderSuccess', { orderNo: data?.orderNo || '--' }))
  }
  catch {
    showCenterToast(t('rechargePage.orderFailed'))
  }
  finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  loadBalance()
})
</script>

<template>
  <div class="recharge-page">
    <AppPageHeader class="recharge-header" :title="t('rechargePage.title')" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

    <section class="card balance-card">
      <div>
        <p class="card-label">
          {{ t('rechargePage.currentBalance') }}
        </p>
        <p class="card-value">
          {{ formatCurrency(balance) }}
        </p>
      </div>
      <span class="card-chip">{{ CURRENCY_SYMBOL }}</span>
    </section>

    <section class="card">
      <h2>{{ t('rechargePage.channelTitle') }}</h2>
      <div class="pill-group">
        <button
          v-for="item in channels" :key="item.id" type="button" class="pill"
          :class="{ active: selectedChannel === item.id }" @click="selectedChannel = item.id"
        >
          {{ item.name }}
        </button>
      </div>
    </section>

    <section class="card">
      <h2>{{ t('rechargePage.amountTitle') }}</h2>
      <div class="amount-grid">
        <button
          v-for="item in amountOptions" :key="item" type="button" class="amount-item"
          :class="{ active: selectedAmount === item }" @click="chooseAmount(item as number | 'custom')"
        >
          <span v-if="item !== 'custom'">{{ item }}</span>
          <span v-else>{{ t('rechargePage.custom') }}</span>
        </button>
      </div>
      <van-field
        v-if="selectedAmount === 'custom'" v-model="customAmount" type="number" :label="t('rechargePage.customAmount')"
        :placeholder="t('rechargePage.customAmountPlaceholder')" class="custom-input"
      />
    </section>

    <section class="card">
      <h2>{{ t('rechargePage.payMethodTitle') }}</h2>
      <div class="pay-list">
        <button
          v-for="method in payMethods" :key="method.id" type="button" class="pay-item"
          :class="{ active: selectedPay === method.id }" @click="selectedPay = method.id"
        >
          <div class="pay-left">
            <span class="pay-logo">{{ method.logo }}</span>
            <div>
              <p class="pay-name">
                {{ method.name }}
              </p>
              <p class="pay-sub">
                {{ t(method.subKey) }}
              </p>
            </div>
          </div>
          <span class="pay-check">
            <van-icon name="success" />
          </span>
        </button>
      </div>
    </section>

    <van-button
      type="primary"
      round
      block
      class="confirm-btn"
      :loading="submitLoading"
      :disabled="!canSubmit"
      @click="handleSubmitRecharge"
    >
      {{ t('rechargePage.submit') }}
    </van-button>
  </div>
</template>

<route lang="json5">
{
  name: "Recharge"
}
</route>

<style scoped>
.recharge-page {
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
  color: #fff0c9;
  padding: 0 12px calc(90px + env(safe-area-inset-bottom));
}

.recharge-header {
  margin-bottom: 12px;
}

.card {
  position: relative;
  overflow: hidden;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  padding: 14px 16px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
  margin-bottom: 12px;
}

.card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.balance-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-label {
  font-size: 13px;
  color: rgba(255, 229, 186, 0.62);
  margin: 0 0 6px;
}

.card-value {
  font-size: 26px;
  font-weight: 800;
  margin: 0;
  color: #ffd87f;
  letter-spacing: 0.04em;
}

.card-chip {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.card h2 {
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #ffd98b;
  margin: 0 0 12px;
}

.pill-group {
  display: flex;
  gap: 10px;
}

.pill {
  flex: 1;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
  padding: 12px;
  font-weight: 700;
  color: #fff0c9;
}

.pill.active {
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  border-color: rgba(255, 248, 214, 0.34);
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.amount-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.amount-item {
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
  padding: 14px 0;
  font-weight: 700;
  color: #fff0c9;
  font-size: 14px;
}

.amount-item.active {
  border-color: rgba(255, 248, 214, 0.34);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.18), rgba(116, 24, 0, 0.28));
  color: #ffd87f;
}

.custom-input {
  margin-top: 12px;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.18);
  background: rgba(255, 248, 214, 0.05);
}

:deep(.custom-input .van-field__label),
:deep(.custom-input .van-field__control) {
  color: #fff0c9;
}

:deep(.custom-input .van-field__control::placeholder) {
  color: rgba(255, 229, 186, 0.4);
}

.pay-list {
  display: grid;
  gap: 10px;
}

.pay-item {
  width: 100%;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.05);
  padding: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  text-align: left;
}

.pay-item.active {
  border-color: rgba(255, 248, 214, 0.34);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.16), rgba(116, 24, 0, 0.24));
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.pay-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pay-logo {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
}

.pay-name {
  font-weight: 700;
  margin: 0;
  color: #fff0c9;
}

.pay-sub {
  margin: 2px 0 0;
  font-size: 12px;
  color: rgba(255, 229, 186, 0.58);
}

.pay-check {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 1px solid rgba(212, 175, 55, 0.24);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffd87f;
  background: rgba(255, 248, 214, 0.08);
}

.pay-item:not(.active) .pay-check {
  color: transparent;
}

.confirm-btn {
  margin-top: 18px;
  border: none;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%) !important;
  color: #5a1b00 !important;
  font-weight: 800;
  letter-spacing: 0.08em;
  box-shadow: 0 12px 22px rgba(75, 25, 0, 0.28);
}
</style>
