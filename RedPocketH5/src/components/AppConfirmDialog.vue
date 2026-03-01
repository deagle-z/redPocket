<script setup lang="ts">
interface Props {
  show: boolean
  title?: string
  cancelText?: string
  confirmText?: string
  closeOnClickOverlay?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '确认退出',
  cancelText: '取消',
  confirmText: '确认',
  closeOnClickOverlay: true,
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'cancel'): void
  (e: 'confirm'): void
}>()

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
          <div class="confirm-body">
            <h3 class="confirm-title">
              {{ title }}
            </h3>
            <div class="confirm-content">
              <slot>
                确定要执行该操作吗？
              </slot>
            </div>
          </div>
          <div class="confirm-actions">
            <button type="button" class="action-btn cancel-btn" @click="onCancel">
              {{ cancelText }}
            </button>
            <button type="button" class="action-btn confirm-btn" @click="onConfirm">
              {{ confirmText }}
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
  background: rgba(0, 0, 0, 0.58);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.confirm-dialog {
  width: min(100%, 520px);
  border-radius: 24px;
  overflow: hidden;
  background: #fff;
}

.confirm-body {
  padding: 26px 20px 20px;
  text-align: center;
}

.confirm-title {
  margin: 0;
  color: #111827;
  font-size: 20px;
  line-height: 1.1;
  font-weight: 700;
}

.confirm-content {
  margin-top: 10px;
  color: #6b7280;
  font-size: 16px;
  line-height: 1.35;
}

.confirm-actions {
  display: flex;
  align-items: center;
  height: 56px;
}

.action-btn {
  flex: 1;
  height: 100%;
  border: none;
  font-size: 17px;
  line-height: 1;
  font-weight: 600;
}

.cancel-btn {
  background: #fff;
  color: #4b5563;
}

.confirm-btn {
  background: #4fae63;
  color: #fff;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
