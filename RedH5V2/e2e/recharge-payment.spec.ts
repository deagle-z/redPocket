import { expect, test, type Page, type Route } from '@playwright/test'

const authPayload = JSON.stringify({
  value: 'e2e-token',
  expiresAt: Date.now() + 60 * 60 * 1000,
})

interface MockCryptoPayment {
  address: string
  expectedAmount: string
  network: string
  platformRate: string
  qrContent: string
  token: string
}

async function fulfillJson(route: Route, data: unknown) {
  await route.fulfill({
    contentType: 'application/json',
    body: JSON.stringify({ code: 0, data }),
  })
}

async function mockRechargeApis(page: Page, cryptoPlatformAmount?: number) {
  let activeCryptoPayment: MockCryptoPayment | null = null
  let cryptoStatusRequestCount = 0

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
    const quotedAmount = cryptoPlatformAmount ?? amount

    return fulfillJson(route, {
      platformAmount: quotedAmount,
      options: [
        {
          network: 'TRC20',
          token: 'USDT',
          contractAddress: 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t',
          receiveAddress: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
          estimatedAmount: quotedAmount.toFixed(2),
          platformRate: '1.000000',
          platformAmount: quotedAmount,
        },
        {
          network: 'BITCOIN',
          token: 'BTC',
          contractAddress: '',
          receiveAddress: 'bc1qexamplemerchantaddress000000000000000',
          estimatedAmount: '0.00100000',
          platformRate: '100000.00',
          platformAmount: quotedAmount,
        },
        {
          network: 'ETHEREUM',
          token: 'ETH',
          contractAddress: '',
          receiveAddress: '0x1111111111111111111111111111111111111111',
          estimatedAmount: '0.050000000000000000',
          platformRate: '2000.00',
          platformAmount: quotedAmount,
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
    const token = String(body.token)
    const network = String(body.network)
    const paymentByToken = {
      USDT: {
        network: 'TRC20',
        address: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
        expectedAmount: amount.toFixed(6),
        platformRate: '1.000000',
        qrContent: 'TEXAMPLEUSDTTRC20ADDRESS000000000',
      },
      BTC: {
        network: 'BITCOIN',
        address: 'bc1qexamplemerchantaddress000000000000000',
        expectedAmount: '0.00100042',
        platformRate: '100000.00',
        qrContent: 'bitcoin:bc1qexamplemerchantaddress000000000000000?amount=0.00100042',
      },
      ETH: {
        network: 'ETHEREUM',
        address: '0x1111111111111111111111111111111111111111',
        expectedAmount: '0.050000000000000042',
        platformRate: '2000.00',
        qrContent: 'ethereum:0x1111111111111111111111111111111111111111@1?value=50000000000000042',
      },
    } as const
    const payment = paymentByToken[token as keyof typeof paymentByToken]
    if (!payment || payment.network !== network) {
      throw new Error(`Unsupported crypto test asset: ${token}/${network}`)
    }
    activeCryptoPayment = { ...payment, network, token }

    await fulfillJson(route, {
      orderNo: 'E2E-CRYPTO-001',
      amount,
      currency: token,
      channel: token,
      payMethod: token,
      status: 0,
      bonusAmount: 0,
      devCallback: false,
      needConfirmUnfinishedActivityCycle: false,
      cryptoPayment: {
        network,
        token,
        contractAddress: token === 'USDT' ? 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t' : '',
        receiveAddress: payment.address,
        expectedAmount: payment.expectedAmount,
        expireTime: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
        qrContent: payment.qrContent,
        platformRate: payment.platformRate,
      },
    })
  })

  await page.route(/\/(?:api\/)?v1\/app\/rechargeOrder\/crypto\/[^/]+\/status(?:\?.*)?$/, route => {
    cryptoStatusRequestCount += 1
    if (!activeCryptoPayment) throw new Error('Status requested before a crypto payment order was created.')

    const expireTime = new Date(Date.now() + 15 * 60 * 1000).toISOString()
    return fulfillJson(route, {
      orderNo: 'E2E-CRYPTO-001',
      status: 0,
      rechargeStatus: 0,
      expireTime,
      remainingSeconds: 900,
      cryptoPayment: {
        network: activeCryptoPayment.network,
        token: activeCryptoPayment.token,
        contractAddress: activeCryptoPayment.token === 'USDT' ? 'TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t' : '',
        receiveAddress: activeCryptoPayment.address,
        expectedAmount: activeCryptoPayment.expectedAmount,
        expireTime,
        qrContent: activeCryptoPayment.qrContent,
        platformRate: activeCryptoPayment.platformRate,
      },
    })
  })

  return {
    getCryptoStatusRequestCount: () => cryptoStatusRequestCount,
  }
}

async function openRecharge(page: Page, cryptoPlatformAmount?: number) {
  const apiControls = await mockRechargeApis(page, cryptoPlatformAmount)
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

  return apiControls
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
    await expect(page.locator('.ppmx-recharge-amount-grid .ppmx-recharge-amount')).toHaveCount(8)
    await expect(page.locator('.ppmx-recharge-amount-grid .ppmx-recharge-amount').first()).toContainText('$50.00')
    await expect(page.locator('.ppmx-recharge-amount-grid')).not.toContainText(/Custom|自定义|Personalizado/)

    const requestPromise = page.waitForRequest(/\/(?:api\/)?v1\/app\/rechargeOrder\/v2$/)
    await page.getByTestId('recharge-submit').click()
    const request = await requestPromise
    const payload = request.postDataJSON() as Record<string, unknown>

    expect(payload.channel).toBe('HOPOPAY')
    expect(payload.payMethod).toBe('cashapp')
    expect(payload.amount).toBe(49.99)
    await expect(page.locator('#paymentModal')).toBeVisible()
    await expect(page.locator('#paymentModal')).toContainText('E2E-RECHARGE-001')
  })

  test('keeps cryptocurrency as the crypto-pay redirect flow', async ({ page }) => {
    const apiControls = await openRecharge(page)

    await page.getByTestId('pay-channel-crypto').click()
    await page.getByTestId('recharge-submit').click()

    await expect(page).toHaveURL(/#\/crypto-pay/)
    expect(page.url()).toContain('amount=49.99')
    expect(page.url()).toContain('channelCode=USDT_TRC20')

    await expect(page.locator('.ppmx-crypto-pay-option')).toHaveCount(3)
    const bitcoinOption = page.locator('.ppmx-crypto-pay-option').filter({ hasText: 'BTC' })
    await bitcoinOption.click()

    const requestPromise = page.waitForRequest(/\/(?:api\/)?v1\/app\/rechargeOrder\/crypto$/)
    await page.getByRole('button', { name: /Create payment order|生成支付订单|Crear orden de pago/ }).click()
    const request = await requestPromise
    const payload = request.postDataJSON() as Record<string, unknown>

    expect(payload.network).toBe('BITCOIN')
    expect(payload.token).toBe('BTC')
    await expect(page.locator('.ppmx-crypto-pay-fields input').nth(0)).toHaveValue('0.00100042 BTC')
    await expect(page.locator('.ppmx-crypto-pay-fields input').nth(1)).toHaveValue('bc1qexamplemerchantaddress000000000000000')
    await expect(page.locator('.ppmx-crypto-pay-qr img')).toBeVisible()

    const persistedOrder = await page.evaluate(() => window.localStorage.getItem('h5_crypto_pending_order_v1:1001'))
    expect(persistedOrder).toContain('E2E-CRYPTO-001')
    expect(persistedOrder).toContain('bitcoin:bc1qexamplemerchantaddress000000000000000?amount=0.00100042')

    await page.getByRole('button', { name: /Cancel payment|取消支付|Cancelar pago/ }).click()
    const cancelDialog = page.locator('.van-dialog')
    await expect(cancelDialog).toContainText(/manual adjustment|人工补单|ajuste manual/)
    await cancelDialog.getByRole('button', { name: /Keep waiting|继续等待|Seguir esperando/ }).click()

    const statusRequestsBeforePageShow = apiControls.getCryptoStatusRequestCount()
    await page.evaluate(() => window.dispatchEvent(new Event('pageshow')))
    await expect.poll(() => apiControls.getCryptoStatusRequestCount()).toBeGreaterThan(statusRequestsBeforePageShow)

    await page.reload()
    await expect(page.locator('.ppmx-crypto-pay-fields input').nth(0)).toHaveValue('0.00100042 BTC')
    await expect(page.locator('.ppmx-crypto-pay-qr img')).toBeVisible()
  })

  test('blocks order creation when the server quote changes the platform amount', async ({ page }) => {
    await openRecharge(page, 49)

    await page.getByTestId('pay-channel-crypto').click()
    await page.getByTestId('recharge-submit').click()

    await expect(page).toHaveURL(/#\/crypto-pay/)
    await expect(page.locator('.ppmx-crypto-pay-state')).toContainText('49.99')
    await expect(page.locator('.ppmx-crypto-pay-state')).toContainText('49.00')
    await expect(page.getByRole('button', { name: /Create payment order|生成支付订单|Crear orden de pago/ })).toHaveCount(0)
  })
})
