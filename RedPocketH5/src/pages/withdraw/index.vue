<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { CURRENCY_SYMBOL, formatCurrency } from '@/utils/currency'
const { t } = useI18n()

const router = useRouter()

const balance = ref(0)
const frozen = ref(0)

const channels = [
  { id: 'hipay', name: 'HIPAY' },
  { id: 'gtpay', name: 'GTPAY(maya)' },
]

const payMethods = [
  {
    id: 'gcash',
    name: 'GCash Wallet',
    fee: '0.0% + 10.00',
    logo: 'G',
  },
  {
    id: 'maya',
    name: 'Maya',
    fee: '0.5% + 5.00',
    logo: 'M',
  },
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

const selectedChannel = ref(channels[0].id)
const selectedPay = ref(payMethods[0].id)
const selectedAmount = ref<number | 'custom'>(amountOptions[0] as number)
const customAmount = ref('')

const receiverAccount = ref('')
const receiverName = ref('')
const receiverEmail = ref('')

function chooseAmount(value: number | 'custom') {
  selectedAmount.value = value
  if (value !== 'custom') {
    customAmount.value = ''
  }
}

function goBack() {
  router.back()
}

function showHelpTip() {
  showToast(t('withdrawPage.helpTip'))
}
</script>

<template>
  <div class="withdraw-page theme-withdraw">
    <AppPageHeader :title="t('withdrawPage.title')" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

    <section class="card balance-card">
      <div>
        <p class="card-label">
          {{ t('withdrawPage.currentBalance') }}
        </p>
        <p class="card-value">
          {{ formatCurrency(balance) }}
        </p>
        <p class="card-sub">
          {{ t('withdrawPage.frozenBalance', { amount: formatCurrency(frozen) }) }}
        </p>
      </div>
      <span class="card-chip">{{ CURRENCY_SYMBOL }}</span>
    </section>

    <section class="card">
      <h2>{{ t('withdrawPage.channelTitle') }}</h2>
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
      <h2>{{ t('withdrawPage.payMethodTitle') }}</h2>
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
                {{ t('withdrawPage.feePrefix', { fee: method.fee }) }}
              </p>
            </div>
          </div>
          <span class="pay-check">
            <van-icon name="success" />
          </span>
        </button>
      </div>
    </section>

    <section class="card">
      <h2 class="section-title">
        {{ t('withdrawPage.receiverTitle') }}
      </h2>
      <div class="form-list">
        <van-field v-model="receiverAccount" :label="t('withdrawPage.receiverAccount')" :placeholder="t('withdrawPage.receiverAccountPlaceholder')" class="form-item" />
        <van-field v-model="receiverName" :label="t('withdrawPage.receiverName')" :placeholder="t('withdrawPage.receiverNamePlaceholder')" class="form-item" />
        <van-field v-model="receiverEmail" :label="t('withdrawPage.receiverEmail')" :placeholder="t('withdrawPage.receiverEmailPlaceholder')" class="form-item" />
      </div>
    </section>

    <section class="card">
      <h2 class="section-title">
        {{ t('withdrawPage.amountTitle') }}
      </h2>
      <van-field
        v-model="customAmount" type="number" :label="t('withdrawPage.amountLabel')" :placeholder="t('withdrawPage.amountPlaceholder')" class="custom-input"
        @focus="selectedAmount = 'custom'"
      />
      <div class="amount-grid">
        <button
          v-for="item in amountOptions" :key="item" type="button" class="amount-item"
          :class="{ active: selectedAmount === item }" @click="chooseAmount(item as number | 'custom')"
        >
          <span v-if="item !== 'custom'">{{ CURRENCY_SYMBOL }}{{ item }}</span>
          <span v-else>{{ t('withdrawPage.custom') }}</span>
        </button>
      </div>

      <div class="balance-breakdown">
        <div class="balance-header">
          <div class="balance-title">
            <span class="balance-icon">◎</span>
            {{ t('withdrawPage.availableBalance') }}
          </div>
          <div class="balance-amount">
            {{ formatCurrency(balance) }}
          </div>
        </div>
        <div class="balance-row">
          <div>
            <p class="row-title">
              {{ t('withdrawPage.normalBalance') }}
            </p>
            <p class="row-sub">
              {{ t('withdrawPage.codingRemain') }}
            </p>
          </div>
          <div class="row-right">
            <span>{{ CURRENCY_SYMBOL }}0.00</span>
            <span class="row-badge">{{ t('withdrawPage.withdrawable') }}</span>
          </div>
        </div>
        <div class="balance-row">
          <div>
            <p class="row-title">
              {{ t('withdrawPage.bonusBalance') }}
            </p>
            <p class="row-sub">
              {{ t('withdrawPage.codingRemain') }}
            </p>
          </div>
          <div class="row-right">
            <span>{{ CURRENCY_SYMBOL }}0.00</span>
            <span class="row-badge">{{ t('withdrawPage.withdrawable') }}</span>
          </div>
        </div>
        <div class="balance-row">
          <div>
            <p class="row-title">
              {{ t('withdrawPage.freezing') }}
            </p>
          </div>
          <div class="row-right">
            <span>-{{ CURRENCY_SYMBOL }}0.00</span>
          </div>
        </div>
      </div>

      <van-button type="primary" round block class="submit-btn">
        {{ t('withdrawPage.submit') }}
      </van-button>

      <div class="fee-card">
        <div class="fee-row">
          <span>{{ t('withdrawPage.serviceFee') }}</span>
          <span>{{ CURRENCY_SYMBOL }}0.00</span>
        </div>
        <div class="fee-row total">
          <span>{{ t('withdrawPage.actualDeduct') }}</span>
          <span>{{ CURRENCY_SYMBOL }}0.00</span>
        </div>
      </div>
    </section>

    <section class="card tips">
      <h2 class="tips-title">
        {{ t('withdrawPage.tipsTitle') }}
      </h2>
      <ol>
        <li>{{ t('withdrawPage.tips1', { amount: `${CURRENCY_SYMBOL}100` }) }}</li>
        <li>{{ t('withdrawPage.tips2') }}</li>
        <li>{{ t('withdrawPage.tips3') }}</li>
        <li>{{ t('withdrawPage.tips4') }}</li>
        <li>{{ t('withdrawPage.tips5') }}</li>
      </ol>
    </section>
  </div>
</template>

<route lang="json5">
{
  name: "Withdraw"
}
</route>

<style scoped>
.withdraw-page {
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
  border: 1px solid rgba(212, 175, 55, 0.34);
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
  padding: 14px 16px;
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

.card-label,
.card-sub,
.row-sub,
.tips li,
.pay-sub {
  color: rgba(255, 229, 186, 0.6);
}

.card-value,
.balance-amount {
  margin: 0;
  color: #ffd87f;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.card-sub {
  margin: 6px 0 0;
  font-size: 12px;
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
}

.card h2,
.section-title,
.tips-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #ffd98b;
}

.pill-group,
.pay-list,
.form-list {
  display: grid;
  gap: 10px;
}

.pill-group {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.pill,
.pay-item,
.amount-item,
.balance-breakdown,
.fee-card {
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
}

.pill,
.amount-item {
  padding: 12px;
  font-weight: 700;
  color: #fff0c9;
}

.pill.active,
.amount-item.active,
.pay-item.active {
  border-color: rgba(255, 248, 214, 0.34);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.16), rgba(116, 24, 0, 0.24));
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.pay-item {
  width: 100%;
  padding: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  text-align: left;
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

.pay-name,
.row-title {
  margin: 0;
  color: #fff0c9;
  font-weight: 700;
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
}

.pay-item:not(.active) .pay-check {
  color: transparent;
}

:deep(.form-item),
:deep(.custom-input) {
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.16);
  background: rgba(255, 248, 214, 0.05);
}

:deep(.form-item + .form-item) {
  margin-top: 10px;
}

:deep(.form-item .van-field__label),
:deep(.form-item .van-field__control),
:deep(.custom-input .van-field__label),
:deep(.custom-input .van-field__control) {
  color: #fff0c9;
}

:deep(.form-item .van-field__control::placeholder),
:deep(.custom-input .van-field__control::placeholder) {
  color: rgba(255, 229, 186, 0.4);
}

.amount-grid {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.amount-item {
  padding: 14px 0;
}

.balance-breakdown,
.fee-card {
  margin-top: 14px;
  padding: 12px;
}

.balance-header,
.balance-row,
.fee-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.balance-header {
  margin-bottom: 10px;
}

.balance-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #ffd98b;
  font-weight: 700;
}

.balance-icon,
.row-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
}

.balance-icon {
  width: 22px;
  height: 22px;
  background: rgba(212, 175, 55, 0.16);
}

.balance-row + .balance-row,
.fee-row + .fee-row {
  border-top: 1px solid rgba(212, 175, 55, 0.12);
  margin-top: 10px;
  padding-top: 10px;
}

.row-right,
.fee-row {
  color: #fff0c9;
  gap: 8px;
}

.row-badge {
  padding: 4px 8px;
  background: rgba(212, 175, 55, 0.14);
  color: #ffd87f;
  font-size: 11px;
}

.submit-btn {
  margin-top: 14px;
  border: none;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%) !important;
  color: #5a1b00 !important;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.fee-row.total {
  font-weight: 800;
  color: #ffd87f;
}

.tips ol {
  margin: 0;
  padding-left: 18px;
}

.tips li + li {
  margin-top: 6px;
}
</style>
