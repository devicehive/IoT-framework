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

import core, controlpanel as cp

DBusGMainLoop(set_as_default=True)


# import json

SATECHI_NAMES = ['DELIGHT', 'SATECHI-1']
SATECHI_COLOR_CHAR = 'fff3'

DBUS_BRIDGE_NAME = 'com.devicehive.alljoyn.bridge'
DBUS_BRIDGE_PATH = '/com/devicehive/alljoyn/bridge'

DBUS_BUS_NAME = 'com.devicehive.alljoyn.SmartHome'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/SmartHome'

SMART_PLUG_SVC = 'org.allseen.SmartHome.SmartPlug'

bus = dbus.SystemBus()
bus_name = dbus.service.BusName(DBUS_BUS_NAME, bus)

class BusContainer(object):
  def __init__(self, bus, root):
    self._bus = bus
    self._root = root.rstrip('/')

  @property
  def bus(self):
    return self._bus

  def relative(self, path):
    return "%s/%s" % (self._root, path.lstrip('/'))



class SmartPlug():
  def __init__(self, busname, name):

    self._id = 'c50ded5d5dfc4de28eb296e921d9a6e2' #uuid.uuid4().hex
    self._name = name

    about_props = {

      'AppId': dbus.ByteArray(bytes.fromhex(self.id)),
      'DefaultLanguage': 'en',
      'DeviceName': self.name,
      'DeviceId': self.id,
      'AppName': 'Controlee',
      'Manufacturer': 'DeviceHive',
      'DateOfManufacture': '2015-07-09',
      'ModelNumber': 'smart Plug',
      'SupportedLanguages': ['en'],
      'Description': 'DeviceHive Alljoyn Bridge Device',
      'SoftwareVersion': '1.0',
      'HardwareVersion': '1.0',
      'SupportUrl': 'devicehive.com',
      'AJSoftwareVersion': '14.06.00a Tag "v14.06.00a"'

    }

    self._container = BusContainer(busname, DBUS_BUS_PATH + '/' + self.id)

    self._services = [
       core.AboutService(self._container, about_props)
      ,core.ConfigService(self._container, self.name)
      ,cp.ControlPanelService(self._container, self.name)
      ,cp.ContainerService(self._container, self.name, 'en', 'ROOT CONTAINER')
      ,cp.PropertyService(self._container, self.name, 'en/State', 'State')
    ]

    print("Registered %s on dbus" % self.name)

  @property
  def id(self):
      return self._id

  @property
  def name(self):
      return self._name

  def publish(self, bridge):      
    service = self._services[0]
    bridge.AddService(self._container.bus.get_name(), self._container.relative('').rstrip('/'), SMART_PLUG_SVC,
      # ignore_reply=True
      reply_handler=lambda id: print("ID: %s" % id),
      error_handler=lambda err: print("Error: %s" % err)
      )

    # for service in self._services:
    #   for interface in service.exports:
    #     bridge.AddService(service.object_path, self._container.bus.get_name(), service.path, interface, service.introspect())


    # bridge.StartAllJoyn(self._container.bus.get_name())

    print("Published %s on bridge" % self.name)



def worker():    
    try:
        
        bridge = dbus.Interface(bus.get_object(DBUS_BRIDGE_NAME, DBUS_BRIDGE_PATH), dbus_interface='com.devicehive.alljoyn.bridge')
        time.sleep(2)

        plug = SmartPlug(bus_name, 'ACPlug')
        plug.publish(bridge)
        
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
