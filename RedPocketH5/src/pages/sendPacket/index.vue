<script setup lang="ts">
import { showToast } from 'vant'
import { sendLuckyPacket } from '@/api/user'
import language1Icon from '@/assets/svg/language-1.svg'

const amountPresets = [100, 200, 300, 500, 1000, 2000]
const mineOptions = [1, 2, 3, 4, 5, 6, 7, 8, 9]

const selectedAmountPreset = ref<number | null>(null)
const selectedMine = ref<number | null>(2)
const amountInput = ref('')
const submitLoading = ref(false)

const canSubmit = computed(() => {
  const amount = Number(amountInput.value)
  return !!selectedMine.value && Number.isFinite(amount) && amount > 0 && !submitLoading.value
})

function selectAmountPreset(value: number) {
  selectedAmountPreset.value = value
  amountInput.value = String(value)
}

function onAmountInput(event: Event) {
  const input = event.target as HTMLInputElement
  const next = input.value.replace(/\D/g, '').slice(0, 8)
  amountInput.value = next
  selectedAmountPreset.value = null
}

function selectMine(value: number) {
  selectedMine.value = value
}

async function submitPacket() {
  if (!canSubmit.value)
    return
  const amount = Number(amountInput.value)
  const thunder = Number(selectedMine.value)
  if (!amount || amount <= 0 || !thunder) {
    showToast('请输入有效金额并选择雷值')
    return
  }

  submitLoading.value = true
  try {
    const { data } = await sendLuckyPacket({
      amount,
      thunder,
      chatId: 0,
    })
    showToast(`红包发送成功 #${data?.id ?? '-'}`)
  }
  catch {
    showToast('发送红包失败')
  }
  finally {
    submitLoading.value = false
  }
}
</script>

<template>
  <div class="send-page">
    <section class="packet-type-card">
      <div class="packet-icon" aria-hidden="true">
        <img :src="language1Icon" class="packet-icon-img" alt="">
      </div>
      <div class="packet-copy">
        <h3 class="packet-title">
          公开发红包
        </h3>
        <p class="packet-subtitle">
          所有人可见，所有人可抢
        </p>
      </div>
    </section>

    <section class="section-block">
      <header class="section-header">
        <span class="section-emoji" aria-hidden="true">🧨</span>
        <h3 class="section-title">
          红包金额
        </h3>
      </header>
      <div class="soft-card amount-card">
        <div class="amount-input-row">
          <span class="amount-label">当前金额</span>
          <div class="amount-value-wrap">
            <input
              :value="amountInput"
              type="text"
              inputmode="numeric"
              class="amount-input"
              placeholder="请输入金额"
              @input="onAmountInput"
            >
          </div>
          <span class="amount-currency">₱</span>
        </div>

        <div class="preset-grid">
          <button
            v-for="value in amountPresets"
            :key="value"
            type="button"
            class="preset-item"
            :class="{ active: selectedAmountPreset === value }"
            @click="selectAmountPreset(value)"
          >
            ₱{{ value }}
          </button>
        </div>
      </div>
    </section>

    <section class="section-block">
      <header class="section-header">
        <span class="section-emoji" aria-hidden="true">🎯</span>
        <h3 class="section-title">
          雷值设置
        </h3>
      </header>
      <div class="soft-card mine-card">
        <div class="mine-title-row">
          <van-icon name="aim" class="mine-title-icon" />
          <h4 class="mine-title">
            选择雷值数字
          </h4>
        </div>
        <div class="mine-subtitle">
          抢到此数字尾号的用户将中雷
          <van-icon name="fire-o" />
          <strong>{{ selectedMine ?? '-' }}</strong>
        </div>
        <div class="mine-grid">
          <button
            v-for="num in mineOptions"
            :key="num"
            type="button"
            class="mine-item"
            :class="{ active: selectedMine === num }"
            @click="selectMine(num)"
          >
            {{ num }}
          </button>
        </div>
      </div>
    </section>

    <button
      type="button"
      class="submit-btn"
      :disabled="!canSubmit"
      @click="submitPacket"
    >
      <van-loading v-if="submitLoading" size="14" color="#fff" />
      <van-icon name="gift-o" />
      立即发布红包
    </button>

    <p class="tips-text">
      发红包，闯好运，赢6个取1.8倍收益！本次发包您将有机会获得最大奖励！
    </p>
    <p class="notice-text">
      注意：发红包需至少有一笔充值记录。
    </p>
  </div>
</template>

