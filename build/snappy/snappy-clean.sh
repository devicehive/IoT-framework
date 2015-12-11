#!/bin/bash

PACKAGE=${1}
FLAGS=-fv #-i # "-i" is used for interactive confirmation

if [ "x${PACKAGE}x" = "xx" ]; then
    echo "No package name provided!"
    echo "Usage: $0 <package name>"
    exit 1
fi

while true; do
    read -p "Do you really want to clean '$PACKAGE' package [y/n]?" yn
    case $yn in
        [Yy]* )
            break
            ;;
        [Nn]* )
            exit 1
            ;;
        * )
            echo "Please answer Yes or No."
            ;;
    esac
done

sudo rm ${FLAGS} /etc/systemd/system/${PACKAGE}_*.service
sudo rm ${FLAGS} /etc/systemd/system/multi-user.target.wants/${PACKAGE}_*.service
sudo rm ${FLAGS} /var/lib/snappy/*/*/${PACKAGE}_*
sudo rm ${FLAGS} /etc/dbus-1/system.d/${PACKAGE}_*.conf
