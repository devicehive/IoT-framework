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

DBusGMainLoop(set_as_default=True)


# import json

SATECHI_NAMES = ['DELIGHT', 'SATECHI-1']
SATECHI_COLOR_CHAR = 'fff3'

DBUS_BRIDGE_NAME = 'com.devicehive.alljoyn.bridge'
DBUS_BRIDGE_PATH = '/com/devicehive/alljoyn/bridge'

DBUS_BUS_NAME = 'com.devicehive.alljoyn.SmartHome'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/SmartHome'

ALLJOYN_SMART_TOILET_NAME = 'org.allseen.SmartHome.ToiletService'
ALLJOYN_SMART_TOILET_PATH = '/org/allseen/SmartHome/Toilet'

ALLJOYN_CONFIG_NAME = 'org.alljoyn.Config'
ALLJOYN_CONFIG_PATH = '/Config'

ALLJOYN_CONTROLPANEL_NAME = 'org.alljoyn.ControlPanel.ControlPanel'
ALLJOYN_CONTROLPANEL_PATH = '/ControlPanel/DHBulb/rootContainer'

ABOUT_IFACE = 'org.alljoyn.About'
CONFIG_SERVICE_IFACE = 'org.alljoyn.Config'
SMART_TOILET_IFACE = 'org.allseen.SmartHome.ToiletService'
CONTROLPANEL_SERVICE_IFACE = 'org.alljoyn.ControlPanel.ControlPanel'

bus = dbus.SystemBus()
bus_name = dbus.service.BusName(DBUS_BUS_NAME, bus)

def flatten(d, parent_key='', sep='.'):
    items = []
    for k, v in d.items():
        new_key = parent_key + sep + k if parent_key else k
        try:
            items.extend(flatten(v, new_key, sep=sep).items())
        except:
            items.append((new_key, v))
    return dict(items)

class BusContainer(object):
  def __init__(self, bus, root):
    self._bus = bus
    self._root = root.rstrip('/')

  @property
  def bus(self):
    return self._bus

  def relative(self, path):
    return "%s/%s" % (self._root, path.lstrip('/'))


class ServiceInterface(dbus.service.Object):
  def __init__(self, container, path, exports):
    self._path = path
    self._exports = exports
    dbus.service.Object.__init__(self, container.bus, container.relative(path))

  @property
  def path(self):
      return self._path

  def introspect(self):    
    introspection = self.Introspect(self._object_path, self._connection)
    # print(introspection)
    return introspection

  @property
  def object_path(self):
      return self._object_path

  @property
  def exports(self):
      return self._exports
  
  

class PropertiesServiceInterface(ServiceInterface):
  def __init__(self, container, path, exports, properties):
    self._path = path
    self._properties = flatten(properties)
    ServiceInterface.__init__(self, container, path, exports)

  ## dbus.PROPERTIES_IFACE
  @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ss', out_signature='v')
  def Get(self, interface, property):
      prop = interface + '.' + property
      print("Properties.Get is called %s" % prop)
      if prop in self._properties:
          return self._properties[prop]
      else:
          raise Exception('Unsupported property: %s.%s' % (interface, prop))

  @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ssv')
  def Set(self, interface, property, value):
      prop = interface + '.' + property
      print("Properties.Set is called %s with %s" % (prop, value))
      if prop in self._properties:
          self._properties[prop] = value
      else:
          raise Exception('Unsupported property: %s.%s' % (interface, prop))

  @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='s', out_signature='a{sv}')
  def GetAll(self, interface):
      prefix = interface + '.'
      return  {k[len(prefix):]: v for k, v in self._properties.items() if k.startswith(prefix)}

