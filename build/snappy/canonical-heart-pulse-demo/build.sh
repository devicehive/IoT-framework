#!/bin/bash

GOOS=linux go build -o bin/x86_64/heart-pulse-demo ../../../examples/heart-pulse-demo.go
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/armhf/heart-pulse-demo ../../../examples/heart-pulse-demo.go
