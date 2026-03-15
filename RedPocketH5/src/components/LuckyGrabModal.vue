<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { showToast } from 'vant'
import { grabLuckyPacket } from '@/api/user'
import { formatCurrency } from '@/utils/currency'
import imgCoin from '@/assets/svg/coin.svg'

interface Props {
  show: boolean
  luckyId?: number
  grabIndex?: number
  senderName?: string
  closeOnClickOverlay?: boolean
  showResultToast?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  luckyId: 0,
  grabIndex: 0,
  senderName: '',
  closeOnClickOverlay: true,
  showResultToast: true,
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'success', payload: { luckyId: number, grabIndex: number, data: any }): void
  (e: 'close'): void
}>()

const { t } = useI18n()

const spinning = ref(false)
const opened = ref(false)
const loading = ref(false)
const resultReady = ref(false)
const resultAmountText = ref('')
const resultAmountValue = ref(0)
const isAmountHidden = ref(false)
const isThunderHit = ref(false)
const loseMoneyText = ref(formatCurrency(0))
const displayAmountText = computed(() => {
  if (!resultReady.value)
    return ''
  return isThunderHit.value ? t('grabModal.thunderShort') : resultAmountText.value
})
const blessingText = computed(() => {
  if (loading.value)
    return t('grabModal.loading')
  if (!resultReady.value)
    return t('grabModal.openingHint')
  if (isThunderHit.value)
    return t('grabModal.thunder', { amount: resultAmountText.value, loseMoney: loseMoneyText.value })
  return t('grabModal.win')
})

function resetState() {
  spinning.value = false
  opened.value = false
  loading.value = false
  resultReady.value = false
  resultAmountText.value = ''
  resultAmountValue.value = 0
  isAmountHidden.value = false
  isThunderHit.value = false
  loseMoneyText.value = formatCurrency(0)
}

function closeModal() {
  if (loading.value)
    return
  emit('update:show', false)
  emit('close')
}

function onOverlayClick(event: MouseEvent) {
  if (!props.closeOnClickOverlay)
    return
  if (event.target !== event.currentTarget)
    return
  closeModal()
}

function sleep(ms: number) {
  return new Promise(resolve => window.setTimeout(resolve, ms))
}

function handleOpen() {
  if (loading.value || spinning.value || opened.value || resultReady.value)
    return
  void runOpenFlow()
}

async function runOpenFlow() {
  spinning.value = true
  const [ok] = await Promise.all([
    submitGrab(),
    sleep(300),
  ])
  spinning.value = false
  opened.value = !!ok
}

function formatAmount(value: number) {
  return formatCurrency(Number(value || 0))
}

async function submitGrab(): Promise<boolean> {
  const luckyId = Number(props.luckyId || 0)
  const grabIndex = Number(props.grabIndex || 0)
  if (!luckyId) {
    showToast(t('grabModal.invalidParam'))
    return false
  }
  loading.value = true
  try {
    const { data } = await grabLuckyPacket({
      luckyId,
      grabIndex: grabIndex > 0 ? grabIndex : undefined,
    })
    isThunderHit.value = data?.isThunder === 1 || data?.isThunder === '1'
    isAmountHidden.value = data?.isAmountHidden === 1 || data?.isAmountHidden === '1'
    const rawAmount = Number(data?.amount ?? data?.grabAmount ?? 0)
    const rawLoseMoney = Number(data?.loseMoney ?? 0)
    resultAmountValue.value = rawAmount
    resultAmountText.value = formatAmount(rawAmount)
    loseMoneyText.value = formatAmount(rawLoseMoney)
    resultReady.value = true
    if (props.showResultToast)
      showToast(data?.message || t('grabModal.grabSuccess'))
    const emitData = isAmountHidden.value
      ? {
          ...data,
          amount: 0,
          grabAmount: 0,
        }
      : data
    emit('success', { luckyId, grabIndex, data: emitData })
    return true
  }
  catch {
    if (props.showResultToast)
      showToast(t('grabModal.grabFailed'))
    return false
  }
  finally {
    loading.value = false
  }
}

watch(
  () => props.show,
  () => {
    resetState()
  },
)

onBeforeUnmount(() => {
  resetState()
})
</script>

