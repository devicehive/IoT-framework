# DeviceHive IoT toolkit for AllJoyn

[AllSeen Alliance]: https://allseenalliance.org
[AllJoyn]: https://allseenalliance.org/framework

## Overview

Driven by the [AllSeen Alliance], [AllJoyn] is an open source software framework that makes it easy for devices and apps to discover and communicate with each other. 


DeviceHive IoT Toolkit provides AllJoyn bridge that enables developers to expose non AllJpyn devices and applications as AllJoyn virtual devices without need to use AllJoyn C/C++ libraries. 
Devices can be described with simple DSL and logic can be implemented on any programming language.  

![AllJoyn Bridge Diagram](alljoyn-bridge.png?raw=true)

DeviceHive AllJoyn Bridge can be used on any Embedded Linux device.  


## Running in development

To be able to compile and run devicehive-alljoyn bridge:

1. Clone and build `/alljoyn` submodule recursively
2. Set `LD_LIBRARY_PATH` environment variable to access `libajtcl.so` file
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

