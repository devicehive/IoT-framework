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

# import json

SATECHI_NAMES = ['DELIGHT', 'SATECHI-1']
SATECHI_COLOR_CHAR = 'fff3'

DBUS_BLE_NAME = 'com.devicehive.bluetooth'
DBUS_BLE_PATH = '/com/devicehive/bluetooth'

DBUS_BRIDGE_NAME = 'com.devicehive.alljoyn.bridge'
DBUS_BRIDGE_PATH = '/com/devicehive/alljoyn/bridge'


DBUS_BUS_NAME = 'com.devicehive.alljoyn.allseen.LSF.LampService'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/allseen/LSF/Lamp'

ALLJOYN_LIGHT_PATH = 'org.allseen.LSF.LampService'
ALLJOYN_LIGHT_NAME = '/org/allseen/LSF/Lamp'

ABOUT_IFACE = 'org.alljoyn.About'
LAMP_SERVICE_IFACE = 'org.allseen.LSF.LampService'
LAMP_PARAMETERS_IFACE = 'org.allseen.LSF.LampParameters'
LAMP_DETAILS_IFACE = 'org.allseen.LSF.LampDetails'
LAMP_STATE_IFACE = 'org.allseen.LSF.LampState'

LAMPS = {}

class LampService(dbus.service.Object):
    def __init__(self, mac, ble):
        self.mac = mac        
        self.ble = ble
        self.m_service_path = DBUS_BUS_PATH + '/' + mac
        self.m_service_name = DBUS_BUS_NAME
        bus_name = dbus.service.BusName(DBUS_BUS_NAME, dbus.SystemBus())   
        dbus.service.Object.__init__(self, bus_name, self.m_service_path)
        self.init()
        for l in self.locations:
            print(l)
        print('Registered lamp %s on %s' % (self.mac, self.m_service_path))
    

    @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ss', out_signature='v')
    def Get(self, interface, prop):
        if interface == ABOUT_IFACE:
            if prop == 'Version':
                return '1.0.0'
            else:
                raise Exception('Unsupported property: %s.%s' % (interface, prop))
        else:
            return

    @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ssv')
    def Set(self, interface, prop, value):
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
                if value:
                    self.ble.GattWrite(self.mac, SATECHI_COLOR_CHAR, '0f0d0300ffffffc800c800c8000059ffff')
                else:
                    self.ble.GattWrite(self.mac, SATECHI_COLOR_CHAR, '0f0d0300ffffff0000c800c8000091ffff')   
        else:
            raise Exception('Unsupported property: %s.%s' % (interface, prop))
      

    @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='s', out_signature='a{sv}')
    def GetAll(self, interface):
        if interface == LAMP_SERVICE_IFACE:
            return { 'Version': self.Version,
                     'LampServiceVersion': self.LampServiceVersion,
                     'LampFaults': self.LampFaults
                   }
        else:
            raise dbus.exceptions.DBusException(
                'com.example.UnknownInterface',
                'The Foo object does not implement the %s interface' % interface_name)


    ## org.alljoyn.About Interface

    @dbus.service.method(ABOUT_IFACE, in_signature='s', out_signature='a{sv}')
    def GetAboutData(self, languageTag):
        return {
            'AppId': '8e01a0b4-2331-45c8-b359-21fdf41dd3bc',
            'DefaultLanguage': 'en',
            'DeviceId': socket.gethostname(),
            'AppName': 'SatchiLight',
            'Manufacturer': 'DeviceHive',
            'ModelNumber': '1',
            'SupportedLanguages': ['en'],
            'Description': 'Description',
            'SoftwareVersion': '1.0.0',
            'AJSoftwareVersion': '1.0.0'


        }

    @dbus.service.method(ABOUT_IFACE, in_signature='', out_signature='a(oas)')
    def GetObjectDescription(self):
        return {}


    @dbus.service.signal(ABOUT_IFACE, signature='qqa(oas)a{sv}')
    def Announce(self, version, port, objectDescription, metaData):
        pass


    ## org.allseen.LSF.LampService Interface

    @dbus.service.method(LAMP_SERVICE_IFACE, in_signature='u', out_signature='uu')
    def ClearLampFault(self, LampFaultCode):
        pass


    ## org.allseen.LSF.LampState Interface

    @dbus.service.signal(LAMP_STATE_IFACE, signature='s')
    def LampStateChanged(self, LampID):
        pass


    @dbus.service.method(LAMP_STATE_IFACE, in_signature='ta{sv}u', out_signature='u')
    def TransitionLampState(self, Timestamp, NewState, TransitionPeriod):
        return 'LampResponseCode'


    @dbus.service.method(LAMP_STATE_IFACE, in_signature='a{sv}a{sv}uuut', out_signature='u')
    def ApplyPulseEffect(self, FromState, ToState, period, duration, numPulses, timestamp):
        return 'LampResponseCode'


    # @dbus.service.method(LAMP_STATE_IFACE, in_signature='', out_signature='s')
    def Introspect(self, object_path, connection):
        return """
<node name="/org/allseen/LSF/Lamp" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
xsi:noNamespaceSchemaLocation="https://allseenalliance.org/schemas/introspect.xsd">
    <interface name="org.freedesktop.DBus.Introspectable">
      <method name="Introspect">
        <arg direction="out" type="s" />
      </method>
    </interface>
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


    # init
    def init(self):        
        self.Version = '1.0.0'
        self.LampServiceVersion = '1.0.0'
        self.LampFaults = []

    # free all resources
    def deinit(self):
        self.remove_from_connection()
        # print('Destroyed %s' % self.mac)

class Lamp:
    def __init__(self, mac, name):
        self.mac = mac
        self.name = name
        self.ble = None
        self.status = 'DISCOVERED'

    def connect(self):
        self.status = 'CONNECTED'
        self._dbus = LampService(self.mac, self.ble)

        # expose to alljoyn 
        bus = dbus.SystemBus()
        bridge = dbus.Interface(bus.get_object(DBUS_BRIDGE_NAME, DBUS_BRIDGE_PATH), dbus_interface='com.devicehive.alljoyn.bridge')
        bridge.AddService(self._dbus.m_service_path, self._dbus.m_service_name, ALLJOYN_LIGHT_PATH, ALLJOYN_LIGHT_NAME)
        bridge.StartService(self._dbus.m_service_name)


    def destroy(self):
        if self.status == 'CONNECTED':
            self._dbus.deinit()



def peripheral_discovered_handler(mac, name, rssi):
    # print('Discovered %s - %s' % (mac, name))
    if name in SATECHI_NAMES:        

        if mac not in LAMPS:
            print('Lamp Discovered %s - %s' % (mac, name))
            LAMPS[mac] = Lamp(mac, name)
        else:
            # TODO: cehck status for reconnect
            pass

def peripheral_connected_handler(mac):
    print('BLE: Connected %s ' % mac)
    if mac in LAMPS:
        threading.Thread(target=LAMPS[mac].connect).start()
        
def peripheral_disconnected_handler(mac):
    print('BLE: Disonnected %s ' % mac)
    if mac in LAMPS:
        LAMPS[mac].destroy()
        del LAMPS[mac]

def worker(run_event):
    try:
        bus = dbus.SystemBus()
        ble = dbus.Interface(bus.get_object(DBUS_BLE_NAME, DBUS_BLE_PATH), dbus_interface='com.devicehive.bluetooth')

        ble.connect_to_signal('PeripheralDiscovered', peripheral_discovered_handler)
        ble.connect_to_signal('PeripheralConnected', peripheral_connected_handler)
        ble.connect_to_signal('PeripheralDisconnected', peripheral_disconnected_handler)

        while run_event.is_set():

            # discover
            print("Searching for new lamps..")
            ble.ScanStart()
            time.sleep(1)
            ble.ScanStop()

            # connect
            for mac, lamp in LAMPS.copy().items():
                if lamp.status == 'DISCOVERED':
                    try:
                        lamp.ble = ble
                        lamp.status = 'CONNECTING'
                        print('Connecting to %s' % mac)
                        ble.Connect(lamp.mac, False)
                    except dbus.exceptions.DBusException as error: 
                        lamp.status = 'DISCOVERED'
                        print(error)
                        traceback.print_exc()

            time.sleep(5)

        print("Worker: Exit")
        return    
    except Exception as err:
        print(err)
        traceback.print_exc()
        os._exit(1)

def main():

    run_event = threading.Event()
    run_event.set()

    worker_thread = threading.Thread(target=worker, args=(run_event,))
    worker_thread.start()

    # init d-bus
    GObject.threads_init()    
    # lamps = [LampService(mac) for mac in argv]

    # start mainloop
    loop = GObject.MainLoop()
    try:
        loop.run()
    except (KeyboardInterrupt, SystemExit):
        # for lamp in lamps:
        #     lamp.deinit()
        loop.quit()
        run_event.clear()
        worker_thread.join()

if __name__ == "__main__":
    main()