/*
  DeviceHive IoT-Framework business logic

  Copyright (C) 2016 DataArt

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at
 
      http://www.apache.org/licenses/LICENSE-2.0
 
  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/ 

package main

import (
	"sync"

	"github.com/godbus/dbus"
)

type signalHandlerFunc func(args ...interface{})
type cloudCommandHandler func(map[string]interface{}) (map[string]interface{}, error)

type signalHandler struct {
	handler  signalHandlerFunc
	priority uint64
}

type signalItem struct {
	handler   signalHandler
	signal    *dbus.Signal
	timestamp uint64
}

type signalQueue struct {
	items     []signalItem
	queueSync sync.Mutex
	cond      *sync.Cond
}

func (q signalQueue) Len() int {
	return len(q.items)
}

func (q signalQueue) Less(i, j int) bool {
	return q.items[i].timestamp*q.items[i].handler.priority > q.items[j].timestamp*q.items[j].handler.priority
}

func (q signalQueue) Swap(i, j int) {
	if (i > len(q.items)-1) || (j > len(q.items)-1) || (i < 0) || (j < 0) {
		return
	}

	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *signalQueue) Push(x interface{}) {
	q.cond.L.Lock()
	q.items = append(q.items, x.(signalItem))
	q.cond.L.Unlock()
	q.cond.Signal()
}

func (q *signalQueue) Pop() interface{} {
	q.cond.L.Lock()

	old := q.items
	n := len(old)

	x := old[n-1]
	q.items = old[0 : n-1]

	q.cond.L.Unlock()
	q.cond.Signal()

	return x
}
