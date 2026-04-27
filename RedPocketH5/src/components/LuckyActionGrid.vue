<script setup lang="ts">
import type { LuckyPlayType } from '@/utils/lucky-play'
import type { LuckyPacketAction } from '@/utils/lucky-actions'
import CoinAmount from './CoinAmount.vue'
import {
  formatLuckyActionAmount,
  formatLuckyActionLabel,
  isLuckyActionThunder,
} from '@/utils/lucky-actions'

const props = withDefaults(defineProps<{
  actions: LuckyPacketAction[]
  status: string
  playType: LuckyPlayType
  variant?: 'inline' | 'detail'
  keyPrefix?: string | number
}>(), {
  variant: 'inline',
  keyPrefix: '',
})

const emit = defineEmits<{
  grab: [action: LuckyPacketAction]
}>()

const { t } = useI18n()
const isOngoing = computed(() => props.status === 'ongoing')

function isDisabled(action: LuckyPacketAction) {
  return !isOngoing.value || Boolean(action.isGrabbed)
}

function getActionClass(action: LuckyPacketAction) {
  return {
    grabbed: Boolean(action.isGrabbed),
    mined: Boolean(action.isGrabMine),
    locked: !isOngoing.value && !action.isGrabbed,
    thunder: isLuckyActionThunder(action),
  }
}

function shouldShowAmount(action: LuckyPacketAction) {
  return Number(action.amount || 0) > 0
    && !isLuckyActionThunder(action)
    && !action.displayLoading
    && (Boolean(action.isGrabbed) || !isOngoing.value)
}

function handleGrab(action: LuckyPacketAction) {
  if (isDisabled(action))
    return
  emit('grab', action)
}
</script>

<template>
  <div
    class="lucky-action-grid"
    :class="[
      variant === 'detail' ? 'packet-actions detail-actions' : 'packet-actions-inline',
    ]"
  >
    <button
      v-for="action in actions"
      :key="`${keyPrefix}-${action.seqNo}`"
      type="button"
      class="action-pill"
      :class="getActionClass(action)"
      :disabled="isDisabled(action)"
      @click="handleGrab(action)"
    >
      <span v-if="isLuckyActionThunder(action) && playType !== 'parity'" aria-hidden="true">💣</span>
      <span v-else-if="isLuckyActionThunder(action) && playType === 'parity'" class="parity-lose-coin" aria-hidden="true">-</span>
      <span v-else-if="action.isGrabMine" class="mine-text">🎁 </span>
      <span v-else-if="playType === 'parity' && !action.isGrabbed && isOngoing" class="choice-mark">{{ t('homeLucky.parityChoiceMark') }}</span>
      <CoinAmount v-if="shouldShowAmount(action)" :text="formatLuckyActionAmount(Number(action.amount || 0))" />
      <template v-else>
        {{ formatLuckyActionLabel(t, action, isOngoing) }}
      </template>
    </button>
  </div>
</template>

<style scoped>
.packet-actions-inline {
  margin-top: 3px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 4px;
}

.detail-actions {
  padding: 10px 12px 12px;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 6px;
}

.action-pill {
  border: none;
  border-radius: 999px;
  background: linear-gradient(180deg, #9e1010 0%, #6a0000 100%);
  color: #fff3de;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px 3px;
  font-size: 7px;
  line-height: 1.15;
  text-align: center;
  word-break: break-word;
  min-height: 28px;
}

.detail-actions .action-pill {
  font-size: 8px;
}

.action-pill.grabbed,
.action-pill.locked {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 248, 214, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.14);
}

.action-pill.thunder {
  color: #ffe088;
}

.parity-lose-coin {
  font-size: 1.15em;
  line-height: 1;
  font-weight: 700;
  opacity: 0.6;
}

.mine-text {
  margin-right: 3px;
  color: #ffd45d;
}

.choice-mark {
  color: #8fd5ff;
}

@media (width <= 340px) {
  .detail-actions {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
