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
        <!-- Ground/shelf -->
        <rect x="14" y="84" width="132" height="28" rx="2" fill="rgba(212,175,55,0.12)" />
        <!-- Envelope body -->
        <rect x="46" y="28" width="68" height="64" rx="3" fill="rgba(140,0,0,0.5)" />
        <!-- Envelope front -->
        <rect x="52" y="34" width="56" height="50" rx="2" fill="rgba(100,0,0,0.6)" />
        <!-- Gold dot lines -->
        <rect x="60" y="45" width="40" height="5" rx="2" fill="rgba(212,175,55,0.35)" />
        <rect x="60" y="57" width="40" height="5" rx="2" fill="rgba(212,175,55,0.25)" />
        <rect x="60" y="69" width="28" height="5" rx="2" fill="rgba(212,175,55,0.2)" />
        <!-- Decorative triangles -->
        <path d="M68 95h24v6H68z" rx="1" fill="rgba(212,175,55,0.15)" />
        <path d="M18 30l5-7 5 7z" fill="rgba(212,175,55,0.2)" />
        <path d="M124 22l6-8 6 8z" fill="rgba(212,175,55,0.2)" />
        <rect x="129" y="22" width="14" height="22" rx="2" fill="rgba(212,175,55,0.12)" />
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
  border-radius: 12px;
  background: rgba(80, 0, 0, 0.3);
  border: 1px solid rgba(212, 175, 55, 0.18);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
}

.empty-illustration {
  width: 130px;
  height: 80px;
  opacity: 0.75;
}

.empty-svg {
  width: 100%;
  height: 100%;
  display: block;
}

.empty-text {
  margin: 0;
  color: rgba(255, 229, 186, 0.5);
  font-size: 12px;
  line-height: 1;
  letter-spacing: 0.04em;
}
</style>
