<script setup lang="ts">
defineOptions({ name: 'PpmxPaymentSuccessNotice' })

defineProps<{
  show: boolean
  title: string
  message: string
  amountText?: string
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="ppmx-payment-notice">
      <aside v-if="show" class="ppmx-payment-notice" role="status" aria-live="polite">
        <span class="ppmx-payment-notice__icon" aria-hidden="true">
          <i class="fa-solid fa-circle-check" />
        </span>
        <span class="ppmx-payment-notice__content">
          <strong>{{ title }}</strong>
          <small>{{ message }}</small>
        </span>
        <span v-if="amountText" class="ppmx-payment-notice__amount">
          {{ amountText }}
        </span>
      </aside>
    </Transition>
  </Teleport>
</template>

<style scoped>
.ppmx-payment-notice {
  position: fixed;
  top: calc(14px + env(safe-area-inset-top));
  left: 50%;
  z-index: 10020;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  width: min(460px, calc(100vw - 28px));
  min-height: 70px;
  padding: 13px 15px;
  color: var(--app-text);
  background:
    radial-gradient(circle at 15% 0%, rgba(255, 215, 0, 0.18), transparent 34%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.04)),
    rgba(10, 10, 13, 0.92);
  border: 1px solid rgba(255, 215, 0, 0.28);
  border-radius: 20px;
  box-shadow:
    0 20px 60px rgba(0, 0, 0, 0.52),
    0 0 0 1px rgba(255, 255, 255, 0.05) inset;
  transform: translateX(-50%);
  backdrop-filter: blur(18px) saturate(145%);
  -webkit-backdrop-filter: blur(18px) saturate(145%);
}

.ppmx-payment-notice__icon {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  color: #211400;
  background: linear-gradient(180deg, #ffe889, #efbc3e);
  border-radius: 14px;
  box-shadow: 0 12px 30px rgba(255, 215, 0, 0.22);
}

.ppmx-payment-notice__content {
  display: grid;
  min-width: 0;
  gap: 3px;
}

.ppmx-payment-notice__content strong {
  overflow: hidden;
  font-size: 0.95rem;
  font-weight: 900;
  line-height: 1.2;
  color: var(--app-text);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ppmx-payment-notice__content small {
  overflow: hidden;
  font-size: 0.76rem;
  font-weight: 700;
  line-height: 1.35;
  color: var(--app-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ppmx-payment-notice__amount {
  color: var(--app-gold);
  font-size: 0.95rem;
  font-weight: 900;
  white-space: nowrap;
}

.ppmx-payment-notice-enter-active,
.ppmx-payment-notice-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.22s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.ppmx-payment-notice-enter-from,
.ppmx-payment-notice-leave-to {
  opacity: 0;
  transform: translate(-50%, -12px) scale(0.98);
}

html[data-theme="light"] .ppmx-payment-notice {
  color: #151316;
  background:
    radial-gradient(circle at 15% 0%, rgba(255, 215, 0, 0.22), transparent 34%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.94), rgba(255, 250, 240, 0.82));
  border-color: rgba(225, 178, 52, 0.34);
  box-shadow: 0 18px 48px rgba(74, 56, 20, 0.16);
}

html[data-theme="light"] .ppmx-payment-notice__amount {
  color: #b9152a;
}

@media (max-width: 480px) {
  .ppmx-payment-notice {
    top: calc(10px + env(safe-area-inset-top));
    grid-template-columns: auto minmax(0, 1fr);
  }

  .ppmx-payment-notice__amount {
    grid-column: 2;
  }
}
</style>
