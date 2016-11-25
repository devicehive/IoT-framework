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

func (q *PriorityQueue) Push(x interface{}) {
	q.cond.L.Lock()
	q.items = append(q.items, x.(QueueItem))
	q.cond.L.Unlock()
	q.cond.Signal()
}

func (q *PriorityQueue) Pop() interface{} {
	q.cond.L.Lock()

	old := q.items
	n := len(old)

	x := old[n-1]
	q.items = old[0 : n-1]

	q.cond.L.Unlock()
	q.cond.Signal()

	return x
}
