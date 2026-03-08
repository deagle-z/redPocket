<script setup lang="ts">
const props = withDefaults(defineProps<{
  text?: string
  minHeight?: number
}>(), {
  minHeight: 220,
})

const { t } = useI18n()
const displayText = computed(() => props.text || t('common.noMore'))
</script>

<template>
  <div class="app-empty" :style="{ minHeight: `${props.minHeight}px` }">
    <div class="empty-illustration" aria-hidden="true">
      <svg viewBox="0 0 160 120" class="empty-svg">
        <rect x="14" y="84" width="132" height="28" rx="2" fill="#d5d9de" />
        <rect x="46" y="28" width="68" height="64" rx="3" fill="#e2e5e9" />
        <rect x="52" y="34" width="56" height="50" rx="2" fill="#f3f5f7" />
        <rect x="60" y="45" width="40" height="8" rx="2" fill="#d6d9de" />
        <rect x="60" y="59" width="40" height="8" rx="2" fill="#d6d9de" />
        <rect x="60" y="73" width="40" height="8" rx="2" fill="#d6d9de" />
        <rect x="68" y="95" width="24" height="6" rx="1" fill="#e7eaee" />
        <path d="M18 30l5-7 5 7z" fill="#eef1f4" />
        <path d="M124 22l6-8 6 8z" fill="#eef1f4" />
        <rect x="129" y="22" width="14" height="22" rx="2" fill="#eef1f4" />
      </svg>
    </div>
    <p class="empty-text">
      {{ displayText }}
    </p>
  </div>
</template>

<style scoped>
.app-empty {
  width: 100%;
  border-radius: 8px;
  background: #f8f9fb;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
}

.empty-illustration {
  width: 160px;
  height: 100px;
}

.empty-svg {
  width: 100%;
  height: 100%;
  display: block;
}

.empty-text {
  margin: 0;
  color: #6b7280;
  font-size: 12px;
  line-height: 1;
}
</style>
