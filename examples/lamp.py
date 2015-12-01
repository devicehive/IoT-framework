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
    if (name == 'SATECHILED-0'):
        print("Discovered %s (%s) " % (mac, name))
        ble.ScanStop(ignore_reply=True)
        ble.Disconnect(mac, ignore_reply=True)
        ble.Connect(mac, False, ignore_reply=True)

def device_connected(mac):
    print("Connected to %s" % (mac))    
    try:
        ble.GattWrite(mac, "fff3", "0f0d0300ffffff0000c800c8000091ffff", ignore_reply=True)
        # ble.GattWrite(mac, "fff3", "0f0d0300ffffffc800c800c8000059ffff", ignore_reply=True)
    except dbus.DBusException as e:
        print(e)

def main():
    print('Scanning ...')
    ble.connect_to_signal("PeripheralDiscovered", device_discovered)
    ble.connect_to_signal("PeripheralConnected", device_connected)
    ble.ScanStart(ignore_reply=True)

    GObject.MainLoop().run()

if __name__ == '__main__':
    main()