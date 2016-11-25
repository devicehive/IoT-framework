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

package pqueue

import (
	"container/heap"
	"sync"
	"time"
)

// type Message []byte
type Message map[string]interface{}

type QueueItem struct {
	Msg       Message
	Timestamp uint64
	Priority  uint64
}

type PriorityQueue struct {
	items []QueueItem
	cond  *sync.Cond

	capacity uint64
	out      chan Message
}

func (pq *PriorityQueue) Out() chan Message {
	return pq.out
}

func NewPriorityQueue(capacity uint64, listener chan Message) (*PriorityQueue, error) {
	pq := PriorityQueue{}
	if listener == nil {
		return &pq, ListenerShouldNotBeNil
	}

	pq.items = []QueueItem{}
	pq.cond = &sync.Cond{L: &sync.Mutex{}}
	pq.capacity = capacity
	pq.out = listener

	go func() {
		defer func() { recover() }()
		for {
			pq.cond.L.Lock()
			for pq.Len() == 0 {
				pq.cond.Wait()
			}
			pq.cond.L.Unlock()
			item := heap.Pop(&pq).(QueueItem)
			pq.out <- item.Msg
		}
	}()
	return &pq, nil
}

func (pq *PriorityQueue) Send(m Message, priority uint64) (removed []QueueItem) {
	for uint64(pq.Len()) > pq.capacity-1 {
		item := heap.Remove(pq, pq.Len()-1).(QueueItem)
		removed = append(removed, item)
	}
	heap.Push(pq, QueueItem{
		Msg:       m,
		Timestamp: uint64(time.Now().Unix()),
		Priority:  priority,
	})
	return
}
