import type { DeviceInfoPayload } from '@/utils/deviceInfo'
import { reportCurrentTgDeviceInfo } from '@/api/user'
import { getDeviceInfo } from '@/utils/deviceInfo'
import { getTtlStorage, setTtlStorage } from '@/utils/storage'

export const DEVICE_REPORT_STORAGE_PREFIX = 'ppmx_device_reported_v1'

function getTodayKey(now = new Date()) {
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function getStorageKey(userID: number | string) {
  return `${DEVICE_REPORT_STORAGE_PREFIX}:${String(userID)}`
}

function getMsUntilTomorrow(now = new Date()) {
  const tomorrow = new Date(now)
  tomorrow.setHours(24, 0, 0, 0)
  return Math.max(1, tomorrow.getTime() - now.getTime())
}

export function markDeviceInfoReportedToday(userID: number | string, now = new Date()) {
  setTtlStorage(getStorageKey(userID), getTodayKey(now), getMsUntilTomorrow(now))
}

export async function reportDeviceInfoOncePerDay(
  userID: number | string,
  loadDeviceInfo: () => Promise<DeviceInfoPayload> = getDeviceInfo,
) {
  const storageKey = getStorageKey(userID)
  const today = getTodayKey()

  if (getTtlStorage<string>(storageKey) === today) {
    return false
  }

  const deviceInfo = await loadDeviceInfo()
  await reportCurrentTgDeviceInfo(deviceInfo)
  markDeviceInfoReportedToday(userID)
  return true
}
