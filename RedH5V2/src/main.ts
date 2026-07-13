import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import { i18n } from './plugins/i18n'
import { initRuntimeConfig } from './utils/runtime'
import { initTracker } from './utils/tracker'
import { initIOSKeyboardViewport } from './utils/iosKeyboardViewport'
import { useAppStore } from './stores/app'
import './assets/styles/main.css'
import './pwa'

function removeInitialLoading() {
  const loader = document.getElementById('app-loading')
  if (!loader) return

  window.requestAnimationFrame(() => {
    loader.classList.add('is-leaving')
    window.setTimeout(() => loader.remove(), 420)
  })
}

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)
useAppStore(pinia).initTheme()
const cleanupIOSKeyboardViewport = initIOSKeyboardViewport()
app.mount('#app')
removeInitialLoading()

initTracker()
void initRuntimeConfig()

if (import.meta.hot) {
  import.meta.hot.dispose(() => cleanupIOSKeyboardViewport())
}
