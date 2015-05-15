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
