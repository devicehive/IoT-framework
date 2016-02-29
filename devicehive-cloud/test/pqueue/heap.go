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
