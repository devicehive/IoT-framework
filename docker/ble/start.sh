#!/bin/sh -e

# this script is used to start devicehive-ble service
# D-Bus daemon should be available via volumes

# com.devicehive.bluetooth.conf configuration should be prepared by init.sh script
# we have to put it to the shared volume so D-Bus daemon recognizes our service
[ -e /etc/dbus-1/system.d/com.devicehive.bluetooth.conf ] || \
	cp -f /usr/share/dbus-1/system.d/com.devicehive.bluetooth.conf /etc/dbus-1/system.d/

/usr/bin/devicehive-ble $@
