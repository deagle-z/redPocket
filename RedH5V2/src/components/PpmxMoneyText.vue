<script setup lang="ts">
import { formatMoney } from '@/utils/money'
import { APP_CURRENCY_SYMBOL } from '@/config/market'

const props = withDefaults(defineProps<{
  cents: number
  currency?: string
  signed?: boolean
  suffix?: string
}>(), {
  currency: APP_CURRENCY_SYMBOL,
  signed: false,
  suffix: '',
})

const formattedValue = computed(() => {
  const sign = props.signed && props.cents > 0 ? '+' : props.signed && props.cents < 0 ? '-' : ''

  return `${sign}${formatMoney(Math.abs(props.cents), { currency: props.currency })}`
})
</script>

<template>
  <strong class="ppmx-money-text">
    <span>{{ formattedValue }}</span>
    <em v-if="props.suffix">{{ props.suffix }}</em>
  </strong>
</template>
