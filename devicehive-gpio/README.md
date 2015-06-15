# DeviceHive GPIO framework
DeviceHive GPIO framework and usage examples.

> DBus Spec can be found [here](DBUS-SPEC.md)

### Running
Run
daemon/gpio-daemon <file with profile>  

### Usage Examples

#### GPIO Blink

This program does blinking of the LED lamps connected to corresponding gpio pins.

```bash
gpio-blink-python [PIN#] ...
```
Where PIN# is one of pins on the device connected to LEDs. To list all available pins run app without parameters:
```bash
gpio-blink-python
```

Example:
```bash
gpio-blink-python PIN9 PIN10 PIN11 PIN12 PIN13 PIN14 PIN15 PIN16
```


#### GPIO Click

This program turns LED on/off when on a button click.

```bash
gpio-click-python BUTTON_PIN LED_PIN
```
Where PIN# is one of pins on the device connected to LEDs. To list all available pins run app without parameters:
```bash
gpio-click-python
```

Example:
```bash
gpio-click-python PIN9 PIN10
```


#### GPIO Switch

This program uses 2 buttons to control a set of LEDs. One of the LEDs is turned off and buttons shift turned off led left or right.

```bash
gpio-switch-python
```

> It is important to connect LEDs and buttons to specific pins for this example. For BBB use 
> Buttons: PIN7, PIN8
> LEDS: PIN9, PIN10, PIN11, PIN12, PIN13, PIN14, PIN15, PIN16

## Device Profiles

As GPIO pinout is different on every device/board and there is no way to enumerate available pins programatically. So gpio daemon uses pin mapping for each device. Mappings are located in `/src/profiles/gpio.yaml` file.  Each profile is mapped to the name returned by `/sys/firmware/devicetree/base/model` on the current system.

