#!/usr/bin/python3

import time
import threading
import traceback
import sys
import os
import socket

import dbus.service
from dbus.mainloop.glib import DBusGMainLoop
try:
    from gi.repository import GObject
except ImportError:
    import gobject as GObject

DBusGMainLoop(set_as_default=True)

bus = dbus.SystemBus()

# import json

SATECHI_NAMES = ['DELIGHT', 'SATECHI-1']
SATECHI_COLOR_CHAR = 'fff3'

DBUS_BRIDGE_NAME = 'com.devicehive.alljoyn.bridge'
DBUS_BRIDGE_PATH = '/com/devicehive/alljoyn/bridge'


DBUS_BUS_NAME = 'com.devicehive.alljoyn.allseen.LSF.LampService'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/allseen/LSF/Lamp'

ALLJOYN_LIGHT_NAME = 'org.allseen.LSF.LampService'
ALLJOYN_LIGHT_PATH = '/org/allseen/LSF/Lamp'

ABOUT_IFACE = 'org.alljoyn.About'
LAMP_SERVICE_IFACE = 'org.allseen.LSF.LampService'
LAMP_PARAMETERS_IFACE = 'org.allseen.LSF.LampParameters'
LAMP_DETAILS_IFACE = 'org.allseen.LSF.LampDetails'
LAMP_STATE_IFACE = 'org.allseen.LSF.LampState'

LAMPS = {}

INTROSPECT = """
<node name="/org/allseen/LSF/Lamp">
<interface name="org.freedesktop.DBus.Properties">
  <method name="Get">
    <arg type="s" direction="in"/>
    <arg type="s" direction="in"/>
    <arg type="v" direction="out"/>
  </method>
  <method name="Set">
    <arg type="s" direction="in"/>
    <arg type="s" direction="in"/>
    <arg type="v" direction="in"/>
  </method>
  <method name="GetAll">
    <arg type="s" direction="in"/>
    <arg type="a{sv}" direction="out"/>
  </method>
</interface>
<interface name="org.allseen.LSF.LampService">
  <property name="Version" type="u" access="read"/>
  <property name="LampServiceVersion" type="u" access="read"/>
  <method name="ClearLampFault">
    <arg name="LampFaultCode" type="u" direction="in"/>
    <arg name="LampResponseCode" type="u" direction="out"/>
    <arg name="LampFaultCode" type="u" direction="out"/>
  </method>
  <property name="LampFaults" type="au" access="read"/>
</interface>
<interface name="org.allseen.LSF.LampParameters">
  <property name="Version" type="u" access="read"/>
  <property name="Energy_Usage_Milliwatts" type="u" access="read"/>
  <property name="Brightness_Lumens" type="u" access="read"/>
</interface>
<interface name="org.allseen.LSF.LampDetails">
  <property name="Version" type="u" access="read"/>
  <property name="Make" type="u" access="read"/>
  <property name="Model" type="u" access="read"/>
  <property name="Type" type="u" access="read"/>
  <property name="LampType" type="u" access="read"/>
  <property name="LampBaseType" type="u" access="read"/>
  <property name="LampBeamAngle" type="u" access="read"/>
  <property name="Dimmable" type="b" access="read"/>
  <property name="Color" type="b" access="read"/>
  <property name="VariableColorTemp" type="b" access="read"/>
  <property name="HasEffects" type="b" access="read"/>
  <property name="MinVoltage" type="u" access="read"/>
  <property name="MaxVoltage" type="u" access="read"/>
  <property name="Wattage" type="u" access="read"/>
  <property name="IncandescentEquivalent" type="u" access="read"/>
  <property name="MaxLumens" type="u" access="read"/>
  <property name="MinTemperature" type="u" access="read"/>
  <property name="MaxTemperature" type="u" access="read"/>
  <property name="ColorRenderingIndex" type="u" access="read"/>
  <property name="LampID" type="s" access="read"/>
</interface>
<interface name="org.allseen.LSF.LampState">
  <property name="Version" type="u" access="read"/>
  <method name="TransitionLampState">
    <arg name="Timestamp" type="t" direction="in"/>
    <arg name="NewState" type="a{sv}" direction="in"/>
    <arg name="TransitionPeriod" type="u" direction="in"/>
    <arg name="LampResponseCode" type="u" direction="out"/>
  </method>
  <method name="ApplyPulseEffect">
    <arg name="FromState" type="a{sv}" direction="in"/>
    <arg name="ToState" type="a{sv}" direction="in"/>
    <arg name="period" type="u" direction="in"/>
    <arg name="duration" type="u" direction="in"/>
    <arg name="numPulses" type="u" direction="in"/>
    <arg name="timestamp" type="t" direction="in"/>
    <arg name="LampResponseCode" type="u" direction="out"/>
  </method>
  <signal name="LampStateChanged">
    <arg name="LampID" type="s"/>
  </signal>
  <property name="OnOff" type="b" access="readwrite"/>
  <property name="Hue" type="u" access="readwrite"/>
  <property name="Saturation" type="u" access="readwrite"/>
  <property name="ColorTemp" type="u" access="readwrite"/>
  <property name="Brightness" type="u" access="readwrite"/>
</interface>
</node>
        """

