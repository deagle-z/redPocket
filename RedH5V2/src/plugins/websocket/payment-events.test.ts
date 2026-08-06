import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const readSource = (path: string) =>
  readFileSync(new URL(`../../../${path}`, import.meta.url), 'utf8')

describe('payment WebSocket integration contract', () => {
  it('adds WebSocket plugin entrypoints and environment configuration', () => {
    const indexSource = readSource('src/plugins/websocket/index.ts')
    const envSource = readSource('env.d.ts')
    const envDevelopment = readSource('.env.development')
    const envStaging = readSource('.env.staging')
    const envProduction = readSource('.env.production')

    expect(indexSource).toContain('connectWebSocket')
    expect(indexSource).toContain('closeWebSocket')
    expect(indexSource).toContain('VITE_WS_URL')
    expect(indexSource).toContain('VITE_APP_WS_URL')
    expect(indexSource).toContain('VITE_APP_WS_UID')
    expect(envSource).toContain('VITE_WS_URL?: string')
    expect(envSource).toContain('VITE_APP_WS_URL?: string')
    expect(envSource).toContain('VITE_APP_WS_UID?: string')
    expect(envDevelopment).toContain('VITE_WS_URL=')
    expect(envStaging).toContain('VITE_WS_URL=')
    expect(envProduction).toContain('VITE_WS_URL=')
  })

  it('wires recharge_success into the root app lifecycle with pending notification compensation', () => {
    const appSource = readSource('src/App.vue')

    expect(appSource).toContain('wsClient')
    expect(appSource).toContain('connectWebSocket')
    expect(appSource).toContain('closeWebSocket')
    expect(appSource).toContain("wsClient.on('recharge_success'")
    expect(appSource).toContain('wsClient.onOpen')
    expect(appSource).toContain('wsClient.onSyncExpired')
    expect(appSource).toContain('getPendingRechargeNotifications')
    expect(appSource).toContain('ackRechargeNotification')
    expect(appSource).toContain('processedRechargeOrders')
    expect(appSource).toContain('userStore.loadUserInfo')
    expect(appSource).toContain('trackRechargeSuccess')
    expect(appSource).toContain('PpmxPaymentSuccessNotice')
  })

  it('wires prize_pool_balance into the realtime prize pool store and wheel jackpot display', () => {
    const appSource = readSource('src/App.vue')
    const wheelSource = readSource('src/pages/wheel/index.vue')
    const storeSource = readSource('src/stores/modules/prizePool.ts')

    expect(appSource).toContain("wsClient.on('prize_pool_balance'")
    expect(appSource).toContain('handlePrizePoolBalanceMessage')
    expect(appSource).toContain('updatePrizePoolBalance')
    expect(storeSource).toContain('usePrizePoolStore')
    expect(storeSource).toContain('updatePrizePoolBalance')
    expect(storeSource).toContain('luckyBalance')
    expect(wheelSource).toContain('usePrizePoolStore')
    expect(wheelSource).toContain('jackpotRealtimeBalance')
  })

  it('adds Purchase pixel tracking and localized payment success copy', () => {
    const pixelSource = readSource('src/utils/facebookPixel.ts')
    const zhSource = readSource('src/locales/zh-CN.ts')
    const enSource = readSource('src/locales/en-US.ts')
    const esSource = readSource('src/locales/es-PE.ts')

    expect(pixelSource).toContain('trackPurchase')
    expect(pixelSource).toContain("'Purchase'")
    expect(pixelSource).toContain('orderNo')
    expect(zhSource).toContain('paymentSuccessTitle')
    expect(zhSource).toContain('paymentSuccessMessage')
    expect(enSource).toContain('paymentSuccessTitle')
    expect(enSource).toContain('paymentSuccessMessage')
    expect(esSource).toContain('paymentSuccessTitle')
    expect(esSource).toContain('paymentSuccessMessage')
  })
})
