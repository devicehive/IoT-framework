#!/bin/sh -e

# this script is used to start D-Bus daemon
# for reference check start() function from /etc/init.d/dbus

/usr/bin/dbus-uuidgen --ensure=/etc/machine-id

# important do not fork D-Bus daemon!
# (to keep container running)
/usr/bin/dbus-daemon --nofork --nopidfile $@
