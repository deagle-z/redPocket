import { beforeEach, describe, expect, it, vi } from 'vitest'

import { post } from '@/utils/http'
import {
  bindCurrentTgEmail,
  bindCurrentTgPhone,
  checkRegisterPhone,
  forgotPasswordByPhone,
  registerByPhone,
  sendTgSmsCode,
} from './user'

vi.mock('@/utils/http', () => ({
  post: vi.fn(async () => ({ accessToken: 'registered-token' })),
}))

vi.mock('@/utils/sourceChannel', () => ({
  getSourceChannelCode: vi.fn(() => 'STORED'),
  normalizeSourceChannelCode: vi.fn((value?: string) => String(value || '').trim().toUpperCase()),
}))

describe('user API registration payloads', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('sends the browser device fingerprint with phone registration', async () => {
    await registerByPhone({
      phone: '12015550123',
      country: 'US',
      firstName: 'Player',
      password: 'password123',
      inviteCode: 'INVITE01',
      sourceChannelCode: 'fb01',
      deviceFingerprint: 'visitor-device-id',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/registerByPhone', {
      phone: '12015550123',
      country: 'US',
      firstName: 'Player',
      password: 'password123',
      inviteCode: 'INVITE01',
      sourceChannelCode: 'FB01',
      channelCode: 'FB01',
      deviceFingerprint: 'visitor-device-id',
    })
  })

  it('checks phone availability before registration continues', async () => {
    await checkRegisterPhone({
      phone: '+12015550123',
      country: 'US',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/checkRegisterPhone', {
      phone: '+12015550123',
      country: 'US',
    }, {
      meta: {
        auth: false,
        showError: false,
      },
    })
  })

  it('uses the RedPocketH5 phone password reset contract', async () => {
    await sendTgSmsCode('12015550123', 'US')

    expect(post).toHaveBeenCalledWith('/v1/app/tg/sendSMSCode', {
      phone: '12015550123',
      country: 'US',
    }, {
      meta: {
        auth: false,
        showError: false,
      },
    })

    await forgotPasswordByPhone({
      phone: '12015550123',
      country: 'US',
      code: '123456',
      newPassword: 'newpass123',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/forgotPasswordByPhone', {
      phone: '12015550123',
      country: 'US',
      code: '123456',
      newPassword: 'newpass123',
    }, {
      meta: {
        auth: false,
        showError: false,
      },
    })
  })

  it('uses the RedPocketH5 current user email binding contract', async () => {
    await bindCurrentTgEmail({ email: 'john.smith@example.com' })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/bindEmail', {
      email: 'john.smith@example.com',
    })
  })

  it('uses the RedPocketH5 current user phone binding contract', async () => {
    await bindCurrentTgPhone({
      phone: '12015550123',
      country: 'US',
      code: '123456',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/bindPhone', {
      phone: '12015550123',
      country: 'US',
      code: '123456',
    })
  })
})