class AboutService(PropertiesServiceInterface):
  def __init__(self, container, properties):
    PropertiesServiceInterface.__init__(self, container, '/About', 
      ['org.alljoyn.About'],
      {'org.alljoyn.About' : properties})

  ## org.alljoyn.About Interface

  @dbus.service.method(ABOUT_IFACE, in_signature='s', out_signature='a{sv}')
  def GetAboutData(self, languageTag):
      print("GetAboutData is called")
      return PropertiesServiceInterface.GetAll(self, ABOUT_IFACE)

  @dbus.service.method(ABOUT_IFACE, in_signature='', out_signature='a(oas)')
  def GetObjectDescription(self):
      print('GetObjectDescription - empty')
      return {}

  @dbus.service.signal(ABOUT_IFACE, signature='qqa(oas)a{sv}')
  def Announce(self, version, port, objectDescription, metaData):
      print('Announce - empty')
      pass

  def Introspect(self, object_path, connection):
    return """
      <node name="/About" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:noNamespaceSchemaLocation="http://www.allseenalliance.org/schemas/introspect.xsd">
         <interface name="org.alljoyn.About">
            <property name="Version" type="q" access="read"/>
            <method name="GetAboutData">
               <arg name="languageTag" type="s" direction="in"/>
               <arg name="aboutData" type="a{sv}" direction="out"/>
            </method>
            <method name="GetObjectDescription">
               <arg name="objectDescription" type="a(oas)" direction="out"/>
            </method>
            <signal name="Announce">
               <arg name="version" type="q"/>
               <arg name="port" type="q"/>
               <arg name="objectDescription" type="a(oas)"/>
               <arg name="metaData" type="a{sv}"/>
            </signal>
         </interface>
      </node>
    """

class ConfigService(PropertiesServiceInterface):
  def __init__(self, container, name):
    self._name = name
    PropertiesServiceInterface.__init__(self, container, '/Config', 
      ['org.alljoyn.Config'], {'org.alljoyn.Config' : {'Version': 1}})

  ## org.alljoyn.Config Interface
  @dbus.service.method(CONFIG_SERVICE_IFACE, in_signature='s', out_signature='a{sv}')
  def GetConfigurations(self, languageTag):
      print('GetConfigurations')
      
      return {            
          'DefaultLanguage': 'en',
          'DeviceName': self._name
      }

  def Introspect(self, object_path, connection):
    return """
      <node name="/Config" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
      xsi:noNamespaceSchemaLocation="http://www.allseenalliance.org/schemas/introspect.xsd">

         <interface name="org.alljoyn.Config">
            <property name="Version" type="q" access="read"/>
            <method name="FactoryReset">
               <annotation name="org.freedesktop.DBus.Method.NoReply" value="true"/>
            </method>
            <method name="Restart">
               <annotation name="org.freedesktop.DBus.Method.NoReply" value="true"/>
            </method>
            <method name="SetPasscode">
               <arg name="daemonRealm" type="s" direction="in"/>
               <arg name="newPasscode" type="ay" direction="in"/>
            </method>
            <method name="GetConfigurations">
               <arg name="languageTag" type="s" direction="in"/>
               <arg name="configData" type="a{sv}" direction="out"/>
            </method>
            <method name="UpdateConfigurations">
               <arg name="languageTag" type="s" direction="in"/>
               <arg name="configMap" type="a{sv}" direction="in"/>
            </method>
            <method name="ResetConfigurations">
               <arg name="languageTag" type="s" direction="in"/>
               <arg name="fieldList" type="as" direction="in"/>
            </method>
         </interface>
      </node>
    """

class ControlPanelService(PropertiesServiceInterface):
  def __init__(self, container, name):
    PropertiesServiceInterface.__init__(self, container, "/ControlPanel/%s/rootContainer" % name, 
      ['org.alljoyn.ControlPanel.ControlPanel', 'org.freedesktop.DBus.Properties'],
      {'org.alljoyn.ControlPanel.ControlPanel' : {'Version': dbus.UInt16(1)}})

  def Introspect(self, object_path, connection):
    return """
      <node name=\"""" + self._path +  """\" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:noNamespaceSchemaLocation="https://www.allseenalliance.org/schemas/introspect.xsd">
         <interface name="org.alljoyn.ControlPanel.ControlPanel">
            <property name="Version" type="q" access="read"/>
         </interface>
        <interface name="org.freedesktop.DBus.Properties">
          <method name="Get">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="in" name="propname" type="s"/>
            <arg direction="out" name="value" type="v"/>
          </method>
          <method name="GetAll">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="out" name="props" type="a{sv}"/>
          </method>
          <method name="Set">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="in" name="propname" type="s"/>
            <arg direction="in" name="value" type="v"/>
          </method>
        </interface>
      </node>
    """

