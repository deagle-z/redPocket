import { beforeEach, describe, expect, it, vi } from 'vitest'
import { WsClient } from './ws-client'

class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3
  static instances: MockWebSocket[] = []

  readyState = 0
  sent: string[] = []
  onopen: ((event: Event) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent<string>) => void) | null = null

  constructor(public url: string) {
    MockWebSocket.instances.push(this)
  }

  send(data: string) {
    this.sent.push(data)
  }

  close() {
    this.readyState = MockWebSocket.CLOSED
  }

  open() {
    this.readyState = MockWebSocket.OPEN
    this.onopen?.(new Event('open'))
  }

  message(data: string) {
    this.onmessage?.(new MessageEvent('message', { data }))
  }
}

function installStorage() {
  const data = new Map<string, string>()
  vi.stubGlobal('localStorage', {
    getItem: vi.fn((key: string) => data.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) => data.set(key, value)),
    removeItem: vi.fn((key: string) => data.delete(key)),
  })
}

async function flushMessage() {
  await Promise.resolve()
  await Promise.resolve()
}

describe('WsClient', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.unstubAllGlobals()
    MockWebSocket.instances = []
    installStorage()
    vi.stubGlobal('WebSocket', MockWebSocket)
  })

  it('connects with token and sends sync from last sequence', () => {
    localStorage.setItem('ws_last_seq', '9')
    const client = new WsClient({
      url: 'wss://example.test/ws',
      uid: 'user-1',
      getToken: () => 'token-1',
    })

    client.connect()
    const socket = MockWebSocket.instances[0]
    socket.open()

    expect(socket.url).toBe('wss://example.test/ws?uid=user-1&token=token-1')
    expect(socket.sent).toContain(JSON.stringify({ type: 'sync', lastSeq: 9, scope: '' }))
  })

  it('dispatches type and event messages, stores seq, and sends ack', async () => {
    const client = new WsClient({ url: 'wss://example.test/ws' })
    const handler = vi.fn()
    client.on('recharge_success', handler)

    client.connect()
    const socket = MockWebSocket.instances[0]
    socket.open()
    socket.message(JSON.stringify({
      type: 'recharge_success',
      seq: 12,
      data: { orderNo: 'RC1' },
    }))
    await flushMessage()

    expect(handler).toHaveBeenCalledWith({
      type: 'recharge_success',
      seq: 12,
      data: { orderNo: 'RC1' },
    })
    expect(localStorage.setItem).toHaveBeenCalledWith('ws_last_seq', '12')
    expect(socket.sent).toContain(JSON.stringify({ type: 'ack', seq: 12 }))

    socket.message(JSON.stringify({ event: 'recharge_success', data: { orderNo: 'RC2' } }))
    await flushMessage()

    expect(handler).toHaveBeenCalledTimes(2)
  })

  it('responds to ping and reports sync expiration', async () => {
    const client = new WsClient({ url: 'wss://example.test/ws' })
    const syncExpired = vi.fn()
    client.onSyncExpired(syncExpired)

    client.connect()
    const socket = MockWebSocket.instances[0]
    socket.open()
    socket.message('ping')
    socket.message(JSON.stringify({ type: 'sync_expired' }))
    await flushMessage()

    expect(socket.sent).toContain('pong')
    expect(syncExpired).toHaveBeenCalledTimes(1)
  })
})
