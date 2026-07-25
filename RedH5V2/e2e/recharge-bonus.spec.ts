import { expect, test, type Page, type Route } from '@playwright/test'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const authPayload = JSON.stringify({
  value: 'recharge-bonus-e2e-token',
  expiresAt: Date.now() + 60 * 60 * 1000,
})

async function fulfillJson(route: Route, data: unknown) {
  await route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify({ code: 0, data }),
  })
}

async function mockRechargeApis(page: Page) {
  await page.route(/\/(?:api\/)?v1\/app\/tg\/currentUserInfo$/, route => fulfillJson(route, {
    id: 1001,
    username: 'bonus_e2e_user',
    country: 'MX',
    balance: 250000,
    sportBalance: 100000,
    hasWithdrawAccount: true,
  }))

  await page.route(/\/(?:api\/)?v1\/app\/countries$/, route => fulfillJson(route, [
    {
      id: 1,
      countryCode: 'MX',
      countryNameEn: 'Mexico',
      countryNameCn: '墨西哥',
      currencyCode: 'MXN',
      currencySymbol: '$',
      sort: 1,
    },
  ]))

  await page.route(/\/(?:api\/)?v1\/app\/country\/MX\/recharge$/, route => fulfillJson(route, {
    rechargeFields: [],
    channels: [
      {
        id: 1,
        channelCode: 'BANK_MX',
        channelName: 'Transferencia bancaria',
        providerType: 'bank',
        icon: null,
        sort: 1,
        methods: [
          {
            id: 1,
            methodCode: 'SPEI',
            methodName: 'SPEI',
            icon: null,
            sort: 1,
          },
        ],
      },
    ],
  }))

  await page.route(/\/(?:api\/)?v1\/app\/country\/MX\/rechargeFields$/, route => fulfillJson(route, []))
  await page.route(/\/(?:api\/)?v1\/app\/recharge\/isFirst\/v2$/, route => fulfillJson(route, {
    hasFirst: false,
    isFirstRecharge: true,
    rechargeNumber: 1,
    giftRates: [18, 20, 25, 30],
    giftMinAmount: 50,
  }))
  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/pendingNotifications$/, route => fulfillJson(route, []))
}

test('shows the API gift percentage for balance deposits and hides it for sports deposits', async ({ page }, testInfo) => {
  const runtimeErrors: string[] = []
  page.on('console', (message) => {
    if (message.type() === 'error') runtimeErrors.push(message.text())
  })
  page.on('pageerror', error => runtimeErrors.push(error.message))

  await mockRechargeApis(page)
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

  await page.goto('/#/recharge')

  const walletSelector = page.locator('.ppmx-recharge-section--wallet .ppmx-select-input__button')
  const promotion = page.locator('.ppmx-recharge-promo')

  await expect(walletSelector).toContainText('E-wallet')
  await expect(promotion).toBeVisible()
  await expect(promotion).toContainText('18%')
  await expect(promotion).toContainText('$9.00')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  const overflowingAmountButtons = await page.evaluate(() => Array.from(document.querySelectorAll('.ppmx-recharge-amount'))
    .filter(button => button.scrollWidth > button.clientWidth)
    .map(button => ({
      text: button.textContent?.trim(),
      clientWidth: button.clientWidth,
      scrollWidth: button.scrollWidth,
    })))
  expect(overflowingAmountButtons).toEqual([])

  await promotion.scrollIntoViewIfNeeded()
  expect(await page.evaluate(() => {
    const promotionElement = document.querySelector('.ppmx-recharge-promo')
    const submitElement = document.querySelector('.ppmx-recharge-submit')
    if (!promotionElement || !submitElement) return false

    const promotionRect = promotionElement.getBoundingClientRect()
    const submitRect = submitElement.getBoundingClientRect()
    return promotionRect.bottom <= submitRect.top || submitRect.bottom <= promotionRect.top
  })).toBe(true)
  await page.screenshot({
    path: join(tmpdir(), `redh5v2-recharge-bonus-${testInfo.project.name}.png`),
    fullPage: false,
  })

  await walletSelector.click()
  await page.getByRole('button', { name: 'Sports wallet', exact: true }).click()

  await expect(walletSelector).toContainText('Sports wallet')
  await expect(promotion).toBeHidden()
  expect(runtimeErrors).toEqual([])
})
