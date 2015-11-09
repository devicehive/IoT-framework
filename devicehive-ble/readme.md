# DeviceHive BLE D-Bus Daemon

This daemon provides access to BLE peripheral devices via D-Bus.

## Consumer example:

```python
#!/usr/bin/env python

import dbus
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject
import array

DBusGMainLoop(set_as_default=True)

def get_ble():
    obj = dbus.SystemBus().get_object("com.devicehive.bluetooth", "/com/devicehive/bluetooth")
    return dbus.Interface(obj, "com.devicehive.bluetooth")

ble = get_ble()
def device_discovered(mac, name, rssi):
    if (name == 'DELIGHT'):
        ble.ScanStop()
        ble.Connect(mac)

def device_connected(mac):
    print "Connected to %s" % (mac)
    ble.GattWrite(mac, "fff1", "0f0d0300ffffffc800c800c8000059ffff")

def main():
    ble.ScanStart()
    ble.connect_to_signal("PeripheralDiscovered", device_discovered)
    ble.connect_to_signal("PeripheralConnected", device_connected)

    GObject.MainLoop().run()

if __name__ == '__main__':
    main()
```
