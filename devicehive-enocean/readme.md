#DeviceHive EnOcean D-Bus Daemon

EnOcean daemon for dbus. All EnOcean actuators harvest energy from enviroment and use it to send short message via air to reciver that connect to machine with this framework. This message can be  obtain with:

message_received

dbus signal of this framework. So, usage is extreamly simple:

````
enocean_manager = bus.get_object(DBUS_BUS_ENOCEAN_NAME, '/com/devicehive/enocean')
enocean = dbus.Interface(enocean_manager, DBUS_BUS_ENOCEAN_NAME)
enocean.connect_to_signal('message_received', message_received) 
````
just unparse pure data in `message_received` callback and use it.


# Running
Current daemon has dependency on `enocean` python3 library:
```
pip3 install enocean
```

