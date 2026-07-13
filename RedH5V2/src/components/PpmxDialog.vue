<script setup lang="ts">
defineOptions({ name: 'PpmxDialog' })

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    message?: string
    icon?: string
    variant?: 'danger' | 'gold' | 'neutral'
    confirmText?: string
    cancelText?: string
    showCancel?: boolean
    confirmLoading?: boolean
    closeOnBackdrop?: boolean
    ariaLabel?: string
  }>(),
  {
    title: '',
    message: '',
    icon: 'fa-triangle-exclamation',
    variant: 'danger',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    showCancel: true,
    confirmLoading: false,
    closeOnBackdrop: true,
    ariaLabel: 'PP.BET dialog',
  },
)

const emit = defineEmits<{
  cancel: []
  close: []
  confirm: []
  'update:modelValue': [value: boolean]
}>()

function closeDialog() {
  if (props.confirmLoading) return

  emit('update:modelValue', false)
  emit('close')
}

function cancelDialog() {
  if (props.confirmLoading) return

  emit('cancel')
  closeDialog()
}

function confirmDialog() {
  if (props.confirmLoading) return

  emit('confirm')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="ppmx-dialog-fade">
      <div
        v-if="modelValue"
        class="ppmx-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="ariaLabel || title"
      >
        <button
          v-if="closeOnBackdrop"
          class="ppmx-dialog__backdrop"
          type="button"
          :aria-label="cancelText"
          @click="cancelDialog"
        />
        <div v-else class="ppmx-dialog__backdrop" />

        <section class="ppmx-dialog__card" :class="`is-${variant}`">
          <button
            class="ppmx-dialog__close"
            type="button"
            :aria-label="cancelText"
            :disabled="confirmLoading"
            @click="cancelDialog"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <slot name="icon">
            <div class="ppmx-dialog__icon">
              <i class="fa-solid" :class="icon" />
            </div>
          </slot>

          <div class="ppmx-dialog__body">
            <slot name="title">
              <h2 v-if="title">
                {{ title }}
              </h2>
            </slot>
            <slot>
              <p v-if="message">
                {{ message }}
              </p>
            </slot>
          </div>

          <div class="ppmx-dialog__actions">
            <slot name="actions">
              <button
                v-if="showCancel"
                class="ppmx-dialog__button ppmx-dialog__button--cancel"
                type="button"
                :disabled="confirmLoading"
                @click="cancelDialog"
              >
                {{ cancelText }}
              </button>
              <button
                class="ppmx-dialog__button ppmx-dialog__button--confirm"
                type="button"
                :disabled="confirmLoading"
                @click="confirmDialog"
              >
                <i v-if="confirmLoading" class="fa-solid fa-spinner fa-spin" />
                <span>{{ confirmText }}</span>
              </button>
            </slot>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.ppmx-dialog {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: grid;
  min-height: var(--app-visual-height, 100dvh);
  place-items: center;
  padding: 22px 22px calc(22px + var(--app-keyboard-inset, 0px));
  color: var(--app-text, #f5f5f7);
  font-family:
    "Montserrat",
    system-ui,
    -apple-system,
    "Segoe UI",
    sans-serif;
}

.ppmx-dialog__backdrop {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 50% 28%, rgba(200, 16, 46, 0.18), transparent 45%),
    rgba(0, 0, 0, 0.7);
  border: 0;
  backdrop-filter: blur(8px) saturate(130%);
  -webkit-backdrop-filter: blur(8px) saturate(130%);
}

.ppmx-dialog__card {
  position: relative;
  display: grid;
  gap: 18px;
  width: min(100%, 430px);
  max-height: calc(var(--app-visual-height, 100dvh) - 32px);
  padding: 30px;
  overflow-y: auto;
  background:
    radial-gradient(circle at 14% 0%, rgba(255, 215, 0, 0.12), transparent 36%),
    radial-gradient(circle at 100% 12%, rgba(200, 16, 46, 0.22), transparent 42%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.09), rgba(255, 255, 255, 0.028)),
    rgba(9, 9, 12, 0.98);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 28px;
  box-shadow:
    0 32px 100px rgba(0, 0, 0, 0.82),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

html.is-ios-keyboard-open .ppmx-dialog {
  align-items: start;
  overflow-y: auto;
}

.ppmx-dialog__card::before {
  position: absolute;
  top: 0;
  right: 34px;
  left: 34px;
  height: 1px;
  content: "";
  background: linear-gradient(90deg, transparent, rgba(255, 215, 0, 0.72), transparent);
}

.ppmx-dialog__close {
  position: absolute;
  top: 18px;
  right: 18px;
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  color: rgba(245, 245, 247, 0.58);
  cursor: pointer;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
}

.ppmx-dialog__close:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.09);
}

