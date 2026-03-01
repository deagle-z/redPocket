<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import AppPageHeader from '@/components/AppPageHeader.vue'

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
  showToast('如有问题请联系客服')
}
</script>

<template>
  <div class="withdraw-page theme-withdraw">
    <AppPageHeader title="提现" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

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
