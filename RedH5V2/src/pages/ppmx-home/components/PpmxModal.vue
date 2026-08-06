<script setup lang="ts">
import PpmxBaseModal from '@/components/PpmxBaseModal.vue'

defineOptions({ name: 'PpmxModal' })

const model = defineModel<boolean>({ required: true })

const props = withDefaults(
  defineProps<{
    ariaLabel?: string
    cardClass?: string
    closeAriaLabel?: string
    closeOnBackdrop?: boolean
  }>(),
  {
    ariaLabel: 'PP.PE modal',
    cardClass: '',
    closeAriaLabel: 'Cerrar',
    closeOnBackdrop: true,
  },
)

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <PpmxBaseModal
    v-model="model"
    :aria-label="props.ariaLabel"
    :card-class="props.cardClass"
    :close-aria-label="props.closeAriaLabel"
    :close-on-backdrop="props.closeOnBackdrop"
    @close="emit('close')"
  >
    <slot />
  </PpmxBaseModal>
</template>