<style scoped>
.send-page {
  min-height: 100vh;
  background: var(--color-bg-page);
  padding: 14px var(--page-padding-x) calc(26px + env(safe-area-inset-bottom));
}

.packet-type-card {
  display: flex;
  align-items: center;
  gap: 12px;
  border: 1px solid #c7cff9;
  border-radius: 18px;
  background: #f1f3ff;
  padding: 16px;
}

.packet-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 27px;
  background: #dae0ff;
  color: #4d5dff;
}

.packet-icon-img {
  width: 28px;
  height: 28px;
  object-fit: contain;
}

.packet-title {
  margin: 0;
  font-size: 15px;
  color: #111d2f;
  font-weight: 700;
}

.packet-subtitle {
  margin: 4px 0 0;
  color: #69758a;
  font-size: 11px;
}

.section-block {
  margin-top: 14px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 2px;
}

.section-emoji {
  font-size: 18px;
  line-height: 1;
}

.section-title {
  margin: 0;
  font-size: 14px;
  line-height: 1;
  color: #1a1a2e;
  font-weight: 700;
}

.soft-card {
  margin-top: 10px;
  border-radius: 16px;
  background: #f5f6fa;
}

.amount-card {
  padding: 16px;
}

.amount-input-row {
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 52px;
  border: 1.5px solid #2dc84d;
  border-radius: 12px;
  background: #fff;
  padding: 0 16px;
}

.amount-label {
  font-size: 12px;
  color: #1a1a2e;
  flex: 0 0 auto;
  font-weight: 500;
}

.amount-value-wrap {
  display: flex;
  align-items: center;
  flex: 1 1 auto;
  min-width: 0;
}

.amount-input {
  width: 100%;
  border: none;
  outline: none;
  text-align: right;
  font-size: 24px;
  line-height: 1.1;
  font-weight: 700;
  color: #1a1a2e;
  background: transparent;
  letter-spacing: 0.01em;
}

.amount-input::placeholder {
  color: #9ca3af;
  font-size: 16px;
}

.amount-currency {
  color: #9ca3af;
  font-size: 13px;
  font-weight: 600;
  flex: 0 0 auto;
}

.preset-grid {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.preset-item {
  flex: 0 0 calc((100% - 12px) / 2);
  height: 44px;
  border-radius: 999px;
  border: none;
  background: #e5e7eb;
  color: #1a1a2e;
  font-size: 16px;
  line-height: 1.1;
  font-weight: 700;
  cursor: pointer;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.preset-item.active {
  background: #2dc84d;
  color: #fff;
}

.mine-card {
  padding: 16px;
}

.mine-title-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.mine-title-icon {
  color: #2dc84d;
  font-size: 16px;
}

.mine-title {
  margin: 0;
  color: #2dc84d;
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
}

.mine-subtitle {
  margin: 10px 0 0;
  text-align: center;
  color: #6d7687;
  font-size: 11px;
  line-height: 1;
}

.mine-subtitle :deep(.van-icon) {
  margin: 0 2px;
  color: #ff9500;
  font-size: 12px;
  vertical-align: -1px;
}

.mine-subtitle strong {
  color: #ff9500;
  font-weight: 600;
}

.mine-grid {
  margin-top: 10px;
  display: flex;
  gap: 4px;
}

.mine-item {
  flex: 0 0 calc((100% - 32px) / 9);
  height: 34px;
  border: none;
  border-radius: 8px;
  background: #e5e7eb;
  color: #111827;
  font-size: 14px;
  line-height: 1.1;
  font-weight: 700;
  cursor: pointer;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.mine-item.active {
  background: #2dc84d;
  color: #fff;
}

.submit-btn {
  margin-top: 20px;
  width: 100%;
  height: 52px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(90deg, #2dc84d 0%, #5dd87a 100%);
  color: #fff;
  font-size: 14px;
  line-height: 1;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  box-shadow: none;
  transition: opacity 0.2s ease;
}

.submit-btn :deep(.van-icon) {
  font-size: 16px;
}

.submit-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.tips-text {
  margin: 10px 12px 0;
  text-align: center;
  color: #6b7280;
  font-size: 11px;
  line-height: 1.35;
  font-weight: 500;
}

.notice-text {
  margin: 8px 12px 0;
  text-align: center;
  color: #9ca3af;
  font-size: 11px;
  line-height: 1.35;
  font-weight: 600;
}

@media (max-width: 390px) {
  .amount-input {
    font-size: 22px;
  }

  .mine-item {
    font-size: 13px;
  }
}
</style>

<route lang="json5">
{
  name: 'SendPacket'
}
</route>
