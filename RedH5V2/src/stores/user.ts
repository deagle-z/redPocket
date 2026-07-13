import type {
  CurrentUserInfo,
  LoginByPhoneParams,
  RegisterByPhoneParams,
} from '@/api/user'
import {
  getCurrentUserInfo,
  loginByPhone as loginByPhoneApi,
  registerByPhone as registerByPhoneApi,
} from '@/api/user'
import { AUTH_EXPIRED_EVENT } from '@/utils/authEvents'
import { clearAuthToken, getAuthToken, setAuthToken } from '@/utils/authToken'
import { setLocalRechargeCount } from '@/utils/rechargeCount'

export interface UserInfo {
  id: number | string
  username: string
  nickname?: string
  avatar?: string
  country?: string
  balance?: number | string
  sportBalance?: number | string
  hasWithdrawAccount?: boolean
  firstName?: string
  lastName?: string
  phone?: string
  email?: string
  rechargeCount?: number
  rebateWithdrawDisabled?: number
}

interface LoginOptions {
  remember?: boolean
}

function extractToken(response?: { accessToken?: string; token?: string }) {
  return response?.accessToken || response?.token || ''
}

function mapCurrentUserInfo(data: CurrentUserInfo): UserInfo {
  const id = data.id ?? data.uid ?? data.userId ?? ''
  const username = data.username || data.phone || data.email || String(id)
  const nickname = data.nickname || data.tgName || data.firstName || username

  return {
    id,
    username,
    nickname,
    avatar: data.avatar,
    country: data.country,
    balance: data.balance,
    sportBalance: data.sportBalance,
    hasWithdrawAccount: data.hasWithdrawAccount,
    firstName: data.firstName,
    lastName: data.lastName,
    phone: data.phone,
    email: data.email,
    rechargeCount: data.rechargeCount,
    rebateWithdrawDisabled: data.rebateWithdrawDisabled,
  }
}

export const useUserStore = defineStore('user', () => {
  const token = ref(getAuthToken())
  const userInfo = ref<UserInfo | null>(null)

  const isLogin = computed(() => Boolean(token.value))

  function setToken(value: string, remember = true) {
    token.value = value
    setAuthToken(value, remember)
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    clearAuthToken()
  }

  if (typeof window !== 'undefined') {
    window.addEventListener(AUTH_EXPIRED_EVENT, logout)
  }

  async function loadUserInfo() {
    const profile = await getCurrentUserInfo()
    userInfo.value = mapCurrentUserInfo(profile)

    if (typeof userInfo.value.rechargeCount === 'number' && userInfo.value.rechargeCount > 0) {
      setLocalRechargeCount(userInfo.value.rechargeCount)
    }

    return userInfo.value
  }

  async function loginByPhone(params: LoginByPhoneParams, options: LoginOptions = {}) {
    try {
      const response = await loginByPhoneApi(params)
      const nextToken = extractToken(response)

      if (!nextToken) {
        throw new Error('Missing login token')
      }

      setToken(nextToken, options.remember ?? true)
      await loadUserInfo()

      return userInfo.value
    } catch (error) {
      logout()
      throw error
    }
  }

  async function registerByPhone(params: RegisterByPhoneParams) {
    try {
      const response = await registerByPhoneApi(params)
      const nextToken = extractToken(response)

      if (!nextToken) {
        throw new Error('Missing registration token')
      }

      setToken(nextToken)
      await loadUserInfo()

      return userInfo.value
    } catch (error) {
      logout()
      throw error
    }
  }

  return {
    token,
    userInfo,
    isLogin,
    setToken,
    logout,
    loadUserInfo,
    loginByPhone,
    registerByPhone,
  }
})
