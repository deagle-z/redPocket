<script setup lang="ts">
defineOptions({ name: 'PpmxBaseModal' })

const props = withDefaults(
  defineProps<{
    ariaLabel?: string
    cardClass?: string
    closeAriaLabel?: string
    closeOnBackdrop?: boolean
    modelValue: boolean
  }>(),
  {
    ariaLabel: 'PP.BET modal',
    cardClass: '',
    closeAriaLabel: 'Cerrar',
    closeOnBackdrop: true,
  },
)

const emit = defineEmits<{
  close: []
  'update:modelValue': [value: boolean]
}>()

function closeModal() {
  emit('update:modelValue', false)
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="props.modelValue" class="ppmx-modal" role="dialog" aria-modal="true" :aria-label="props.ariaLabel">
      <button
        v-if="props.closeOnBackdrop"
        class="ppmx-modal__backdrop"
        type="button"
        :aria-label="props.closeAriaLabel"
        @click="closeModal"
      />
      <div v-else class="ppmx-modal__backdrop" />

      <section class="ppmx-modal__card ppmx-glass-strong" :class="props.cardClass">
        <button class="ppmx-modal__close" type="button" :aria-label="props.closeAriaLabel" @click="closeModal">
          <i class="fa-solid fa-xmark" />
        </button>

        <slot />
      </section>
    </div>
  </Teleport>
</template>
