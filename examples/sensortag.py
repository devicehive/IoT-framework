#!/usr/bin/env python3

import dbus
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject
import array

DBusGMainLoop(set_as_default=True)

def get_ble():
    return dbus.Interface(dbus.SystemBus().get_object("com.devicehive.bluetooth", '/com/devicehive/bluetooth'), "com.devicehive.bluetooth")

ble = get_ble()
def device_discovered(mac, name, rssi):
    print("Discovered %s (%s) %s" % (mac, name, rssi))
    if ((name == 'SensorTag')):
        #ble.ScanStop()
        ble.Connect(mac)

def device_connected(mac):
    print("Connected to %s" % (mac))    
    try:
        ble.GattWrite(mac, "F000AA1204514000b000000000000000", "01")
        ble.GattWrite(mac, "F000AA1304514000b000000000000000", "0A")
        ble.GattNotifications(mac, "F000AA1104514000b000000000000000")

    except dbus.DBusException as e:
        print(e)

def notification_received(mac, uuid, message):
    print("MAC: %s, UUID: %s, Received: %s" % (mac, uuid, message))

def main():
    ble.ScanStart()
    ble.connect_to_signal("DeviceDiscovered", device_discovered)
    ble.connect_to_signal("DeviceConnected", device_connected)
    ble.connect_to_signal("NotificationReceived", notification_received)

    GObject.MainLoop().run()

if __name__ == '__main__':
    main()