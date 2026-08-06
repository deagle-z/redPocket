<script setup lang="ts">
import { showToast } from 'vant'
import { redeemExchangeCode } from '@/api/user'
import { formatMoney, toDisplayCents } from '@/utils/money'
import { APP_CURRENCY_SYMBOL } from '@/config/market'
import { usePpmxLocale } from '@/pages/ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '@/pages/ppmx-home/types'
import PpmxDialog from '@/components/PpmxDialog.vue'

defineOptions({ name: 'PpmxPromoCodeModal' })

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  redeemed: []
}>()

const userStore = useUserStore()
const { pickText } = usePpmxLocale()

const tx = {
  title: { es: 'Canjear código', en: 'Redeem code', zh: '兑换优惠码' },
  sub: {
    es: 'Ingresa tu código de 6 dígitos para recibir tu recompensa.',
    en: 'Enter your 6-digit code to receive your reward.',
    zh: '输入 6 位数字优惠码以领取奖励。',
  },
  placeholder: { es: '000000', en: '000000', zh: '000000' },
  redeem: { es: 'Canjear', en: 'Redeem', zh: '立即兑换' },
  cancel: { es: 'Cerrar', en: 'Close', zh: '关闭' },
  hint: {
    es: 'Cada código se puede activar una sola vez por cuenta.',
    en: 'Each code can be activated only once per account.',
    zh: '每个优惠码每个账户仅可激活一次。',
  },
  invalid: {
    es: 'Ingresa un código de 6 dígitos.',
    en: 'Enter a 6-digit code.',
    zh: '请输入 6 位数字优惠码。',
  },
  submitting: { es: 'Canjeando', en: 'Redeeming', zh: '兑换中' },
  success: {
    es: 'Código canjeado: +{amount}',
    en: 'Code redeemed: +{amount}',
    zh: '兑换成功：+{amount}',
  },
  failed: {
    es: 'No se pudo canjear el código.',
    en: 'Could not redeem the code.',
    zh: '兑换失败，请检查兑换码。',
  },
} satisfies Record<string, LocalizedText>

const promoCode = ref('')
const submitLock = useSubmitLock(submitPromoCodeRequest, { minLockMs: 600 })
const isSubmitting = computed(() => submitLock.locked.value)

function normalizePromoCode(value: string) {
  return value.replace(/\D/g, '').slice(0, 6)
}

function updatePromoCode(event: Event) {
  const input = event.target as HTMLInputElement | null
  promoCode.value = normalizePromoCode(input?.value ?? '')
}

function submitPromoCode() {
  void submitLock.run().catch(() => {})
}

async function submitPromoCodeRequest() {
  const normalizedCode = normalizePromoCode(promoCode.value)
  promoCode.value = normalizedCode

  if (normalizedCode.length !== 6) {
    showToast(pickText(tx.invalid))
    return
  }

  try {
    const result = await redeemExchangeCode(normalizedCode)
    emit('update:modelValue', false)
    promoCode.value = ''
    await userStore.loadUserInfo()
    const amount = formatMoney(toDisplayCents(result.amount), { currency: APP_CURRENCY_SYMBOL })
    showToast(pickText(tx.success).replace('{amount}', amount))
    emit('redeemed')
  } catch (error) {
    const message = error instanceof Error ? error.message : pickText(tx.failed)
    showToast(message || pickText(tx.failed))
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) promoCode.value = ''
  },
)
</script>

<template>
  <PpmxDialog
    :model-value="props.modelValue"
    variant="gold"
    icon="fa-ticket"
    :title="pickText(tx.title)"
    :confirm-text="isSubmitting ? pickText(tx.submitting) : pickText(tx.redeem)"
    :cancel-text="pickText(tx.cancel)"
    :confirm-loading="isSubmitting"
    :aria-label="pickText(tx.title)"
    @update:model-value="emit('update:modelValue', $event)"
    @confirm="submitPromoCode"
  >
    <div id="promoModal" class="ppmx-promo-code-modal">
      <p>{{ pickText(tx.sub) }}</p>
      <input
        id="promoInput"
        :value="promoCode"
        class="ppmx-promo-code-input"
        type="text"
        inputmode="numeric"
        maxlength="6"
        autocomplete="off"
        :placeholder="pickText(tx.placeholder)"
        aria-label="Promo code"
        @input="updatePromoCode"
        @keydown.enter.prevent="submitPromoCode"
      >
      <small>{{ pickText(tx.hint) }}</small>
    </div>
  </PpmxDialog>
</template>
