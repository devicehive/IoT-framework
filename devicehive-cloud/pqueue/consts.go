package pqueue

import "errors"

const StandardQueueCapacity = 2048

var ListenerShouldNotBeNil = errors.New("Out-channel should not be nil")
