import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { expect, test, type Page, type Route } from '@playwright/test'

const authPayload = JSON.stringify({
  value: 'social-card-e2e-token',
  expiresAt: Date.now() + 60 * 60 * 1000,
})

async function fulfillJson(route: Route, data: unknown) {
  await route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify({ code: 0, data }),
  })
}

async function mockHomeApis(page: Page) {
  await page.route(/\/(?:api\/)?v1\/app\/tg\/currentUserInfo$/, route => fulfillJson(route, {
    id: 1001,
    username: 'social_card_e2e_user',
    country: 'PE',
    balance: 1000,
  }))
  await page.route(/\/(?:api\/)?v1\/app\/tg\/deviceInfo$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/pendingNotifications$/, route => fulfillJson(route, []))
  await page.route(/\/(?:api\/)?v1\/app\/appGame\/home$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/attribution\/event$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/banners$/, route => fulfillJson(route, {}))
}

test('keeps social cards visible without navigation behavior', async ({ page }, testInfo) => {
  const consoleErrors: string[] = []
  let openedPage = false

  page.on('console', (message) => {
    if (message.type() === 'error') consoleErrors.push(message.text())
  })
  page.context().on('page', () => {
    openedPage = true
  })

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
  await expect(page).toHaveTitle(/PP\.PE/)
  await page.locator('.ppmx-float-widget--social .ppmx-float-inner').click()

  const modal = page.locator('#socialModal')
  const cards = modal.locator('.ppmx-social-card')
  await expect(modal).toBeVisible()
  await expect(cards).toHaveCount(4)
  await expect(cards.first()).toHaveJSProperty('tagName', 'ARTICLE')
  await expect(cards.first()).toHaveCSS('cursor', 'default')

  const urlBeforeClick = page.url()
  await cards.first().click()
  await expect(modal).toBeVisible()
  expect(page.url()).toBe(urlBeforeClick)
  expect(openedPage).toBe(false)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  expect(consoleErrors).toEqual([])

  await page.screenshot({
    path: join(tmpdir(), `ppmx-social-card-${testInfo.project.name}.png`),
    fullPage: false,
  })
})
