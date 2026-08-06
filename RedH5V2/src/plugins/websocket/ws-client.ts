export interface WsClientOptions {
  url: string
  uid?: string
  token?: string
  getToken?: () => string | null | undefined
  heartbeatEnabled?: boolean
  heartbeatIntervalMs?: number
  reconnectEnabled?: boolean
  reconnectIntervalMs?: number
  maxReconnectAttempts?: number
}

export type WsEventHandler<T = unknown> = (payload: T) => void
export type WsStatusHandler = (event?: Event | CloseEvent) => void

export interface WsParsedMessage {
  event?: string
  data?: unknown
  seq?: number
  origin: unknown
}

const WS_LAST_SEQ_KEY = 'ws_last_seq'

const DEFAULT_OPTIONS = {
  uid: '',
  token: '',
  heartbeatEnabled: true,
  heartbeatIntervalMs: 10000,
  reconnectEnabled: true,
  reconnectIntervalMs: 5000,
  maxReconnectAttempts: 50,
}

export class WsClient {
  private readonly options: Required<Omit<WsClientOptions, 'getToken'>> & Pick<WsClientOptions, 'getToken'>
  private socket: WebSocket | null = null
  private latestUrl = ''
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectAttemptsLeft: number
  private reconnectLock = false
  private lastSeq = 0

  private eventHandlers = new Map<string, Set<WsEventHandler>>()
  private openHandlers = new Set<WsStatusHandler>()
  private closeHandlers = new Set<WsStatusHandler>()
  private errorHandlers = new Set<WsStatusHandler>()
  private syncExpiredHandlers = new Set<() => void>()

  constructor(options: WsClientOptions) {
    this.options = { ...DEFAULT_OPTIONS, ...options }
    this.reconnectAttemptsLeft = this.options.maxReconnectAttempts

    if (typeof localStorage !== 'undefined') {
      const stored = Number(localStorage.getItem(WS_LAST_SEQ_KEY) || '0')
      if (stored > 0) this.lastSeq = stored
    }
  }

  on<T = unknown>(event: string, handler: WsEventHandler<T>) {
    const handlers = this.eventHandlers.get(event) || new Set<WsEventHandler>()
    handlers.add(handler as WsEventHandler)
    this.eventHandlers.set(event, handlers)

    return () => this.off(event, handler as WsEventHandler)
  }

  off(event: string, handler: WsEventHandler) {
    const handlers = this.eventHandlers.get(event)
    if (!handlers) return

    handlers.delete(handler)
    if (handlers.size === 0) this.eventHandlers.delete(event)
  }

  onOpen(handler: WsStatusHandler) {
    this.openHandlers.add(handler)
    return () => this.openHandlers.delete(handler)
  }

  onClose(handler: WsStatusHandler) {
    this.closeHandlers.add(handler)
    return () => this.closeHandlers.delete(handler)
  }

  onError(handler: WsStatusHandler) {
    this.errorHandlers.add(handler)
    return () => this.errorHandlers.delete(handler)
  }

  onSyncExpired(handler: () => void) {
    this.syncExpiredHandlers.add(handler)
    return () => this.syncExpiredHandlers.delete(handler)
  }

  connect() {
    if (!this.options.url) {
      console.info('[ws] skipped: empty websocket url')
      return
    }

    if (
      this.socket
      && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)
    ) {
      return
    }

