<script setup lang="ts">
defineProps<{
  title: string
  iconColor?: string
}>()

const emit = defineEmits<{
  (e: 'back'): void
  (e: 'right-click'): void
}>()
</script>

<template>
  <header class="app-page-header">
    <button class="header-btn" type="button" @click="emit('back')">
      <slot name="left">
        <van-icon name="arrow-left" />
      </slot>
    </button>
    <h1 class="header-title">
      {{ title }}
    </h1>
    <button class="header-btn" type="button" @click="emit('right-click')">
      <slot name="right" />
    </button>
  </header>
</template>

<style scoped>
.app-page-header {
  height: var(--header-height);
  display: grid;
  grid-template-columns: 40px 1fr 40px;
  align-items: center;
}

.header-title {
  margin: 0;
  text-align: center;
  font-size: var(--font-2xl);
  font-weight: 700;
  color: var(--color-text-title);
}

.header-btn {
  height: 32px;
  width: 32px;
  border: none;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  font-size: var(--font-2xl);
  color: var(--color-primary);
}
</style>
