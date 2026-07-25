import { describe, expect, it } from 'vitest'

import { detectDeviceDetails } from './deviceInfo'

describe('device information detection', () => {
  it('extracts Android platform, model, operating system and version', () => {
    expect(detectDeviceDetails({
      userAgent: 'Mozilla/5.0 (Linux; Android 13; SM-S918B Build/TP1A.220624.014) AppleWebKit/537.36',
      navigatorPlatform: 'Linux armv8l',
      maxTouchPoints: 5,
      clientPlatform: 'Android',
      clientModel: '',
      clientPlatformVersion: '',
    })).toEqual({
      devicePlatform: 'android',
      deviceModel: 'SM-S918B',
      deviceOS: 'Android',
      deviceOSVersion: '13',
    })
  })

  it('reports the device identity exposed by iOS Safari without inventing an iPhone model', () => {
    expect(detectDeviceDetails({
      userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15',
      navigatorPlatform: 'iPhone',
      maxTouchPoints: 5,
      clientPlatform: '',
      clientModel: '',
      clientPlatformVersion: '',
    })).toEqual({
      devicePlatform: 'ios',
      deviceModel: 'iPhone',
      deviceOS: 'iOS',
      deviceOSVersion: '17.4',
    })
  })

  it('prefers Client Hints model and platform version when the browser exposes them', () => {
    expect(detectDeviceDetails({
      userAgent: 'Mozilla/5.0 (Linux; Android 14; Generic Build/UP1A) AppleWebKit/537.36',
      navigatorPlatform: 'Linux armv8l',
      maxTouchPoints: 5,
      clientPlatform: 'Android',
      clientModel: 'Pixel 8 Pro',
      clientPlatformVersion: '14.0.0',
    })).toEqual({
      devicePlatform: 'android',
      deviceModel: 'Pixel 8 Pro',
      deviceOS: 'Android',
      deviceOSVersion: '14.0.0',
    })
  })
})
