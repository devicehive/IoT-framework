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
    print("Discovered %s %s" % (name, mac)) 
    if (mac == 'e4c95cc8bbc1'):
        print("Connecting")
        ble.ScanStop()
        ble.Connect(mac)
        time.sleep(10)
        ble.ScanStart()

def device_connected(mac):
    print("Connected to %s" % (mac))
    try:
        ble.GattIndications(mac, "299d64102f6111e281c10800200c9a66", True)

    except dbus.DBusException as e:
        print(e)

def indication_received(mac, uuid, message):
    print("MAC: %s, UUID: %s, Received: %s" % (mac, uuid, message))

def main():
    ble.connect_to_signal("PeripheralDiscovered", device_discovered)
    ble.connect_to_signal("PeripheralConnected", device_connected)
    ble.connect_to_signal("IndiacationReceived", indication_received)

    ble.ScanStart()

    GObject.MainLoop().run()

if __name__ == '__main__':
    main()