<template>
  <teleport to="body">
    <transition name="grab-fade">
      <div v-if="show" class="grab-overlay" @click="onOverlayClick">
        <div
          class="ang-pao"
          :class="{
            'active': spinning,
            'opened': opened,
            'thunder-hit': resultReady && isThunderHit,
            'win-hit': resultReady && !isThunderHit,
          }"
          role="dialog"
          aria-modal="true"
          @click.stop
        >
          <div class="layer-back" />
          <div class="layer-flap">
            <span class="flap-text">{{ t('grabModal.flapText') }}</span>
          </div>

          <div class="gift-card">
            <div v-if="resultReady && !isThunderHit" class="win-mark" aria-hidden="true">
              ✨
            </div>
            <div v-if="resultReady && isThunderHit" class="thunder-mark" aria-hidden="true">
              ⚡
            </div>
            <div class="corner tl" />
            <div class="corner tr" />
            <div class="corner bl" />
            <div class="corner br" />
            <div class="card-content">
              <div class="thai-header">
                {{ senderName || t('grabModal.defaultSender') }}
              </div>
              <div class="sub-header">
                {{ t('grabModal.voucherTitle') }}
              </div>
              <div class="amount">
                <CoinAmount v-if="resultReady && !isThunderHit && displayAmountText" :text="displayAmountText" class="amount-coin" />
                <template v-else>{{ displayAmountText }}</template>
              </div>
              <div class="blessing">
                {{ blessingText }}
              </div>
            </div>
          </div>

          <div class="layer-pocket">
            <div class="thai-pattern" />
          </div>

          <button type="button" class="coin-btn" :disabled="loading || resultReady" @click="handleOpen">
            <img class="coin-btn-icon" :src="imgCoin" alt="">
            <span class="coin-label">{{ t('grabModal.open') }}</span>
          </button>

          <button type="button" class="close-btn" :disabled="loading" @click="closeModal">
            &times;
          </button>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.grab-overlay {
  position: fixed;
  inset: 0;
  z-index: 3500;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  isolation: isolate;
}

