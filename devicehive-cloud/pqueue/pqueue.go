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

func NewPriorityQueue(capacity uint64, listener chan Message) (PriorityQueue, error) {
	pq := PriorityQueue{}
	if listener == nil {
		return pq, ListenerShouldNotBeNil
	}

	pq.items = []QueueItem{}
	pq.cond = &sync.Cond{L: &sync.Mutex{}}
	pq.capacity = capacity
	pq.out = listener

	go observer(&pq)
	return pq, nil
}

func observer(pq *PriorityQueue) {
	defer func() { recover() }()
	for {
		pq.cond.L.Lock()
		for pq.Len() == 0 {
			pq.cond.Wait()
		}
		pq.cond.L.Unlock()
		item := heap.Pop(pq).(QueueItem)
		pq.out <- item.Msg
	}
}

func (pq *PriorityQueue) Send(m Message, priority uint64) (removed []QueueItem) {
	for uint64(pq.Len()) > pq.capacity-1 {
		item := heap.Remove(pq, pq.Len()-1).(QueueItem)
		//log.Printf("Removing %+v from queue", item)
		removed = append(removed, item)
	}
	heap.Push(pq, QueueItem{
		Msg:       m,
		Timestamp: uint64(time.Now().Unix()),
		Priority:  priority,
	})
	return
}
