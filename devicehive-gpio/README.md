# DeviceHive GPIO framework
DeviceHive GPIO framework and usage examples.

> DBus Spec can be found [here](DBUS-SPEC.md)

### Running
Run
daemon/gpio-daemon <file with profile>  

## Device Profiles

As GPIO pinout is different on every device/board and there is no way to enumerate available pins programatically. So gpio daemon uses pin mapping for each device. Mappings are located in `profiles` folder.  Each profile is mapped to the name returned by `/sys/firmware/devicetree/base/model` on the current system.

