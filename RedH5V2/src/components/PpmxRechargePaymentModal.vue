<script setup lang="ts">
import type { RechargeOrderAppBack } from '@/api/user'
import { APP_CURRENCY } from '@/config/market'
import { showFailToast, showSuccessToast } from 'vant'
import PpmxButton from '@/components/PpmxButton.vue'

defineOptions({ name: 'PpmxRechargePaymentModal' })

const props = withDefaults(defineProps<{
  amountText: string
  currencyFallback?: string
  methodText: string
  modelValue: boolean
  order: RechargeOrderAppBack | null
}>(), {
  currencyFallback: APP_CURRENCY,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const { t } = useI18n()

const paymentUrl = computed(() => props.order?.payUrl ?? '')

function closeModal() {
  emit('update:modelValue', false)
}

function openPaymentUrl() {
  if (!paymentUrl.value) return

  window.open(paymentUrl.value, '_blank', 'noopener,noreferrer')
}

async function copyPaymentUrl() {
  if (!paymentUrl.value) return

  try {
    await navigator.clipboard.writeText(paymentUrl.value)
    showSuccessToast(t('recharge.paymentLinkCopied'))
  } catch {
    showFailToast(t('recharge.paymentLinkCopyFailed'))
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="ppmx-withdraw-fade">
      <div
        v-if="props.modelValue"
        id="paymentModal"
        class="ppmx-recharge-modal ppmx-recharge-payment-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="paymentModalTitle"
      >
        <button
          class="ppmx-recharge-modal__backdrop ppmx-recharge-payment-modal__backdrop"
          type="button"
          :aria-label="t('common.close')"
          @click="closeModal"
        />

        <section
          class="ppmx-recharge-modal__card ppmx-glass-strong ppmx-recharge-payment-modal__card"
          role="document"
        >
          <button
            class="ppmx-recharge-modal__close"
            type="button"
            :aria-label="t('common.close')"
            @click="closeModal"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <div class="ppmx-recharge-modal__scroll dep-scroll ppmx-recharge-payment-modal__scroll">
            <header class="ppmx-recharge-modal__head ppmx-recharge-payment-modal__head">
              <span class="ppmx-eyebrow">PAYMENT</span>
              <h1 id="paymentModalTitle">{{ t('recharge.paymentModalTitle') }}</h1>
              <p>{{ t('recharge.paymentModalSub') }}</p>
            </header>

            <section class="ppmx-recharge-payment-modal__hero" aria-live="polite">
              <span class="ppmx-recharge-payment-modal__icon">
                <i class="fa-solid fa-arrow-up-right-from-square" />
              </span>
              <div>
                <span>{{ t('recharge.payAmount') }}</span>
                <strong>{{ props.amountText }} <em>{{ props.order?.currency || props.currencyFallback }}</em></strong>
              </div>
            </section>

            <dl class="ppmx-recharge-payment-modal__details">
              <div>
                <dt>{{ t('recharge.paymentOrderNo') }}</dt>
                <dd>{{ props.order?.orderNo || '-' }}</dd>
              </div>
              <div>
                <dt>{{ t('recharge.payMethodTitle') }}</dt>
                <dd>{{ props.order?.payMethod || props.methodText || '-' }}</dd>
              </div>
            </dl>

            <div class="ppmx-recharge-payment-modal__actions">
              <PpmxButton
                block
                size="lg"
                type="button"
                :disabled="!paymentUrl"
                @click="openPaymentUrl"
              >
                {{ t('recharge.openPaymentPage') }}
                <i class="fa-solid fa-arrow-up-right-from-square" />
              </PpmxButton>
              <button
                class="ppmx-btn-ghost"
                type="button"
                :disabled="!paymentUrl"
                @click="copyPaymentUrl"
              >
                <i class="fa-solid fa-copy" />
                <span>{{ t('recharge.copyPaymentLink') }}</span>
              </button>
            </div>

            <section class="wd-rules ppmx-recharge-payment-modal__rules">
              <h3 class="wd-rules-t">
                <i class="fa-solid fa-circle-info" />
                <span>{{ t('recharge.rulesTitle') }}</span>
              </h3>
              <ul class="wd-rules-list">
                <li>{{ t('recharge.rule1') }}</li>
                <li>{{ t('recharge.rule2') }}</li>
                <li>{{ t('recharge.rule3') }}</li>
              </ul>
            </section>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>
