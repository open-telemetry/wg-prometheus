package loadbalancer

import (
	"container/heap"
)

// CollectorId is unique for a Instance, however it may not always map to same pod due to pod restart.
// Currently it is ordinal in sts, in the future we can have more than one sts for a single config
// due to topology, resource etc. In that case we may have extra fields like StsName.
// Hence we didn't do type CollectorId int.
type CollectorId struct {
	Ordinal int
}

type CollectorState struct {
	Id              CollectorId
	PodIP           string
	ScheduledTarget *ScheduledTarget
}

func newCollectorState(id CollectorId) *CollectorState {
	return &CollectorState{Id: id, ScheduledTarget: newScheduledTarget()}
}

func (cs *CollectorState) DeepCopy() *CollectorState {
	cs2 := *cs
	cs2.ScheduledTarget = cs.ScheduledTarget.DeepCopy()
	return &cs2
}

func (cs *CollectorState) score() int {
	return len(cs.ScheduledTarget.Targets)
}

// CollectorsPQ wraps collectorsPQ
type CollectorsPQ struct {
	inner   *collectorsPQ
	mapping map[*CollectorState]*collectorPQItem
}

func newCollectorPQ(collectors []*CollectorState) *CollectorsPQ {
	p := CollectorsPQ{inner: &collectorsPQ{}, mapping: make(map[*CollectorState]*collectorPQItem)}
	for _, v := range collectors {
		p.Push(v)
	}
	return &p
}

func newCollectorPQFromMap(collectors map[CollectorId]*CollectorState) *CollectorsPQ {
	p := CollectorsPQ{inner: &collectorsPQ{}, mapping: make(map[*CollectorState]*collectorPQItem)}
	for _, v := range collectors {
		p.Push(v)
	}
	return &p
}

func (pq *CollectorsPQ) Push(s *CollectorState) {
	item := &collectorPQItem{s: s}
	pq.mapping[s] = item
	heap.Push(pq.inner, item)
}

// Pop returns collector with least scheduled targets.
func (pq *CollectorsPQ) Pop() *CollectorState {
	// FIXED: use heap. push pop etc.
	// https://stackoverflow.com/a/60122593
	item := heap.Pop(pq.inner).(*collectorPQItem)
	delete(pq.mapping, item.s)
	return item.s
}

func (pq *CollectorsPQ) Update(c *CollectorState) {
	item := pq.mapping[c]
	pq.inner.update(item)
}

// wrap to avid pqindex inside CollectorState and fails assertion in unit test.
type collectorPQItem struct {
	s       *CollectorState
	pqindex int // for supporting update
}

func (c *collectorPQItem) score() int {
	return c.s.score()
}

func (pq *CollectorsPQ) Len() int {
	return pq.inner.Len()
}

// collectorsPQ implements heap.Interface and holds Items.
// It is copied from https://golang.org/pkg/container/heap/
type collectorsPQ []*collectorPQItem

func (pq collectorsPQ) Len() int { return len(pq) }

func (pq collectorsPQ) Less(i, j int) bool {
	return pq[i].score() < pq[j].score()
}

func (pq collectorsPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].pqindex = i
	pq[j].pqindex = j
}

func (pq *collectorsPQ) Push(x interface{}) {
	n := len(*pq)
	item := x.(*collectorPQItem)
	item.pqindex = n
	*pq = append(*pq, item)
}

func (pq *collectorsPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil    // avoid memory leak
	item.pqindex = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *collectorsPQ) update(item *collectorPQItem) {
	heap.Fix(pq, item.pqindex)
}
