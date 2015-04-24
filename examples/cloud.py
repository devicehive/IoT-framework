#!/usr/bin/env python3

import dbus
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject
import array

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
        cloud.SendNotification("echo", s)
        cloud.UpdateCommand(dbus.UInt32(id), "OK", s)
    except dbus.DBusException as e:
        print(e)

def main():
    cloud.connect_to_signal("CommandReceived", command_received)

    GObject.MainLoop().run()

if __name__ == '__main__':
    main()