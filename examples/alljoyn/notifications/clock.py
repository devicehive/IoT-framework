#!/usr/bin/python3

import collections
import os
import socket
import sys
import threading
import time
import traceback
import uuid
from threading import Timer
import dbus.service
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject

scriptDir = os.path.dirname(os.path.realpath(__file__))
sys.path.append(scriptDir + "/../common")

import core
import notification

DBusGMainLoop(set_as_default=True)

DBUS_BRIDGE_NAME = 'com.devicehive.alljoyn.bridge'
DBUS_BRIDGE_PATH = '/com/devicehive/alljoyn/bridge'

DBUS_BUS_NAME = 'com.devicehive.alljoyn.SmartHome'
DBUS_BUS_PATH = '/com/devicehive/alljoyn/SmartHome'

CLOCK_SVC = 'org.allseen.SmartHome.Clock'

bus = dbus.SystemBus()
bus_name = dbus.service.BusName(DBUS_BUS_NAME, bus)

FORMAT = "%a, %d %b %Y %H:%M:%S"


class ClockService(core.PropertiesServiceInterface):
    def __init__(self, container, notification_handler):
        core.PropertiesServiceInterface.__init__(self, container, "/Service", {
            CLOCK_SVC: {'Time': ''}})

        self._timer = None
        self._alarm = None
        self._notification_handler = notification_handler

    def start(self):
        self._timer = Timer(1.0, self.tick)
        self._timer.start()

    def tick(self):
        try:
            timestr = time.strftime(FORMAT, time.localtime())
            self.Set(CLOCK_SVC, 'Time', dbus.String(timestr, variant_level=1))
            print("Time: %s" % timestr)

            if self._alarm is not None:
                alarmstr = time.strftime(FORMAT, self._alarm)
                print("Alarm: %s" % alarmstr)
                if self._alarm <= time.localtime():
                    self.Alarm(alarmstr)
                    self._alarm = None

        except Exception as err:
            print(err)
            traceback.print_exc()
            os._exit(1)

        self.start()

    def stop(self):
        if self._timer is not None:
            self._timer.cancel()

    def IntrospectionXml(self):
        return """
            <interface name="org.allseen.SmartHome.Clock">
              <property name="Time" type="s" access="read"/>
              <method name="SetAlarm">
                 <arg name="time" type="s" direction="in"/>
              </method>
              <signal name="Alarm">
                <arg type="s"/>
              </signal>
           </interface>

        """ + core.PropertiesServiceInterface.IntrospectionXml(self)

    @dbus.service.method(CLOCK_SVC, in_signature='s', out_signature='')
    def SetAlarm(self, timestr):
        self._alarm = time.strptime(timestr, FORMAT)
        print("Alarm set to: %s" + timestr)

    @dbus.service.signal(CLOCK_SVC, signature='s')
    def Alarm(self, timestr):
        print("ALARM: %s" % timestr)
        self._notification_handler(timestr)

#     @dbus.service.method("org.allseen.Introspectable", in_signature='', out_signature='as')
#     def GetDescriptionLanguages(self):
#         return ["en"]

