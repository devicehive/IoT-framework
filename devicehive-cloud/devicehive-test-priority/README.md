#Instruments for testing SendNotification priority queue


## Parameters in source code
* in configuration file (or `testConf()` in `conf/conf.go`) you can change `SendNotificatonQueueCapacity` (for example `23`, default=`2048`)
* see `const sendCommandSleeperSeconds` in `ws/send.go` to simulate duration of sending a notification (for example `10`, default=`0`)

## Utilites
* `devicehive-high-send` — sends notification with HIGH priority immediately
* `devicehive-low-sender-loop` — sends notifications with LOW priority with rate is equal one notification per second

