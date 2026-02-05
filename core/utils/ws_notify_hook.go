package utils

import "errors"

var wsNotifyDevices func(string, []string, interface{})

// RegisterWsNotify sets the notifier used to send messages to ws devices.
func RegisterWsNotify(fn func(string, []string, interface{})) {
	wsNotifyDevices = fn
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
