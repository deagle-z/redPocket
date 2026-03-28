<script setup lang="ts">
import CoinAmount from './CoinAmount.vue'

const props = defineProps<{ text: string }>()

// Split on currency amounts: optional sign + coin marker + optional space + digits
const parts = computed(() => props.text.split(/([-+]?\u0E3F\s*\d+\.?\d*)/g))
</script>

<template>
  <span>
    <template v-for="(part, i) in parts" :key="i">
      <CoinAmount v-if="/\u0E3F/.test(part)" :text="part" />
      <template v-else>{{ part }}</template>
    </template>
  </span>
</template>
