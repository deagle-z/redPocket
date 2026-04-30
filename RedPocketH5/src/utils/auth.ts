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

function getTokenTenantId() {
  const rawToken = localStorage.getItem(STORAGE_TOKEN_KEY) || token.value || ''
  const parts = rawToken.split('.')
  if (parts.length < 2)
    return 0

  try {
    const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = payload.padEnd(payload.length + (4 - payload.length % 4) % 4, '=')
    const binary = window.atob(padded)
    const bytes = Uint8Array.from(binary, char => char.charCodeAt(0))
    const claims = JSON.parse(new TextDecoder().decode(bytes))
    return Number(claims?.tenantId || 0)
  }
  catch {
    return 0
  }
}

function isMerchantTenant() {
  return getTokenTenantId() > 0
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

export { isLogin, getToken, getTokenTenantId, isMerchantTenant, setToken, clearToken, getAuthCountry, setAuthCountry }
