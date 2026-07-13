<script setup lang="ts">
const props = withDefaults(defineProps<{
  closeAriaLabel?: string
  eyebrow?: string
  id?: string
  modelValue: boolean
  title: string
  titleId?: string
}>(), {
  closeAriaLabel: 'Cerrar',
  eyebrow: '',
  id: undefined,
  titleId: 'withdrawModalTitle',
})

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
    <Transition name="ppmx-withdraw-fade">
      <div
        v-if="props.modelValue"
        :id="props.id"
        class="ppmx-withdraw-modal"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="props.titleId"
      >
        <button
          class="ppmx-withdraw-modal__backdrop"
          type="button"
          :aria-label="props.closeAriaLabel"
          @click="closeModal"
        />
        <section class="ppmx-withdraw-card" role="document">
          <button
            class="ppmx-withdraw-close"
            type="button"
            :aria-label="props.closeAriaLabel"
            @click="closeModal"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <div class="ppmx-withdraw-scroll">
            <header class="ppmx-withdraw-head">
              <span v-if="props.eyebrow" class="ppmx-eyebrow">{{ props.eyebrow }}</span>
              <h2 :id="props.titleId">{{ props.title }}</h2>
            </header>

            <slot />
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>
