#!/bin/bash
cat > /etc/dbus-1/system.d/com.devicehive.conf <<EOF
<!DOCTYPE busconfig PUBLIC
 "-//freedesktop//DTD D-BUS Bus Configuration 1.0//EN"
 "http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
<busconfig>
	<policy user="root">
		<allow own_prefix="com.devicehive" />
		<allow send_type="method_call" />
	</policy>
</busconfig>
EOF
pkill dbus-daemon
sudo systemctl restart dbus