class ContainerService(PropertiesServiceInterface):
  def __init__(self, container, name, path, label):
    PropertiesServiceInterface.__init__(self, container, "/ControlPanel/%s/rootContainer/%s" % (name, path), 
      ['org.alljoyn.ControlPanel.Container', 'org.freedesktop.DBus.Properties'],
      {'org.alljoyn.ControlPanel.Container' : {
        'Version': dbus.UInt16(1), 
        'States': dbus.UInt32(1), 
        'OptParams': dbus.Dictionary({
          dbus.UInt16(0): label,
          dbus.UInt16(2): [dbus.UInt16(1)]
        }, variant_level=1, signature='qv')
      }})

  def Introspect(self, object_path, connection):
    return """
      <node name=\"""" + self._path +  """\" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:noNamespaceSchemaLocation="https://www.allseenalliance.org/schemas/introspect.xsd">
          <interface name="org.alljoyn.ControlPanel.Container">
            <property name="Version" type="q" access="read"/>
            <property name="States" type="u" access="read"/>
            <property name="OptParams" type="a{qv}" access="read"/>
            <signal name="MetadataChanged" />
          </interface>
          <interface name="org.freedesktop.DBus.Properties">
          <method name="Get">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="in" name="propname" type="s"/>
            <arg direction="out" name="value" type="v"/>
          </method>
          <method name="GetAll">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="out" name="props" type="a{sv}"/>
          </method>
          <method name="Set">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="in" name="propname" type="s"/>
            <arg direction="in" name="value" type="v"/>
          </method>
        </interface>
      </node>
    """
  @dbus.service.signal('org.alljoyn.ControlPanel.Container', signature='')
  def MetadataChanged(self):
      pass


class PropertyService(PropertiesServiceInterface):
  def __init__(self, container, name, path, label):
    PropertiesServiceInterface.__init__(self, container, "/ControlPanel/%s/rootContainer/%s" % (name, path), 
      ['org.alljoyn.ControlPanel.Property', 'org.freedesktop.DBus.Properties'],
      {'org.alljoyn.ControlPanel.Property' : {
        'Version': dbus.UInt16(1), 
        'States': dbus.UInt32(1), 
        'Value': dbus.String('Off', variant_level = 1),
        'OptParams': dbus.Dictionary({
          dbus.UInt16(0): label
        }, variant_level=1, signature='qv')
      }})

  def Introspect(self, object_path, connection):
    return """
      <node name=\"""" + self._path +  """\" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:noNamespaceSchemaLocation="https://www.allseenalliance.org/schemas/introspect.xsd">
          <interface name="org.alljoyn.ControlPanel.Property">
            <property name="Version" type="q" access="read"/>
            <property name="States" type="u" access="read"/>
            <property name="OptParams" type="a{qv}" access="read"/>
            <property name="Value" type="v" access="readwrite"/>
            <signal name="MetadataChanged" />
            <signal name="ValueChanged">
               <arg type="v"/>
            </signal>
         </interface>
          <interface name="org.freedesktop.DBus.Properties">
          <method name="Get">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="in" name="propname" type="s"/>
            <arg direction="out" name="value" type="v"/>
          </method>
          <method name="GetAll">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="out" name="props" type="a{sv}"/>
          </method>
          <method name="Set">
            <arg direction="in" name="interface" type="s"/>
            <arg direction="in" name="propname" type="s"/>
            <arg direction="in" name="value" type="v"/>
          </method>
        </interface>
      </node>
    """
  @dbus.service.signal('org.alljoyn.ControlPanel.Property', signature='')
  def MetadataChanged(self):
      pass
  @dbus.service.signal('org.alljoyn.ControlPanel.Property', signature='')
  def ValueChanged(self, value):
      pass

class SmartPlug():
  def __init__(self, busname, name):

    self._id = 'c50ded5d5dfc4de28eb296e921d9a6e2' #uuid.uuid4().hex
    self._name = name

    about = {

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
      AboutService(self._container, about)
      ,ConfigService(self._container, self.name)
      ,ControlPanelService(self._container, self.name)
      ,ContainerService(self._container, self.name, 'en', 'ROOT CONTAINER')
      ,PropertyService(self._container, self.name, 'en/State', 'State')
    ]

    print("Registered %s on dbus" % self.name)

  @property
  def id(self):
      return self._id

  @property
  def name(self):
      return self._name

  def publish(self, bridge):      
    for service in self._services:
      for interface in service.exports:
        bridge.AddService(service.object_path, self._container.bus.get_name(), service.path, interface, service.introspect())

    bridge.StartAllJoyn(self._container.bus.get_name())

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
