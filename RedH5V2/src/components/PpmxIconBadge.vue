<script setup lang="ts">
type PpmxIconBadgeTone = 'green' | 'blue' | 'gold' | 'red' | 'teal' | 'violet' | 'pink'

const props = withDefaults(defineProps<{
  badgeClass?: string
  icon: string
  iconPrefix?: string
  tone?: PpmxIconBadgeTone
}>(), {
  badgeClass: 'ppmx-account-hub-card__icon',
  iconPrefix: 'fa-solid',
  tone: 'gold',
})
</script>

<template>
  <span class="ppmx-icon-badge" :class="[props.badgeClass, `is-${props.tone}`]" aria-hidden="true">
    <i :class="[props.iconPrefix, props.icon]" />
  </span>
</template>
