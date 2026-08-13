import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { expect, test, type Page, type Route } from '@playwright/test'

const authPayload = JSON.stringify({
  value: 'home-recharge-copy-e2e-token',
  expiresAt: Date.now() + 60 * 60 * 1000,
})

const localeCases = [
  { locale: 'es-PE', copy: 'La primera recarga recibe un 8% de reembolso, la segunda un 9%, la tercera un 10% y desde la cuarta un 18%. Cuantas más recargas hagas, mayor será el reembolso.' },
  { locale: 'en-US', copy: 'The first deposit gets an 8% rebate, the second 9%, the third 10%, and the fourth and subsequent deposits 18%. The more times you deposit, the higher your rebate.' },
  { locale: 'zh-CN', copy: '首次充值享8%返利，二次充值享9%返利，三次充值享10%返利，第四次及以上充值享18%返利。充值的次数越多享受的返利越高！' },
] as const

async function fulfillJson(route: Route, data: unknown) {
  await route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify({ code: 0, data }),
  })
}

async function mockHomeApis(page: Page) {
  await page.route(/\/(?:api\/)?v1\/app\/tg\/currentUserInfo$/, route => fulfillJson(route, {
    id: 1001,
    username: 'home_recharge_copy_user',
    country: 'PE',
    balance: 1000,
  }))
  await page.route(/\/(?:api\/)?v1\/app\/tg\/deviceInfo$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/pendingNotifications$/, route => fulfillJson(route, []))
  await page.route(/\/(?:api\/)?v1\/app\/appGame\/home$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/attribution\/event$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/banners$/, route => fulfillJson(route, {}))
}

test('shows the 8%, 9%, 10%, and 18% deposit rebate tiers in every locale', async ({ page }, testInfo) => {
  const runtimeErrors: string[] = []
  page.on('console', (message) => {
    if (message.type() === 'error') runtimeErrors.push(message.text())
  })
  page.on('pageerror', error => runtimeErrors.push(error.message))

  await mockHomeApis(page)
  await page.addInitScript((tokenPayload) => {
    class NoopWebSocket extends EventTarget {
      static CONNECTING = 0
      static OPEN = 1
      static CLOSING = 2
      static CLOSED = 3
      readyState = NoopWebSocket.CLOSED
      close() {}
      send() {}
    }

    Object.defineProperty(window, 'WebSocket', {
      configurable: true,
      writable: true,
      value: NoopWebSocket,
    })
    window.localStorage.setItem('h5_token', tokenPayload)
  }, authPayload)

  await page.goto('/#/')

  for (const { locale, copy } of localeCases) {
    await page.evaluate((nextLocale) => {
      window.localStorage.setItem('locale', JSON.stringify({ value: nextLocale, expiresAt: null }))
    }, locale)
    await page.reload()
    await expect(page).toHaveTitle(/PP\.PE/)
    await page.locator('.ppmx-hero-dot').nth(4).click()

    const activeSlide = page.locator('.ppmx-hero-slide.is-active')
    await expect(activeSlide).toHaveCSS('opacity', '1')
    await expect(activeSlide.locator('.ppmx-hero-sub')).toHaveText(copy)
    await expect(activeSlide).not.toContainText('50%')
    await expect(activeSlide).not.toContainText('100%')
  }

  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  expect(runtimeErrors).toEqual([])
  await page.screenshot({
    path: join(tmpdir(), `ppmx-home-recharge-copy-${testInfo.project.name}.png`),
    fullPage: false,
  })
})
