<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showToast } from 'vant'
import { ppmxPromoItems } from '../ppmx-home/pageData'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'
import PpmxPromoDetail from '../ppmx-home/components/PpmxPromoDetail.vue'
import PpmxPromoCodeModal from '@/components/PpmxPromoCodeModal.vue'

definePage({
  name: 'promo-detail',
  meta: {
    titleKey: 'ppmx.promo.title',
    shell: 'ppmx',
    tabbar: false,
  },
})

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const pageChromeRef = ref<InstanceType<typeof PpmxPageChrome> | null>(null)
const selectedPromoId = computed(() => {
  const params = route.params as { id?: string | string[] }
  const param = params.id

  return Array.isArray(param) ? param[0] : param
})
const selectedPromo = computed(() => (
  ppmxPromoItems.find((promo) => promo.id === selectedPromoId.value) ?? null
))
const promoCodeModalOpen = ref(false)

function requireRegisterAuth() {
  return pageChromeRef.value?.requireRegisterAuth() ?? false
}

function clearPromoDetail() {
  void router.push('/promo')
}

function claimPromo() {
  if (!requireRegisterAuth()) return

  if (selectedPromoId.value === 'checkin') {
    void router.push('/?action=checkin')
    return
  }

  if (selectedPromoId.value === 'vip') {
    void router.push('/vip')
    return
  }

  if (selectedPromoId.value === 'registro' || selectedPromoId.value === 'register') {
    promoCodeModalOpen.value = true
    return
  }

  if (selectedPromoId.value === 'invita') {
    void router.push('/team')
    return
  }

  showToast(t('ppmx.promo.activationSoon'))
}

watchEffect(() => {
  if (!selectedPromo.value) void router.replace('/promo')
})
</script>

<template>
  <main id="page-promo" class="ppmx-page ppmx-subpage ppmx-promo-page">
    <PpmxPageChrome ref="pageChromeRef">
      <section class="ppmx-subpage__wide ppmx-promo-stack">
        <PpmxPromoDetail
          v-if="selectedPromo"
          :selected-promo="selectedPromo"
          @back="clearPromoDetail"
          @claim="claimPromo"
        />
      </section>
    </PpmxPageChrome>

    <PpmxPromoCodeModal v-model="promoCodeModalOpen" />
  </main>
</template>
