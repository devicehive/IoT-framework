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

func (q PriorityQueue) Len() int {
	return len(q.items)
}

func (q PriorityQueue) Less(i, j int) bool {
	return q.items[i].Timestamp*q.items[i].Priority > q.items[j].Timestamp*q.items[j].Priority
}

func (q PriorityQueue) Swap(i, j int) {
	if (i > len(q.items)-1) || (j > len(q.items)-1) || (i < 0) || (j < 0) {
		return
	}
	q.items[i], q.items[j] = q.items[j], q.items[i]
}
