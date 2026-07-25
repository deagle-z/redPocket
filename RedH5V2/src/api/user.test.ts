import { beforeEach, describe, expect, it, vi } from 'vitest'

import { post } from '@/utils/http'
import {
  bindCurrentTgEmail,
  bindCurrentTgPhone,
  checkRegisterPhone,
  forgotPasswordByPhone,
  loginByPhone,
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

vi.mock('@/utils/deviceInfo', () => ({
  getDeviceInfo: vi.fn(async () => ({
    deviceFingerprint: 'visitor-device-id',
    devicePlatform: 'android',
    deviceModel: 'SM-S918B',
    deviceOS: 'Android',
    deviceOSVersion: '13',
  })),
}))

describe('user API registration payloads', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('sends the browser device fingerprint with phone registration', async () => {
    await registerByPhone({
      phone: '525512345678',
      country: 'MX',
      firstName: 'Player',
      password: 'password123',
      inviteCode: 'INVITE01',
      sourceChannelCode: 'fb01',
      deviceFingerprint: 'visitor-device-id',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/registerByPhone', {
      phone: '525512345678',
      country: 'MX',
      firstName: 'Player',
      password: 'password123',
      inviteCode: 'INVITE01',
      sourceChannelCode: 'FB01',
      channelCode: 'FB01',
      deviceFingerprint: 'visitor-device-id',
      devicePlatform: 'android',
      deviceModel: 'SM-S918B',
      deviceOS: 'Android',
      deviceOSVersion: '13',
    })
  })

  it('sends complete browser device information with phone login', async () => {
    await loginByPhone({
      phone: '525512345678',
      country: 'MX',
      password: 'password123',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/phoneLogin', {
      phone: '525512345678',
      country: 'MX',
      password: 'password123',
      deviceFingerprint: 'visitor-device-id',
      devicePlatform: 'android',
      deviceModel: 'SM-S918B',
      deviceOS: 'Android',
      deviceOSVersion: '13',
    })
  })

  it('checks phone availability before registration continues', async () => {
    await checkRegisterPhone({
      phone: '+525512345678',
      country: 'MX',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/checkRegisterPhone', {
      phone: '+525512345678',
      country: 'MX',
    }, {
      meta: {
        auth: false,
        showError: false,
      },
    })
  })

  it('uses the RedPocketH5 phone password reset contract', async () => {
    await sendTgSmsCode('525512345678', 'MX')

    expect(post).toHaveBeenCalledWith('/v1/app/tg/sendSMSCode', {
      phone: '525512345678',
      country: 'MX',
    }, {
      meta: {
        auth: false,
        showError: false,
      },
    })

    await forgotPasswordByPhone({
      phone: '525512345678',
      country: 'MX',
      code: '123456',
      newPassword: 'newpass123',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/forgotPasswordByPhone', {
      phone: '525512345678',
      country: 'MX',
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
      phone: '525512345678',
      country: 'MX',
      code: '123456',
    })

    expect(post).toHaveBeenCalledWith('/v1/app/tg/bindPhone', {
      phone: '525512345678',
      country: 'MX',
      code: '123456',
    })
  })
})
