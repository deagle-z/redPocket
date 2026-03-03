import { defineStore } from 'pinia'
import type { ForgotPasswordData, LoginData, RegisterData, TgAuthLoginData, UserState } from '@/api/user'
import { clearToken, setToken } from '@/utils/auth'

import {
  forgotPasswordByEmail,
  getEmailCode,
  getCurrentTgUserInfo,
  sendRegisterEmailCode,
  loginByEmail as userLoginByEmail,
  logout as userLogout,
  register as userRegister,
  tgLogin as userTgLogin,
} from '@/api/user'

const InitUserInfo = {
  uid: 0,
  nickname: '',
  avatar: '',
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserState>({ ...InitUserInfo })

  // Set user's information
  const setInfo = (partial: Partial<UserState>) => {
    userInfo.value = { ...partial }
  }

  const login = async (loginForm: LoginData) => {
    try {
      const { data } = await userLoginByEmail(loginForm)
      const token = data.accessToken || data.token
      if (!token)
        throw new Error('Login token missing')
      setToken(token)
    }
    catch (error) {
      clearToken()
      throw error
    }
  }

  const loginByTelegram = async (payload: TgAuthLoginData) => {
    try {
      const { data } = await userTgLogin(payload)
      setToken(data.accessToken)
    }
    catch (error) {
      clearToken()
      throw error
    }
  }

  const loadCurrentUserInfo = async () => {
    try {
      const { data } = await getCurrentTgUserInfo()
      setInfo({
        uid: Number(data?.uid || 0),
        name: data?.username || '',
        avatar: data?.avatar || '',
      })
    }
    catch (error) {
      clearToken()
      throw error
    }
  }

  const logout = async () => {
    try {
      await userLogout()
    }
    finally {
      clearToken()
      setInfo({ ...InitUserInfo })
    }
  }

  const getCode = async () => {
    try {
      const data = await getEmailCode()
      return data
    }
    catch {}
  }

  const reset = async (form: ForgotPasswordData) => {
    const data = await forgotPasswordByEmail(form)
    return data
  }

  const register = async (form: RegisterData) => {
    const data = await userRegister(form)
    return data
  }

  const sendCode = async (email: string) => {
    const data = await sendRegisterEmailCode(email)
    return data
  }

  return {
    userInfo,
    info: loadCurrentUserInfo,
    loadCurrentUserInfo,
    login,
    loginByTelegram,
    logout,
    getCode,
    reset,
    register,
    sendCode,
  }
}, {
  persist: true,
})

export default useUserStore
