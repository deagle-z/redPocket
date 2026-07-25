import { beforeEach, describe, expect, it, vi } from 'vitest'

import { reportCurrentTgDeviceInfo } from '@/api/user'
import { getTtlStorage, setTtlStorage } from '@/utils/storage'
import { reportDeviceInfoOncePerDay } from './deviceReport'

vi.mock('@/api/user', () => ({
  reportCurrentTgDeviceInfo: vi.fn(async () => undefined),
}))

vi.mock('@/utils/storage', () => ({
  getTtlStorage: vi.fn(),
  setTtlStorage: vi.fn(),
}))

const deviceInfo = {
  deviceFingerprint: 'visitor-device-id',
  devicePlatform: 'android',
  deviceModel: 'Pixel 8 Pro',
  deviceOS: 'Android',
  deviceOSVersion: '14',
}

describe('daily device report', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('reports and marks the first authenticated app open of the day', async () => {
    vi.mocked(getTtlStorage).mockReturnValue(null)
    const loadDeviceInfo = vi.fn(async () => deviceInfo)

    await expect(reportDeviceInfoOncePerDay(41, loadDeviceInfo)).resolves.toBe(true)
    expect(reportCurrentTgDeviceInfo).toHaveBeenCalledWith(deviceInfo)
    expect(setTtlStorage).toHaveBeenCalledOnce()
  })

  it('skips another report when the current day is already marked', async () => {
    const today = new Date()
    const todayKey = [
      today.getFullYear(),
      String(today.getMonth() + 1).padStart(2, '0'),
      String(today.getDate()).padStart(2, '0'),
    ].join('-')
    vi.mocked(getTtlStorage).mockReturnValue(todayKey)
    const loadDeviceInfo = vi.fn(async () => deviceInfo)

    await expect(reportDeviceInfoOncePerDay(41, loadDeviceInfo)).resolves.toBe(false)
    expect(loadDeviceInfo).not.toHaveBeenCalled()
    expect(reportCurrentTgDeviceInfo).not.toHaveBeenCalled()
    expect(setTtlStorage).not.toHaveBeenCalled()
  })

  it('does not mark the day when the backend report fails', async () => {
    vi.mocked(getTtlStorage).mockReturnValue(null)
    vi.mocked(reportCurrentTgDeviceInfo).mockRejectedValueOnce(new Error('network error'))

    await expect(reportDeviceInfoOncePerDay(41, async () => deviceInfo)).rejects.toThrow('network error')
    expect(setTtlStorage).not.toHaveBeenCalled()
  })
})
