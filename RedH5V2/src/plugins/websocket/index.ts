import { getAuthToken } from '@/utils/authToken'
import { WsClient } from './ws-client'

const defaultUrl = import.meta.env.VITE_WS_URL || import.meta.env.VITE_APP_WS_URL || ''
const defaultUid = import.meta.env.VITE_APP_WS_UID || ''

function getToken() {
  return getAuthToken()
}

const wsClient = new WsClient({
  url: defaultUrl,
  uid: defaultUid,
  getToken,
})

export function connectWebSocket(uid?: string) {
  if (!defaultUrl) {
    console.info('[ws] skipped: VITE_WS_URL is empty')
    return wsClient
  }

  if (uid && uid !== defaultUid) {
    const dynamicClient = new WsClient({
      url: defaultUrl,
      uid,
      getToken,
    })
    dynamicClient.connect()
    return dynamicClient
  }

  wsClient.connect()
  return wsClient
}

export function closeWebSocket() {
  wsClient.close()
}

export default wsClient
