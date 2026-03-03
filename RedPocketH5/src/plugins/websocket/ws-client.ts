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

type WsEventHandler<T = any> = (payload: T) => void
type WsStatusHandler = (event?: Event | CloseEvent) => void

interface ParsedMessage {
  event?: string
  data?: any
  origin: any
}

const defaultOptions: Required<Omit<WsClientOptions, 'url'>> = {
  uid: '',
  token: '',
  getToken: undefined,
  heartbeatEnabled: true,
  heartbeatIntervalMs: 10000,
  reconnectEnabled: true,
  reconnectIntervalMs: 5000,
  maxReconnectAttempts: 50,
}

export class WsClient {
  private readonly options: Required<WsClientOptions>
  private socket: WebSocket | null = null
  private latestUrl = ''
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectAttemptsLeft: number
  private reconnectLock = false

  private eventHandlers = new Map<string, Set<WsEventHandler>>()
  private openHandlers = new Set<WsStatusHandler>()
  private closeHandlers = new Set<WsStatusHandler>()
  private errorHandlers = new Set<WsStatusHandler>()

  constructor(options: WsClientOptions) {
    this.options = { ...defaultOptions, ...options }
    this.reconnectAttemptsLeft = this.options.maxReconnectAttempts
  }

  on(event: string, handler: WsEventHandler) {
    const handlers = this.eventHandlers.get(event) || new Set<WsEventHandler>()
    handlers.add(handler)
    this.eventHandlers.set(event, handlers)
    return () => this.off(event, handler)
  }

  off(event: string, handler: WsEventHandler) {
    const handlers = this.eventHandlers.get(event)
    if (!handlers)
      return
    handlers.delete(handler)
    if (handlers.size === 0)
      this.eventHandlers.delete(event)
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

  connect() {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING))
      return

    const url = this.buildSocketUrl()
    this.latestUrl = url
    this.socket = new WebSocket(url)
    this.socket.onopen = evt => this.handleOpen(evt)
    this.socket.onclose = evt => this.handleClose(evt)
    this.socket.onerror = evt => this.handleError(evt)
    this.socket.onmessage = evt => { void this.handleMessage(evt) }
  }

  close() {
    this.clearHeartbeat()
    this.clearReconnect()
    this.reconnectAttemptsLeft = 0
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
  }

  send(data: any) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN)
      return false
    this.socket.send(typeof data === 'string' ? data : JSON.stringify(data))
    return true
  }

  emit(event: string, data?: Record<string, any>) {
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
    const { uid, url } = this.options
    const token = this.options.getToken?.() || this.options.token
    if (!uid && !token)
      return url

    const [base, query = ''] = url.split('?')
    const params = new URLSearchParams(query)
    if (uid)
      params.set('uid', uid)
    if (token)
      params.set('token', token)
    return `${base}?${params.toString()}`
  }

  private handleOpen(evt: Event) {
    this.reconnectAttemptsLeft = this.options.maxReconnectAttempts
    this.reconnectLock = false
    this.startHeartbeat()
    console.warn('[ws] connected:', this.maskToken(this.latestUrl))
    this.openHandlers.forEach(handler => handler(evt))
  }

  private handleClose(evt: CloseEvent) {
    this.clearHeartbeat()
    console.warn('[ws] closed:', {
      code: evt.code,
      reason: evt.reason || '(empty)',
      wasClean: evt.wasClean,
      reconnectLeft: this.reconnectAttemptsLeft,
      url: this.maskToken(this.latestUrl),
    })
    this.closeHandlers.forEach(handler => handler(evt))

    // 1000 = normal close, do not reconnect
    if (evt.code !== 1000)
      this.reconnect()
  }

  private handleError(evt: Event) {
    console.error('[ws] error:', {
      readyState: this.socket?.readyState,
      reconnectLeft: this.reconnectAttemptsLeft,
      url: this.maskToken(this.latestUrl),
      event: evt,
    })
    this.errorHandlers.forEach(handler => handler(evt))
  }

  private async handleMessage(evt: MessageEvent<string | Blob | ArrayBuffer>) {
    const raw = await this.normalizeMessageData(evt.data)
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
    if (!parsed.event)
      return

    const handlers = this.eventHandlers.get(parsed.event)
    if (!handlers || handlers.size === 0)
      return

    handlers.forEach(handler => handler(parsed.origin))
  }

  private parseMessage(raw: string): ParsedMessage {
    try {
      const obj = JSON.parse(raw)
      return {
        event: obj?.type || obj?.event,
        data: obj?.data,
        origin: obj,
      }
    }
    catch {
      return { origin: raw }
    }
  }

  private async normalizeMessageData(data: string | Blob | ArrayBuffer): Promise<string> {
    if (typeof data === 'string')
      return data
    if (data instanceof Blob)
      return await data.text()
    if (data instanceof ArrayBuffer)
      return new TextDecoder().decode(new Uint8Array(data))
    return ''
  }

  private startHeartbeat() {
    if (!this.options.heartbeatEnabled)
      return
    this.clearHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      this.send({
        type: 'pong',
        ts: Math.floor(Date.now() / 1000),
      })
    }, this.options.heartbeatIntervalMs)
  }

  private reconnect() {
    if (!this.options.reconnectEnabled)
      return
    if (this.reconnectLock || this.reconnectAttemptsLeft <= 0)
      return

    this.reconnectLock = true
    this.clearReconnect()
    this.reconnectTimer = setTimeout(() => {
      this.reconnectLock = false
      this.reconnectAttemptsLeft--
      this.connect()
    }, this.options.reconnectIntervalMs)
  }

  private clearHeartbeat() {
    if (!this.heartbeatTimer)
      return
    clearInterval(this.heartbeatTimer)
    this.heartbeatTimer = null
  }

  private clearReconnect() {
    if (!this.reconnectTimer)
      return
    clearTimeout(this.reconnectTimer)
    this.reconnectTimer = null
  }

  private maskToken(url: string) {
    if (!url)
      return url
    return url.replace(/(token=)[^&]+/g, '$1***')
  }
}
