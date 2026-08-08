import { expect, test, type Page, type Route } from '@playwright/test'

const authPayload = JSON.stringify({
  value: 'game-online-e2e-token',
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
    username: 'game_online_e2e_user',
    country: 'PE',
    balance: 1000,
  }))
  await page.route(/\/(?:api\/)?v1\/app\/tg\/deviceInfo$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/pendingNotifications$/, route => fulfillJson(route, []))
  await page.route(/\/(?:api\/)?v1\/app\/attribution\/event$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/banners$/, route => fulfillJson(route, {}))
  await page.route(/\/(?:api\/)?v1\/app\/appGame\/home$/, route => fulfillJson(route, {
    slots: [{
      gameId: 1001,
      gameName: 'Lucky Jaguar',
      categoryCode: 'slots',
      manufacturer: 'PGSOFT',
      gameIcon: '',
      horizontalImage: '',
      sort: 1,
      showIndex: 0,
      fakeOnlineCount: 137,
    }],
  }))
}

test('renders the API fake online count on a game tile', async ({ page }) => {
  const consoleErrors: string[] = []
  page.on('console', (message) => {
    if (message.type() === 'error') consoleErrors.push(message.text())
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
  await expect(page.getByText('Lucky Jaguar', { exact: true })).toBeVisible()

  const onlineCount = page.locator('.ppmx-htile-online')
  await expect(onlineCount).toHaveText('137')
  await expect(onlineCount).toHaveAttribute('aria-label', /137/)
  await onlineCount.hover()

  const [countBox, thumbBox] = await Promise.all([
    onlineCount.boundingBox(),
    onlineCount.locator('xpath=..').boundingBox(),
  ])
  expect(countBox).not.toBeNull()
  expect(thumbBox).not.toBeNull()
  if (!countBox || !thumbBox) throw new Error('Game online count layout is unavailable')
  expect(countBox.x).toBeGreaterThanOrEqual(thumbBox.x)
  expect(countBox.y).toBeGreaterThanOrEqual(thumbBox.y)
  expect(countBox.x + countBox.width).toBeLessThanOrEqual(thumbBox.x + thumbBox.width)
  expect(countBox.y + countBox.height).toBeLessThanOrEqual(thumbBox.y + thumbBox.height)

  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  expect(consoleErrors).toEqual([])
})
