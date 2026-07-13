<script setup lang="ts">
import PpmxButton from '@/components/PpmxButton.vue'

defineOptions({ name: 'PpmxButtonLegacy' })
</script>

<template>
  <PpmxButton v-bind="$attrs">
    <slot />
  </PpmxButton>
</template>