class LampService(dbus.service.Object):
    def __init__(self, mac):
        print('__init__')
        self.mac = mac        
        self.m_service_path = DBUS_BUS_PATH + '/' + mac
        self.m_service_name = DBUS_BUS_NAME
        bus_name = dbus.service.BusName(DBUS_BUS_NAME, bus)   
        dbus.service.Object.__init__(self, bus_name, self.m_service_path)
        for l in self.locations:
            print(l)
        print('Registered lamp %s on %s' % (self.mac, self.m_service_path))
    

    @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ss', out_signature='v')
    def Get(self, interface, prop):
        print('Get')
        print("Properties.Get is called for %s interface %s prop %s" % (self.mac, interface, prop))
        if interface == ABOUT_IFACE:
            if prop == 'Version':
                return LAMPS[self.mac].Version
            else:
                raise Exception('Unsupported property: %s.%s' % (interface, prop))
        elif interface == LAMP_SERVICE_IFACE:
            pass
        elif interface == LAMP_PARAMETERS_IFACE:
            pass
        elif interface == LAMP_DETAILS_IFACE:
            pass
        elif interface == LAMP_STATE_IFACE:
            if prop == 'OnOff':
              return LAMPS[self.mac].OnOff
        return "NotSupported"

    @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ssv')
    def Set(self, interface, prop, value):
        print('Set')
        print("Properties.Set is called for %s" % self.mac)
        if interface == ABOUT_IFACE:
            pass
        if interface == LAMP_SERVICE_IFACE:
            pass
        elif interface == LAMP_PARAMETERS_IFACE:
            pass
        elif interface == LAMP_DETAILS_IFACE:
            pass
        elif interface == LAMP_STATE_IFACE:
            if prop == 'OnOff':
                LAMPS[self.mac].turnOnOff(value)
        else:
            raise Exception('Unsupported property: %s.%s' % (interface, prop))
      

    @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='s', out_signature='a{sv}')
    def GetAll(self, interface):
        print('GetAll')
        print("Properties.GetAll is called for %s, interface %s" % (self.mac, interface))
        if interface == LAMP_SERVICE_IFACE:
            return { 'Version': dbus.UInt32(LAMPS[self.mac].Version),
                     'LampServiceVersion': dbus.UInt32(LAMPS[self.mac].LampServiceVersion),
                     'LampFaults': dbus.UInt32(LAMPS[self.mac].LampFaults)
                   }
        elif interface == LAMP_PARAMETERS_IFACE:
            return { 'Version': dbus.UInt32(LAMPS[self.mac].Version),
                     'Energy_Usage_Milliwatts': dbus.UInt32(LAMPS[self.mac].energyUsage_mW),
                     'Brightness_Lumens': dbus.UInt32(LAMPS[self.mac].brigthnessLumen)
                   }

        elif interface == LAMP_DETAILS_IFACE:
            return { 'Version': LAMPS[self.mac].Version,
                     'LampVersion' : dbus.UInt32(1),
                     'Make': dbus.UInt32(1),
                     'Model': dbus.UInt32(1),
                     'Type': dbus.UInt32(1),
                     'LampType': dbus.UInt32(30),
                     'LampBaseType': dbus.UInt32(8),
                     'LampBeamAngle': dbus.UInt32(100),
                     'Dimmable': False,
                     'Color': False, 
                     'VariableColorTemp': False,
                     'HasEffects': False,
                     'MinVoltage': dbus.UInt32(90),
                     'MaxVoltage': dbus.UInt32(240),
                     'Wattage': dbus.UInt32(9),
                     'IncandescentEquivalent': dbus.UInt32(75),
                     'MaxLumens': dbus.UInt32(900),
                     'MinTemperature': dbus.UInt32(2700),
                     'MaxTemperature': dbus.UInt32(5400),
                     'ColorRenderingIndex': dbus.UInt32(0),
                     'LampID': LAMPS[self.mac].deviceId
                   }
        elif interface == LAMP_STATE_IFACE:
            return { 'Version': LAMPS[self.mac].Version,
## In percentage, 0 means 0. uint32_max-1 means 100. */
                     'Hue': 99999, 
                     'Saturation': 99999,
                     'ColorTemp': 99999,
                     'Brightness': 100,
                     'OnOff': LAMPS[self.mac].OnOff
                   }
        else:
            raise dbus.exceptions.DBusException(
                'com.example.UnknownInterface',
                'The Foo object does not implement the %s interface' % interface_name)


    ## org.alljoyn.About Interface

    @dbus.service.method(ABOUT_IFACE, in_signature='s', out_signature='a{sv}')
    def GetAboutData(self, languageTag):
        print("GetAboutData is called for %s" % self.mac)
        return {
            'AppId': dbus.ByteArray(bytes.fromhex("8e01a0b4233145c8b35921fdf41dd3bc")),
            'DefaultLanguage': 'en',
            'DeviceName': "DeviceHiveVB",
            'DeviceId': LAMPS[self.mac].deviceId,
            'AppName': 'SatchiLight',
            'Manufacturer': 'DeviceHive',
            'DateOfManufacture': '2015-06-17',
            'ModelNumber': '0.0.1',
            'SupportedLanguages': ['en'],
            'Description': 'Simulated Lamp',
            'SoftwareVersion': '1.0',
            'HardwareVersion': '1.0',              
            'SupportUrl': 'http://devicehive.com',
            'AJSoftwareVersion': '14.06.00a Tag "v14.06.00a"',
            'Maxlength': dbus.UInt16(254)
        }

    @dbus.service.method(ABOUT_IFACE, in_signature='', out_signature='a(oas)')
    def GetObjectDescription(self):
        print('GetObjectDescription - empty')
        return {}


    @dbus.service.signal(ABOUT_IFACE, signature='qqa(oas)a{sv}')
    def Announce(self, version, port, objectDescription, metaData):
        print('Announce - empty')
        pass


    ## org.allseen.LSF.LampService Interface

    @dbus.service.method(LAMP_SERVICE_IFACE, in_signature='u', out_signature='uu')
    def ClearLampFault(self, LampFaultCode):
        print('ClearLampFault - empty')
        pass


    ## org.allseen.LSF.LampState Interface

    @dbus.service.signal(LAMP_STATE_IFACE, signature='s')
    def LampStateChanged(self, LampID):
        print('LampStateChanged - empty')
        pass


    @dbus.service.method(LAMP_STATE_IFACE, in_signature='ta{sv}u', out_signature='u')
    def TransitionLampState(self, Timestamp, NewState, TransitionPeriod):
        print('TransitionLampState  %s' % NewState['OnOff'])
        LAMPS[self.mac].turnOnOff(NewState['OnOff'])
        return 1


    @dbus.service.method(LAMP_STATE_IFACE, in_signature='a{sv}a{sv}uuut', out_signature='u')
    def ApplyPulseEffect(self, FromState, ToState, period, duration, numPulses, timestamp):
        print('ApplyPulseEffect - empty')
        return 0


    # @dbus.service.method(LAMP_STATE_IFACE, in_signature='', out_signature='s')
    def Introspect(self, object_path, connection):
        print('Introspect - empty')
        print("Introspect is called for %s" % self.mac)
        return INTROSPECT


