<script setup lang="ts">
withDefaults(defineProps<{
  count?: number
  variant?: 'card' | 'row'
}>(), {
  count: 6,
  variant: 'card',
})
</script>

<template>
  <div class="app-skeleton-section" :class="`app-skeleton-section--${variant}`">
    <div v-for="idx in count" :key="idx" class="app-skeleton-section__item">
      <van-skeleton-image />
      <van-skeleton title :row="variant === 'row' ? 1 : 0" />
    </div>
  </div>
</template>

<style scoped>
.app-skeleton-section {
  width: 100%;
  display: grid;
  gap: 12px;
}

.app-skeleton-section--card {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.app-skeleton-section--row {
  grid-template-columns: 1fr;
}

.app-skeleton-section__item {
  min-width: 0;
  border: 1px solid var(--color-game-card-border);
  border-radius: var(--radius-md);
  background: var(--color-game-card);
  overflow: hidden;
}

.app-skeleton-section--card .app-skeleton-section__item {
  aspect-ratio: 0.76;
}

.app-skeleton-section--row .app-skeleton-section__item {
  min-height: 72px;
  padding: 12px;
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
}

.app-skeleton-section :deep(.van-skeleton-image) {
  width: 100%;
  height: 100%;
  border-radius: 0;
  background: rgba(255, 232, 194, 0.12);
}

.app-skeleton-section--row :deep(.van-skeleton-image) {
  width: 56px;
  height: 48px;
  border-radius: var(--radius-sm);
}

.app-skeleton-section :deep(.van-skeleton__title),
.app-skeleton-section :deep(.van-skeleton__row) {
  background: rgba(255, 232, 194, 0.16);
}
</style>
