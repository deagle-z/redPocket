<script setup lang="ts">
import { showToast } from 'vant'
import { getCurrentTgInviteRuleConfig, sendLuckyPacket } from '@/api/user'
import language1Icon from '@/assets/svg/language-1.svg'

const props = withDefaults(defineProps<{
  variant?: 'page' | 'modal'
  showIntro?: boolean
  showTips?: boolean
  autoReset?: boolean
}>(), {
  variant: 'page',
  showIntro: true,
  showTips: true,
  autoReset: false,
})

const emit = defineEmits<{
  (e: 'success', payload: { id: number, amount: number, thunder: number }): void
}>()

const { t } = useI18n()

const amountPresets = [100, 200, 300, 500, 1000, 2000]
const mineOptions = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
const selectedAmountPreset = ref<number | null>(null)
const selectedMine = ref<number | null>(0)
const amountInput = ref('')
const submitLoading = ref(false)
const lastSubmitAt = ref(0)
const amountMin = ref(100)
const amountMax = ref(5000)
const SUBMIT_THROTTLE_MS = 1000

const rootClass = computed(() => [
  'send-packet-form',
  `send-packet-form--${props.variant}`,
])

const availableAmountPresets = computed(() => {
  return amountPresets.filter(value => value >= amountMin.value && value <= amountMax.value)
})

const amountRangeText = computed(() => {
  return t('sendPacketPage.amountRange', {
    min: amountMin.value,
    max: amountMax.value,
  })
})

const canSubmit = computed(() => {
  const amount = Number(amountInput.value)
  return selectedMine.value !== null
    && Number.isFinite(amount)
    && amount >= amountMin.value
    && amount <= amountMax.value
    && !submitLoading.value
})

function resetForm() {
  selectedAmountPreset.value = null
  selectedMine.value = 0
  amountInput.value = ''
}

function selectAmountPreset(value: number) {
  selectedAmountPreset.value = value
  amountInput.value = String(value)
}

function onAmountInput(event: Event) {
  const input = event.target as HTMLInputElement
  amountInput.value = input.value.replace(/\D/g, '').slice(0, 8)
  selectedAmountPreset.value = null
}

function selectMine(value: number) {
  selectedMine.value = value
}

async function loadSendRangeConfig() {
  try {
    const { data } = await getCurrentTgInviteRuleConfig()
    const minValue = Number(data?.sendMinAmount)
    const maxValue = Number(data?.sendMaxAmount)
    if (Number.isFinite(minValue) && Number.isFinite(maxValue) && minValue > 0 && maxValue > 0 && minValue <= maxValue) {
      amountMin.value = minValue
      amountMax.value = maxValue
    }
  }
  catch {
    // Keep defaults when config loading fails.
  }
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
  if (!Number.isFinite(amount) || amount < amountMin.value || amount > amountMax.value) {
    showToast(amountRangeText.value)
    return
  }
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
    const id = Number(data?.id || 0)
    showToast(t('sendPacketPage.sendSuccess', { id: id || '-' }))
    emit('success', { id, amount, thunder })
    if (props.autoReset)
      resetForm()
  }
  catch {
    showToast(t('sendPacketPage.sendFailed'))
  }
  finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  void loadSendRangeConfig()
})
</script>

<template>
  <div :class="rootClass">
    <section v-if="showIntro" class="packet-type-card">
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
        <span class="section-title">
          {{ t('sendPacketPage.amountTitle') }}
        </span>
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
          <span class="amount-currency"><img class="amount-currency-coin" src="@/assets/svg/coin.svg" alt=""></span>
        </div>
        <p class="amount-range-tip">
          {{ amountRangeText }}
        </p>

        <div class="preset-grid">
          <button
            v-for="value in availableAmountPresets"
            :key="value"
            type="button"
            class="preset-item"
            :class="{ active: selectedAmountPreset === value }"
            @click="selectAmountPreset(value)"
          >
            <CoinAmount :text="`${value}`" />
          </button>
        </div>
      </div>
    </section>

    <section class="section-block">
      <header class="section-header">
        <span class="section-title">
          {{ t('sendPacketPage.mineTitle') }}
        </span>
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
      <van-loading v-if="submitLoading" size="14" color="#5a1b00" />
      <van-icon name="gift-o" />
      {{ t('sendPacketPage.submit') }}
    </button>

    <template v-if="showTips">
      <p class="tips-text">
        {{ t('sendPacketPage.tipsText') }}
      </p>
      <p class="notice-text">
        {{ t('sendPacketPage.noticeText') }}
      </p>
    </template>
  </div>
</template>

<style scoped>
.send-packet-form {
  width: 100%;
}

.send-packet-form--page {
  min-height: 100%;
}

.send-packet-form--modal {
  padding-bottom: 4px;
}

.packet-type-card,
.soft-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.38);
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.28),
    inset 0 0 0 1px rgba(255, 248, 214, 0.1);
}