#     @dbus.service.method("org.allseen.Introspectable", in_signature='s', out_signature='s')
#     def IntrospectWithDescription(self, lang):
#         return """
#         <node name="/Service">
# <description language="en">Description</description>
#            <interface name="org.allseen.SmartHome.Clock">
# <description language="en">Description</description>
#   <method name="SetAlarm">
#     <arg name="time" type="s" direction="in">
#         <description language="en">Description</description>
#     </arg>
#     <description language="en">Description</description>
#   </method>
#   <signal name="Alarm" sessionless="false">
#     <arg type="s">
#         <description language="en">Description</description>
#     </arg>
#     <description language="en">Description</description>
#   </signal>
#   <property name="Time" type="s" access="read">
#     <description language="en">Description</description>
#   </property>
# </interface>
# <interface name="org.freedesktop.DBus.Properties">
# <description language="en">Description</description>
#   <method name="Get">
#     <arg name="interface" type="s" direction="in">
#         <description language="en">Description</description>
#     </arg>
#     <arg name="propname" type="s" direction="in">
#         <description language="en">Description</description>
#     </arg>
#     <arg name="value" type="v" direction="out">
#         <description language="en">Description</description>
#     </arg>
#     <description language="en">Description</description>
#   </method>
#   <method name="GetAll">
#     <arg name="interface" type="s" direction="in">
#         <description language="en">Description</description>
#     </arg>
#     <arg name="props" type="a{sv}" direction="out">
#         <description language="en">Description</description>
#     </arg>
#     <description language="en">Description</description>
#   </method>
#   <method name="Set">
#     <arg name="interface" type="s" direction="in">
#         <description language="en">Description</description>
#     </arg>
#     <arg name="propname" type="s" direction="in">
#         <description language="en">Description</description>
#     </arg>
#     <arg name="value" type="v" direction="in">
#         <description language="en">Description</description>
#     </arg>
#     <description language="en">Description</description>
#   </method>
# </interface>
# <interface name="org.freedesktop.DBus.Introspectable">
# <description language="en">Description</description>
#   <method name="Introspect">
#     <arg name="out" type="s" direction="out">
#         <description language="en">Description</description>
#     </arg>
#     <description language="en">Description</description>
#   </method>
# </interface>
# </node>
#         """


class Clock():
    def __init__(self, busname, name):

        self._id = 'f85bda3742d04ff782c774f01b458cba'
        self._name = name

        about_props = {
            'AppId': dbus.ByteArray(bytes.fromhex(self.id)),
            'DefaultLanguage': 'en',
            'DeviceName': self.name,
            'DeviceId': self.id,
            'AppName': 'Clock',
            'Manufacturer': 'DeviceHive',
            'DateOfManufacture': '2015-10-28',
            'ModelNumber': 'example',
            'SupportedLanguages': ['en'],
            'Description': 'DeviceHive Alljoyn Clock Device',
            # 'SoftwareVersion': '1.0',
            # 'HardwareVersion': '1.0',
            # 'SupportUrl': 'devicehive.com'

        }

        self._container = core.BusContainer(busname, DBUS_BUS_PATH + '/' + self.id)
        self._notifications = notification.Notifications(self._container)
        self._service = ClockService(self._container, lambda time: 
            self._notifications.warning('en' ,"ALARM: %s!" % time)
            )

        self._services = [
            core.AboutService(self._container, about_props),
            # core.ConfigService(self._container, self.name),
            # core.ConfigService(self._container, self.name),
            self._notifications,
            self._service
        ]

        self._service.start()
        print("Registered %s on dbus" % self.name)

    def __del__(self):
        self._service.stop()

    @property
    def id(self):
        return self._id

    @property
    def name(self):
        return self._name

    def publish(self, bridge):
        service = self._services[0]
        bridge.AddService(
            self._container.bus.get_name(),
            self._container.relative('').rstrip('/'), CLOCK_SVC,
            # ignore_reply=True
            reply_handler=lambda id: print("ID: %s" % id),
            error_handler=lambda err: print("Error: %s" % err)
        )

        print("Published %s on bridge" % self.name)


def main():

    # init d-bus
    GObject.threads_init()
    dbus.mainloop.glib.threads_init()

    # start mainloop
    loop = GObject.MainLoop()

    bridge = dbus.Interface(bus.get_object(DBUS_BRIDGE_NAME, DBUS_BRIDGE_PATH), 
        dbus_interface='com.devicehive.alljoyn.bridge')
    clock = Clock(bus_name, 'Clock')
    clock.publish(bridge)

    try:
        loop.run()
    except (KeyboardInterrupt, SystemExit):
        loop.quit()
        clock.__del__()

if __name__ == "__main__":
    main()
