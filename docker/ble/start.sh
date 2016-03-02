#!/bin/sh -e

# this script is used to start devicehive-ble service
# D-Bus daemon should be available via volumes:
#   - to store D-Bus configuration (/etc/dbus-1/system.d/)
#   - to access D-Bus daemon (/var/run/dbus/)

# com.devicehive.bluetooth.conf configuration should be prepared by init.sh script
# we have to put it to the shared volume so D-Bus daemon recognizes our service
DBUS_TEMPLATE=/usr/share/dbus-1/system.d/com.devicehive.bluetooth.conf
DBUS_CONF=/etc/dbus-1/system.d/com.devicehive.bluetooth.conf
[ -e "$DBUS_CONF" ] || cp -f "$DBUS_TEMPLATE" "$DBUS_CONF"

/usr/bin/devicehive-ble $@
