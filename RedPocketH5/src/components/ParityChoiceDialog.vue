<script setup lang="ts">
import type { ParityChoice } from '@/utils/lucky-play'

withDefaults(defineProps<{
  show: boolean
  senderName?: string
}>(), {
  senderName: '',
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'confirm', value: ParityChoice): void
}>()

const { t } = useI18n()

function closeDialog() {
  emit('update:show', false)
}

function choose(value: ParityChoice) {
  emit('confirm', value)
}
</script>

<template>
  <van-popup
    :show="show"
    round
    position="bottom"
    class="parity-choice-popup"
    @update:show="emit('update:show', $event)"
    @click-overlay="closeDialog"
  >
    <section class="parity-choice-sheet">
      <p class="parity-choice-sheet__eyebrow">
        {{ t('homeLucky.parityChoiceEyebrow') }}
      </p>
      <h3 class="parity-choice-sheet__title">
        {{ t('homeLucky.parityChoiceTitle') }}
      </h3>
      <p class="parity-choice-sheet__sub">
        {{ t('homeLucky.parityChoiceSub', { user: senderName || t('grabModal.defaultSender') }) }}
      </p>

      <div class="parity-choice-sheet__grid">
        <button type="button" class="parity-choice-btn odd" @click="choose('odd')">
          <span class="parity-choice-btn__label">{{ t('homeLucky.choiceOdd') }}</span>
          <span class="parity-choice-btn__hint">{{ t('homeLucky.choiceOddHint') }}</span>
        </button>
        <button type="button" class="parity-choice-btn even" @click="choose('even')">
          <span class="parity-choice-btn__label">{{ t('homeLucky.choiceEven') }}</span>
          <span class="parity-choice-btn__hint">{{ t('homeLucky.choiceEvenHint') }}</span>
        </button>
      </div>

      <button type="button" class="parity-choice-sheet__cancel" @click="closeDialog">
        {{ t('common.cancel') }}
      </button>
    </section>
  </van-popup>
</template>

<style scoped>
.parity-choice-sheet {
  padding: 18px 14px calc(14px + env(safe-area-inset-bottom));
  background:
    radial-gradient(circle at 12% 10%, rgba(212, 175, 55, 0.18), transparent 22%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
}

.parity-choice-sheet__eyebrow {
  margin: 0;
  text-align: center;
  color: #ffd98b;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.parity-choice-sheet__title {
  margin: 8px 0 0;
  text-align: center;
  color: #fff0c9;
  font-size: 20px;
  font-weight: 800;
}

.parity-choice-sheet__sub {
  margin: 8px 0 0;
  text-align: center;
  color: rgba(255, 229, 186, 0.72);
  font-size: 12px;
  line-height: 1.45;
}

.parity-choice-sheet__grid {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.parity-choice-btn {
  min-height: 104px;
  border: 1px solid rgba(255, 248, 214, 0.18);
  border-radius: 18px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 6px;
  padding: 14px;
  color: #fff7e8;
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.parity-choice-btn.odd {
  background: linear-gradient(165deg, rgba(177, 24, 24, 0.96), rgba(108, 0, 0, 0.96));
}

.parity-choice-btn.even {
  background: linear-gradient(165deg, rgba(26, 90, 127, 0.96), rgba(10, 47, 88, 0.96));
}

.parity-choice-btn__label {
  font-size: 22px;
  font-weight: 800;
  line-height: 1;
}

.parity-choice-btn__hint {
  color: rgba(255, 247, 232, 0.76);
  font-size: 12px;
  line-height: 1.3;
}

.parity-choice-sheet__cancel {
  margin-top: 12px;
  width: 100%;
  height: 44px;
  border: 1px solid rgba(255, 248, 214, 0.2);
  border-radius: 999px;
  background: rgba(255, 248, 214, 0.08);
  color: #fff0c9;
  font-size: 14px;
  font-weight: 700;
}
</style>
