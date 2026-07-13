<script setup lang="ts">
import PpmxIconBadge from './PpmxIconBadge.vue'

type ClassValue = string | Record<string, boolean> | Array<string | Record<string, boolean>>
type PpmxActionCardTone = 'green' | 'blue' | 'gold' | 'red' | 'teal' | 'violet' | 'pink'

const props = withDefaults(defineProps<{
  baseClass?: string
  bodyClass?: string
  cardClass?: ClassValue
  chevron?: boolean
  description?: string
  disabled?: boolean
  icon: string
  iconClass?: string
  title: string
  tone?: PpmxActionCardTone
}>(), {
  baseClass: 'ppmx-account-hub-card',
  bodyClass: '',
  cardClass: '',
  chevron: true,
  description: '',
  disabled: false,
  iconClass: 'ppmx-account-hub-card__icon',
  tone: 'gold',
})

defineEmits<{
  click: [event: MouseEvent]
}>()
</script>

<template>
  <button
    class="ppmx-action-card"
    :class="[props.baseClass, props.cardClass]"
    type="button"
    :disabled="props.disabled"
    @click="$emit('click', $event)"
  >
    <PpmxIconBadge :icon="props.icon" :tone="props.tone" :badge-class="props.iconClass" />
    <span :class="props.bodyClass">
      <strong>{{ props.title }}</strong>
      <small v-if="props.description">{{ props.description }}</small>
    </span>
    <i v-if="props.chevron" class="fa-solid fa-chevron-right" />
  </button>
</template>
