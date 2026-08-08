<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import QRCode from 'qrcode'
import { showToast } from 'vant'
import PpmxSubpageLayout from '@/components/PpmxSubpageLayout.vue'
import {
  ppmxDownloadFeatures,
  ppmxDownloadSteps,
  ppmxDownloadText,
} from '../ppmx-home/pageData'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import { downloadIosWebClipProfile } from '@/utils/iosWebClip'

definePage({
  name: 'download',
  meta: {
    titleKey: 'ppmx.download.title',
    shell: 'ppmx',
    tabbar: false,
  },
})

const { pickText } = usePpmxLocale()
const PPMX_ANDROID_APK_URL = 'https://pub-93b0b439f98b49c4ba1db81844583907.r2.dev/pepp.apk'
const downloadQrDataUrl = ref('')
const downloadQrError = ref(false)

function downloadApp() {
  window.location.href = PPMX_ANDROID_APK_URL
}

async function renderDownloadQr() {
  downloadQrError.value = false

  try {
    downloadQrDataUrl.value = await QRCode.toDataURL(PPMX_ANDROID_APK_URL, {
      errorCorrectionLevel: 'M',
      margin: 2,
      scale: 8,
      color: {
        dark: '#050507',
        light: '#ffffff',
      },
    })
  } catch {
    downloadQrDataUrl.value = ''
    downloadQrError.value = true
  }
}

async function downloadIosProfile() {
  try {
    await downloadIosWebClipProfile('ppmx-ios.mobileconfig')
  } catch {
    showToast(pickText(ppmxDownloadText.soon))
  }
}

onMounted(() => {
  void renderDownloadQr()
})
</script>

<template>
  <PpmxSubpageLayout
    page-class="ppmx-download-page"
    stack-class="ppmx-download-stack"
    head-class="ppmx-download-head reveal"
    :eyebrow="pickText(ppmxDownloadText.eyebrow)"
  >
    <section class="ppmx-download-hero reveal">
          <div class="ppmx-download-hero__content">
            <span class="ppmx-download-badge" aria-hidden="true">
              <i class="fa-solid fa-mobile-screen-button" />
            </span>
            <h1 class="ppmx-display">{{ pickText(ppmxDownloadText.title) }}</h1>
            <p>{{ pickText(ppmxDownloadText.sub) }}</p>
            <div class="ppmx-download-actions">
              <button class="ppmx-download-button" type="button" @click="downloadIosProfile">
                <i class="fa-brands fa-apple" />
                <span>ios</span>
              </button>
              <button class="ppmx-download-button" type="button" @click="downloadApp">
                <i class="fa-brands fa-android" />
                <span>Android</span>
              </button>
            </div>
          </div>

          <aside class="ppmx-download-qr" :aria-label="pickText(ppmxDownloadText.qr)">
            <span class="ppmx-download-qr__box" :class="{ 'is-disabled': !downloadQrDataUrl }">
              <img
                v-if="downloadQrDataUrl"
                :src="downloadQrDataUrl"
                :alt="pickText(ppmxDownloadText.qr)"
              >
              <i v-else-if="downloadQrError" class="fa-solid fa-triangle-exclamation" aria-hidden="true" />
              <i v-else class="fa-solid fa-qrcode" aria-hidden="true" />
            </span>
            <small>{{ pickText(ppmxDownloadText.qr) }}</small>
          </aside>
    </section>

    <section class="ppmx-download-panel reveal">
          <h2>{{ pickText(ppmxDownloadText.featuresTitle) }}</h2>
          <ul class="ppmx-download-features">
            <li v-for="feature in ppmxDownloadFeatures" :key="feature.id">
              <span class="ppmx-download-feature-icon" aria-hidden="true">
                <i class="fa-solid" :class="feature.icon" />
              </span>
              <span>
                <strong>{{ pickText(feature.title) }}</strong>
                <small>{{ pickText(feature.description) }}</small>
              </span>
            </li>
          </ul>
    </section>

    <section class="ppmx-download-panel reveal">
          <h2>{{ pickText(ppmxDownloadText.stepsTitle) }}</h2>
          <ol class="ppmx-download-steps">
            <li v-for="step in ppmxDownloadSteps" :key="step.id">
              <strong>{{ pickText(step.title) }}</strong>
              <small>{{ pickText(step.description) }}</small>
            </li>
          </ol>
    </section>
  </PpmxSubpageLayout>
</template>
