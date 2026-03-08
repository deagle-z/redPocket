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
