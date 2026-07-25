import { getDeviceFingerprint } from '@/utils/deviceFingerprint'

export interface DeviceInfoPayload {
  deviceFingerprint: string
  devicePlatform: string
  deviceModel: string
  deviceOS: string
  deviceOSVersion: string
}

export interface DeviceEnvironment {
  userAgent: string
  navigatorPlatform: string
  maxTouchPoints: number
  clientPlatform: string
  clientModel: string
  clientPlatformVersion: string
}

interface NavigatorUserAgentData {
  platform?: string
  getHighEntropyValues?: (hints: string[]) => Promise<{
    model?: string
    platformVersion?: string
  }>
}

function normalizeVersion(value: string) {
  return value.trim().replaceAll('_', '.')
}

function detectPlatform(environment: DeviceEnvironment) {
  const clientPlatform = environment.clientPlatform.toLowerCase()
  const navigatorPlatform = environment.navigatorPlatform.toLowerCase()
  const userAgent = environment.userAgent

  if (/android/i.test(clientPlatform) || /android/i.test(userAgent)) return 'android'
  if (
    /ios/i.test(clientPlatform) ||
    /iPhone|iPad|iPod/i.test(userAgent) ||
    (navigatorPlatform === 'macintel' && environment.maxTouchPoints > 1)
  ) {
    return 'ios'
  }
  if (/windows/i.test(clientPlatform) || /Windows NT/i.test(userAgent)) return 'windows'
  if (/macos|mac os/i.test(clientPlatform) || /Macintosh|Mac OS X/i.test(userAgent)) return 'macos'
  if (/chrome os/i.test(clientPlatform) || /CrOS/i.test(userAgent)) return 'chromeos'
  if (/linux/i.test(clientPlatform) || /Linux/i.test(userAgent)) return 'linux'
  return ''
}

function detectAndroidModel(userAgent: string) {
  const match = userAgent.match(
    /Android\s+[^;()]+;\s*(?:[^;()]+;\s*)?([^;()]+?)(?:\s+Build\/[^;()]*)?(?:;|\))/i,
  )
  return match?.[1]?.trim() ?? ''
}

export function detectDeviceDetails(environment: DeviceEnvironment): Omit<DeviceInfoPayload, 'deviceFingerprint'> {
  const devicePlatform = detectPlatform(environment)
  const userAgent = environment.userAgent
  let deviceModel = environment.clientModel.trim()
  let deviceOS = ''
  let deviceOSVersion = normalizeVersion(environment.clientPlatformVersion)

  if (devicePlatform === 'android') {
    deviceOS = 'Android'
    deviceOSVersion ||= normalizeVersion(userAgent.match(/Android\s+([\d.]+)/i)?.[1] ?? '')
    deviceModel ||= detectAndroidModel(userAgent)
  } else if (devicePlatform === 'ios') {
    deviceOS = 'iOS'
    deviceOSVersion ||= normalizeVersion(
      userAgent.match(/(?:CPU (?:iPhone )?OS|iPhone OS)\s+([\d_]+)/i)?.[1] ?? '',
    )
    if (!deviceModel) {
      if (/iPad/i.test(userAgent)) deviceModel = 'iPad'
      else if (/iPod/i.test(userAgent)) deviceModel = 'iPod'
      else if (/iPhone/i.test(userAgent)) deviceModel = 'iPhone'
    }
  } else if (devicePlatform === 'windows') {
    deviceOS = 'Windows'
    deviceOSVersion ||= normalizeVersion(userAgent.match(/Windows NT\s+([\d.]+)/i)?.[1] ?? '')
  } else if (devicePlatform === 'macos') {
    deviceOS = 'macOS'
    deviceOSVersion ||= normalizeVersion(userAgent.match(/Mac OS X\s+([\d_]+)/i)?.[1] ?? '')
  } else if (devicePlatform === 'chromeos') {
    deviceOS = 'Chrome OS'
    deviceOSVersion ||= normalizeVersion(userAgent.match(/CrOS\s+\S+\s+([\d.]+)/i)?.[1] ?? '')
  } else if (devicePlatform === 'linux') {
    deviceOS = 'Linux'
  }

  return {
    devicePlatform,
    deviceModel,
    deviceOS,
    deviceOSVersion,
  }
}

async function getDeviceEnvironment(): Promise<DeviceEnvironment> {
  if (typeof navigator === 'undefined') {
    return {
      userAgent: '',
      navigatorPlatform: '',
      maxTouchPoints: 0,
      clientPlatform: '',
      clientModel: '',
      clientPlatformVersion: '',
    }
  }

  const navigatorWithClientHints = navigator as Navigator & {
    userAgentData?: NavigatorUserAgentData
  }
  const userAgentData = navigatorWithClientHints.userAgentData
  let clientModel = ''
  let clientPlatformVersion = ''

  if (userAgentData?.getHighEntropyValues) {
    try {
      const values = await userAgentData.getHighEntropyValues(['model', 'platformVersion'])
      clientModel = values.model?.trim() ?? ''
      clientPlatformVersion = values.platformVersion?.trim() ?? ''
    } catch {
      // Client Hints may be unavailable; the User-Agent parser below remains the browser-compatible source.
    }
  }

  return {
    userAgent: navigator.userAgent ?? '',
    navigatorPlatform: navigator.platform ?? '',
    maxTouchPoints: navigator.maxTouchPoints ?? 0,
    clientPlatform: userAgentData?.platform?.trim() ?? '',
    clientModel,
    clientPlatformVersion,
  }
}

export async function getDeviceInfo(): Promise<DeviceInfoPayload> {
  const [deviceFingerprint, environment] = await Promise.all([
    getDeviceFingerprint(),
    getDeviceEnvironment(),
  ])

  return {
    deviceFingerprint,
    ...detectDeviceDetails(environment),
  }
}
