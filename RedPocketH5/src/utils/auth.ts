import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'
import { useLocalStorage } from '@vueuse/core'

const token = useLocalStorage(STORAGE_TOKEN_KEY, '')

function isLogin() {
  return !!localStorage.getItem(STORAGE_TOKEN_KEY)
}

function getToken() {
  return token.value
}

function setToken(newToken: string) {
  token.value = newToken
}

function clearToken() {
  token.value = null
}

export { isLogin, getToken, setToken, clearToken }
