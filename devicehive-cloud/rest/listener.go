package rest

import "log"

type DeviceNotificationHandler func(n DeviceNotification)

type DeviceNotificationListener struct {
	deviceHiveURL, deviceID string
	handler                 DeviceNotificationHandler
}

func NewDeviceNotificationListener(deviceHiveURL, deviceID string, handler DeviceNotificationHandler) DeviceNotificationListener {
	return DeviceNotificationListener{
		deviceHiveURL: deviceHiveURL,
		deviceID:      deviceID,
		handler:       handler,
	}
}

func (l *DeviceNotificationListener) Run() {
	l.listen()
}

//TODO: add listening interruption
func (l *DeviceNotificationListener) listen() {
	DeviceNotificationPollAsync(l.deviceHiveURL, l.deviceID, []Parameter{}, func(n DeviceNotification, err error, interrupted bool) {
		switch {
		case interrupted:
			return
		case err != nil:
			log.Printf("DeviceNotificationPollAsync error: %s", err.Error())
		default:
			if l.handler != nil {
				l.handler(n)
			}
		}

		l.listen()
	})
}
