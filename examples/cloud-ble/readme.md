Example of using BLE and cloud parts of IoT framework. This is a simple wrapper
around GATT protocol. You can connect to any BLE device with GATT implementation
from DeviceHive server. This demo shows how to implement it. It awaits for commands
from cloud and perfoms it. This demo can be used as dedicated application if you
need to control BLE device from cloud.

### Building
Run `go build`

### Usage
Run application, then you can send commands from cloud. Command list:

- `connect` - connect to BLE device
- `scan/start` - start scanning, on each found device app will send `PeripheralDiscovered` notification
- `scan/stop` - stop scanning
- `gatt/read` - read characteristic
- `gatt/write` - write characteristic
- `gatt/notifications` - subscribe on characteristic notification. On each BLE notification `NotificationReceived` cloud notification will be sent
- `gatt/notifications/stop` - stop watching on characteristic notification
- `gatt/indications` - subscribe on characteristic indication. On each BLE indication `IndicationReceived` cloud notification will be sent
- `gatt/indications/stop` - stop watching on characteristic indication

You can receive these notfications from app to cloud:

- `PeripheralDiscovered` - each time when app founds new BLE device in air
- `PeripheralConnected` - when app succefuly connects to BLE device
- `PeripheralDisconnected` - when app lost connection with BLE device, including normal disconnect.
- `NotificationReceived` - on each BLE notification
- `IndicationReceived` - on each BLE indication
