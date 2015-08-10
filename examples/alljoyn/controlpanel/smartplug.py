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

    self._container = core.BusContainer(busname, DBUS_BUS_PATH + '/' + self.id)

    controlpanel = cp.ControlPanelService(self._container, self.name)

    rootcontainer = cp.ContainerService(self._container, controlpanel.relative("en"))
    rootcontainer.SetOptParam(cp.CONTAINER_METADATA_LAYOUT_HINTS, [cp.CONTAINER_LAYOUT_VERTICAL, cp.CONTAINER_LAYOUT_HORIZONTAL])
    rootcontainer.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(2003199))
    # rootcontainer.SetOptParam(cp.WIDGET_METADATA_LABEL, 

    statepropertywidget = cp.PropertyService(self._container, rootcontainer.relative("1State"))
    statepropertywidget.SetOptParam(cp.WIDGET_METADATA_LABEL, 'State')
    statepropertywidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1280))
    statepropertywidget.SetOptParam(cp.PROPERTY_METADATA_HINTS, [cp.PROPERTY_WIDGET_HINT_TEXTLABEL])
    statepropertywidget.SetValue(dbus.String("Switch Off", variant_level=2))

    controlscontainer = cp.ContainerService(self._container, rootcontainer.relative("2ControlsContainer"))
    controlscontainer.SetOptParam(cp.CONTAINER_METADATA_LAYOUT_HINTS, [cp.CONTAINER_LAYOUT_HORIZONTAL])
    controlscontainer.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(512))

    onactionwidget = cp.ActionService(self._container, controlscontainer.relative("1On"))
    onactionwidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1024))
    onactionwidget.SetOptParam(cp.WIDGET_METADATA_LABEL, "On")

    offactionwidget = cp.ActionService(self._container, controlscontainer.relative("2Off"))
    offactionwidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1024))
    offactionwidget.SetOptParam(cp.WIDGET_METADATA_LABEL, "Off")
    offactionwidget.SetStates(cp.WIDGET_STATE_DISABLED)

    def lightsOn():
      onactionwidget.SetStates(cp.WIDGET_STATE_DISABLED)
      offactionwidget.SetStates(cp.WIDGET_STATE_ENABLED)
      statepropertywidget.SetValue(dbus.String("Switch On", variant_level=2))

    def lightsOff():
      onactionwidget.SetStates(cp.WIDGET_STATE_DISABLED)
      offactionwidget.SetStates(cp.WIDGET_STATE_ENABLED)
      statepropertywidget.SetValue(dbus.String("Switch Off", variant_level=2))

    onactionwidget.SetHandler(lightsOn)
    offactionwidget.SetHandler(lightsOff)

    measurecontainer = cp.ContainerService(self._container, rootcontainer.relative("3MeasureContainer"))
    measurecontainer.SetOptParam(cp.CONTAINER_METADATA_LAYOUT_HINTS, [cp.CONTAINER_LAYOUT_VERTICAL])
    measurecontainer.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(512))
    measurecontainer.SetOptParam(cp.WIDGET_METADATA_LABEL, "Measure Properties")

    voltagepropertywidget = cp.PropertyService(self._container, measurecontainer.relative("1VoltageProperty"))
    voltagepropertywidget.SetOptParam(cp.WIDGET_METADATA_LABEL, 'Volt(V):')
    voltagepropertywidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1280))
    voltagepropertywidget.SetOptParam(cp.PROPERTY_METADATA_HINTS, [cp.PROPERTY_WIDGET_HINT_TEXTLABEL])
    voltagepropertywidget.SetValue(dbus.String("118.9194", variant_level=2))

    currentpropertywidget = cp.PropertyService(self._container, measurecontainer.relative("2CurrentProperty"))
    currentpropertywidget.SetOptParam(cp.WIDGET_METADATA_LABEL, 'Curr(A):')
    currentpropertywidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1280))
    currentpropertywidget.SetOptParam(cp.PROPERTY_METADATA_HINTS, [cp.PROPERTY_WIDGET_HINT_TEXTLABEL])
    currentpropertywidget.SetValue(dbus.String("0.0000", variant_level=2))

    requencypropertywidget = cp.PropertyService(self._container, measurecontainer.relative("3FrequencyProperty"))
    requencypropertywidget.SetOptParam(cp.WIDGET_METADATA_LABEL, 'Freq(Hz):')
    requencypropertywidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1280))
    requencypropertywidget.SetOptParam(cp.PROPERTY_METADATA_HINTS, [cp.PROPERTY_WIDGET_HINT_TEXTLABEL])
    requencypropertywidget.SetValue(dbus.String("60.0", variant_level=2))

    powerpropertywidget = cp.PropertyService(self._container, measurecontainer.relative("4PowerProperty"))
    powerpropertywidget.SetOptParam(cp.WIDGET_METADATA_LABEL, 'Watt(W):')
    powerpropertywidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1280))
    powerpropertywidget.SetOptParam(cp.PROPERTY_METADATA_HINTS, [cp.PROPERTY_WIDGET_HINT_TEXTLABEL])
    powerpropertywidget.SetValue(dbus.String("0.0000", variant_level=2))

    accumengpropertywidget = cp.PropertyService(self._container, measurecontainer.relative("5AccumulateEnergy"))
    accumengpropertywidget.SetOptParam(cp.WIDGET_METADATA_LABEL, 'ACCU(KWH):')
    accumengpropertywidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1280))
    accumengpropertywidget.SetOptParam(cp.PROPERTY_METADATA_HINTS, [cp.PROPERTY_WIDGET_HINT_TEXTLABEL])
    accumengpropertywidget.SetValue(dbus.String("0.0000", variant_level=2))

    pwrfactorpropertywidget = cp.PropertyService(self._container, measurecontainer.relative("6PowerFactorProperty"))
    pwrfactorpropertywidget.SetOptParam(cp.WIDGET_METADATA_LABEL, 'PF(%):')
    pwrfactorpropertywidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1280))
    pwrfactorpropertywidget.SetOptParam(cp.PROPERTY_METADATA_HINTS, [cp.PROPERTY_WIDGET_HINT_TEXTLABEL])
    pwrfactorpropertywidget.SetValue(dbus.String("000", variant_level=2))

    getpropsactionwidget = cp.ActionService(self._container, measurecontainer.relative("7GetProperties"))
    getpropsactionwidget.SetOptParam(cp.WIDGET_METADATA_BGCOLOR, dbus.UInt32(1024))
    getpropsactionwidget.SetOptParam(cp.WIDGET_METADATA_LABEL, "Get Properties")


    self._services = [
       core.AboutService(self._container, about_props)
      ,core.ConfigService(self._container, self.name)
      , controlpanel
      , rootcontainer      
      , controlscontainer, offactionwidget, onactionwidget
      , measurecontainer, voltagepropertywidget, currentpropertywidget, requencypropertywidget, 
        powerpropertywidget, accumengpropertywidget, pwrfactorpropertywidget, getpropsactionwidget
      , statepropertywidget
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
