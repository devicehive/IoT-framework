#!/usr/bin/python3

import dbus
import time
import sys
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject
import json

DBUS_BUS_ENOCEAN_NAME = 'com.devicehive.enocean'
DBUS_BUS_BLE_NAME = 'com.devicehive.bluetooth'

SWITCH_ADDRESS = '00:2A:1A:B8'
BULB_ADDRESS = '20:C3:8F:F5:49:B4'

bulb_on_value = '0f0d0300ffffffc800c800c8000059ffff'
bulb_off_value = '0f0d0300ffffff0000c800c8000091ffff'
bulb_handle = 0x002b

DBusGMainLoop(set_as_default=True)


def init_bulb():
    bus = dbus.SystemBus()
    ble_manager = bus.get_object(DBUS_BUS_BLE_NAME, '/')
    bulb_path = ble_manager.Create(BULB_ADDRESS, dbus_interface='com.devicehive.BluetoothManager')
    return bus.get_object(DBUS_BUS_BLE_NAME, bulb_path)

bulb = init_bulb()


def turn_bulb(on):
    if on:
        print('ON')
        bulb.Write(bulb_handle, bulb_on_value, False, dbus_interface='com.devicehive.BluetoothDevice')
    else:
        print('OFF')
        bulb.Write(bulb_handle, bulb_off_value, False, dbus_interface='com.devicehive.BluetoothDevice')


def message_received(value):
    res = json.loads(value)
    if res['sender'] == SWITCH_ADDRESS:
        if res['R1']['raw_value'] == 2:
            turn_bulb(True)

        if res['R1']['raw_value'] == 3:
            turn_bulb(False)


def main():
    bus = dbus.SystemBus()
    enocean_manager = bus.get_object(DBUS_BUS_ENOCEAN_NAME, '/com/devicehive/enocean')
    enocean = dbus.Interface(enocean_manager, DBUS_BUS_ENOCEAN_NAME)
    enocean.connect_to_signal('message_received', message_received)

    try:
        GObject.MainLoop().run()
    except KeyboardInterrupt:
        pass


if __name__ == '__main__':
    main()