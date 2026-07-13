<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { ppmxPromoItems } from '../ppmx-home/pageData'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import { highlightPromoText } from '../ppmx-home/utils/promoHighlight'
import type { PpmxPromoItem } from '../ppmx-home/types'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'

definePage({
  name: 'promo',
  meta: {
    titleKey: 'ppmx.promo.title',
    shell: 'ppmx',
    tabbar: false,
  },
})

const { t } = useI18n()
const { pickText } = usePpmxLocale()
const router = useRouter()

// Download-app promo activity is hidden from the promotions list.
const visiblePromoItems = computed(() => ppmxPromoItems.filter((promo) => promo.id !== 'download'))

function showPromoDetail(promo: PpmxPromoItem) {
  void router.push(`/promo/${promo.id}`)
}
</script>

<template>
  <main id="page-promo" class="ppmx-page ppmx-subpage ppmx-promo-page">
    <PpmxPageChrome>
      <section class="ppmx-subpage__wide ppmx-promo-stack">
        <section
          id="promoList"
          class="ppmx-promo-list"
        >
          <header class="ppmx-subpage-head ppmx-promo-head reveal">
            <div class="ppmx-eyebrow">{{ t('ppmx.promo.eyebrow') }}</div>
            <h1 class="ppmx-display">{{ t('ppmx.promo.title') }}</h1>
            <p>{{ t('ppmx.promo.description') }}</p>
          </header>

          <section id="promosGrid" class="ppmx-promo-grid" :aria-label="t('ppmx.promo.gridAria')">
            <article
              v-for="promo in visiblePromoItems"
              :key="promo.id"
              class="ppmx-promo-card reveal"
              :class="`ppmx-promo-card--${promo.tone}`"
              @click="showPromoDetail(promo)"
            >
              <span
                v-if="promo.backgroundImage"
                class="ppmx-promo-card__media"
                :style="{
                  backgroundImage: `url(${promo.backgroundImage})`,
                  backgroundPosition: promo.backgroundPosition,
                }"
              />
              <span class="ppmx-promo-tag">{{ pickText(promo.tag) }}</span>
              <h3 class="ppmx-display" v-html="highlightPromoText(pickText(promo.title))" />
              <p v-html="highlightPromoText(pickText(promo.sub))" />
              <small>
                <i class="fa-solid fa-circle-info" />
                <span>{{ pickText(promo.terms) }}</span>
              </small>
              <button type="button" @click.stop="showPromoDetail(promo)">
                {{ t('ppmx.promo.details') }}
                <i class="fa-solid fa-arrow-right" />
              </button>
            </article>
          </section>
        </section>
      </section>
    </PpmxPageChrome>
  </main>
</template>
