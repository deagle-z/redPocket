<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import PpmxSubpageLayout from '@/components/PpmxSubpageLayout.vue'
import {
  ppmxGuideCards,
  ppmxGuideText,
} from '../ppmx-home/pageData'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'

definePage({
  name: 'guide',
  meta: {
    titleKey: 'ppmx.guide.title',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

const router = useRouter()
const { pickText } = usePpmxLocale()

// Download-app guide activity is hidden from the activity guide list.
const visibleGuideCards = computed(() => ppmxGuideCards.filter((card) => card.id !== 'download'))

function openGuideCard(id: string) {
  void router.push(`/promo/${id}`)
}
</script>

<template>
  <PpmxSubpageLayout
    page-id="page-guide"
    page-class="ppmx-guide-page"
    stack-class="ppmx-guide-stack"
    head-class="ppmx-guide-head reveal"
    width="compact"
    :eyebrow="pickText(ppmxGuideText.eyebrow)"
    :title="pickText(ppmxGuideText.title)"
    :description="pickText(ppmxGuideText.sub)"
  >
    <section class="ppmx-guide-list" :aria-label="pickText(ppmxGuideText.title)">
      <article
        v-for="card in visibleGuideCards"
        :key="card.id"
        class="ppmx-guide-card reveal"
        :class="`is-${card.tone}`"
        role="button"
        tabindex="0"
        @click="openGuideCard(card.id)"
        @keydown.enter.prevent="openGuideCard(card.id)"
        @keydown.space.prevent="openGuideCard(card.id)"
      >
        <span class="ppmx-guide-card__icon" aria-hidden="true">
          <i class="fa-solid" :class="card.icon" />
        </span>
        <span class="ppmx-guide-card__body">
          <strong>{{ pickText(card.title) }}</strong>
          <p>{{ pickText(card.description) }}</p>
          <small>
            <i class="fa-solid" :class="card.tagIcon" />
            {{ pickText(card.tag) }}
          </small>
        </span>
      </article>
    </section>

    <p class="ppmx-guide-foot reveal">{{ pickText(ppmxGuideText.foot) }}</p>
  </PpmxSubpageLayout>
</template>
