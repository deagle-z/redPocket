<script setup lang="ts">
const props = withDefaults(defineProps<{
  fieldClass?: string
  label: string
  required?: boolean
}>(), {
  fieldClass: 'ppmx-recharge-field',
  required: false,
})
</script>

<template>
  <label :class="props.fieldClass">
    <span>
      <slot name="label">{{ props.label }}</slot>
      <em v-if="props.required">*</em>
    </span>
    <slot />
  </label>
</template>
