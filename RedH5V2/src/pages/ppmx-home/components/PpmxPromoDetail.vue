<script setup lang="ts">
import { usePpmxLocale } from '../composables/usePpmxLocale'
import type { PpmxPromoItem } from '../types'
import { highlightPromoText } from '../utils/promoHighlight'

defineOptions({ name: 'PpmxPromoDetail' })

defineProps<{
  selectedPromo: PpmxPromoItem
}>()

const emit = defineEmits<{
  back: []
  claim: []
}>()

const { t } = useI18n()
const { pickText } = usePpmxLocale()
</script>

<template>
  <section
    id="promoDetail"
    class="ppmx-promo-detail reveal"
    :data-promo-detail="selectedPromo.id"
  >
    <span
      v-if="selectedPromo.backgroundImage"
      class="ppmx-promo-detail__media"
      :style="{
        backgroundImage: `url(${selectedPromo.backgroundImage})`,
        backgroundPosition: selectedPromo.backgroundPosition || 'right center',
      }"
    />
    <button class="ppmx-promo-back" type="button" @click="emit('back')">
      <i class="fa-solid fa-arrow-left" />
      <span>{{ t('ppmx.promo.detailBack') }}</span>
    </button>
    <div class="ppmx-promo-detail__body">
      <span class="ppmx-promo-tag">{{ pickText(selectedPromo.tag) }}</span>
      <h2 class="ppmx-display" v-html="highlightPromoText(pickText(selectedPromo.title))" />
      <p v-html="highlightPromoText(pickText(selectedPromo.sub))" />
    </div>
    <div class="ppmx-promo-detail__meta">
      <article>
        <span>{{ t('ppmx.promo.detailParticipate') }}</span>
        <p>{{ pickText(selectedPromo.howTo) }}</p>
      </article>
      <article>
        <span>{{ t('ppmx.promo.detailValidity') }}</span>
        <p>{{ pickText(selectedPromo.valid) }}</p>
      </article>
    </div>
    <div class="ppmx-promo-detail__footer">
      <p>
        <i class="fa-solid fa-circle-info" />
        <span>{{ pickText(selectedPromo.terms) }}</span>
      </p>
      <button class="ppmx-btn-red" type="button" @click="emit('claim')">
        {{ t('ppmx.promo.claim') }}
        <i class="fa-solid fa-arrow-right" />
      </button>
    </div>
  </section>
</template>
