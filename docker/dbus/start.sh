#!/bin/sh -e

# this script is used to start D-Bus daemon
# check start() from /etc/init.d/dbus to reference

/usr/bin/dbus-uuidgen --ensure=/etc/machine-id

mkdir -p /var/run/dbus
/usr/bin/dbus-daemon --system $@
