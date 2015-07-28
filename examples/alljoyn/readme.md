Examples of using GPIO framework. AllJoyn framework.


### real-light.py
Example of proxying BLE(bluetooth low energy) bulb into AllJoyn ligth service. This demo actually convert BLE bulb in to AllJoyn bulb! Change your BLE bulb mac addr in file(at line 486) to mac of your Satechi BLE bulb and run demo. You can modify this demo to control any other BLE bulb, you just need to find out GATT services and characteristic to turn it on and off. After demo has started, you can use any application that works with AllJoyn light service and control your BLE bulbs as AllJoyn bulbs.

### terminal-light.py
Example of creating virtual bulb that shows own status in std output. Usage: just run this demo and use any application which can control AllJoyn bulbs. Application will find our virtual bulb as a real one and you can control it. On each turn on or off action demo will print text with current status to std output.   
