<script setup lang="ts">
import imgCoin from '@/assets/svg/coin.svg'

const props = defineProps<{ text: string }>()

// Strip ฿ symbol wherever it appears, preserve +/- signs
const stripped = computed(() => props.text.replace(/฿\s*/g, '').trim())
</script>

<template>
  <span class="coin-amount-wrap">
    <img class="coin-amount-icon" :src="imgCoin" alt="">
    <span class="coin-amount-text">{{ stripped }}</span>
  </span>
</template>

<style scoped>
.coin-amount-wrap {
  display: inline-flex;
  align-items: center;
  gap: 0.18em;
  vertical-align: middle;
  line-height: inherit;
}

.coin-amount-icon {
  width: 1.15em;
  height: 1.15em;
  flex-shrink: 0;
  display: block;
}

.coin-amount-text {
  line-height: inherit;
}
</style>
