#!/usr/bin/python3

import time, uuid
import threading
import traceback
import sys
import os
import socket
import json
import collections
import dbus.service
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject

scriptDir = os.path.dirname(os.path.realpath(__file__))
sys.path.append( scriptDir + "/../common" )

import core

DBusGMainLoop(set_as_default=True)

bulb_on_value = '0f0d0300ffffffc800c800c8000059ffff'
bulb_off_value = '0f0d0300ffffff0000c800c8000091ffff'
bulb_handle = 'fff3'

DBUS_BUS_NAME = 'com.devicehive.alljoyn.SmartHome'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/SmartHome'

LIGHTS_SVC = 'org.allseen.SmartHome.Lights'
SWITCH_SVC = 'org.allseen.SmartHome.Switch'

bus = dbus.SystemBus()
bus_name = dbus.service.BusName(DBUS_BUS_NAME, bus)

bridge = dbus.Interface(dbus.SystemBus().get_object('com.devicehive.alljoyn.bridge', '/com/devicehive/alljoyn/bridge'), 'com.devicehive.alljoyn.bridge')
ble = dbus.Interface(dbus.SystemBus().get_object('com.devicehive.bluetooth', '/com/devicehive/bluetooth'), 'com.devicehive.bluetooth')
enocean = dbus.Interface(dbus.SystemBus().get_object('com.devicehive.enocean', '/com/devicehive/enocean'), 'com.devicehive.enocean')

class HelloService(core.PropertiesServiceInterface):
  def __init__(self, container, lightsHandler):
    core.PropertiesServiceInterface.__init__(self, container, "/Service", 
      {})

    self._lightsHandler = lightsHandler

  def IntrospectionXml(self):
    return """
        <interface name="org.allseen.SmartHome.Lights">
          <method name="LightsOn">
          </method>
          <method name="LightsOff">            
          </method>
       </interface>
        <interface name="org.allseen.SmartHome.Switch">
          <signal name="SwitchOn">
            <annotation name="com.devicehive.alljoyn.signal" value="sessionless"/>
          </signal>
          <signal name="SwitchOff">
            <annotation name="com.devicehive.alljoyn.signal" value="sessionless"/>
          </signal>
       </interface>
    """ + core.PropertiesServiceInterface.IntrospectionXml(self)

  @dbus.service.method(LIGHTS_SVC, in_signature='', out_signature='')
  def LightsOn(self):
    self._lightsHandler(True)

  @dbus.service.method(LIGHTS_SVC, in_signature='', out_signature='')
  def LightsOff(self):
    self._lightsHandler(False)

  @dbus.service.signal(SWITCH_SVC, signature='')
  def SwitchOn(self):
      print("SwitchOn")


  @dbus.service.signal(SWITCH_SVC, signature='')
  def SwitchOff(self):
      print("SwitchOff")


class Hello():
  def __init__(self, busname, name):

    self._id = '17f3c67c2dea4ea5869a65b751fb5150'; uuid.uuid4().hex
    self._name = name
    self._lamps = set()

    about_props = {

      'AppId': dbus.ByteArray(bytes.fromhex(self.id)),
      'DefaultLanguage': 'en',
      'DeviceName': self.name,
      'DeviceId': self.id,
      'AppName': 'Enocean-AllJoyn-BLE',
      'Manufacturer': 'DeviceHive',
      'DateOfManufacture': '2015-10-28',
      'ModelNumber': 'example',
      'SupportedLanguages': ['en'],
      'Description': 'Example of using BLE and Enocean with AllJoyn EventActions interface.',
      'SoftwareVersion': '1.0',
      'HardwareVersion': '1.0',
      'SupportUrl': 'devicehive.com'

    }
  
    self._container = core.BusContainer(busname, DBUS_BUS_PATH + '/' + self.id)
    self._service = HelloService(self._container, self.lights)

    self._services = [
       core.AboutService(self._container, about_props), self._service
    ]

    print("Registered %s on dbus" % self.name)

  def lights(self, state):
    for mac in self._lamps:
      print("Turning Light %s: %s" % ('On' if state else 'Off', mac))
      ble.GattWrite(mac, bulb_handle, bulb_on_value if state else bulb_off_value, ignore_reply=True)

  @property
  def id(self):
      return self._id

  @property
  def name(self):
      return self._name

  def publish(self, bridge):      
    service = self._services[0]
    bridge.AddService(self._container.bus.get_name(), 
      self._container.relative('').rstrip('/'), LIGHTS_SVC,
      ignore_reply=True
      )

    print("Published %s on bridge" % self.name)

  def device_connected(self, mac):
    print("BLE: Connected to %s" % (mac))    
    self._lamps.add(str(mac))

  def device_discovered(self, mac, name, rssi):
    print("Discovered %s (%s) " % (mac, name))
    if (name != None and name.startswith('SATECHILED')):    
      ble.Connect(str(mac), False, ignore_reply=True)
      self._lamps.add(str(mac))

  def device_disconnected(self, mac):
    print("BLE: Disconnected to %s" % (mac))    
    if str(mac) in self._lamps:
      self._lamps.remove(str(mac))

  def enocean_received(self, value):
      res = json.loads(value)
      # print(res)
      # if res['sender'] == SWITCH_ADDRESS:
      if res['R1']['raw_value'] == 2:
          self._service.SwitchOn()

      if res['R1']['raw_value'] == 3:
          self._service.SwitchOff()


plug = Hello(bus_name, 'Enocean-AllJoyn-BLE')


exiting = threading.Event()
def worker():    
    try:
        
        plug.publish(bridge)
        
        ble.connect_to_signal("PeripheralDiscovered", plug.device_discovered)
        # ble.connect_to_signal("PeripheralConnected", plug.device_connected)
        ble.connect_to_signal("PeripheralDisconnected", plug.device_disconnected)

        enocean.connect_to_signal('message_received', plug.enocean_received)

        while not exiting.is_set():
            ble.ScanStart()
            exiting.wait(5)
            ble.ScanStop()
            exiting.wait(10)

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
        exiting.set()
        loop.quit()
        worker_thread.join()

if __name__ == "__main__":
    main()
