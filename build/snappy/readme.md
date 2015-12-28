## IoT Framework build system for Ubuntu Snappy

Run
```bash
build-framework.sh
```
to create `devicehive-iot-toolkit_1.0.0_multi.snap` file. This snap contains all frameworks and will start all services automatically after installation. Upload snap to Snappy machine and run `snappy install` or just use `snappy-remote` for install.

You can also run
```bash
build-apps.sh
```
to create few demo snap that using our IoT framework

There are a few environment variables you can customize:
- `PLATFORM` could be `x86_64` or `armhf`
- `VARIANT` could be `debug` or `release`

For example:
```bash
VARIANT=debug build-alljoyn.sh
```