.ang-pao {
  position: relative;
  z-index: 10;
  width: 320px;
  height: 480px;
  max-width: calc(100vw - 30px);
  perspective: 1500px;
  transform-style: preserve-3d;
  transform: scale(0.86);
  transition: transform 0.45s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.layer-back {
  position: absolute;
  width: 100%;
  height: 100%;
  background-color: #700000;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6);
  transform: translateZ(-1px);
}

.gift-card {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 90%;
  min-height: 380px;
  background: linear-gradient(135deg, #fffbf0 0%, #f3e6c0 100%);
  border-radius: 8px;
  box-shadow:
    inset 0 0 0 3px #800000,
    inset 0 0 0 6px #d4af37;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  transition: all 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.8) translateZ(2px);
}

.gift-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(#d4af37 1px, transparent 1px);
  background-size: 15px 15px;
  opacity: 0.2;
  z-index: 0;
}

.thunder-mark {
  position: absolute;
  top: 16px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3;
  font-size: 30px;
  line-height: 1;
  color: #ffb300;
  text-shadow:
    0 0 8px rgba(255, 179, 0, 0.7),
    0 0 18px rgba(255, 60, 60, 0.45);
  animation: thunderPulse 0.5s ease-in-out infinite;
}

.win-mark {
  position: absolute;
  top: 16px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3;
  font-size: 28px;
  line-height: 1;
  color: #ffd700;
  text-shadow:
    0 0 8px rgba(255, 215, 0, 0.8),
    0 0 16px rgba(255, 248, 214, 0.7);
  animation: winSparkle 0.9s ease-in-out infinite;
}

.card-content {
  z-index: 2;
  text-align: center;
  width: 100%;
  padding: 20px;
  box-sizing: border-box;
}

.thai-header {
  color: #800000;
  font-size: 22px;
  line-height: 1.1;
  font-weight: 700;
  word-break: break-all;
}

.sub-header {
  margin-top: 8px;
  color: #b8860b;
  font-size: 12px;
  letter-spacing: 2px;
  text-transform: uppercase;
}

.amount {
  margin-top: 20px;
  font-size: 48px;
  line-height: 1.1;
  font-family: 'Times New Roman', serif;
  font-weight: 700;
}

.amount-coin :deep(.coin-amount-text) {
  background: linear-gradient(to bottom, #cfb53b, #8a6e14, #d4af37);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.amount-coin :deep(.coin-amount-icon) {
  width: 1em;
  height: 1em;
}

.ang-pao.thunder-hit .amount {
  color: #ff2d2d;
  background: none;
  -webkit-text-fill-color: currentcolor;
  text-shadow:
    0 0 10px rgba(255, 77, 77, 0.7),
    0 0 20px rgba(255, 200, 0, 0.45);
  animation: thunderTextBlink 0.42s ease-in-out infinite;
}

.ang-pao.thunder-hit .gift-card {
  animation: thunderShake 0.32s linear infinite;
}

.ang-pao.win-hit .gift-card {
  box-shadow:
    0 0 0 2px rgba(255, 215, 0, 0.45),
    0 0 24px rgba(255, 215, 0, 0.35),
    inset 0 0 0 3px #800000,
    inset 0 0 0 6px #d4af37;
  animation: winGlow 1.1s ease-in-out infinite;
}

.ang-pao.win-hit .amount {
  animation: winAmountPop 0.9s ease-in-out infinite;
}

.blessing {
  margin-top: 22px;
  font-size: 14px;
  color: #555;
}

.corner {
  position: absolute;
  width: 15px;
  height: 15px;
  border: 2px solid #b8860b;
  z-index: 2;
}

.tl {
  top: 12px;
  left: 12px;
  border-right: none;
  border-bottom: none;
}

.tr {
  top: 12px;
  right: 12px;
  border-left: none;
  border-bottom: none;
}

.bl {
  bottom: 12px;
  left: 12px;
  border-right: none;
  border-top: none;
}

.br {
  bottom: 12px;
  right: 12px;
  border-left: none;
  border-top: none;
}

.layer-pocket {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 75%;
  background: linear-gradient(160deg, #900000 30%, #600000 100%);
  border-radius: 0 0 12px 12px;
  transform: translateZ(5px);
  z-index: 5;
  border-top: 3px solid #d4af37;
  box-shadow: 0 -5px 15px rgba(0, 0, 0, 0.2);
  overflow: hidden;
  transition: filter 0.5s;
}

.thai-pattern {
  position: absolute;
  width: 120%;
  height: 120%;
  top: -10%;
  left: -10%;
  opacity: 0.1;
  background: repeating-linear-gradient(45deg, #ffd700, #ffd700 2px, transparent 2px, transparent 10px);
}

.layer-flap {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 140px;
  background-color: #850000;
  border-radius: 12px 12px 50% 50%;
  transform: translateZ(6px);
  z-index: 6;
  transform-origin: top;
  transition:
    transform 0.6s ease-in-out,
    z-index 0.6s step-end;
  border-bottom: 4px solid #d4af37;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 34px;
  box-sizing: border-box;
}

.flap-text {
  color: #d4af37;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 2px;
}

.coin-btn {
  position: absolute;
  top: 100px;
  left: 50%;
  width: 90px;
  height: 90px;
  transform: translateX(-50%) translateZ(20px);
  border: 4px solid #fff8d6;
  border-radius: 50%;
  background: radial-gradient(ellipse at center, #ffd700 0%, #b8860b 100%);
  z-index: 100;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  overflow: hidden;
  transform-origin: center;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
  will-change: transform, filter;
  transition: opacity 0.3s ease;
}

.coin-btn::after {
  content: '';
  position: absolute;
  inset: 8px;
  border-radius: 50%;
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.28) 48%, transparent 100%);
  opacity: 0.18;
  pointer-events: none;
}

.coin-btn:disabled {
  cursor: not-allowed;
  opacity: 0.8;
}

.coin-btn-icon {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  display: block;
}

.coin-label {
  margin-top: 2px;
  font-size: 10px;
  color: #6d0000;
  font-weight: 700;
}

@keyframes spinCoin {
  0%,
  100% {
    transform: translateX(-50%) translateZ(20px) scaleX(1) rotateZ(0deg);
    filter: brightness(1);
  }

  24% {
    transform: translateX(-50%) translateZ(20px) scaleX(0.16) rotateZ(-2deg);
    filter: brightness(0.9);
  }

  50% {
    transform: translateX(-50%) translateZ(20px) scaleX(1) rotateZ(-4deg);
    filter: brightness(1.06);
  }

  74% {
    transform: translateX(-50%) translateZ(20px) scaleX(0.16) rotateZ(2deg);
    filter: brightness(0.9);
  }
}

@keyframes coinShine {
  0%,
  100% {
    transform: translateX(-14px);
    opacity: 0.08;
  }

  50% {
    transform: translateX(14px);
    opacity: 0.3;
  }
}

@keyframes thunderShake {
  0%,
  100% {
    transform: translate(-50%, -50%) translateZ(50px) scale(1.05) translateX(0);
  }

  20% {
    transform: translate(-50%, -50%) translateZ(50px) scale(1.05) translateX(-5px) rotate(-1.3deg);
  }

  40% {
    transform: translate(-50%, -50%) translateZ(50px) scale(1.05) translateX(5px) rotate(1.3deg);
  }

  60% {
    transform: translate(-50%, -50%) translateZ(50px) scale(1.05) translateX(-4px) rotate(-1deg);
  }

  80% {
    transform: translate(-50%, -50%) translateZ(50px) scale(1.05) translateX(4px) rotate(1deg);
  }
}

@keyframes thunderTextBlink {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }

  50% {
    opacity: 0.7;
    transform: scale(1.04);
  }
}

@keyframes thunderPulse {
  0%,
  100% {
    transform: translateX(-50%) scale(1);
    opacity: 1;
  }

  50% {
    transform: translateX(-50%) scale(1.14);
    opacity: 0.72;
  }
}

@keyframes winGlow {
  0%,
  100% {
    transform: translate(-50%, -50%) translateZ(50px) scale(1.05);
    filter: saturate(1);
  }

  50% {
    transform: translate(-50%, -50%) translateZ(50px) scale(1.075);
    filter: saturate(1.15);
  }
}

@keyframes winAmountPop {
  0%,
  100% {
    transform: scale(1);
  }

  40% {
    transform: scale(1.08);
  }

  70% {
    transform: scale(1.03);
  }
}

@keyframes winSparkle {
  0%,
  100% {
    transform: translateX(-50%) scale(1) rotate(0deg);
    opacity: 0.92;
  }

  50% {
    transform: translateX(-50%) scale(1.18) rotate(10deg);
    opacity: 1;
  }
}

.ang-pao.active .coin-btn {
  animation: spinCoin 0.6s linear infinite;
}

.ang-pao.active .coin-btn::after {
  animation: coinShine 0.6s linear infinite;
}

.ang-pao.opened .layer-flap {
  transform: rotateX(180deg);
  z-index: 1;
  background-color: #600000;
  border-bottom: none;
  border-top: 4px solid #d4af37;
}

.ang-pao.opened .coin-btn {
  opacity: 0;
  pointer-events: none;
  transform: translateX(-50%) translateZ(20px) scale(0);
  transition:
    transform 0.4s ease-in,
    opacity 0.2s;
}

.ang-pao.opened .gift-card {
  opacity: 1;
  transform: translate(-50%, -50%) translateZ(50px) scale(1.05);
  z-index: 200;
}

.ang-pao.opened .layer-pocket {
  filter: brightness(0.7);
}

.close-btn {
  position: absolute;
  bottom: -70px;
  left: 50%;
  width: 44px;
  height: 44px;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-radius: 50%;
  transform: translateX(-50%) translateZ(20px);
  color: #fff;
  font-size: 24px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  background: transparent;
}

.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.grab-fade-enter-active,
.grab-fade-leave-active {
  transition: opacity 0.24s ease;
}

.grab-fade-enter-from,
.grab-fade-leave-to {
  opacity: 0;
}

@media (max-width: 420px) {
  .ang-pao {
    width: 292px;
    height: 440px;
  }

  .amount {
    font-size: 40px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .ang-pao.active .coin-btn,
  .ang-pao.active .coin-btn::after,
  .ang-pao.thunder-hit .gift-card,
  .ang-pao.win-hit .gift-card,
  .ang-pao.win-hit .amount,
  .thunder-mark,
  .win-mark {
    animation: none !important;
  }
}
</style>
