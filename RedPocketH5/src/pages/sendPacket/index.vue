<script setup lang="ts">
import { showToast } from 'vant'
import { sendLuckyPacket } from '@/api/user'
import language1Icon from '@/assets/svg/language-1.svg'
import { CURRENCY_SYMBOL } from '@/utils/currency'
const { t } = useI18n()

const amountPresets = [100, 200, 300, 500, 1000, 2000]
const mineOptions = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]

const selectedAmountPreset = ref<number | null>(null)
const selectedMine = ref<number | null>(0)
const amountInput = ref('')
const submitLoading = ref(false)
const SUBMIT_THROTTLE_MS = 1000
const lastSubmitAt = ref(0)

const canSubmit = computed(() => {
  const amount = Number(amountInput.value)
  return selectedMine.value !== null && Number.isFinite(amount) && amount > 0 && !submitLoading.value
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
  const now = Date.now()
  if (now - lastSubmitAt.value < SUBMIT_THROTTLE_MS)
    return
  lastSubmitAt.value = now
  const amount = Number(amountInput.value)
  const thunder = Number(selectedMine.value)
  if (!amount || amount <= 0 || !Number.isInteger(thunder) || thunder < 0 || thunder > 9) {
    showToast(t('sendPacketPage.invalidInput'))
    return
  }

  submitLoading.value = true
  try {
    const { data } = await sendLuckyPacket({
      amount,
      thunder,
      chatId: 0,
    })
    showToast(t('sendPacketPage.sendSuccess', { id: data?.id ?? '-' }))
  }
  catch {
    showToast(t('sendPacketPage.sendFailed'))
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
          {{ t('sendPacketPage.packetTypeTitle') }}
        </h3>
        <p class="packet-subtitle">
          {{ t('sendPacketPage.packetTypeSub') }}
        </p>
      </div>
    </section>

    <section class="section-block">
      <header class="section-header">
        <span class="section-emoji" aria-hidden="true">🧨</span>
        <h3 class="section-title">
          {{ t('sendPacketPage.amountTitle') }}
        </h3>
      </header>
      <div class="soft-card amount-card">
        <div class="amount-input-row">
          <span class="amount-label">{{ t('sendPacketPage.currentAmount') }}</span>
          <div class="amount-value-wrap">
            <input
              :value="amountInput"
              type="text"
              inputmode="numeric"
              class="amount-input"
              :placeholder="t('sendPacketPage.amountPlaceholder')"
              @input="onAmountInput"
            >
          </div>
          <span class="amount-currency">{{ CURRENCY_SYMBOL }}</span>
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
            {{ CURRENCY_SYMBOL }}{{ value }}
          </button>
        </div>
      </div>
    </section>

    <section class="section-block">
      <header class="section-header">
        <span class="section-emoji" aria-hidden="true">🎯</span>
        <h3 class="section-title">
          {{ t('sendPacketPage.mineTitle') }}
        </h3>
      </header>
      <div class="soft-card mine-card">
        <div class="mine-title-row">
          <van-icon name="aim" class="mine-title-icon" />
          <h4 class="mine-title">
            {{ t('sendPacketPage.mineSelectTitle') }}
          </h4>
        </div>
        <div class="mine-subtitle">
          {{ t('sendPacketPage.mineSub') }}
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
      {{ t('sendPacketPage.submit') }}
    </button>

    <p class="tips-text">
      {{ t('sendPacketPage.tipsText') }}
    </p>
    <p class="notice-text">
      {{ t('sendPacketPage.noticeText') }}
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
  font-size: 18px;
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
  padding: 12px;
}

.amount-input-row {
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 46px;
  border: 1.5px solid var(--color-primary);
  border-radius: 12px;
  background: #fff;
  padding: 0 12px;
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
  font-size: 18px;
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
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preset-item {
  flex: 0 0 calc((100% - 8px) / 2);
  height: 38px;
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
  background: var(--color-primary);
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
  color: var(--color-primary);
  font-size: 16px;
}

.mine-title {
  margin: 0;
  color: var(--color-primary);
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
  flex: 0 0 calc((100% - 32px) / 10);
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
  background: var(--color-primary);
  color: #fff;
}

.submit-btn {
  margin-top: 20px;
  width: 100%;
  height: 52px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(90deg, var(--color-primary) 0%, var(--color-primary-gradient-end) 100%);
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
    font-size: 18px;
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
