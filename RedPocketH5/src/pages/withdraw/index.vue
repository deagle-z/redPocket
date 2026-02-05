<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

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
    fee: '0.0% + 固定10.00',
    logo: 'G',
  },
  {
    id: 'maya',
    name: 'Maya',
    fee: '0.5% + 固定5.00',
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
  <div class="withdraw-page">
    <header class="page-header">
      <button class="icon-btn" type="button" @click="goBack">
        <van-icon name="arrow-left" />
      </button>
      <h1>提现</h1>
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
        <p class="card-sub">
          提现中冻结: R${{ frozen.toFixed(2) }}
        </p>
      </div>
      <span class="card-chip">R$</span>
    </section>

    <section class="card">
      <h2>提现渠道</h2>
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
      <h2>支付方式</h2>
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
                手续费: {{ method.fee }}
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
        收款信息
      </h2>
      <div class="form-list">
        <van-field v-model="receiverAccount" label="收款账号" placeholder="请输入您的收款账号" class="form-item" />
        <van-field v-model="receiverName" label="收款人姓名" placeholder="请输入收款人姓名" class="form-item" />
        <van-field v-model="receiverEmail" label="邮箱地址" placeholder="请输入邮箱地址" class="form-item" />
      </div>
    </section>

    <section class="card">
      <h2 class="section-title">
        提现金额
      </h2>
      <van-field
        v-model="customAmount" type="number" label="金额" placeholder="请输入提现金额" class="custom-input"
        @focus="selectedAmount = 'custom'"
      />
      <div class="amount-grid">
        <button
          v-for="item in amountOptions" :key="item" type="button" class="amount-item"
          :class="{ active: selectedAmount === item }" @click="chooseAmount(item as number | 'custom')"
        >
          <span v-if="item !== 'custom'">R${{ item }}</span>
          <span v-else>自定义</span>
        </button>
      </div>

      <div class="balance-breakdown">
        <div class="balance-header">
          <div class="balance-title">
            <span class="balance-icon">◎</span>
            可用余额
          </div>
          <div class="balance-amount">
            R${{ balance.toFixed(2) }}
          </div>
        </div>
        <div class="balance-row">
          <div>
            <p class="row-title">
              正常余额
            </p>
            <p class="row-sub">
              剩余打码量
            </p>
          </div>
          <div class="row-right">
            <span>R$0.00</span>
            <span class="row-badge">可提现</span>
          </div>
        </div>
        <div class="balance-row">
          <div>
            <p class="row-title">
              赠金余额
            </p>
            <p class="row-sub">
              剩余打码量
            </p>
          </div>
          <div class="row-right">
            <span>R$0.00</span>
            <span class="row-badge">可提现</span>
          </div>
        </div>
        <div class="balance-row">
          <div>
            <p class="row-title">
              提现中冻结
            </p>
          </div>
          <div class="row-right">
            <span>-R$0.00</span>
          </div>
        </div>
      </div>

      <van-button type="primary" round block class="submit-btn">
        提交
      </van-button>

      <div class="fee-card">
        <div class="fee-row">
          <span>手续费</span>
          <span>R$0.00</span>
        </div>
        <div class="fee-row total">
          <span>实际扣除</span>
          <span>R$0.00</span>
        </div>
      </div>
    </section>

    <section class="card tips">
      <h2 class="tips-title">
        提现说明:
      </h2>
      <ol>
        <li>提现最低金额为 R$100</li>
        <li>正常余额需打满充值的 1 倍流水后方可提现</li>
        <li>赠金需打满 20 倍流水后提现，流水从第一笔充值开始计算</li>
        <li>请仔细核对收款信息，错误信息将导致提现失败</li>
        <li>提现将在 1-30 分钟内处理完成，如有问题请联系客服</li>
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
@import url('https://fonts.googleapis.com/css2?family=Sora:wght@400;600;700&display=swap');

.withdraw-page {
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
  padding-bottom: 24px;
  min-height: 100vh;
  background:
    radial-gradient(circle at top right, rgba(61, 155, 79, 0.12), transparent 45%),
    linear-gradient(180deg, #f7f9fc 0%, #eef2f8 100%);
  position: relative;
}

.withdraw-page::before {
  content: '';
  position: absolute;
  top: 70px;
  left: -30px;
  width: 140px;
  height: 140px;
  background: rgba(61, 155, 79, 0.08);
  border-radius: 32px;
  transform: rotate(-8deg);
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

.card-sub {
  font-size: 13px;
  color: var(--text-sub);
  margin: 6px 0 0;
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

.section-title {
  color: var(--accent);
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

.form-list {
  display: grid;
  gap: 8px;
}

.form-item {
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--stroke);
}

.custom-input {
  margin-bottom: 12px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--stroke);
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
  padding: 12px 0;
  font-weight: 600;
  color: var(--text-main);
  font-size: 14px;
}

.amount-item.active {
  border-color: var(--accent);
  background: var(--accent-soft);
  color: var(--accent);
}

.balance-breakdown {
  margin-top: 16px;
  border-radius: 12px;
  background: #fff;
  border: 1px solid var(--stroke);
  overflow: hidden;
}

.balance-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid var(--stroke);
  font-weight: 600;
}

.balance-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.balance-icon {
  color: var(--accent);
}

.balance-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid var(--stroke);
  font-size: 14px;
}

.balance-row:last-child {
  border-bottom: none;
}

.row-title {
  margin: 0;
  font-weight: 600;
}

.row-sub {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--text-sub);
}

.row-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.row-badge {
  font-size: 12px;
  color: var(--accent);
  border: 1px solid var(--accent);
  border-radius: 999px;
  padding: 2px 8px;
}

.submit-btn {
  margin-top: 16px;
  background: var(--accent);
  border: none;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.fee-card {
  margin-top: 14px;
  border-radius: 12px;
  border: 1px solid var(--stroke);
  padding: 12px 14px;
  background: #fff;
}

.fee-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--text-sub);
  font-size: 14px;
}

.fee-row.total {
  margin-top: 8px;
  color: var(--accent);
  font-weight: 600;
}

.tips {
  background: #fff;
}

.tips-title {
  color: var(--accent);
  font-size: 16px;
  margin-bottom: 8px;
}

.tips ol {
  margin: 0;
  padding-left: 18px;
  color: var(--text-sub);
  font-size: 13px;
  line-height: 1.6;
}
</style>
