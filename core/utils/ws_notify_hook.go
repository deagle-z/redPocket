package utils

import "errors"

var wsNotifyDevices func(string, []string, interface{})
var wsBroadcastAll func(string, interface{})
var wsNotifyUser func(int, int64, int64, string, interface{}) (bool, error)

// RegisterWsNotify sets the notifier used to send messages to ws devices.
func RegisterWsNotify(fn func(string, []string, interface{})) {
	wsNotifyDevices = fn
}

// RegisterWsBroadcast sets the notifier used to broadcast messages to all ws clients.
func RegisterWsBroadcast(fn func(string, interface{})) {
	wsBroadcastAll = fn
}

// RegisterWsNotifyUser sets the notifier used to send messages to a specific online user.
func RegisterWsNotifyUser(fn func(int, int64, int64, string, interface{}) (bool, error)) {
	wsNotifyUser = fn
}

// SendWsTask dispatches data to the given device IDs via websocket.
func SendWsTask(deviceIDs []string, data interface{}) error {
	return SendWsTaskWithType("withdrawal", deviceIDs, data)
}

// SendWsTaskWithType dispatches data to the given device IDs via websocket.
func SendWsTaskWithType(msgType string, deviceIDs []string, data interface{}) error {
	if wsNotifyDevices == nil {
		return errors.New("ws notifier not ready")
	}
	wsNotifyDevices(msgType, deviceIDs, data)
	return nil
}

// BroadcastWsWithType dispatches data to all websocket clients.
func BroadcastWsWithType(msgType string, data interface{}) error {
	if wsBroadcastAll == nil {
		return errors.New("ws broadcast notifier not ready")
	}
	wsBroadcastAll(msgType, data)
	return nil
}

// SendWsUserWithType dispatches data to a specific online user.
func SendWsUserWithType(userType int, tenantID int64, userID int64, msgType string, data interface{}) (bool, error) {
	if wsNotifyUser == nil {
		return false, errors.New("ws user notifier not ready")
	}
	return wsNotifyUser(userType, tenantID, userID, msgType, data)
}
