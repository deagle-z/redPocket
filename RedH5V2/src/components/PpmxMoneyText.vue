<script setup lang="ts">
import { formatMoney } from '@/utils/money'

const props = withDefaults(defineProps<{
  cents: number
  currency?: string
  signed?: boolean
  suffix?: string
}>(), {
  currency: '$',
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
