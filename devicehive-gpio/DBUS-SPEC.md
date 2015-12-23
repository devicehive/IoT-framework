# DeviceHive GPIO D-Bus Spec

### Interface com.devicehive.gpio.Service
- Bus Name: `com.devicehive.gpio`
- Path: `/com/devicehive/gpio`

#### Methods
The following methods are supported by the service:
- `GetAllPins()` - returns dict of created pins/ports
- `CreatePin(pin, port)` - Register pin to expose a physical port
- `CreatePins(pins)` - Register multiple pins/ports (as a dict)
- `DeletePin(pin)` - Unregister exposed pin
- `DeleteAllPins()` - Unregister all exposed pins


#### Introspection
```xml
<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
<node name="/com/devicehive/gpio">
  <interface name="com.devicehive.gpio.Service">
    <method name="GetAllPins">
      <arg direction="out" type="a{ss}" />
    </method>
    <method name="CreatePin">
      <arg direction="in"  type="s" name="pin" />
      <arg direction="in"  type="s" name="port" />
    </method>
    <method name="CreatePins">
      <arg direction="in"  type="a{ss}" name="pins" />
    </method>
    <method name="DeletePin">
      <arg direction="in"  type="s" name="pin" />
    </method>
    <method name="DeleteAllPins">
    </method>
  </interface>
  <interface name="org.freedesktop.DBus.ObjectManager">
    <method name="GetManagedObjects">
      <arg direction="out" type="a{oa{sa{sv}}}" />
    </method>
    <signal name="InterfacesRemoved">
      <arg type="o" name="object_path" />
      <arg type="a{sa{sv}}" name="interfaces_and_properties" />
    </signal>
    <signal name="InterfacesAdded">
      <arg type="o" name="object_path" />
      <arg type="a{sa{sv}}" name="interfaces_and_properties" />
    </signal>
  </interface>
  <interface name="org.freedesktop.DBus.Introspectable">
    <method name="Introspect">
      <arg direction="out" type="s" />
    </method>
  </interface>
  <interface name="org.freedesktop.DBus.Properties">
    <method name="GetAll">
      <arg direction="in"  type="s" name="interface" />
      <arg direction="out" type="a{sv}" />
    </method>
    <method name="Set">
      <arg direction="in"  type="s" name="interface" />
      <arg direction="in"  type="s" name="prop" />
      <arg direction="in"  type="v" name="value" />
    </method>
    <signal name="PropertiesChanged">
      <arg type="s" name="interface" />
      <arg type="a{sv}" name="values" />
      <arg type="as" name="unchanged" />
    </signal>
    <method name="Get">
      <arg direction="in"  type="s" name="interface" />
      <arg direction="in"  type="s" name="prop" />
      <arg direction="out" type="v" />
    </method>
  </interface>
</node>
```


### Interface com.devicehive.gpio.Pin
- Bus Name: `com.devicehive.gpio`
- Path: `/com/devicehive/gpio/{PIN}`

#### Methods
The following methods are supported by the service:
- `Start(mode)` - initialize pin. For digital pins mode can be `"out"` for output, `"in"` for input,
         `"rising"` or `"falling"` or `"both"` for input with enabled notifications.
          analog pins receive report period in miliseconds as mode, `"500ms"`.
          It sends notifications with analog value every choosen period of time.
- `Stop()` - deinitialize pin and free all resources
- `SetValue(value)` - set pin state, where value is ether `"0"` or `"1"`
- `Set()` - set pin state to `"1"`
- `Clear()` - set pin state to `"0"`
- `Get()` - read pin state or value for analog inputs
- `Toggle()` - toggle pin state from `"0"` to `"1"` or from `"1"` to `"0"`


#### Introspection:
```xml
<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
<node name="/com/devicehive/gpio/1">
  <interface name="org.freedesktop.DBus.Introspectable">
    <method name="Introspect">
      <arg direction="out" type="s" />
    </method>
  </interface>
  <interface name="com.devicehive.gpio.Pin">
    <method name="Stop">
    </method>
    <method name="Get">
      <arg direction="out" type="s" />
    </method>
    <method name="Set">
    </method>
    <method name="SetValue">
      <arg direction="in"  type="s" name="value" />
    </method>
    <method name="Clear">
    </method>
    <signal name="ValueChanged">
      <arg type="s" name="pin" />
      <arg type="s" name="value" />
    </signal>
    <method name="Start">
      <arg direction="in"  type="s" name="mode" />
    </method>
    <method name="Toggle">
    </method>
  </interface>
</node>
```
