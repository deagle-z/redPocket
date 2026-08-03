import { expect, test, type Page, type Route } from '@playwright/test'

const authPayload = JSON.stringify({
  value: 'announcement-e2e-token',
  expiresAt: Date.now() + 60 * 60 * 1000,
})

async function fulfillJson(route: Route, data: unknown) {
  await route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify({ code: 0, data }),
  })
}

async function mockHomeApis(page: Page) {
  let bannerRequestCount = 0

  await page.route(/\/(?:api\/)?v1\/app\/tg\/currentUserInfo$/, route => fulfillJson(route, {
    id: 1001,
    username: 'popup_e2e_user',
    country: 'MX',
    balance: 1000,
  }))
  await page.route(/\/(?:api\/)?v1\/app\/tg\/deviceInfo$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/pendingNotifications$/, route => fulfillJson(route, []))
  await page.route(/\/(?:api\/)?v1\/app\/appGame\/home$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/attribution\/event$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/banners$/, async (route) => {
    bannerRequestCount += 1
    await fulfillJson(route, {
      popup: [{
        id: 7001,
        tenantId: 1,
        bannerName: 'Platform announcement',
        bannerCode: null,
        position: 'popup',
        platform: 'h5',
        bannerType: 'popup',
        jumpType: 'none',
        displayType: 'popup',
        openMode: 'current',
        sort: 1,
        status: 1,
        startTime: null,
        endTime: null,
        languageCode: 'es-MX',
        countryCode: 'MX',
        title: 'Refresh announcement',
        subTitle: null,
        description: 'Shown once per platform entry.',
        buttonText: null,
        imageUrl: '',
        thumbUrl: null,
        bgImageUrl: null,
        iconUrl: null,
        videoUrl: null,
        jumpValue: null,
        textColor: null,
        buttonColor: null,
        bgColor: null,
      }],
    })
  })

  return () => bannerRequestCount
}

test('shows again after refresh but not after returning from another page', async ({ page }) => {
  const getBannerRequestCount = await mockHomeApis(page)

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
  await expect(page.getByRole('dialog', { name: 'Refresh announcement' })).toBeVisible()
  expect(getBannerRequestCount()).toBe(1)

  await page.locator('.ppmx-announcement__close').click()
  await expect(page.getByRole('dialog', { name: 'Refresh announcement' })).toBeHidden()

  await page.evaluate(() => {
    window.location.hash = '#/profile'
  })
  await expect(page).toHaveURL(/#\/profile$/)
  await page.evaluate(() => {
    window.location.hash = '#/'
  })
  await expect(page).toHaveURL(/#\/$/)
  await expect(page.getByRole('dialog', { name: 'Refresh announcement' })).toHaveCount(0)
  expect(getBannerRequestCount()).toBe(1)

  await page.reload()
  await expect(page.getByRole('dialog', { name: 'Refresh announcement' })).toBeVisible()
  expect(getBannerRequestCount()).toBe(2)
})
