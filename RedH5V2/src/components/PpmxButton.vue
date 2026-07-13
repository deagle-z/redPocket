<script setup lang="ts">
defineOptions({ name: 'PpmxButton' })

const props = withDefaults(defineProps<{
  block?: boolean
  disabled?: boolean
  loading?: boolean
  size?: 'sm' | 'md' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  variant?: 'red' | 'gold' | 'ghost'
}>(), {
  block: false,
  disabled: false,
  loading: false,
  size: 'md',
  type: 'button',
  variant: 'red',
})

const buttonClass = computed(() => [
  `ppmx-button--${props.variant}`,
  `ppmx-button--${props.size}`,
  { 'is-block': props.block, 'is-loading': props.loading },
])

const buttonStyle = computed(() => {
  if (props.variant === 'red') {
    return {
      '--ppmx-button-bg': 'linear-gradient(180deg, #e21f33 0%, #c8102e 58%, #8f071f 100%)',
      '--ppmx-button-color': '#fff',
      '--ppmx-button-border': '1px solid rgba(255, 255, 255, 0.16)',
      '--ppmx-button-shadow': '0 18px 42px -18px rgba(226, 31, 51, 0.95), inset 0 1px 0 rgba(255, 255, 255, 0.22)',
    }
  }

  if (props.variant === 'gold') {
    return {
      '--ppmx-button-bg': 'linear-gradient(180deg, #ffd700 0%, #d8a200 100%)',
      '--ppmx-button-color': '#1a0d00',
      '--ppmx-button-border': '1px solid rgba(255, 255, 255, 0.22)',
      '--ppmx-button-shadow': '0 18px 42px -24px rgba(255, 215, 0, 0.78), inset 0 1px 0 rgba(255, 255, 255, 0.35)',
    }
  }

  return {
    '--ppmx-button-bg': 'rgba(255, 255, 255, 0.025)',
    '--ppmx-button-color': '#fff',
    '--ppmx-button-border': '1px solid rgba(255, 255, 255, 0.16)',
    '--ppmx-button-shadow': 'none',
  }
})
</script>

<template>
  <button
    class="ppmx-button"
    :class="buttonClass"
    :disabled="props.disabled || props.loading"
    :style="buttonStyle"
    :type="props.type"
  >
    <i v-if="props.loading" class="fa-solid fa-circle-notch fa-spin" />
    <slot />
  </button>
</template>
