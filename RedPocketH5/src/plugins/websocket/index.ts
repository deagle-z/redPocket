import { WsClient } from './ws-client'
import { getToken } from '@/utils/auth'

const defaultUrl = import.meta.env.VITE_WS_URL || import.meta.env.VITE_APP_WS_URL || ''
const defaultUid = import.meta.env.VITE_APP_WS_UID || ''

const wsClient = new WsClient({
  url: defaultUrl,
  uid: defaultUid,
  getToken,
})

export function connectWebSocket(uid?: string) {
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

export default wsClient
