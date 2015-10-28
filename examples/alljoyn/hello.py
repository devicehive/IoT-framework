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
sys.path.append( scriptDir + "/common" )

import core, controlpanel as cp

DBusGMainLoop(set_as_default=True)


DBUS_BRIDGE_NAME = 'com.devicehive.alljoyn.bridge'
DBUS_BRIDGE_PATH = '/com/devicehive/alljoyn/bridge'

DBUS_BUS_NAME = 'com.devicehive.alljoyn.SmartHome'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/SmartHome'

SMART_PLUG_SVC = 'org.allseen.SmartHome.Hello'

bus = dbus.SystemBus()
bus_name = dbus.service.BusName(DBUS_BUS_NAME, bus)


class HelloService(core.PropertiesServiceInterface):
  def __init__(self, container):
    core.PropertiesServiceInterface.__init__(self, container, "/Service", 
      {SMART_PLUG_SVC : {'Name': 'AllJoyn'}})
    

  def IntrospectionXml(self):
    return """
        <interface name="org.allseen.SmartHome.Hello">
          <property name="Name" type="s" access="readwrite"/>
          <method name="Greet">
             <arg name="greeting" type="s" direction="out"/>
          </method>
       </interface>
    """ + core.PropertiesServiceInterface.IntrospectionXml(self)

  @dbus.service.method(SMART_PLUG_SVC, in_signature='', out_signature='s')
  def Greet(self):
    return "Hello, %s!" % self.Get(SMART_PLUG_SVC, "Name")


class Hello():
  def __init__(self, busname, name):

    self._id = uuid.uuid4().hex
    self._name = name

    about_props = {

      'AppId': dbus.ByteArray(bytes.fromhex(self.id)),
      'DefaultLanguage': 'en',
      'DeviceName': self.name,
      'DeviceId': self.id,
      'AppName': 'Hello',
      'Manufacturer': 'DeviceHive',
      'DateOfManufacture': '2015-10-28',
      'ModelNumber': 'example',
      'SupportedLanguages': ['en'],
      'Description': 'DeviceHive Alljoyn Hello Device',
      'SoftwareVersion': '1.0',
      'HardwareVersion': '1.0',
      'SupportUrl': 'devicehive.com'

    }
  
    self._container = core.BusContainer(busname, DBUS_BUS_PATH + '/' + self.id)

    self._services = [
       core.AboutService(self._container, about_props)
      # ,core.ConfigService(self._container, self.name)      
      ,core.ConfigService(self._container, self.name)
      ,HelloService(self._container)
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

    print("Published %s on bridge" % self.name)



def worker():    
    try:

        bridge = dbus.Interface(bus.get_object(DBUS_BRIDGE_NAME, DBUS_BRIDGE_PATH), dbus_interface='com.devicehive.alljoyn.bridge')
        plug = Hello(bus_name, 'Hello')
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
