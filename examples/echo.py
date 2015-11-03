#!/usr/bin/env python3

import dbus
import json
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GObject

DBusGMainLoop(set_as_default=True)

def get_cloud():
    obj = dbus.SystemBus().get_object("com.devicehive.cloud", '/com/devicehive/cloud')
    return dbus.Interface(obj, "com.devicehive.cloud")

cloud = get_cloud()
def command_received(id, name, parameters):
    print("new Command received {id:%s, name:%s, parameters:%s}" % (id, name, parameters))
    try:
        p = json.loads(parameters)
        s = {"command": name, "id": id, "parameters": p}
        cloud.SendNotification("echo", json.dumps(s), 0)
        cloud.UpdateCommand(dbus.UInt64(id), "OK", parameters)
    except dbus.DBusException as e:
        print(e)

def main():
    cloud.connect_to_signal("CommandReceived", command_received)
    print("ready to process commands...")
    GObject.MainLoop().run()

if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print("stopped")
