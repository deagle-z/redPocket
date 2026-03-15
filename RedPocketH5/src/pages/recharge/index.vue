<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import type { AppCountryItem, AppPayMethodItem, AppRechargeChannelItem, RechargeField } from '@/api/user'
import {
  createRechargeOrder,
  getAppCountries,
  getCountryRechargeFields,
  getCountryRechargeInfo,
  getCurrentTgUserInfo,
} from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { CURRENCY_CODE, formatCurrency } from '@/utils/currency'

const { t } = useI18n()
const router = useRouter()

const balance = ref(0)

// 国家列表
const countries = ref<AppCountryItem[]>([])
const selectedCountry = ref<AppCountryItem | null>(null)

// 充值通道与支付方式（按国家动态加载）
const channels = ref<AppRechargeChannelItem[]>([])
const selectedChannel = ref<AppRechargeChannelItem | null>(null)
const payMethods = computed<AppPayMethodItem[]>(() => selectedChannel.value?.methods ?? [])
const selectedPay = ref<AppPayMethodItem | null>(null)
const rechargeInfoLoading = ref(false)

// 充值字段
const rechargeFields = ref<RechargeField[]>([])
const fieldValues = ref<Record<string, string>>({})

const amountOptions = [100, 200, 500, 1000, 5000, 10000, 20000, 50000, 'custom']
const selectedAmount = ref<number | 'custom'>(amountOptions[0] as number)
const customAmount = ref('')
const submitLoading = ref(false)

const displayAmount = computed(() => {
  if (selectedAmount.value === 'custom')
    return customAmount.value ? Number(customAmount.value) : 0
  return selectedAmount.value
})

const canSubmit = computed(() =>
  Number(displayAmount.value) > 0
  && !submitLoading.value
  && !!selectedChannel.value
  && !rechargeInfoLoading.value,
)

function chooseAmount(value: number | 'custom') {
  selectedAmount.value = value
  if (value !== 'custom')
    customAmount.value = ''
}

function selectChannel(ch: AppRechargeChannelItem) {
  selectedChannel.value = ch
  selectedPay.value = ch.methods[0] ?? null
}

function goBack() {
  router.back()
}

