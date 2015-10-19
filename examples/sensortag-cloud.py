#!/usr/bin/python3

import os, sys, time, threading, traceback
import struct, math, json
import dbus.service
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject

DBusGMainLoop(set_as_default=True)

ble = dbus.Interface(dbus.SystemBus().get_object("com.devicehive.bluetooth", '/com/devicehive/bluetooth'), "com.devicehive.bluetooth")
cloud = dbus.Interface(dbus.SystemBus().get_object("com.devicehive.cloud", '/com/devicehive/cloud'), "com.devicehive.cloud")

sensors = {}
DEFAULT_PRIORITY = 100

def accelerometer_handler(mac, uuid, message):
    data = bytearray.fromhex(message)
    data = struct.unpack('<3b', data)

    x = data[0] / 64.0
    y = data[1] / 64.0
    z = data[2] / -64.0

    absolute = math.sqrt(x*x + y*y + z*z)
    return {'x': x, 'y': y, 'z': z, 'abs': absolute}

def accelerometer_CC2650_handler(mac, uuid, message):
    data = bytearray.fromhex(message)
    data = struct.unpack('<3h', data[6:12])

    x = data[0] * 2.0 / 32768.0
    y = data[1] * 2.0 / 32768.0
    z = data[2] * 2.0 / 32768.0

    absolute = math.sqrt(x*x + y*y + z*z)
    return {'x': x, 'y': y, 'z': z, 'abs': absolute}


init = {
    'SensorTag' : {
        'F000AA1104514000b000000000000000' : {
            'notification' : 'accelerometer',
            'write': [('F000AA1204514000b000000000000000', '01'), 
                      ('F000AA1304514000b000000000000000', '20')],
            'handler': accelerometer_handler
        },  
    },
    'CC2650 SensorTag' : {
        'F000AA8104514000b000000000000000' : {
            'notification' : 'accelerometer',
            'write': [('F000AA8204514000b000000000000000', '3800'), 
                      ('F000AA8304514000b000000000000000', '20')],
            'handler': accelerometer_CC2650_handler
            },
        }
}


def device_discovered(mac, name, rssi):    
    if 'SensorTag' not in name: return
    print("Discovered %s (%s) %s" % (mac, name, rssi))
    if mac in sensors and sensors[mac][1]: return

    sensors[mac] = (name, False)
    ble.Connect(mac, False, ignore_reply = True)

def device_connected(mac):    
    if mac not in sensors: return
    print("Connected: %s (%s)" % (sensors[mac][0], mac))
    name = sensors[mac][0]
    sensors[mac] = (name, True)

    if name not in init: 
        print("Could not find init config for %s" % name)
        return

    for char, config in init[name].items():
        print("Configuring: %s (%s) - %s" % (name, mac, char))
        for charname, value in config['write']:
            ble.GattWrite(mac, charname, value, ignore_reply = True)
        ble.GattNotifications(mac, char, True, ignore_reply = True)

def device_disconnected(mac):
    if mac not in sensors: return
    print("Disconnected: %s (%s)" % (sensors[mac][0], mac))
    del sensors[mac]
    pass

def notification_received(mac, uuid, message):
    if mac not in sensors: return # unknown device
    name = sensors[mac][0]
    if name not in init or uuid not in init[name]: return # no handlers for this notification

    handler = init[name][uuid]['handler']
    result = handler(mac, uuid, message)

    notification = init[name][uuid]['notification']

    print("MAC: %s, UUID: %s => %s" % (mac, uuid, result))
    cloud.SendNotification(notification, json.dumps({
        'SensorTag': mac,
        'Value': result
        }), DEFAULT_PRIORITY
        , error_handler=lambda err: print(err)
        , reply_handler=lambda *args: None)


ble.connect_to_signal("PeripheralDiscovered", device_discovered)
ble.connect_to_signal("PeripheralConnected", device_connected)
ble.connect_to_signal("PeripheralDisconnected", device_disconnected)
ble.connect_to_signal("NotificationReceived", notification_received)

exiting = threading.Event()
def worker():
    while not exiting.is_set():
        ble.ScanStart()
        exiting.wait(5)
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