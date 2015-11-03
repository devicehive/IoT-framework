## Examples for DeviceHive D-Bus Framework 

These are examples of usage IoT framework. To run each of them you have to install IoT framework on target machine. You may use this examples to find out how to use components of IoT ferameowork. For running each of this examples you will need to upload it to target machine and run there for Python examples. For `go` example you will need to compile them before.

### echo.py
Simple application which uses cloud part of IoT framework. It replyes with echoes
on each sent command. Usage: run the echo.py, send command to device (for example
via web admin page) and check command result - it should be the same. At the same
time corresponding "echo" notification should be sent.

### cpu-stats.go
Simple application which reads current system state(CPU, momory usage) and sends it to cloud with some period of time. Usage: run this demo and check DeviceHive server, it should recieve notifications with current machine status.

### dash-xylo.go
Simple application that sends some melody for playback to Play-I robot. These robots uses BLE for connectivity purpose and they have special characterstic which can be written and robot will playback recorder data. 

### heart-pulse-demo.go
TODO

### iot-demo.go
TODO

### lamp.py
Simple application which uses bluetooth low energy part of IoT Framework. It scans for BLE bulb with name `DELIGHT` and since it found send command to turn on or off. Usage: uncomment action that you need to perform at line 22-23 and run this demo.

### pod.py
TODO

### scales.py
TODO

### sensortag.py
Simple application which shows how to connect to TI SensorTag. It connects to it, enable accelerometer, receives notification with accelerometer data and prints it to std output. Usage: run this demo on machine with BLE, turn on TI Sensor tag in BLE radius and demo will print accelerometer data from it.

### Directories alljoyn, cloud-ble, enocean, gpio
See corresponding readme file inside of each directory.


