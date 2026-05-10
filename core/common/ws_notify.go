package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

const (
	wsPingInterval = 25 * time.Second
	wsPongTimeout  = 60 * time.Second
	wsSendBuffer   = 64
	wsConnWindow   = 60 * time.Second
	wsConnMax      = 30
	wsReconnectMin = 2 * time.Second
	wsMsgBufMax    = 200
	wsMsgBufExpire = 5 * time.Minute
	wsMsgBufKey    = "ws:msgbuf:"
	wsMsgSeqKey    = "ws:seq:"
	wsDefaultScope = "all"
)

type wsMessage struct {
	Type  string      `json:"type"`
	Seq   int64       `json:"seq,omitempty"`
	Scope string      `json:"scope,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Ts    int64       `json:"ts"`
}

type wsIncoming struct {
	Type  string          `json:"type"`
	Seq   int64           `json:"seq"`
	Scope string          `json:"scope"`
	Data  json.RawMessage `json:"data,omitempty"`
	Ts    int64           `json:"ts"`
}

type syncReqData struct {
	LastSeq int64  `json:"lastSeq"`
	Scope   string `json:"scope"`
}

type helloData struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type wsClient struct {
	conn       *websocket.Conn
	send       chan []byte
	lastPong   int64
	deviceID   string
	deviceName string
	ip         string
	userId     int64
	userType   int
	tenantId   int64
	token      string
}

type wsBindInfo struct {
	client   *wsClient
	deviceID string
}

type wsUserNotify struct {
	userType int
	tenantID int64
	userID   int64
	msgType  string
	data     interface{}
}

type wsHub struct {
	clients    map[*wsClient]struct{}
	register   chan *wsClient
	unregister chan *wsClient
	broadcast  chan []byte
	bindDevice chan wsBindInfo
	devices    map[string]*wsClient
	devicesMu  sync.RWMutex
	users      map[string]map[*wsClient]struct{}
	usersMu    sync.RWMutex
}

var (
	wsHubOnce sync.Once
	hub       *wsHub
	wsLimiter = &wsConnLimiter{
		stats: make(map[string]*wsConnStat),
	}
)

type wsConnStat struct {
	windowStart time.Time
	count       int
	lastAttempt time.Time
}

type wsConnLimiter struct {
	mu    sync.Mutex
	stats map[string]*wsConnStat
}

func startWsHub() {
	wsHubOnce.Do(func() {
		hub = &wsHub{
			clients:    make(map[*wsClient]struct{}),
			register:   make(chan *wsClient),
			unregister: make(chan *wsClient),
			broadcast:  make(chan []byte),
			bindDevice: make(chan wsBindInfo),
			devices:    make(map[string]*wsClient),
			users:      make(map[string]map[*wsClient]struct{}),
		}
		utils.RegisterWsNotify(NotifyDevicesWithType)
		utils.RegisterWsBroadcast(BroadcastAllWithType)
		utils.RegisterWsNotifyUser(NotifyUserWithType)
		go hub.run()
	})
}

func (h *wsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
			h.addUserClient(client)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.removeUserClient(client)
			if client.deviceID != "" {
				utils.RemoveOnlineDevice(client.deviceID)
				h.devicesMu.Lock()
				delete(h.devices, client.deviceID)
				h.devicesMu.Unlock()
			}
		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		case bind := <-h.bindDevice:
			if bind.deviceID == "" {
				continue
			}
			h.devicesMu.Lock()
			h.devices[bind.deviceID] = bind.client
			h.devicesMu.Unlock()
		}
	}
}

func wsUserKey(userType int, tenantID int64, userID int64) string {
	return fmt.Sprintf("%d:%d:%d", userType, tenantID, userID)
}

func (h *wsHub) addUserClient(client *wsClient) {
	if client == nil || client.userId <= 0 {
		return
	}
	key := wsUserKey(client.userType, client.tenantId, client.userId)
	h.usersMu.Lock()
	defer h.usersMu.Unlock()
	if h.users[key] == nil {
		h.users[key] = make(map[*wsClient]struct{})
	}
	h.users[key][client] = struct{}{}
}

func (h *wsHub) removeUserClient(client *wsClient) {
	if client == nil || client.userId <= 0 {
		return
	}
	key := wsUserKey(client.userType, client.tenantId, client.userId)
	h.usersMu.Lock()
	defer h.usersMu.Unlock()
	clients := h.users[key]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.users, key)
	}
}

// NotifyAll broadcasts a notification message to all connected clients.
func NotifyAll(data interface{}) {
	BroadcastAllWithType("notify", data)
}

// wsPushToBuffer assigns a seq, writes to Redis buffer, and returns the serialized payload.
func wsPushToBuffer(scope string, msg *wsMessage) []byte {
	if utils.RD == nil {
		payload, _ := json.Marshal(msg)
		return payload
	}
	ctx := context.Background()
	seq, err := utils.RD.Incr(ctx, wsMsgSeqKey+scope).Result()
	if err != nil {
		payload, _ := json.Marshal(msg)
		return payload
	}
	msg.Seq = seq
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	pipe := utils.RD.Pipeline()
	pipe.RPush(ctx, wsMsgBufKey+scope, string(payload))
	pipe.LTrim(ctx, wsMsgBufKey+scope, -int64(wsMsgBufMax), -1)
	pipe.Expire(ctx, wsMsgBufKey+scope, wsMsgBufExpire)
	_, _ = pipe.Exec(ctx)
	return payload
}

// BroadcastAllWithType broadcasts a typed message to all connected clients.
func BroadcastAllWithType(msgType string, data interface{}) {
	startWsHub()
	msg := wsMessage{
		Type:  msgType,
		Scope: wsDefaultScope,
		Data:  data,
		Ts:    time.Now().Unix(),
	}
	payload := wsPushToBuffer(wsDefaultScope, &msg)
	if payload == nil {
		return
	}
	hub.broadcast <- payload
}

// NotifyDevices sends a withdrawal task to specific devices by device ID.
func NotifyDevices(deviceIDs []string, data interface{}) {
	NotifyDevicesWithType("withdrawal", deviceIDs, data)
}

// NotifyDevicesWithType sends a task to specific devices by device ID.
func NotifyDevicesWithType(msgType string, deviceIDs []string, data interface{}) {
	startWsHub()
	msg := wsMessage{
		Type: msgType,
		Data: data,
		Ts:   time.Now().Unix(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	for _, deviceID := range deviceIDs {
		if deviceID == "" {
			continue
		}
		hub.devicesMu.RLock()
		client, ok := hub.devices[deviceID]
		hub.devicesMu.RUnlock()
		if !ok {
			continue
		}
		select {
		case client.send <- payload:
		default:
			hub.devicesMu.Lock()
			delete(hub.devices, deviceID)
			hub.devicesMu.Unlock()
		}
	}
}

func NotifyUserWithType(userType int, tenantID int64, userID int64, msgType string, data interface{}) (bool, error) {
	startWsHub()
	msg := wsMessage{
		Type: msgType,
		Data: data,
		Ts:   time.Now().Unix(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}
	key := wsUserKey(userType, tenantID, userID)
	hub.usersMu.RLock()
	clients := make([]*wsClient, 0, len(hub.users[key]))
	for client := range hub.users[key] {
		clients = append(clients, client)
	}
	hub.usersMu.RUnlock()
	if len(clients) == 0 {
		return false, nil
	}
	delivered := false
	for _, client := range clients {
		select {
		case client.send <- payload:
			delivered = true
		default:
			hub.usersMu.Lock()
			delete(hub.users[key], client)
			if len(hub.users[key]) == 0 {
				delete(hub.users, key)
			}
			hub.usersMu.Unlock()
		}
	}
	return delivered, nil
}

func bindDevice(client *wsClient, deviceID string) {
	startWsHub()
	hub.bindDevice <- wsBindInfo{client: client, deviceID: deviceID}
}

func getRequestHost(rawHost string) string {
	if host, _, err := net.SplitHostPort(rawHost); err == nil && host != "" {
		return host
	}
	if strings.Contains(rawHost, ":") {
		return strings.Split(rawHost, ":")[0]
	}
	return rawHost
}

func extractWsToken(c *gin.Context) string {
	token := strings.TrimSpace(c.Query("token"))
	if token != "" {
		return token
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
}

func wsTokenOnline(userType int, userID int64, token string) bool {
	if token == "" || utils.RD == nil {
		return false
	}
	var key string
	switch userType {
	case 1, 2, 3:
		key = utils.KeyRdOnline + utils.MD5(token)
	case 4:
		key = utils.KeyRdTenantOnline + utils.MD5(token)
	case 5:
		key = utils.KeyRdTgOnline + utils.MD5(token)
	default:
		return false
	}
	data := utils.RD.Get(context.Background(), key)
	if data != nil && data.Err() == nil && data.Val() != "" {
		return true
	}
	// 兼容旧key格式（部分租户场景用userId做hash）
	if userType == 4 {
		oldKey := utils.KeyRdTenantOnline + utils.MD5(fmt.Sprintf("%d", userID))
		oldData := utils.RD.Get(context.Background(), oldKey)
		return oldData != nil && oldData.Err() == nil && oldData.Val() == token
	}
	// 兼容旧key格式（部分tg场景用userId做hash）
	if userType == 5 {
		oldKey := utils.KeyRdTgOnline + utils.MD5(fmt.Sprintf("%d", userID))
		oldData := utils.RD.Get(context.Background(), oldKey)
		return oldData != nil && oldData.Err() == nil && oldData.Val() == token
	}
	return false
}

func validateWsToken(c *gin.Context, token string) (int64, int, int64, error) {
	if token == "" {
		return 0, 0, 0, fmt.Errorf("token is required")
	}
	accessSecret := utils.CsConfig.DefaultHost.AccessSecret
	hostInfo := utils.GetTempHostInfo(utils.GetRequestHost(c))
	if strings.TrimSpace(hostInfo.AccessSecret) != "" {
		accessSecret = hostInfo.AccessSecret
	}
	userID, userType, hostName, _, err := utils.ParseToken(accessSecret, token)
	if err != nil && accessSecret != utils.CsConfig.DefaultHost.AccessSecret {
		// 兼容历史token：尝试默认host密钥
		userID, userType, hostName, _, err = utils.ParseToken(utils.CsConfig.DefaultHost.AccessSecret, token)
		if err == nil {
			accessSecret = utils.CsConfig.DefaultHost.AccessSecret
		}
	}
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid token")
	}
	if userID <= 0 {
		return 0, 0, 0, fmt.Errorf("invalid user")
	}
	reqHost := getRequestHost(c.Request.Host)
	if hostName != "" && reqHost != "" && !strings.EqualFold(hostName, reqHost) {
		return 0, 0, 0, fmt.Errorf("token host mismatch")
	}
	if !wsTokenOnline(userType, userID, token) {
		return 0, 0, 0, fmt.Errorf("token is expired")
	}
	var tenantID int64
	if userType == 4 {
		user := utils.GetTempTenantUser(hostInfo.TablePrefix, userID)
		tenantID = user.TenantId
	}
	if userType == 5 {
		_, _, tenantID, _ = utils.ParseAppToken(accessSecret, token)
	}
	return userID, userType, tenantID, nil
}

func touchWsOnline(client *wsClient) {
	if client == nil {
		return
	}
	switch client.userType {
	case 1, 2, 3:
		utils.TouchAdminOnlineUser(client.userId)
	case 4:
		utils.TouchTenantOnlineUser(client.tenantId, client.userId)
	case 5:
		utils.TouchTgOnlineUser(client.tenantId, client.userId)
	}
}

func allowWsConnect(ip string) (bool, string) {
	now := time.Now()
	wsLimiter.mu.Lock()
	defer wsLimiter.mu.Unlock()

	// 轻量清理，避免map持续增长
	for k, v := range wsLimiter.stats {
		if now.Sub(v.lastAttempt) > 5*wsConnWindow {
			delete(wsLimiter.stats, k)
		}
	}

	stat, ok := wsLimiter.stats[ip]
	if !ok {
		wsLimiter.stats[ip] = &wsConnStat{
			windowStart: now,
			count:       1,
			lastAttempt: now,
		}
		return true, ""
	}
	if now.Sub(stat.lastAttempt) < wsReconnectMin {
		stat.lastAttempt = now
		return false, fmt.Sprintf("reconnect too fast, retry in %ds", int(wsReconnectMin/time.Second))
	}
	if now.Sub(stat.windowStart) >= wsConnWindow {
		stat.windowStart = now
		stat.count = 0
	}
	stat.count++
	stat.lastAttempt = now
	if stat.count > wsConnMax {
		return false, fmt.Sprintf("too many connections, max %d per %ds", wsConnMax, int(wsConnWindow/time.Second))
	}
	return true, ""
}

// WsHandler upgrades the connection and registers a websocket client.
func WsHandler(c *gin.Context) {
	startWsHub()
	clientIP := c.ClientIP()
	if ok, msg := allowWsConnect(clientIP); !ok {
		c.String(http.StatusTooManyRequests, msg)
		return
	}
	token := extractWsToken(c)
	userID, userType, tenantID, err := validateWsToken(c, token)
	if err != nil {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}
	server := websocket.Server{
		Handshake: func(_ *websocket.Config, _ *http.Request) error {
			return nil
		},
		Handler: func(conn *websocket.Conn) {
			client := &wsClient{
				conn:     conn,
				send:     make(chan []byte, wsSendBuffer),
				lastPong: time.Now().UnixNano(),
				ip:       clientIP,
				userId:   userID,
				userType: userType,
				tenantId: tenantID,
				token:    token,
			}
			touchWsOnline(client)
			hub.register <- client
			go client.writePump()
			client.readPump()
			hub.unregister <- client
		},
	}
	server.ServeHTTP(c.Writer, c.Request)
}

func (c *wsClient) readPump() {
	defer func() {
		_ = c.conn.Close()
	}()
	for {
		var raw string
		if err := websocket.Message.Receive(c.conn, &raw); err != nil {
			return
		}
		touchWsOnline(c)
		if raw == "" {
			continue
		}
		if raw == "pong" || raw == "ping" {
			atomic.StoreInt64(&c.lastPong, time.Now().UnixNano())
			utils.UpdateOnlineDevice(c.deviceID, c.deviceName, c.ip)
			continue
		}

		var msg wsIncoming
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			continue
		}
		if msg.Type == "pong" || msg.Type == "ping" {
			atomic.StoreInt64(&c.lastPong, time.Now().UnixNano())
			if c.deviceID == "" && len(msg.Data) > 0 {
				var hello helloData
				if err := json.Unmarshal(msg.Data, &hello); err == nil && hello.DeviceID != "" {
					c.deviceID = hello.DeviceID
					c.deviceName = hello.DeviceName
					bindDevice(c, c.deviceID)
				}
			}
			utils.UpdateOnlineDevice(c.deviceID, c.deviceName, c.ip)
			continue
		}
		if msg.Type == "hello" {
			var hello helloData
			if err := json.Unmarshal(msg.Data, &hello); err == nil {
				if hello.DeviceID != "" {
					c.deviceID = hello.DeviceID
				}
				if hello.DeviceName != "" {
					c.deviceName = hello.DeviceName
				}
				if c.deviceID != "" {
					bindDevice(c, c.deviceID)
				}
				utils.UpdateOnlineDevice(c.deviceID, c.deviceName, c.ip)
			}
			continue
		}
		if msg.Type == "ack" {
			// 客户端确认已收到 seq，更新 lastPong 保持连接活跃
			atomic.StoreInt64(&c.lastPong, time.Now().UnixNano())
			continue
		}
		if msg.Type == "sync" {
			var syncReq syncReqData
			if err := json.Unmarshal(msg.Data, &syncReq); err == nil {
				scope := syncReq.Scope
				if scope == "" {
					scope = wsDefaultScope
				}
				go c.handleSync(syncReq.LastSeq, scope)
			}
		}
	}
}

// handleSync 从 Redis 缓冲中取出 seq > lastSeq 的消息，逐条补发给客户端。
func (c *wsClient) handleSync(lastSeq int64, scope string) {
	if utils.RD == nil {
		return
	}
	ctx := context.Background()
	items, err := utils.RD.LRange(ctx, wsMsgBufKey+scope, 0, -1).Result()
	if err != nil || len(items) == 0 {
		// 缓冲已过期或为空，通知前端做 HTTP 全量刷新
		payload, _ := json.Marshal(wsMessage{Type: "sync_expired", Ts: time.Now().Unix()})
		select {
		case c.send <- payload:
		default:
		}
		return
	}

	type seqItem struct {
		seq int64
		raw []byte
	}
	var pending []seqItem
	for _, item := range items {
		var m struct {
			Seq int64 `json:"seq"`
		}
		if err := json.Unmarshal([]byte(item), &m); err == nil && m.Seq > lastSeq {
			pending = append(pending, seqItem{seq: m.Seq, raw: []byte(item)})
		}
	}
	if len(pending) == 0 {
		return
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].seq < pending[j].seq })
	for _, item := range pending {
		select {
		case c.send <- item.raw:
		default:
			return // 客户端发送缓冲满，停止补发
		}
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(wsPingInterval)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := websocket.Message.Send(c.conn, string(msg)); err != nil {
				return
			}
		case <-ticker.C:
			touchWsOnline(c)
			last := time.Unix(0, atomic.LoadInt64(&c.lastPong))
			if time.Since(last) > wsPongTimeout {
				return
			}
			ping := wsMessage{Type: "ping", Ts: time.Now().Unix()}
			payload, err := json.Marshal(ping)
			if err != nil {
				return
			}
			if err := websocket.Message.Send(c.conn, string(payload)); err != nil {
				return
			}
		}
	}
}
