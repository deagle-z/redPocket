<script setup lang="ts">
import { ppmxSocialItems, ppmxSocialText } from '@/pages/ppmx-home/data'
import { usePpmxLocale } from '@/pages/ppmx-home/composables/usePpmxLocale'
import type { PpmxSocialItem } from '@/pages/ppmx-home/types'

defineOptions({ name: 'PpmxSocialModal' })

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const { pickText } = usePpmxLocale()
const socialText = ppmxSocialText

function closeSocialModal() {
  emit('update:modelValue', false)
}

function openSocialChannel(social: PpmxSocialItem) {
  window.open(social.url, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="ppmx-social-fade">
      <div
        v-if="props.modelValue"
        id="socialModal"
        class="ppmx-social-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="socialModalTitle"
      >
        <button
          class="ppmx-social-modal__backdrop"
          type="button"
          :aria-label="pickText(socialText.close)"
          @click="closeSocialModal"
        />

        <section class="ppmx-social-card-wrap" role="document">
          <button
            class="ppmx-social-close"
            type="button"
            :aria-label="pickText(socialText.close)"
            @click="closeSocialModal"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <div class="ppmx-social-icon" aria-hidden="true">
            <i class="fa-solid fa-share-nodes" />
          </div>
          <h2 id="socialModalTitle">{{ pickText(socialText.title) }}</h2>
          <p>{{ pickText(socialText.sub) }}</p>

          <div class="ppmx-social-grid">
            <button
              v-for="social in ppmxSocialItems"
              :key="social.id"
              class="ppmx-social-card"
              :class="`is-${social.tone}`"
              type="button"
              :aria-label="`${pickText(socialText.open)} ${social.label}`"
              @click="openSocialChannel(social)"
            >
              <span class="ppmx-social-mark" aria-hidden="true">
                <i class="fa-brands" :class="social.icon" />
              </span>
              <strong>{{ social.label }}</strong>
              <small>{{ social.handle }}</small>
            </button>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>
