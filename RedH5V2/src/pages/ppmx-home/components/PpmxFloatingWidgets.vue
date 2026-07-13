<script setup lang="ts">
defineOptions({ name: 'PpmxFloatingWidgets' })

withDefaults(
  defineProps<{
    hiding?: boolean
  }>(),
  {
    hiding: false,
  },
)
</script>

<template>
  <div class="ppmx-floating-widgets" :class="{ 'is-scroll-hiding': hiding }">
    <slot />
  </div>
</template>
