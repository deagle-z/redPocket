<script setup lang="ts">
interface Props {
  show: boolean
  title?: string
  cancelText?: string
  confirmText?: string
  closeOnClickOverlay?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  cancelText: '',
  confirmText: '',
  closeOnClickOverlay: true,
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'cancel'): void
  (e: 'confirm'): void
}>()

const { t } = useI18n()
const dialogTitle = computed(() => props.title || t('common.confirmTitle'))
const dialogCancelText = computed(() => props.cancelText || t('common.cancel'))
const dialogConfirmText = computed(() => props.confirmText || t('common.confirm'))
const dialogDefaultContent = computed(() => t('common.confirmAction'))

function closeDialog() {
  emit('update:show', false)
}

function onCancel() {
  emit('cancel')
  closeDialog()
}

function onConfirm() {
  emit('confirm')
  closeDialog()
}

function onClickOverlay() {
  if (!props.closeOnClickOverlay)
    return
  closeDialog()
}
</script>

<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="show" class="confirm-overlay" @click="onClickOverlay">
        <div class="confirm-dialog" role="dialog" aria-modal="true" @click.stop>
          <!-- Gold top accent stripe -->
          <div class="dialog-accent" />
          <!-- Corner brackets -->
          <div class="dialog-corner tl" />
          <div class="dialog-corner tr" />
          <div class="confirm-body">
            <h3 class="confirm-title">
              {{ dialogTitle }}
            </h3>
            <div class="confirm-content">
              <slot>
                {{ dialogDefaultContent }}
              </slot>
            </div>
          </div>
          <div class="confirm-actions">
            <button type="button" class="action-btn cancel-btn" @click="onCancel">
              {{ dialogCancelText }}
            </button>
            <div class="action-divider" />
            <button type="button" class="action-btn confirm-btn" @click="onConfirm">
              {{ dialogConfirmText }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 3000;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.confirm-dialog {
  position: relative;
  width: min(100%, 340px);
  border-radius: 18px;
  overflow: hidden;
  background: linear-gradient(160deg, #7c0000 0%, #560000 55%, #3a0000 100%);
  border: 1px solid rgba(212, 175, 55, 0.5);
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.6),
    inset 0 0 0 1px rgba(255, 248, 214, 0.1),
    0 0 0 1px rgba(212, 175, 55, 0.25);
}

/* Gold dot watermark */
.confirm-dialog::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px);
  background-size: 18px 18px;
  opacity: 0.05;
  pointer-events: none;
  z-index: 0;
}

/* Gold top accent stripe */
.dialog-accent {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 15%, #ffd700 40%, #d4af37 60%, #b8860b 85%, transparent 100%);
  z-index: 4;
}

/* Corner brackets */
.dialog-corner {
  position: absolute;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(212, 175, 55, 0.65);
  z-index: 3;
  pointer-events: none;
}

.dialog-corner.tl {
  top: 10px;
  left: 10px;
  border-right: none;
  border-bottom: none;
}

.dialog-corner.tr {
  top: 10px;
  right: 10px;
  border-left: none;
  border-bottom: none;
}

.confirm-body {
  position: relative;
  z-index: 1;
  padding: 28px 22px 20px;
  text-align: center;
}

.confirm-title {
  margin: 0;
  color: #ffd98b;
  font-size: 17px;
  line-height: 1.1;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.4);
}

.confirm-content {
  margin-top: 12px;
  color: rgba(255, 229, 186, 0.72);
  font-size: 14px;
  line-height: 1.45;
}

.confirm-actions {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: stretch;
  border-top: 1px solid rgba(212, 175, 55, 0.22);
  height: 52px;
}

.action-divider {
  width: 1px;
  background: rgba(212, 175, 55, 0.22);
  flex-shrink: 0;
}

.action-btn {
  flex: 1;
  height: 100%;
  border: none;
  font-size: 15px;
  line-height: 1;
  font-weight: 600;
  background: transparent;
  letter-spacing: 0.02em;
  transition: background 0.15s ease;
}

.cancel-btn {
  color: rgba(255, 229, 186, 0.55);
}

.cancel-btn:active {
  background: rgba(255, 255, 255, 0.05);
}

.confirm-btn {
  background: transparent;
  color: #ffd98b;
  text-shadow: 0 0 6px rgba(212, 175, 55, 0.4);
}

.confirm-btn:active {
  background: rgba(212, 175, 55, 0.12);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.22s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
