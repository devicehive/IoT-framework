## Running in development

To be able to compile and run devicehive-alljoyn bridge:


1. Clone and build `/alljoyn` submodule recursivel
2. Set environment variable to `libajtcl.so` file
```
LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$GOPATH/src/github.com/devicehive/IoT-framework/devicehive-alljoyn/alljoyn/core/ajtcl
```
3. Create dbus config file `/etc/dbus-1/system.d/com.devicehive.conf` to allow bridge service register on system bus:

```xml
<!DOCTYPE busconfig PUBLIC "-//freedesktop//DTD D-BUS Bus Configuration1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">

<busconfig>
 <policy  context="default">
    <allow own_prefix="com.devicehive" />
    <allow own="com.devicehive.alljoyn.bridge" />
    <allow send_destination="*" />
    
  </policy>

</busconfig>
```