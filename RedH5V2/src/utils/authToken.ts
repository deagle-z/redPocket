import {
  getSessionTtlStorage,
  getTtlStorage,
  removeSessionTtlStorage,
  removeTtlStorage,
  setSessionTtlStorage,
  setTtlStorage,
} from './storage'

export const AUTH_TOKEN_KEY = 'h5_token'

const AUTH_TOKEN_TTL = 7 * 24 * 60 * 60 * 1000

export function getAuthToken() {
  return getTtlStorage<string>(AUTH_TOKEN_KEY) ?? getSessionTtlStorage<string>(AUTH_TOKEN_KEY) ?? ''
}

export function setAuthToken(value: string, remember = true) {
  clearAuthToken()

  if (remember) {
    setTtlStorage(AUTH_TOKEN_KEY, value, AUTH_TOKEN_TTL)
  } else {
    setSessionTtlStorage(AUTH_TOKEN_KEY, value)
  }
}

export function clearAuthToken() {
  removeTtlStorage(AUTH_TOKEN_KEY)
  removeSessionTtlStorage(AUTH_TOKEN_KEY)
}
