## Examples for DeviceHive D-Bus Framework 

This directory contains various examples of IoT framework usage.
To run each of them you have to install IoT framework on target machine.
Python examples might be run as is, `go` examples should be compiled first.

### echo.py
Simple application which uses cloud part of IoT framework. It replyes with echoes
on each sent command. Usage: run the `echo.py`, send command to device (for example
via web admin page) and check command result - it should be the same. At the same
time corresponding "echo" notification should be sent.

### cpu-stats.go
Simple application which reads current system state (CPU, memory usage) and sends
it to cloud with some period of time. Usage: run this demo and check DeviceHive
server, it should receive corresponding "stats" notifications.

### dash-xylo.go
Simple application that sends some melody for playback to Play-I robot.
These robots uses BLE for connectivity purpose and they have special characterstic.
To make the robot play recorder melody you need to write special value to this
characteristic.

### heart-pulse-demo.go
TODO

### iot-demo.go
TODO

### lamp.py
Simple application which uses Bluetooth low energy part of IoT Framework.
It scans for BLE bulb with name `DELIGHT` and once it found sends command
to turn it ON or OFF. Usage: uncomment action that you need to perform at
line 22-23 and run this demo.

### pod.py
TODO

### scales.py
TODO

### sensortag.py
Simple application which shows how to connect to TI SensorTag. Once connected it
enables accelerometer, receives notification with accelerometer data and prints it
to standard output. Usage: run this demo on machine with BLE, turn on TI Sensor
tag and demo will print accelerometer data.

### Directories alljoyn, cloud-ble, enocean, gpio
See corresponding readme file inside of each directory.
