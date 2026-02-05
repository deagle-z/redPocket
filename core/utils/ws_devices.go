package utils

import (
	"log"
	"sync"
	"time"
)

type OnlineDevice struct {
	DeviceID    string    `json:"device_id"`
	Name        string    `json:"name"`
	IP          string    `json:"ip"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeen    time.Time `json:"last_seen"`
}

var onlineDevices = struct {
	sync.RWMutex
	items map[string]*OnlineDevice
}{
	items: make(map[string]*OnlineDevice),
}

func UpdateOnlineDevice(deviceID, deviceName, ip string) {
	if deviceID == "" {
		return
	}
	now := time.Now()
	onlineDevices.Lock()
	defer onlineDevices.Unlock()
	if existing, ok := onlineDevices.items[deviceID]; ok {
		if deviceName != "" {
			existing.Name = deviceName
		}
		existing.IP = ip
		existing.LastSeen = now
		log.Printf("online device update: id=%s name=%s ip=%s total=%d", existing.DeviceID, existing.Name, existing.IP, len(onlineDevices.items))
		return
	}
	onlineDevices.items[deviceID] = &OnlineDevice{
		DeviceID:    deviceID,
		Name:        deviceName,
		IP:          ip,
		ConnectedAt: now,
		LastSeen:    now,
	}
	log.Printf("online device add: id=%s name=%s ip=%s total=%d", deviceID, deviceName, ip, len(onlineDevices.items))
}

func RemoveOnlineDevice(deviceID string) {
	onlineDevices.Lock()
	defer onlineDevices.Unlock()
	delete(onlineDevices.items, deviceID)
	log.Printf("online device remove: id=%s total=%d", deviceID, len(onlineDevices.items))
}

func ListOnlineDevices() []OnlineDevice {
	onlineDevices.RLock()
	defer onlineDevices.RUnlock()
	out := make([]OnlineDevice, 0, len(onlineDevices.items))
	for _, item := range onlineDevices.items {
		out = append(out, *item)
	}
	return out
}
