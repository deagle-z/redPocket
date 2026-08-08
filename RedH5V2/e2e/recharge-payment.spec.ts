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
    country: 'PE',
    balance: 250000,
    sportBalance: 100000,
    hasWithdrawAccount: true,
  }))

  await page.route(/\/(?:api\/)?v1\/app\/tg\/deviceInfo$/, route => fulfillJson(route, null))
  await page.route(/\/(?:api\/)?v1\/app\/attribution\/event$/, route => fulfillJson(route, null))

  await page.route(/\/(?:api\/)?v1\/app\/countries$/, route => fulfillJson(route, [
    {
      id: 1,
      countryCode: 'PE',
      countryName: 'Peru',
      countryNameEn: 'Peru',
      countryNameCn: '秘鲁',
      currencyCode: 'PEN',
      currencySymbol: 'S/',
      sort: 1,
    },
  ]))

  await page.route(/\/(?:api\/)?v1\/app\/country\/PE\/recharge$/, route => fulfillJson(route, {
    rechargeFields: [],
    channels: [
      {
        id: 1,
        channelCode: 'HOPOPAY',
        channelName: 'Cash',
        providerType: 'cash',
        icon: null,
        sort: 1,
        methods: [
          {
            id: 1,
            methodCode: 'cashapp',
            methodName: 'Cash App',
            icon: null,
            sort: 1,
          },
        ],
      },
      {
        id: 2,
        channelCode: 'USDT_TRC20',
        channelName: 'Crypto',
        providerType: 'native',
        icon: null,
        sort: 2,
        methods: [],
      },
    ],
  }))

  await page.route(/\/(?:api\/)?v1\/app\/country\/PE\/rechargeFields$/, route => fulfillJson(route, []))

  await page.route(/\/(?:api\/)?v1\/app\/recharge\/isFirst\/v2$/, route => fulfillJson(route, {
    isFirstRecharge: true,
    rechargeNumber: 1,
    giftRates: [25, 15, 10, 10, 10, 10, 10, 10, 10, 10, 10],
    giftMinAmount: 50,
  }))

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/pendingNotifications$/, route => fulfillJson(route, []))

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/v2$/, async route => {
    const body = route.request().postDataJSON() as Record<string, unknown>

    await fulfillJson(route, {
      orderNo: 'E2E-RECHARGE-001',
      amount: body.amount,
      currency: 'PEN',
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

  await page.goto('/#/profile?recharge=open')
  await expect(page.locator('#rechargeModal')).toBeVisible()
}

test.describe('recharge payment channels', () => {
  test('renders the recharge identity fields in the selected language', async ({ page }) => {
    let rechargeFieldRequestCount = 0
    page.on('request', (request) => {
      if (new URL(request.url()).pathname.endsWith('/rechargeFields')) {
        rechargeFieldRequestCount += 1
      }
    })
    await page.addInitScript((localePayload) => {
      window.localStorage.setItem('locale', localePayload)
    }, JSON.stringify({ value: 'zh-CN', expiresAt: null }))

    await openRecharge(page)

    const fields = page.locator('.ppmx-recharge-section--fields')
    await expect(fields).toBeVisible()
    await expect(fields.locator('label').nth(0)).toContainText('证件类型')
    await expect(fields.locator('label').nth(1)).toContainText('证件号码')
    await expect(fields.locator('label').nth(2)).toContainText('证件姓名')
    await fields.locator('input').nth(0).fill('71234567')
    await fields.locator('input').nth(1).fill('JUAN PEREZ')
    await expect(fields).not.toContainText('Tipo de documento')
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
    expect(rechargeFieldRequestCount).toBe(0)
  })

  test('renders API payment channels and opens v2 payment popup', async ({ page }) => {
    await openRecharge(page)

    await expect(page.getByTestId('pay-channel-HOPOPAY')).toBeVisible()
    await expect(page.getByTestId('pay-channel-crypto')).toBeVisible()
    await expect(page.getByTestId('pay-channel-HOPOPAY')).toContainText('Cash')
    await expect(page.getByTestId('pay-channel-crypto')).toContainText(/Crypto|加密货币/)
    await expect(page.locator('.ppmx-recharge-amount-grid .ppmx-recharge-amount')).toHaveCount(7)
    await expect(page.locator('.ppmx-recharge-amount-grid .ppmx-recharge-amount').first()).toContainText('S/50.00')
    await expect(page.locator('.ppmx-recharge-amount-grid')).not.toContainText(/Custom|自定义|Personalizado/)
    await page.getByTestId('recharge-field-identity').locator('input').fill('71234567')
    await page.getByTestId('recharge-field-identityName').locator('input').fill('JUAN PEREZ')

    const requestPromise = page.waitForRequest(/\/(?:api\/)?v1\/app\/rechargeOrder\/v2$/)
    await page.getByTestId('recharge-submit').click()
    await expect(page.locator('.van-dialog')).toBeVisible()
    await page.locator('.van-dialog__confirm').click()
    const request = await requestPromise
    const payload = request.postDataJSON() as Record<string, unknown>

    expect(payload.channel).toBe('HOPOPAY')
    expect(payload.payMethod).toBe('cashapp')
    expect(payload.amount).toBe(50)
    await expect(page.locator('#paymentModal')).toBeVisible()
    await expect(page.locator('#paymentModal')).toContainText('E2E-RECHARGE-001')
  })

  test('keeps cryptocurrency as the crypto-pay redirect flow', async ({ page }) => {
    await openRecharge(page)

    await page.getByTestId('pay-channel-crypto').click()
    await page.getByTestId('recharge-submit').click()

    await expect(page).toHaveURL(/#\/crypto-pay/)
    expect(page.url()).toContain('amount=50.00')
    expect(page.url()).toContain('channelCode=USDT_TRC20')
  })
})
