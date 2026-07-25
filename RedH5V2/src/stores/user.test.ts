import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { loginByPhone as loginByPhoneApi } from '@/api/user'
import { AUTH_EXPIRED_EVENT } from '@/utils/authEvents'
import { useUserStore } from './user'

vi.mock('@/api/user', () => ({
  getCurrentUserInfo: vi.fn(async () => ({
    uid: 1001,
    username: 'player_1001',
    firstName: 'Player',
    avatar: 'avatar.png',
    country: 'MX',
    balance: 1200,
    sportBalance: 300,
    hasWithdrawAccount: true,
  })),
  loginByPhone: vi.fn(async () => ({ accessToken: 'access-token-1' })),
  registerByPhone: vi.fn(async () => ({ token: 'registered-token' })),
  reportCurrentTgDeviceInfo: vi.fn(async () => undefined),
}))

function installWindowStorage() {
  const localData = new Map<string, string>()
  const sessionData = new Map<string, string>()
  const eventTarget = new EventTarget()
  const createStorage = (data: Map<string, string>) => ({
    get length() {
      return data.size
    },
    clear: vi.fn(() => data.clear()),
    getItem: vi.fn((key: string) => data.get(key) ?? null),
    key: vi.fn((index: number) => Array.from(data.keys())[index] ?? null),
    removeItem: vi.fn((key: string) => data.delete(key)),
    setItem: vi.fn((key: string, value: string) => data.set(key, value)),
  })
  const localStorage = createStorage(localData)
  const sessionStorage = createStorage(sessionData)

  vi.stubGlobal('window', {
    localStorage,
    sessionStorage,
    addEventListener: eventTarget.addEventListener.bind(eventTarget),
    removeEventListener: eventTarget.removeEventListener.bind(eventTarget),
    dispatchEvent: eventTarget.dispatchEvent.bind(eventTarget),
  })

  return { localStorage, sessionStorage }
}

describe('user store auth actions', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    vi.clearAllMocks()
    setActivePinia(createPinia())
    installWindowStorage()
  })

  it('writes unified token key after phone login', async () => {
    const store = useUserStore()

    await store.loginByPhone({
      phone: '525512345678',
      password: 'password123',
      country: 'MX',
    })

    expect(store.token).toBe('access-token-1')
    expect(store.userInfo?.country).toBe('MX')
    expect(store.userInfo?.sportBalance).toBe(300)
    expect(store.userInfo?.hasWithdrawAccount).toBe(true)
    expect(window.localStorage.getItem('h5_token')).toContain('access-token-1')
  })

  it('uses session token storage when remember me is disabled', async () => {
    const store = useUserStore()

    await store.loginByPhone({
      phone: '525512345678',
      password: 'password123',
      country: 'MX',
    }, { remember: false })

    expect(store.token).toBe('access-token-1')
    expect(window.localStorage.getItem('h5_token')).toBeNull()
    expect(window.sessionStorage.getItem('h5_token')).toContain('access-token-1')
  })

  it('clears local and session token storage on logout', async () => {
    const store = useUserStore()

    await store.loginByPhone({
      phone: '525512345678',
      password: 'password123',
      country: 'MX',
    }, { remember: false })
    store.logout()

    expect(store.token).toBe('')
    expect(store.userInfo).toBeNull()
    expect(window.localStorage.getItem('h5_token')).toBeNull()
    expect(window.sessionStorage.getItem('h5_token')).toBeNull()
  })

  it('logs out immediately when the global auth expired event is dispatched', async () => {
    const store = useUserStore()

    await store.loginByPhone({
      phone: '525512345678',
      password: 'password123',
      country: 'MX',
    })

    window.dispatchEvent(new Event(AUTH_EXPIRED_EVENT))

    expect(store.token).toBe('')
    expect(store.userInfo).toBeNull()
    expect(window.localStorage.getItem('h5_token')).toBeNull()
    expect(window.sessionStorage.getItem('h5_token')).toBeNull()
  })

  it('uses registration token directly and loads current user info', async () => {
    const store = useUserStore()

    await store.registerByPhone({
      phone: '525512345678',
      password: 'password123',
      country: 'MX',
      firstName: 'Player',
    })

    expect(store.isLogin).toBe(true)
    expect(store.token).toBe('registered-token')
    expect(store.userInfo?.username).toBe('player_1001')
    expect(loginByPhoneApi).toHaveBeenCalledTimes(0)
  })
})
