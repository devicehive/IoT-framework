#!/usr/bin/python3

import time, uuid
import threading
import traceback
import sys
import os
import socket
import collections

import dbus.service
from dbus.mainloop.glib import DBusGMainLoop
try:
    from gi.repository import GObject
except ImportError:
    import gobject as GObject


scriptDir = os.path.dirname(os.path.realpath(__file__))
sys.path.append( scriptDir + "/../common" )

import core, lighting as lsf

DBusGMainLoop(set_as_default=True)


# import json

DBUS_BRIDGE_NAME = 'com.devicehive.alljoyn.bridge'
DBUS_BRIDGE_PATH = '/com/devicehive/alljoyn/bridge'

DBUS_BUS_NAME = 'org.allseen.LSF.LampService'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/allseen/LSF/Lamp'

SVC_NAME = 'org.allseen.LSF.LampService'

bus = dbus.SystemBus()
bus_name = dbus.service.BusName(DBUS_BUS_NAME, bus)


class Lamp():
  def __init__(self, busname, name):

    self._id = '75b715e7e1a8411eb7c4b2719d3d0bc5' #uuid.uuid4().hex
    self._name = name

    about_props = {

      'AppId': dbus.ByteArray(bytes.fromhex(self.id)),
      'DefaultLanguage': 'en',
      'DeviceName': self.name,
      'DeviceId': self.id,
      'AppName': 'TerminalLight',
      'Manufacturer': 'DeviceHive',
      'DateOfManufacture': '2015-07-09',
      'ModelNumber': 'Simulated Light',
      'SupportedLanguages': ['en'],
      'Description': 'DeviceHive Alljoyn Bridge Device',
      'SoftwareVersion': '1.0',
      'HardwareVersion': '1.0',
      'SupportUrl': 'devicehive.com'

    }

    self._container = core.BusContainer(busname, DBUS_BUS_PATH + '/' + self.id)

    lamp_service = lsf.LampService(self._container, self.name)

    self._services = [
       core.AboutService(self._container, about_props),
       core.ConfigService(self._container, self.name),
       lamp_service

    ]

    print("Registered %s on dbus" % self.name)

  @property
  def id(self):
      return self._id

  @property
  def name(self):
      return self._name

  def publish(self, bridge):      
    
    bridge.AddService(self._container.bus.get_name(), self._container.relative('').rstrip('/'), SVC_NAME,
      # ignore_reply=True
      reply_handler=lambda id: print("ID: %s" % id),
      error_handler=lambda err: print("Error: %s" % err)
      )

    print("Published %s on bridge" % self.name)


def worker():    
    try:
        
        bridge = dbus.Interface(bus.get_object(DBUS_BRIDGE_NAME, DBUS_BRIDGE_PATH), dbus_interface='com.devicehive.alljoyn.bridge')
        time.sleep(2)

        lamp = Lamp(bus_name, 'DeviceHive Terminal Light')
        lamp.publish(bridge)
        
        return    

    except Exception as err:
        print(err)
        traceback.print_exc()
        os._exit(1)

def main():

    # init d-bus
    GObject.threads_init()    
    dbus.mainloop.glib.threads_init()
    # lamps = [LampService(mac) for mac in argv]

    # start mainloop
    loop = GObject.MainLoop()

    worker_thread = threading.Thread(target=worker,)
    worker_thread.start()

    try:
        loop.run()
    except (KeyboardInterrupt, SystemExit):
        # for lamp in lamps:
        #     lamp.deinit()
        loop.quit()
        worker_thread.join()

if __name__ == "__main__":
    main()
