package common

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
	"net/http"
)

const (
	wsPingInterval = 25 * time.Second
	wsPongTimeout  = 60 * time.Second
	wsSendBuffer   = 64
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
)

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

// WsHandler upgrades the connection and registers a websocket client.
func WsHandler(c *gin.Context) {
	startWsHub()
	clientIP := c.ClientIP()
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
