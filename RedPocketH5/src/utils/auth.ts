import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'
import { useLocalStorage } from '@vueuse/core'

const token = useLocalStorage(STORAGE_TOKEN_KEY, '')
const COUNTRY_STORAGE_KEY = 'auth_country'

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

function getAuthCountry() {
  return localStorage.getItem(COUNTRY_STORAGE_KEY) || ''
}

function setAuthCountry(country: string) {
  localStorage.setItem(COUNTRY_STORAGE_KEY, country)
}

export { isLogin, getToken, setToken, clearToken, getAuthCountry, setAuthCountry }
