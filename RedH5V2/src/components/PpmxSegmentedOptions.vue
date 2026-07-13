<script setup lang="ts">
export interface PpmxSegmentedOption {
  disabled?: boolean
  icon?: string
  label: string
  value: string
}

const props = withDefaults(defineProps<{
  itemClass?: string
  modelValue: string
  options: PpmxSegmentedOption[]
  rowClass?: string
}>(), {
  itemClass: 'ppmx-records-chip',
  rowClass: 'ppmx-records-chip-row',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="ppmx-segmented-options" :class="props.rowClass">
    <button
      v-for="option in props.options"
      :key="option.value"
      type="button"
      :class="[props.itemClass, { 'is-active': props.modelValue === option.value }]"
      :disabled="option.disabled"
      @click="emit('update:modelValue', option.value)"
    >
      <i v-if="option.icon" class="fa-solid" :class="option.icon" />
      {{ option.label }}
    </button>
  </div>
</template>
