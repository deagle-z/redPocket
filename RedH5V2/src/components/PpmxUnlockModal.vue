<script setup lang="ts">
import { usePpmxLocale } from '@/pages/ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '@/pages/ppmx-home/types'

defineOptions({ name: 'PpmxUnlockModal' })

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  activate: []
  reject: []
}>()

const { pickText } = usePpmxLocale()

const unlockText = {
  title: {
    es: '¡Deposita para desbloquear los juegos!',
    en: 'Deposit to unlock the games!',
    zh: '充值即可解锁游戏！',
  },
  sub: {
    es: '¡Tu recompensa de juego exclusiva! ¡Actívala al instante después de depositar!',
    en: 'Your exclusive game reward! Activate it instantly after depositing!',
    zh: '专属游戏奖励，充值后立即激活！',
  },
  go: { es: 'Activar ahora', en: 'Activate now', zh: '立即激活' },
  no: { es: 'Rechazar cruelmente', en: 'Reject cruelly', zh: '狠心拒绝' },
} satisfies Record<string, LocalizedText>

function reject() {
  emit('update:modelValue', false)
  emit('reject')
}

function activate() {
  emit('update:modelValue', false)
  emit('activate')
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="props.modelValue"
      class="ppmx-unlock-modal"
      role="dialog"
      aria-modal="true"
      :aria-label="pickText(unlockText.title)"
    >
      <button class="ppmx-unlock-modal__backdrop" type="button" :aria-label="pickText(unlockText.no)" @click="reject" />

      <section class="ppmx-unlock-card">
        <span class="ppmx-unlock-glow" />
        <div class="ppmx-unlock-body">
          <div class="ppmx-unlock-hero">
            <span class="ppmx-unlock-coin"><i class="fa-solid fa-gift" /></span>
            <span class="ppmx-unlock-lock"><i class="fa-solid fa-lock" /></span>
            <span class="ppmx-unlock-spark s1" />
            <span class="ppmx-unlock-spark s2" />
            <span class="ppmx-unlock-spark s3" />
          </div>
          <h2 class="ppmx-unlock-title">{{ pickText(unlockText.title) }}</h2>
          <p class="ppmx-unlock-sub">{{ pickText(unlockText.sub) }}</p>
          <div class="ppmx-unlock-actions">
            <button class="ppmx-unlock-btn ppmx-unlock-btn--go" type="button" @click="activate">
              <i class="fa-solid fa-bolt" />
              <span>{{ pickText(unlockText.go) }}</span>
            </button>
            <button class="ppmx-unlock-btn ppmx-unlock-btn--no" type="button" @click="reject">
              <span>{{ pickText(unlockText.no) }}</span>
            </button>
          </div>
        </div>
      </section>
    </div>
  </Teleport>
</template>
