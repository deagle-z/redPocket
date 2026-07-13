<script setup lang="ts">
type PpmxGuestPanelButtonVariant = 'red' | 'gold' | 'ghost'

const props = withDefaults(defineProps<{
  id?: string
  icon?: string
  title: string
  message: string
  primaryText: string
  primaryVariant?: PpmxGuestPanelButtonVariant
  secondaryText?: string
  secondaryVariant?: PpmxGuestPanelButtonVariant
  ariaLabel?: string
}>(), {
  icon: 'fa-user-lock',
  primaryVariant: 'red',
  secondaryVariant: 'ghost',
  secondaryText: '',
  ariaLabel: '',
})

defineEmits<{
  primary: []
  secondary: []
}>()

function buttonClass(variant: PpmxGuestPanelButtonVariant) {
  return `ppmx-btn-${variant}`
}
</script>

<template>
  <section :id="props.id" class="ppmx-account-guest" :aria-label="props.ariaLabel || props.title">
    <i class="fa-solid" :class="props.icon" />
    <h2>{{ props.title }}</h2>
    <p>{{ props.message }}</p>
    <div>
      <button
        v-if="props.secondaryText"
        :class="buttonClass(props.secondaryVariant)"
        type="button"
        @click="$emit('secondary')"
      >
        {{ props.secondaryText }}
      </button>
      <button :class="buttonClass(props.primaryVariant)" type="button" @click="$emit('primary')">
        {{ props.primaryText }}
      </button>
    </div>
  </section>
</template>
