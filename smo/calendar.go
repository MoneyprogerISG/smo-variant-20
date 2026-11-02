package smo

import (
	"container/heap"
	"sort"
)

func eventLess(a, b *Event) bool {
	if a.Time() != b.Time() {
		return a.Time() < b.Time()
	} else {
		return (a.Type() == RELEASE) && (b.Type() == GENERATION)
	}
}

type eventHeap []*Event

func (h eventHeap) Len() int {
	return len(h)
}

func (h eventHeap) Less(i, j int) bool {
	return eventLess(h[i], h[j])
}

func (h eventHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *eventHeap) Push(x any) {
	event := x.(*Event)
	*h = append(*h, event)
}

func (h *eventHeap) Pop() any {
	old := *h
	n := len(old)
	last := old[n-1]
	*h = old[:n-1]
	return last
}

type EventCalendar struct {
	events eventHeap
}

func (c *EventCalendar) Schedule(e *Event) {
	heap.Push(&c.events, e)
}

func (c *EventCalendar) IsEmpty() bool {
	return len(c.events) == 0
}

func (c *EventCalendar) Next() *Event {
	if c.IsEmpty() {
		return nil
	}

	x := heap.Pop(&c.events)
	return x.(*Event)
}

func (c *EventCalendar) Snapshot() []*Event {
	cp := make([]*Event, len(c.events))
	copy(cp, c.events)
	sort.Slice(cp, func(i, j int) bool {
		return eventLess(cp[i], cp[j])
	})

	return cp
}
