import { expect, test, type Page, type Route } from '@playwright/test'

const authPayload = JSON.stringify({
  value: 'e2e-token',
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
    username: 'e2e_user',
    phone: '+12025550123',
    country: 'US',
    balance: 250000,
    sportBalance: 100000,
    hasWithdrawAccount: true,
  }))

  await page.route(/\/(?:api\/)?v1\/app\/countries$/, route => fulfillJson(route, [
    {
      id: 1,
      countryCode: 'US',
      countryName: 'United States',
      countryNameEn: 'United States',
      countryNameCn: '美国',
      currencyCode: 'USD',
      currencySymbol: '$',
      sort: 1,
    },
  ]))

  await page.route(/\/(?:api\/)?v1\/app\/country\/US\/recharge$/, route => fulfillJson(route, {
    rechargeFields: [],
    channels: [],
  }))

  await page.route(/\/(?:api\/)?v1\/app\/country\/US\/rechargeFields$/, route => fulfillJson(route, []))

  await page.route(/\/(?:api\/)?v1\/app\/recharge\/isFirst\/v2$/, route => fulfillJson(route, {
    isFirstRecharge: true,
    rechargeNumber: 1,
    giftRates: [0.25, 0.15, 0.1],
    giftMinAmount: 50,
  }))

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/pendingNotifications$/, route => fulfillJson(route, []))

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/v2$/, async route => {
    const body = route.request().postDataJSON() as Record<string, unknown>

    await fulfillJson(route, {
      orderNo: 'E2E-RECHARGE-001',
      amount: body.amount,
      currency: 'USD',
      channel: body.channel,
      payMethod: body.payMethod || body.channel,
      payUrl: 'https://pay.example.test/order/E2E-RECHARGE-001',
      status: 0,
      bonusAmount: 0,
      devCallback: false,
      needConfirmUnfinishedActivityCycle: false,
    })
  })

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/crypto\/options(?:\?.*)?$/, route => {
    const requestUrl = new URL(route.request().url())
    const amount = Number(requestUrl.searchParams.get('amount'))
    if (!Number.isFinite(amount)) {
      throw new Error('Crypto recharge options request must include a numeric amount.')
    }

    return fulfillJson(route, {
      platformAmount: amount,
      options: [
        {
          network: 'TRC20',
          token: 'USDT',
          contractAddress: 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',
          receiveAddress: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
          estimatedAmount: amount.toFixed(2),
          platformRate: '1.000000',
          platformAmount: amount,
        },
      ],
    })
  })

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/crypto$/, async route => {
    const body = route.request().postDataJSON() as Record<string, unknown>
    const amount = Number(body.amount)
    if (!Number.isFinite(amount)) {
      throw new Error('Crypto recharge order request must include a numeric amount.')
    }

    await fulfillJson(route, {
      orderNo: 'E2E-CRYPTO-001',
      amount,
      currency: 'USD',
      channel: 'USDT_TRC20',
      payMethod: 'USDT_TRC20',
      status: 0,
      bonusAmount: 0,
      devCallback: false,
      needConfirmUnfinishedActivityCycle: false,
      cryptoPayment: {
        network: 'TRC20',
        token: 'USDT',
        contractAddress: 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',
        receiveAddress: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
        expectedAmount: amount.toFixed(2),
        expireTime: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
        qrContent: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
        platformRate: '1.000000',
      },
    })
  })

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/crypto\/[^/]+\/status(?:\?.*)?$/, route => fulfillJson(route, {
    orderNo: 'E2E-CRYPTO-001',
    status: 0,
    rechargeStatus: 0,
    expireTime: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
    remainingSeconds: 900,
    cryptoPayment: {
      network: 'TRC20',
      token: 'USDT',
      contractAddress: 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',
      receiveAddress: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
      expectedAmount: '100.00',
      expireTime: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
      qrContent: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
      platformRate: '1.000000',
    },
  }))
}

async function openRecharge(page: Page) {
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
}

test.describe('recharge payment channels', () => {
  test('renders fixed payment channels and opens v2 payment popup', async ({ page }) => {
    await openRecharge(page)

    await expect(page.getByTestId('pay-channel-cash')).toBeVisible()
    await expect(page.getByTestId('pay-channel-apple')).toBeVisible()
    await expect(page.getByTestId('pay-channel-google')).toBeVisible()
    await expect(page.getByTestId('pay-channel-crypto')).toBeVisible()
    await expect(page.getByTestId('pay-channel-cash')).toContainText('Cash')
    await expect(page.getByTestId('pay-channel-apple')).toContainText('Apple Pay')
    await expect(page.getByTestId('pay-channel-google')).toContainText('Google Pay')
    await expect(page.getByTestId('pay-channel-crypto')).toContainText(/Crypto|加密货币/)
    await expect(page.locator('.ppmx-recharge-amount-grid .ppmx-recharge-amount')).toHaveCount(10)
    await expect(page.locator('.ppmx-recharge-amount-grid .ppmx-recharge-amount').first()).toContainText('S/30.00')
    await expect(page.locator('.ppmx-recharge-amount-grid')).not.toContainText(/Custom|自定义|Personalizado/)

    const requestPromise = page.waitForRequest(/\/(?:api\/)?v1\/app\/rechargeOrder\/v2$/)
    await page.getByTestId('recharge-submit').click()
    const request = await requestPromise
    const payload = request.postDataJSON() as Record<string, unknown>

    expect(payload.channel).toBe('HOPOPAY')
    expect(payload.payMethod).toBe('cashapp')
    expect(payload.amount).toBe(29.99)
    await expect(page.locator('#paymentModal')).toBeVisible()
    await expect(page.locator('#paymentModal')).toContainText('E2E-RECHARGE-001')
  })

  test('keeps cryptocurrency as the crypto-pay redirect flow', async ({ page }) => {
    await openRecharge(page)

    await page.getByTestId('pay-channel-crypto').click()
    await page.getByTestId('recharge-submit').click()

    await expect(page).toHaveURL(/#\/crypto-pay/)
    expect(page.url()).toContain('amount=29.99')
    expect(page.url()).toContain('channelCode=USDT_TRC20')
  })
})
