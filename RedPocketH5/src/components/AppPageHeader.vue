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
  height: 44px;
  display: grid;
  grid-template-columns: 36px 1fr 36px;
  align-items: center;
  padding: 0 4px;
}

.header-title {
  margin: 0;
  text-align: center;
  font-size: 16px;
  font-weight: 700;
  color: #ffd98b;
  letter-spacing: 0.04em;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.35);
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
  font-size: 18px;
  color: #d4af37;
  filter: drop-shadow(0 0 3px rgba(212, 175, 55, 0.35));
  transition: opacity 0.15s ease;
}

.header-btn:active {
  opacity: 0.7;
}
</style>
