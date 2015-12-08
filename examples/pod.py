#!/usr/bin/env python3

import dbus
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject
import array
import time

DBusGMainLoop(set_as_default=True)

def get_ble():
    obj = dbus.SystemBus().get_object("com.devicehive.bluetooth", "/com/devicehive/bluetooth")
    return dbus.Interface(obj, "com.devicehive.bluetooth")

ble = get_ble()
def device_discovered(mac, name, rssi):
    print("Discovered %s" % (name)) 
    if (name == 'POD'):
        ble.ScanStop()
        time.sleep(5)
        ble.Connect(mac, False)

def device_connected(mac):
    print("Connected to %s" % (mac))
    try:
        ble.GattWrite(mac, "ffb2", "aa550f038401e0")
        ble.GattNotifications(mac, "ffb2", True)

    except dbus.DBusException as e:
        print(e)

def notification_received(mac, uuid, message):
    print("MAC: %s, UUID: %s, Received: %s" % (mac, uuid, message))

def main():
    ble.connect_to_signal("PeripheralDiscovered", device_discovered)
    ble.connect_to_signal("PeripheralConnected", device_connected)
    ble.connect_to_signal("NotificationReceived", notification_received)

    ble.ScanStart()

    GObject.MainLoop().run()

if __name__ == '__main__':
    main()
