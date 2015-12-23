# DeviceHive GPIO framework
DeviceHive GPIO framework and usage examples.

D-Bus specifications can be found [here](./DBUS-SPEC.md)

### Running
Run `./devicehive-gpio <file with profile>`

## Device Profiles

As GPIO pinout is different on every device/board and there is no way to enumerate
available pins programatically. So gpio daemon uses pin mapping for each device.
Mappings are located in `profiles` folder. Each profile is mapped to the name
returned by `/sys/firmware/devicetree/base/model` on the current system.

## Testing

Here are a set of useful commands to test GPIO daemon.

To get list of current pins:

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio com.devicehive.gpio.Service.GetAllPins`

To create a pin:

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio com.devicehive.gpio.Service.CreatePin string:"PIN3" string:"38"`

To delete all pins:

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio com.devicehive.gpio.Service.DeleteAllPins`

To start pin OUT mode:

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio/PIN3 com.devicehive.gpio.Pin.Start string:"out"`

To get current pin value:

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio/PIN3 com.devicehive.gpio.Pin.Get`

To set pin value to `"1"`:

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio/PIN3 com.devicehive.gpio.Pin.Set`

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio/PIN3 com.devicehive.gpio.Pin.SetValue string:"1"`

To set pin value to `"0"`:

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio/PIN3 com.devicehive.gpio.Pin.Clear`

`sudo dbus-send --system --dest=com.devicehive.gpio --print-reply --type=method_call /com/devicehive/gpio/PIN3 com.devicehive.gpio.Pin.SetValue string:"0"`

Watch for `ValueChanged` signal:

`sudo dbus-monitor --system --monitor "type='signal',interface='com.devicehive.gpio.Pin'"`