# """
#     <interface name="org.alljoyn.About">
#       <property name="Version" type="q" access="read"/>
#       <method name="GetAboutData">
#          <arg name="languageTag" type="s" direction="in"/>
#          <arg name="aboutData" type="a{sv}" direction="out"/>
#       </method>
#       <method name="GetObjectDescription">
#          <arg name="objectDescription" type="a(oas)" direction="out"/>
#       </method>
#       <signal name="Announce">
#          <arg name="version" type="q"/>
#          <arg name="port" type="q"/>
#          <arg name="objectDescription" type="a(oas)"/>
#          <arg name="metaData" type="a{sv}"/>
#       </signal>
#     </interface>
# """

    # free all resources
    def deinit(self):
        print('deinit')
        self.remove_from_connection()
        # print('Destroyed %s' % self.mac)

class Lamp:
    def __init__(self, mac, name):
        self.Version = 1
        self.mac = mac
        self.name = name        
        self.status = 'DISCOVERED'
        self.LampServiceVersion = 1
        self.LampFaults = []
        self.energyUsage_mW = 15
        self.brigthnessLumen = 100
        self.deviceId = "8e01a0b4233145c8e35921fdf41dd3bc";
        self.OnOff = False

    def connect(self):
        self.status = 'CONNECTED'
        self._dbus = LampService(self.mac)

        print("Calling alljoyn bridge")

        try:
        # expose to alljoyn             
            bridge = dbus.Interface(bus.get_object(DBUS_BRIDGE_NAME, DBUS_BRIDGE_PATH), dbus_interface='com.devicehive.alljoyn.bridge')
            bridge.AddService(self._dbus.m_service_path, self._dbus.m_service_name, ALLJOYN_LIGHT_PATH, ALLJOYN_LIGHT_NAME, INTROSPECT)
            bridge.StartAllJoyn(self._dbus.m_service_name)
            print("%s exposed to Alljoyn" % self.mac)
        except Exception as err:
                print(err)
                traceback.print_exc()
        
    def turnOnOff(self, state):
        self.OnOff = state
        if state:
            print("***LAMP NOW IS ON***")
        else:
            print("***LAMP NOW IS OFF***")

    def destroy(self):
        if self.status == 'CONNECTED':
            self._dbus.deinit()

def worker():
    try:

        time.sleep(2)

        # single lamp for now
        mac = '12345678'
        LAMPS[mac] = Lamp(mac, 'Virtual Lamp')
        # threading.Thread(target=LAMPS[mac].connect).start()
        LAMPS[mac].connect()
        
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