function showCenterToast(message: string) {
  showToast({ message, position: 'middle', teleport: '#app', wordBreak: 'break-word' })
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

async function loadRechargeInfo(code: string) {
  rechargeInfoLoading.value = true
  channels.value = []
  selectedChannel.value = null
  selectedPay.value = null
  rechargeFields.value = []
  fieldValues.value = {}
  try {
    const [infoRes, fieldsRes] = await Promise.all([
      getCountryRechargeInfo(code),
      getCountryRechargeFields(code),
    ])
    channels.value = infoRes.data?.channels ?? []
    if (channels.value.length)
      selectChannel(channels.value[0])
    const fields = fieldsRes.data ?? []
    rechargeFields.value = fields
    const init: Record<string, string> = {}
    for (const f of fields)
      init[f.fieldKey] = f.defaultValue ?? ''
    fieldValues.value = init
  }
  catch {
    // ignore
  }
  finally {
    rechargeInfoLoading.value = false
  }
}

async function handleSelectCountry(country: AppCountryItem) {
  if (selectedCountry.value?.countryCode === country.countryCode)
    return
  selectedCountry.value = country
  await loadRechargeInfo(country.countryCode)
}

async function loadCountries() {
  try {
    const { data } = await getAppCountries()
    countries.value = data ?? []
    if (countries.value.length) {
      selectedCountry.value = countries.value[0]
      await loadRechargeInfo(countries.value[0].countryCode)
    }
  }
  catch {
    countries.value = []
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

  // 校验必填字段
  for (const f of rechargeFields.value) {
    if (f.isRequired === 1 && !fieldValues.value[f.fieldKey]?.trim()) {
      showCenterToast(f.errorTips || `${f.fieldLabel} 不能为空`)
      return
    }
  }

  submitLoading.value = true
  try {
    const { data } = await createRechargeOrder({
      amount,
      channel: selectedChannel.value?.channelCode ?? '',
      payMethod: selectedPay.value?.methodCode ?? '',
      currency: CURRENCY_CODE,
      extraFields: rechargeFields.value.length ? { ...fieldValues.value } : undefined,
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
  loadCountries()
})
</script>

<template>
  <div class="recharge-page">
    <AppPageHeader class="recharge-header" :title="t('rechargePage.title')" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

    <!-- 国家选择行 -->
    <div class="country-bar">
      <div class="country-scroll">
        <button
          v-for="country in countries"
          :key="country.countryCode"
          type="button"
          class="country-pill"
          :class="{ active: selectedCountry?.countryCode === country.countryCode }"
          @click="handleSelectCountry(country)"
        >
          <span class="country-code">{{ country.countryCode }}</span>
          <span class="country-name">{{ country.countryNameEn }}</span>
        </button>
      </div>
    </div>

    <section class="card balance-card">
      <div>
        <p class="card-label">
          {{ t('rechargePage.currentBalance') }}
        </p>
        <p class="card-value">
          <CoinAmount :text="formatCurrency(balance)" />
        </p>
      </div>
      <span class="card-chip"><img class="chip-coin" src="@/assets/svg/coin.svg" alt=""></span>
    </section>

    <!-- 充值通道 -->
    <section class="card">
      <h2>{{ t('rechargePage.channelTitle') }}</h2>
      <van-loading v-if="rechargeInfoLoading" size="20" color="#d4af37" class="section-loading" />
      <template v-else>
        <div v-if="channels.length" class="pill-group">
          <button
            v-for="ch in channels"
            :key="ch.channelCode"
            type="button"
            class="pill"
            :class="{ active: selectedChannel?.channelCode === ch.channelCode }"
            @click="selectChannel(ch)"
          >
            <img v-if="ch.icon" :src="ch.icon" class="channel-icon" alt="">
            {{ ch.channelName }}
          </button>
        </div>
        <p v-else class="empty-tip">
          {{ t('rechargePage.noChannel') || '暂无可用通道' }}
        </p>
      </template>
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

    <!-- 支付方式 -->
    <section v-if="payMethods.length" class="card">
      <h2>{{ t('rechargePage.payMethodTitle') }}</h2>
      <div class="pay-list">
        <button
          v-for="method in payMethods" :key="method.methodCode" type="button" class="pay-item"
          :class="{ active: selectedPay?.methodCode === method.methodCode }"
          @click="selectedPay = method"
        >
          <div class="pay-left">
            <div class="pay-logo">
              <img v-if="method.icon" :src="method.icon" class="pay-logo-img" alt="">
              <span v-else>{{ method.methodName.charAt(0) }}</span>
            </div>
            <div>
              <p class="pay-name">
                {{ method.methodName }}
              </p>
              <p class="pay-sub">
                {{ method.methodCode }}
              </p>
            </div>
          </div>
          <span class="pay-check">
            <van-icon name="success" />
          </span>
        </button>
      </div>
    </section>

    <!-- 充值字段 -->
    <section v-if="rechargeFields.length" class="card">
      <h2>{{ t('rechargePage.fillInfo') || '填写信息' }}</h2>
      <template v-for="field in rechargeFields" :key="field.fieldKey">
        <van-field
          v-model="fieldValues[field.fieldKey]"
          :type="field.fieldType === 'number' ? 'number' : field.fieldType === 'textarea' ? 'textarea' : 'text'"
          :label="field.fieldLabel"
          :placeholder="field.fieldPlaceholder || ''"
          :required="field.isRequired === 1"
          :maxlength="field.maxLength ?? undefined"
          class="custom-input recharge-field"
          rows="3"
        />
      </template>
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
  margin-bottom: 10px;
}

/* 国家选择条 */
.country-bar {
  margin-bottom: 12px;
  overflow: hidden;
}

.country-scroll {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
  scrollbar-width: none;
}

.country-scroll::-webkit-scrollbar {
  display: none;
}

.country-pill {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid rgba(212, 175, 55, 0.28);
  background: rgba(255, 248, 214, 0.06);
  color: rgba(255, 240, 201, 0.7);
  font-size: 13px;
  white-space: nowrap;
  transition: all 0.15s;
}

.country-pill.active {
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  border-color: transparent;
  color: #5a1b00;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(75, 25, 0, 0.3);
}

.country-code {
  font-weight: 700;
  font-size: 12px;
}

.country-name {
  font-size: 12px;
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
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.chip-coin {
  width: 100%;
  height: 100%;
  display: block;
}

.card h2 {
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #ffd98b;
  margin: 0 0 12px;
}

.section-loading {
  display: flex;
  justify-content: center;
  padding: 8px 0;
}

.empty-tip {
  font-size: 13px;
  color: rgba(255, 229, 186, 0.45);
  text-align: center;
  margin: 4px 0 0;
}

.pill-group {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.pill {
  flex: 1;
  min-width: 80px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
  padding: 12px 8px;
  font-weight: 700;
  color: #fff0c9;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.pill.active {
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  border-color: rgba(255, 248, 214, 0.34);
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.channel-icon {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  object-fit: contain;
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
  overflow: hidden;
  flex-shrink: 0;
}

.pay-logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
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

.recharge-field {
  margin-top: 10px;
}

.recharge-field:first-of-type {
  margin-top: 0;
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
