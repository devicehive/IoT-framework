#!/usr/bin/python3

import os, sys, time, threading, traceback
import dbus.service
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject

DBusGMainLoop(set_as_default=True)

ble = dbus.Interface(dbus.SystemBus().get_object("com.devicehive.bluetooth", '/com/devicehive/bluetooth'), "com.devicehive.bluetooth")

def device_discovered(mac, name, rssi):    
    print("Discovered %s (%s) %s" % (mac, name, rssi))
    if ((name == 'SensorTag')):        
        ble.Connect(mac)

def device_connected(mac):
    print("Connected to %s" % (mac))    
    try:
        ble.GattWrite(mac, "F000AA1204514000b000000000000000", "01")
        ble.GattWrite(mac, "F000AA1304514000b000000000000000", "0A")
        ble.GattNotifications(mac, "F000AA1104514000b000000000000000", True)

    except dbus.DBusException as e:
        print(e)

def notification_received(mac, uuid, message):
    print("MAC: %s, UUID: %s, Received: %s" % (mac, uuid, message))


ble.connect_to_signal("DeviceDiscovered", device_discovered)
ble.connect_to_signal("DeviceConnected", device_connected)
ble.connect_to_signal("NotificationReceived", notification_received)

exiting = threading.Event()
def worker():
    while not exiting.is_set():
        ble.ScanStart()
        exiting.wait(2)
        ble.ScanStop()
        exiting.wait(10)

def main():

    # init d-bus
    GObject.threads_init()    
    dbus.mainloop.glib.threads_init()

    # start mainloop
    loop = GObject.MainLoop()

    worker_thread = threading.Thread(target=worker,)
    worker_thread.start()

    try:
        loop.run()
    except (KeyboardInterrupt, SystemExit):
        exiting.set()
        loop.quit()
        worker_thread.join()

if __name__ == "__main__":
    main()