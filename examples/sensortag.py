#!/usr/bin/env python3

import dbus
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject
import struct, codecs

DBusGMainLoop(set_as_default=True)

def get_ble():
    bus = dbus.SystemBus()
    obj = bus.get_object("com.devicehive.bluetooth",
                         "/com/devicehive/bluetooth")
    return dbus.Interface(obj, "com.devicehive.bluetooth")

ble = get_ble()
def device_discovered(mac, name, rssi):
    print("Discovered %s (%s) %s" % (mac, name, rssi))
    if name == 'SensorTag':
        ble.ScanStop()
        ble.Connect(mac, True)

def device_connected(mac):
    print("Connected to %s" % (mac))
    try:
        ble.GattWrite(mac, "F000AA1204514000b000000000000000", "01")
        ble.GattWrite(mac, "F000AA1304514000b000000000000000", "0A")
        ble.GattNotifications(mac, "F000AA1104514000b000000000000000", True)
    except dbus.DBusException as e:
        print(e)

def device_disconnected(mac):
    print("Disconnected %s" % (mac))

def notification_received(mac, uuid, message):
    if uuid == "F000AA1104514000b000000000000000":
        x, y, z = struct.unpack('bbb', codecs.decode(message, 'hex'))
        print("MAC:%s, Accelerometer:(X:%+.3f, Y:%+.3f, Z:%+.3f)" % (mac, x/16.0, y/16.0, z/16.0))
    else:
        print("MAC:%s, UUID:%s, Received:%s" % (mac, uuid, message))

def main():
    ble.connect_to_signal("PeripheralDiscovered", device_discovered)
    ble.connect_to_signal("PeripheralConnected", device_connected)
    ble.connect_to_signal("PeripheralDisconnected", device_disconnected)
    ble.connect_to_signal("NotificationReceived", notification_received)
    ble.ScanStart()

    GObject.MainLoop().run()

if __name__ == '__main__':
    main()
