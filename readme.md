# DeviceHive D-Bus Framework


Develop apps connected to devices, peripherials and cloud using commonly available linux dbus interface.

![](framework.png?raw=true)

## Cloning

This repository uses submodules so to clonse fresh copy use command
```
git clone --recurisve https://github.com/devicehive/IoT-framework.git
```

Or if doing pull to existing repo use
```
git pull
git submodule update --init --recursive
```


## Building 

To build `Ubuntu Snappy` package navigate to `build/snappy` and run 
```
# for framework
./build-framework.sh

#for sample apps
./build-apps.sh
```

