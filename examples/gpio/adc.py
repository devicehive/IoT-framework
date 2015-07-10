#!/usr/bin/env python3

import dbus
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject
import array
import threading

DBusGMainLoop(set_as_default=True)

def get_cloud():
    return dbus.Interface(dbus.SystemBus().get_object("com.devicehive.cloud", '/com/devicehive/cloud'), "com.devicehive.cloud")

cloud = get_cloud()
def command_received(id, name, parameters):
    print("Command Received id : %s, name : %s, parameters : %s" % (id, name, parameters))
    try:
        # TODO: use json.dumps instead
        s = "{\"command\" : \"%s\", \"id\" : %s, \"parameters\": %s}" % (name, id, parameters)
        print(s) 
        cloud.SendNotification("echo", s, 0)
        cloud.UpdateCommand(dbus.UInt32(id), "OK", s)
    except dbus.DBusException as e:
        print(e)

def adc_changed(port, value):
    global pin_out
    print("pin_value_changed port: {} value: {}".format(port, value))
    s = "{\"notification\" : \"emg\", \"parameters\": \"%s\"}" % value
    cloud.SendNotification("emg", s, 0)
            
def main():
    pin_in = dbus.SystemBus().get_object("com.devicehive.gpio",
        "/com/devicehive/gpio/PIN85")
    pin_in.init("500", dbus_interface='com.devicehive.gpio.GpioPin')
    pin_in.connect_to_signal("pin_value_changed", adc_changed)
    
    cloud.connect_to_signal("CommandReceived", command_received)
    GObject.MainLoop().run()
    try:
        GObject.MainLoop().run()
    except (KeyboardInterrupt, SystemExit):
        pin_in.deinit()

if __name__ == '__main__':
    main()
