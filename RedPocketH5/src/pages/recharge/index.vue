<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

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
    sub: '扫码支付',
    logo: 'G',
  },
  {
    id: 'maya',
    name: 'Maya',
    sub: '电子钱包',
    logo: 'M',
  },
]

const selectedChannel = ref(channels[0].id)
const selectedAmount = ref<number | 'custom'>(amountOptions[0] as number)
const customAmount = ref('')
const selectedPay = ref(payMethods[0].id)

const displayAmount = computed(() => {
  if (selectedAmount.value === 'custom') {
    return customAmount.value ? Number(customAmount.value) : 0
  }
  return selectedAmount.value
})

function chooseAmount(value: number | 'custom') {
  selectedAmount.value = value
  if (value !== 'custom') {
    customAmount.value = ''
  }
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="recharge-page">
    <header class="page-header">
      <button class="icon-btn" type="button" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <h1>充值</h1>
      <button class="icon-btn" type="button">
        <van-icon name="question-o" />
      </button>
    </header>

    <section class="card balance-card">
      <div>
        <p class="card-label">
          当前余额
        </p>
        <p class="card-value">
          R${{ balance.toFixed(2) }}
        </p>
      </div>
      <span class="card-chip">R$</span>
    </section>

    <section class="card">
      <h2>充值渠道</h2>
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
      <h2>选择充值金额</h2>
      <div class="amount-grid">
        <button
          v-for="item in amountOptions" :key="item" type="button" class="amount-item"
          :class="{ active: selectedAmount === item }" @click="chooseAmount(item as number | 'custom')"
        >
          <span v-if="item !== 'custom'">{{ item }}</span>
          <span v-else>自定义</span>
        </button>
      </div>
      <van-field
        v-if="selectedAmount === 'custom'" v-model="customAmount" type="number" label="自定义金额"
        placeholder="请输入充值金额" class="custom-input"
      />
    </section>

    <section class="card">
      <h2>选择支付方式</h2>
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
                {{ method.sub }}
              </p>
            </div>
          </div>
          <span class="pay-check">
            <van-icon name="success" />
          </span>
        </button>
      </div>
    </section>

    <van-button type="primary" round block class="confirm-btn" :disabled="displayAmount <= 0">
      确认充值
    </van-button>
  </div>
</template>

<route lang="json5">
{
  name: "Recharge"
}
</route>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Sora:wght@400;600;700&display=swap');

.recharge-page {
  --page-bg: #f3f6fb;
  --card-bg: #ffffff;
  --text-main: #141a22;
  --text-sub: #6b7280;
  --accent: #3d9b4f;
  --accent-soft: #e6f5ea;
  --stroke: #e4e9f0;
  --shadow: 0 12px 28px rgba(21, 32, 56, 0.08);
  font-family: 'Sora', sans-serif;
  color: var(--text-main);
  padding-bottom: 16px;
  min-height: calc(100vh - 32px);
  background:
    radial-gradient(circle at top right, rgba(61, 155, 79, 0.12), transparent 45%),
    linear-gradient(180deg, #f7f9fc 0%, #eef2f8 100%);
  position: relative;
}

.recharge-page::before {
  content: '';
  position: absolute;
  top: 80px;
  right: -40px;
  width: 120px;
  height: 120px;
  background: rgba(61, 155, 79, 0.08);
  border-radius: 28px;
  transform: rotate(12deg);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.page-header h1 {
  font-size: 20px;
  font-weight: 700;
}

.icon-btn {
  background: transparent;
  border: none;
  font-size: 20px;
  color: var(--accent);
  padding: 6px;
}

.card {
  background: var(--card-bg);
  border-radius: 14px;
  padding: 14px 16px;
  box-shadow: var(--shadow);
  border: 1px solid rgba(228, 233, 240, 0.7);
  margin-bottom: 14px;
}

.balance-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-label {
  font-size: 14px;
  color: var(--text-sub);
  margin: 0 0 6px;
}

.card-value {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.card-chip {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: var(--accent-soft);
  color: var(--accent);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.card h2 {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 12px;
}

.pill-group {
  display: flex;
  gap: 12px;
}

.pill {
  flex: 1;
  border-radius: 10px;
  border: 1px solid var(--stroke);
  background: #fff;
  padding: 10px 12px;
  font-weight: 600;
  color: var(--text-main);
}

.pill.active {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
  box-shadow: 0 6px 14px rgba(61, 155, 79, 0.2);
}

.amount-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.amount-item {
  border-radius: 12px;
  border: 1px solid var(--stroke);
  background: #fff;
  padding: 14px 0;
  font-weight: 600;
  color: var(--text-main);
  font-size: 14px;
}

.amount-item.active {
  border-color: var(--accent);
  background: var(--accent-soft);
  color: var(--accent);
}

.custom-input {
  margin-top: 12px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--stroke);
}

.pay-list {
  display: grid;
  gap: 12px;
}

.pay-item {
  width: 100%;
  border-radius: 12px;
  border: 1px solid var(--stroke);
  background: #fff;
  padding: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  text-align: left;
}

.pay-item.active {
  border-color: var(--accent);
  box-shadow: 0 8px 18px rgba(61, 155, 79, 0.16);
}

.pay-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pay-logo {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: #101828;
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.pay-name {
  font-weight: 600;
  margin: 0;
}

.pay-sub {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--text-sub);
}

.pay-check {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: 1px solid var(--stroke);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  background: #fff;
}

.pay-item:not(.active) .pay-check {
  color: transparent;
}

.confirm-btn {
  margin-top: 18px;
  background: var(--accent);
  border: none;
  font-weight: 600;
  letter-spacing: 0.5px;
}
</style>
