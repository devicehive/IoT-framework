# How to build docker images

All `iot-cloud` and `iot-ble` containers don't have D-Bus installed.
Instead D-Bus daemon from `iot-dbus` container is used.

To make all services up use the following command:

```{.sh}
$ cd IoT-framework/docker
$ docker-compose up
```

## Manual run

First of all we need to run D-Bus service (see below how to build images):

```{.sh}
$ docker run -d --name=dbus devicehive/iot-dbus:2.0
```

Then it's possible to run cloud service:

```{.sh}
$ docker run -d --volumes-from dbus --name=cloud devicehive/iot-cloud:2.0
```

It's it. Now you can send commands via Admin page.

At the same time it's possible to connect to `dbus` container to monitor D-Bus messages:

```{.sh}
$ docker exec -it dbus /bin/sh
# dbus-monitor --system
```

The `iot-ble` container should be run in privileged mode with "host" network.
See this docker [issue](https://github.com/docker/docker/issues/16208) for more details.


## Build iot-dbus
```{.sh}
$ cd IoT-framework/docker
$ docker build -t devicehive/iot-dbus:2.0 dbus
```

## Build iot-cloud
You can put valid data to `IoT-framework/docker/cloud/config.yml` before building image.
Or just override this configuration later (via volumes) to provide valid credentials.

```{.sh}
$ cd IoT-framework/docker
$ docker build -t devicehive/iot-cloud:2.0 cloud
```

## Build iot-ble
```{.sh}
$ cd IoT-framework/docker
$ docker build -t devicehive/iot-ble:2.0 ble
```

## Beagle Bone Black

First of all we need to install docker on BBB. Follow [these](https://www.element14.com/community/people/markfink/blog/2015/02/05/using-docker-on-beaglebone-black) instructions to prepare Arch Linux SD card (should be at least 8GB) and install docker:

Replace `sdX` in the following instructions with the device name
for the SD card as it appears on your computer.

```{.sh}
# Zero the beginning of the SD card:
sudo dd if=/dev/zero of=/dev/sdX bs=1M count=8

# Start fdisk to partition the SD card:
sudo fdisk /dev/sdX

# At the fdisk prompt, delete old partitions and create a new one:
# Type 'o'. This will clear out any partitions on the drive.
# Type 'p' to list partitions. There should be no partitions left.
# Now type 'n', then 'p' for primary, '1' for the first partition on the drive,
# '2048' for the first sector, and then press 'ENTER' to accept the default last sector.
# Write the partition table and exit by typing 'w'.

# Create and mount the ext4 filesystem:
sudo mkfs.ext4 /dev/sdX1
mkdir bbb
sudo mount /dev/sdX1 bbb

# Download and extract the root filesystem:
wget http://archlinuxarm.org/os/ArchLinuxARM-am33x-latest.tar.gz
sudo bsdtar -xpf ArchLinuxARM-am33x-latest.tar.gz -C bbb
sync

# Install the U-Boot bootloader:
sudo dd if=bbb/boot/MLO of=/dev/sdX count=1 seek=1 conv=notrunc bs=128k
sudo dd if=bbb/boot/u-boot.img of=/dev/sdX count=2 seek=1 conv=notrunc bs=384k
sudo umount bbb
sync

# Insert the SD card into the BeagleBone, connect ethernet, press S2 button, and apply 5V power.

# Use the serial console or SSH to the IP address given to the board by your router.
# Login as the default user alarm with the password 'alarm'.
# The default root password is 'root'.
```

Install docker:

```{.sh}
# sync and update:
pacman -Syyu
pacman-db-upgrade

# An optional step is to configure a suitable hostname
hostnamectl set-hostname bbb-docker

# And now install the docker package itself:
pacman -S docker docker-compose systemd git

# Enable docker to run as a service:
systemctl enable docker.service

# Enable 'alarm' user to run docker:
gpasswd -a alarm docker
reboot
```

Get IoT-framework:

```{.sh}
export GOPATH=~/go
mkdir -p $GOPATH/src/github.com/devicehive
cd $GOPATH/src/github.com/devicehive
git clone -b v2 --single-branch https://github.com/devicehive/IoT-framework
cd $GOPATH/src/github.com/devicehive/IoT-framework/docker
docker-compose -f docker-compose.armhf.yml up
```
