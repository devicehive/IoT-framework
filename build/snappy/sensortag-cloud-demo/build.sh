#!/bin/bash

DIR="$(dirname "$(readlink -f "$0")")"

mkdir -p $DIR/bin
cp $DIR/../../../examples/sensortag-cloud.py $DIR/bin

