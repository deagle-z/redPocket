package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
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
)

type wsMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
	Ts   int64       `json:"ts"`
}

type wsIncoming struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	Ts   int64           `json:"ts"`
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
	token      string
}

type wsBindInfo struct {
	client   *wsClient
	deviceID string
}

type wsHub struct {
	clients    map[*wsClient]struct{}
	register   chan *wsClient
	unregister chan *wsClient
	broadcast  chan []byte
	bindDevice chan wsBindInfo
	devices    map[string]*wsClient
	devicesMu  sync.RWMutex
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
		}
		utils.RegisterWsNotify(NotifyDevicesWithType)
		go hub.run()
	})
}

func (h *wsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
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

// NotifyAll broadcasts a notification message to all connected clients.
func NotifyAll(data interface{}) {
	startWsHub()
	msg := wsMessage{
		Type: "notify",
		Data: data,
		Ts:   time.Now().Unix(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws notify marshal error: %v", err)
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
		log.Printf("ws notify marshal error: %v", err)
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
	return false
}

func validateWsToken(c *gin.Context, token string) (int64, int, error) {
	if token == "" {
		return 0, 0, fmt.Errorf("token is required")
	}
	userID, userType, hostName, _, err := utils.ParseToken(utils.CsConfig.DefaultHost.AccessSecret, token)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid token")
	}
	if userID <= 0 {
		return 0, 0, fmt.Errorf("invalid user")
	}
	reqHost := getRequestHost(c.Request.Host)
	if hostName != "" && reqHost != "" && !strings.EqualFold(hostName, reqHost) {
		return 0, 0, fmt.Errorf("token host mismatch")
	}
	if !wsTokenOnline(userType, userID, token) {
		return 0, 0, fmt.Errorf("token is expired")
	}
	return userID, userType, nil
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
	userID, userType, err := validateWsToken(c, token)
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
				token:    token,
			}
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
		log.Printf("websocket raw: %s", raw)
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
				log.Printf("websocket hello: device_id=%s device_name=%s ip=%s", c.deviceID, c.deviceName, c.ip)
				if c.deviceID != "" {
					bindDevice(c, c.deviceID)
				}
				utils.UpdateOnlineDevice(c.deviceID, c.deviceName, c.ip)
			}
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
			if err := websocket.Message.Send(c.conn, msg); err != nil {
				return
			}
		case <-ticker.C:
			last := time.Unix(0, atomic.LoadInt64(&c.lastPong))
			if time.Since(last) > wsPongTimeout {
				return
			}
			ping := wsMessage{Type: "ping", Ts: time.Now().Unix()}
			payload, err := json.Marshal(ping)
			if err != nil {
				return
			}
			if err := websocket.Message.Send(c.conn, payload); err != nil {
				return
			}
		}
	}
}