.ppmx-dialog__icon {
  display: grid;
  width: 58px;
  height: 58px;
  place-items: center;
  color: #fff;
  background: linear-gradient(180deg, #e43b45, #b9152a);
  border-radius: 18px;
  box-shadow:
    0 16px 34px rgba(200, 16, 46, 0.34),
    inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.ppmx-dialog__card.is-gold .ppmx-dialog__icon {
  color: #1f1302;
  background: linear-gradient(180deg, #ffe889, #efbc3e);
  box-shadow: 0 16px 34px rgba(255, 215, 0, 0.18);
}

.ppmx-dialog__card.is-neutral .ppmx-dialog__icon {
  color: rgba(245, 245, 247, 0.88);
  background: rgba(255, 255, 255, 0.08);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.12);
}

.ppmx-dialog__body {
  display: grid;
  gap: 10px;
  padding-right: 24px;
}

.ppmx-dialog__body h2 {
  margin: 0;
  color: #fff;
  font-size: clamp(1.3rem, 4vw, 1.62rem);
  font-weight: 900;
  line-height: 1.12;
  letter-spacing: 0;
}

.ppmx-dialog__body p {
  margin: 0;
  color: rgba(245, 245, 247, 0.62);
  font-size: 0.94rem;
  font-weight: 700;
  line-height: 1.65;
}

.ppmx-dialog__actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 4px;
}

.ppmx-dialog__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 50px;
  padding: 0 16px;
  color: inherit;
  font: inherit;
  font-size: 0.95rem;
  font-weight: 900;
  cursor: pointer;
  border-radius: 16px;
  transition:
    transform 0.18s cubic-bezier(0.2, 0.7, 0.2, 1),
    filter 0.2s ease;
}

.ppmx-dialog__button:active {
  transform: scale(0.98);
}

.ppmx-dialog__button:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.ppmx-dialog__button--cancel {
  color: rgba(245, 245, 247, 0.78);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.07), rgba(255, 255, 255, 0.025)),
    rgba(255, 255, 255, 0.035);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.ppmx-dialog__button--confirm {
  gap: 8px;
  color: #fff;
  background: linear-gradient(180deg, #e43b45, #b9152a);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow:
    0 16px 34px rgba(200, 16, 46, 0.34),
    inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.ppmx-dialog__card.is-gold .ppmx-dialog__button--confirm {
  color: #201404;
  background: linear-gradient(180deg, #ffe889, #efbc3e);
  box-shadow: 0 16px 34px rgba(255, 215, 0, 0.18);
}

.ppmx-dialog-fade-enter-active,
.ppmx-dialog-fade-leave-active {
  transition: opacity 0.22s ease;
}

.ppmx-dialog-fade-enter-active .ppmx-dialog__card,
.ppmx-dialog-fade-leave-active .ppmx-dialog__card {
  transition: transform 0.26s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.22s ease;
}

.ppmx-dialog-fade-enter-from,
.ppmx-dialog-fade-leave-to {
  opacity: 0;
}

.ppmx-dialog-fade-enter-from .ppmx-dialog__card,
.ppmx-dialog-fade-leave-to .ppmx-dialog__card {
  opacity: 0;
  transform: translateY(14px) scale(0.97);
}

html[data-theme="light"] .ppmx-dialog__close {
  color: rgba(23, 23, 27, 0.54);
  background: rgba(255, 255, 255, 0.72);
  border-color: rgba(26, 26, 32, 0.1);
}

html[data-theme="light"] .ppmx-dialog__close:hover {
  color: #17171b;
  background: #fff;
}

html[data-theme="light"] .ppmx-dialog__card.is-neutral .ppmx-dialog__icon {
  color: #5a5a62;
  background: rgba(26, 26, 32, 0.06);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

html[data-theme="light"] .ppmx-dialog__card.is-gold .ppmx-dialog__button--confirm {
  color: #fff;
  background: linear-gradient(180deg, #ff5a5a, #ff1a1a);
  box-shadow: 0 16px 34px rgba(255, 26, 26, 0.22);
}

html[data-theme="light"] .ppmx-dialog__button--cancel:hover {
  color: #17171b;
  background: rgba(255, 255, 255, 0.96);
}

@media (max-width: 480px) {
  .ppmx-dialog {
    padding: 18px;
  }

  .ppmx-dialog__card {
    gap: 16px;
    padding: 26px 22px 22px;
    border-radius: 24px;
  }

  .ppmx-dialog__actions {
    grid-template-columns: 1fr;
  }
}
</style>