    const url = this.buildSocketUrl()
    this.latestUrl = url
    this.socket = new WebSocket(url)
    this.socket.onopen = event => this.handleOpen(event)
    this.socket.onclose = event => this.handleClose(event)
    this.socket.onerror = event => this.handleError(event)
    this.socket.onmessage = event => {
      void this.handleMessage(event)
    }
  }

  close() {
    this.clearHeartbeat()
    this.clearReconnect()
    this.reconnectAttemptsLeft = 0

    if (!this.socket) return

    this.socket.close()
    this.socket = null
  }

  send(data: unknown) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return false

    this.socket.send(typeof data === 'string' ? data : JSON.stringify(data))
    return true
  }

  emit(event: string, data?: Record<string, unknown>) {
    return this.send({
      type: event,
      data: data || {},
      ts: Math.floor(Date.now() / 1000),
    })
  }

  get isConnected() {
    return this.socket?.readyState === WebSocket.OPEN
  }

  private buildSocketUrl() {
    const token = this.options.getToken?.() || this.options.token
    const { uid, url } = this.options

    if (!uid && !token) return url

    const [base, query = ''] = url.split('?')
    const params = new URLSearchParams(query)
    if (uid) params.set('uid', uid)
    if (token) params.set('token', token)

    return `${base}?${params.toString()}`
  }

  private handleOpen(event: Event) {
    this.reconnectAttemptsLeft = this.options.maxReconnectAttempts
    this.reconnectLock = false
    this.startHeartbeat()
    console.info('[ws] connected:', this.maskToken(this.latestUrl))
    this.send({ type: 'sync', lastSeq: this.lastSeq, scope: '' })
    this.openHandlers.forEach(handler => handler(event))
  }

  private handleClose(event: CloseEvent) {
    this.clearHeartbeat()
    this.closeHandlers.forEach(handler => handler(event))

    if (event.code !== 1000) this.reconnect()
  }

  private handleError(event: Event) {
    console.warn('[ws] error:', {
      readyState: this.socket?.readyState,
      reconnectLeft: this.reconnectAttemptsLeft,
      url: this.maskToken(this.latestUrl),
    })
    this.errorHandlers.forEach(handler => handler(event))
  }

  private async handleMessage(event: MessageEvent<string | Blob | ArrayBuffer>) {
    const raw = await this.normalizeMessageData(event.data)

    if (raw === 'ping') {
      this.send('pong')
      return
    }

    const parsed = this.parseMessage(raw)

    if (parsed.event === 'ping') {
      this.send({
        type: 'pong',
        ts: Math.floor(Date.now() / 1000),
      })
      return
    }

    if (parsed.event === 'sync_expired') {
      this.syncExpiredHandlers.forEach(handler => handler())
      return
    }

    if (parsed.seq && parsed.seq > 0) {
      if (parsed.seq > this.lastSeq) {
        this.lastSeq = parsed.seq
        if (typeof localStorage !== 'undefined') {
          localStorage.setItem(WS_LAST_SEQ_KEY, String(parsed.seq))
        }
      }
      this.send({ type: 'ack', seq: parsed.seq })
    }

    if (!parsed.event) return

    const handlers = this.eventHandlers.get(parsed.event)
    if (!handlers || handlers.size === 0) return

    handlers.forEach(handler => handler(parsed.origin))
  }

  private parseMessage(raw: string): WsParsedMessage {
    try {
      const obj = JSON.parse(raw) as Record<string, unknown>
      return {
        event: typeof obj.type === 'string' ? obj.type : typeof obj.event === 'string' ? obj.event : undefined,
        data: obj.data,
        seq: obj.seq ? Number(obj.seq) : undefined,
        origin: obj,
      }
    } catch {
      return { origin: raw }
    }
  }

  private async normalizeMessageData(data: string | Blob | ArrayBuffer) {
    if (typeof data === 'string') return data
    if (data instanceof Blob) return await data.text()
    if (data instanceof ArrayBuffer) return new TextDecoder().decode(new Uint8Array(data))
    return ''
  }

  private startHeartbeat() {
    if (!this.options.heartbeatEnabled) return

    this.clearHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      this.send({
        type: 'pong',
        ts: Math.floor(Date.now() / 1000),
      })
    }, this.options.heartbeatIntervalMs)
  }

  private reconnect() {
    if (!this.options.reconnectEnabled || this.reconnectLock || this.reconnectAttemptsLeft <= 0) return

    this.reconnectLock = true
    this.clearReconnect()
    this.reconnectTimer = setTimeout(() => {
      this.reconnectLock = false
      this.reconnectAttemptsLeft -= 1
      this.connect()
    }, this.options.reconnectIntervalMs)
  }

  private clearHeartbeat() {
    if (!this.heartbeatTimer) return

    clearInterval(this.heartbeatTimer)
    this.heartbeatTimer = null
  }

  private clearReconnect() {
    if (!this.reconnectTimer) return

    clearTimeout(this.reconnectTimer)
    this.reconnectTimer = null
  }

  private maskToken(url: string) {
    return url.replace(/(token=)[^&]+/g, 'S/1***')
  }
}
