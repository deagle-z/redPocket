<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
  text?: string
  actionText?: string
  minHeight?: number
}>(), {
  title: 'Unable to load',
  text: 'Please try again later',
  actionText: 'Retry',
  minHeight: 220,
})

const emit = defineEmits<{
  (e: 'retry'): void
}>()
</script>

<template>
  <div class="app-retry-state" :style="{ minHeight: `${minHeight}px` }">
    <div class="app-retry-state__mark" aria-hidden="true">
      <van-icon name="replay" />
    </div>
    <h3>{{ title }}</h3>
    <p>{{ text }}</p>
    <button type="button" @click="emit('retry')">
      {{ actionText }}
    </button>
  </div>
</template>

<style scoped>
.app-retry-state {
  width: 100%;
  border: 1px solid rgba(212, 175, 55, 0.18);
  border-radius: var(--radius-xl);
  background: radial-gradient(circle at 50% 0%, rgba(212, 175, 55, 0.16), transparent 34%), rgba(80, 0, 0, 0.3);
  color: var(--color-game-text);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 22px 18px;
  text-align: center;
}

.app-retry-state__mark {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1px solid rgba(255, 218, 127, 0.32);
  background: rgba(31, 0, 0, 0.42);
  color: var(--color-gold-light);
  display: grid;
  place-items: center;
  font-size: 23px;
  box-shadow: var(--shadow-gold);
}

.app-retry-state h3 {
  margin: 2px 0 0;
  color: #fff2c8;
  font-size: var(--font-md);
  line-height: 1.2;
  font-weight: 800;
}

.app-retry-state p {
  max-width: 240px;
  margin: 0;
  color: var(--color-game-text-soft);
  font-size: var(--font-xs);
  line-height: 1.4;
}

.app-retry-state button {
  min-width: 104px;
  height: 36px;
  margin-top: 4px;
  border: 1px solid rgba(255, 245, 195, 0.7);
  border-radius: var(--radius-md);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.32), transparent 44%), var(--color-game-gold-gradient);
  color: #3a0800;
  font-size: var(--font-sm);
  font-weight: 900;
  box-shadow:
    0 8px 18px rgba(128, 45, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.62);
}
</style>