.packet-type-card {
  display: flex;
  align-items: center;
  gap: 12px;
  border-radius: 18px;
  background:
    radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px),
    linear-gradient(160deg, rgba(116, 0, 0, 0.96) 0%, rgba(74, 0, 0, 0.96) 100%);
  background-size: 18px 18px, 100% 100%;
  padding: 16px;
}

.packet-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  box-shadow:
    0 8px 18px rgba(75, 25, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.34);
}

.packet-icon-img {
  width: 28px;
  height: 28px;
  object-fit: contain;
  filter: brightness(0) saturate(100%) invert(16%) sepia(38%) saturate(1338%) hue-rotate(357deg) brightness(92%) contrast(97%);
}

.packet-title {
  margin: 0;
  color: #fff0c9;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.packet-subtitle {
  margin: 5px 0 0;
  color: rgba(255, 229, 186, 0.72);
  font-size: 12px;
  line-height: 1.4;
}

.section-block {
  margin-top: 14px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 2px;
}

.section-title {
  color: #ffd98b;
  font-size: 14px;
  line-height: 1;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.soft-card {
  margin-top: 10px;
  border-radius: 18px;
  background:
    linear-gradient(180deg, rgba(255, 248, 214, 0.06) 0%, rgba(255, 248, 214, 0) 100%),
    linear-gradient(160deg, rgba(128, 0, 0, 0.95) 0%, rgba(76, 0, 0, 0.96) 100%);
}

.amount-card,
.mine-card {
  padding: 14px;
}

.amount-input-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-height: 48px;
  border: 1px solid rgba(212, 175, 55, 0.48);
  border-radius: 14px;
  background: rgba(42, 0, 0, 0.56);
  padding: 0 12px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.amount-label {
  font-size: 12px;
  color: rgba(255, 229, 186, 0.8);
  flex: 0 0 auto;
  font-weight: 600;
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
  font-size: 20px;
  line-height: 1.1;
  font-weight: 800;
  color: #fff3de;
  background: transparent;
  letter-spacing: 0.01em;
}

.amount-input::placeholder {
  color: rgba(255, 229, 186, 0.34);
  font-size: 16px;
}

.amount-currency {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
}

.amount-currency-coin {
  width: 22px;
  height: 22px;
  display: block;
}

.preset-grid {
  margin-top: 10px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.amount-range-tip {
  margin: 10px 2px 0;
  color: rgba(255, 229, 186, 0.74);
  font-size: 12px;
  line-height: 1.4;
}

.preset-item,
.mine-item {
  border: none;
  cursor: pointer;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease;
}

.preset-item {
  height: 38px;
  border-radius: 999px;
  background: rgba(255, 248, 214, 0.08);
  border: 1px solid rgba(255, 248, 214, 0.12);
  color: #ffe9bf;
  font-size: 15px;
  line-height: 1.1;
  font-weight: 700;
}

.preset-item.active {
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  border-color: rgba(255, 248, 214, 0.55);
  color: #5a1b00;
  box-shadow:
    0 8px 16px rgba(75, 25, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.34);
}

.mine-title-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.mine-title-icon {
  color: #ffd98b;
  font-size: 16px;
}

.mine-title {
  margin: 0;
  color: #fff0c9;
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
}

.mine-subtitle {
  margin: 10px 0 0;
  text-align: center;
  color: rgba(255, 229, 186, 0.7);
  font-size: 11px;
  line-height: 1.4;
}

.mine-subtitle :deep(.van-icon) {
  margin: 0 2px;
  color: #ffcf5c;
  font-size: 12px;
  vertical-align: -1px;
}

.mine-subtitle strong {
  color: #ffdf87;
  font-weight: 700;
}

.mine-grid {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

.mine-item {
  height: 36px;
  border-radius: 10px;
  background: rgba(255, 248, 214, 0.08);
  border: 1px solid rgba(255, 248, 214, 0.12);
  color: #fff0c9;
  font-size: 15px;
  line-height: 1.1;
  font-weight: 800;
}

.mine-item.active {
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  border-color: rgba(255, 248, 214, 0.55);
  color: #5a1b00;
  box-shadow:
    0 8px 16px rgba(75, 25, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.34);
}

.submit-btn {
  margin-top: 18px;
  width: 100%;
  height: 50px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 14px;
  line-height: 1;
  font-weight: 800;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow:
    0 10px 20px rgba(75, 25, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.34);
}

.submit-btn :deep(.van-icon) {
  font-size: 16px;
}

.submit-btn:disabled {
  opacity: 0.56;
}

.tips-text,
.notice-text {
  text-align: center;
  line-height: 1.45;
}

.tips-text {
  margin: 12px 12px 0;
  color: rgba(255, 229, 186, 0.72);
  font-size: 11px;
  font-weight: 600;
}

.notice-text {
  margin: 8px 12px 0;
  color: rgba(255, 229, 186, 0.48);
  font-size: 11px;
  font-weight: 600;
}

@media (max-width: 390px) {
  .preset-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .amount-input {
    font-size: 18px;
  }
}
</style>
