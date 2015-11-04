# DeviceHive GPIO DBus Spec

### Interface com.devicehive.gpio.GpioService
Bus Name: `com.devicehive.gpio`
Path: `/com/devicehive/gpio`

#### Methods:
`list()` - returns list of available gpio pins

`add(pin, port)` - Register pin to expose a physical port

`rm(pin)` - Unregister exposed pin

`clear()` - Unregister all exposed pins

`add_profile(pin_port_pairs)` - Register multiple pins


#### Introspection:
```xml
<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
<node name="/com/devicehive/gpio">
  <interface name="com.devicehive.gpio.GpioService">
    <method name="clear">
    </method>
    <method name="list">
      <arg direction="out" type="as" />
    </method>
    <method name="add_profile">
      <arg direction="in"  type="a{ss}" name="pins" />
    </method>
    <method name="add">
      <arg direction="in"  type="s" name="pin" />
      <arg direction="in"  type="s" name="port" />
    </method>
    <method name="remove">
      <arg direction="in"  type="v" name="pin" />
      <arg direction="in"  type="v" name="in_signature" />
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


### Interface com.devicehive.gpio.GpioPin
Bus Name: `com.devicehive.gpio`
Path: `/com/devicehive/gpio/{PIN}`

#### Methods:
`init(mode)` - initialize pin. For digital pins mode can be `out` for output, `in` for input, 
               `rising` or `falling` or `both` for input with enabled notifications
               analog pins recieve period in miliseconds as mode. It sends notifications with
               analog value every choosen period of time. 

`deinit()` - deinitialize pin and free all resource

`set_value(value)` - set pin state, where value is ether `0` or `1`

`set()` - set pin state to `1`

`clear()` - set pin state to `0`

`get()` - read pin state or value for analog inputs

`toggle()` - toggle pin state from `0` to `1` or from `1` to `0`


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
  <interface name="com.devicehive.gpio.GpioPin">
    <method name="deinit">
    </method>
    <method name="get">
      <arg direction="out" type="s" />
    </method>
    <method name="set">
    </method>
    <method name="set_value">
      <arg direction="in"  type="s" name="value" />
    </method>
    <method name="clear">
    </method>
    <signal name="pin_value_changed">
      <arg type="s" name="pin" />
      <arg type="v" name="value" />
    </signal>
    <method name="init">
      <arg direction="in"  type="s" name="mode" />
    </method>
    <method name="toggle">
    </method>
  </interface>
</node>

```

