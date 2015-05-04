package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	cloud, err := NewdbusWrapper("/com/devicehive/cloud", "com.devicehive.cloud")
	if err != nil {
		log.Panic(err)
	}

	ble, err := NewdbusWrapper("/com/devicehive/bluetooth", "com.devicehive.bluetooth")
	if err != nil {
		log.Panic(err)
	}

	ble.RegisterHandler("PeripheralDiscovered", 100, func(args ...interface{}) {
		cloud.SendNotification("PeripheralDiscovered", map[string]interface{}{
			"mac":  args[0].(string),
			"name": args[1].(string),
			"rssi": args[2].(int16),
		})
	})

	ble.RegisterHandler("PeripheralConnected", 100, func(args ...interface{}) {
		cloud.SendNotification("PeripheralConnected", map[string]interface{}{
			"mac": args[0].(string),
		})
	})

	ble.RegisterHandler("PeripheralDisconnected", 100, func(args ...interface{}) {
		cloud.SendNotification("PeripheralDisconnected", map[string]interface{}{
			"mac": args[0].(string),
		})
	})

	ble.RegisterHandler("NotificationReceived", 1, func(args ...interface{}) {
		cloud.SendNotification("NotificationReceived", map[string]interface{}{
			"mac":   args[0].(string),
			"uuid":  args[1].(string),
			"value": args[2].(string),
		})
	})

	ble.RegisterHandler("IndicationReceived", 1, func(args ...interface{}) {
		cloud.SendNotification("IndicationReceived", map[string]interface{}{
			"mac":   args[0].(string),
			"uuid":  args[1].(string),
			"value": args[2].(string),
		})
	})

	cloudHandlers := make(map[string]cloudCommandHandler)

	cloudHandlers["connect"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		return nil, ble.BleConnect(p["mac"].(string))
	}

	cloudHandlers["scan/start"] = func(map[string]interface{}) (map[string]interface{}, error) {
		return nil, ble.BleScanStart()
	}

	cloudHandlers["scan/stop"] = func(map[string]interface{}) (map[string]interface{}, error) {
		ble.BleScanStop()
		return nil, ble.BleScanStop()
	}

	cloudHandlers["gatt/read"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		return ble.BleGattRead(p["mac"].(string), p["uuid"].(string))
	}

	cloudHandlers["gatt/write"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattWrite(p["mac"].(string), p["uuid"].(string), p["value"].(string))
		return nil, err
	}

	cloudHandlers["gatt/notifications"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattNotifications(p["mac"].(string), p["uuid"].(string), true)
		return nil, err
	}

	cloudHandlers["gatt/notifications/stop"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattNotifications(p["mac"].(string), p["uuid"].(string), false)
		return nil, err
	}

	cloudHandlers["gatt/indications"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattIndications(p["mac"].(string), p["uuid"].(string), true)
		return nil, err
	}

	cloudHandlers["gatt/indications/stop"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattIndications(p["mac"].(string), p["uuid"].(string), false)
		return nil, err
	}

	cloud.RegisterHandler("CommandReceived", 1, func(args ...interface{}) {
		id := args[0].(uint32)
		command := args[1].(string)
		params := args[2].(string)

		var dat map[string]interface{}
		b := []byte(params)
		json.Unmarshal(b, &dat)

		if h, ok := cloudHandlers[command]; ok {
			res, err := h(dat)

			if err != nil {
				cloud.CloudUpdateCommand(id, fmt.Sprintf("ERROR: %s", err.Error()), nil)
			} else {
				cloud.CloudUpdateCommand(id, "success", res)
			}

		} else {
			log.Printf("Unhandled command: %s", command)
		}
	})

	select {}
}
