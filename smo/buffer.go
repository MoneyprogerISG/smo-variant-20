package smo

type Buffer struct {
	capacity int
	requests []*Request
}

func NewBuffer(capacity int) *Buffer {
	if capacity < 0 {
		panic("buffer: negative capacity")
	}

	return &Buffer{
		capacity: capacity,
		requests: make([]*Request, 0, capacity),
	}
}

func (b *Buffer) Add(r *Request, now float64) int {
	if b.IsFull() {
		panic("buffer: Add to full buffer")
	}

	r.SetBufferInTime(now)

	b.requests = append(b.requests, r)

	return len(b.requests)
}

func (b *Buffer) RemoveLast() *Request {
	if b.IsEmpty() {
		return nil
	}

	last := b.requests[len(b.requests)-1]
	b.requests[len(b.requests)-1] = nil
	b.requests = b.requests[:len(b.requests)-1]

	return last
}

func (b *Buffer) TakeFirst() *Request {
	if b.IsEmpty() {
		return nil
	}

	first := b.requests[0]
	b.requests[0] = nil
	b.requests = b.requests[1:]

	return first
}

func (b *Buffer) IsFull() bool {
	return b.Size() >= b.capacity
}

func (b *Buffer) IsEmpty() bool {
	return len(b.requests) == 0
}

func (b *Buffer) Size() int {
	return len(b.requests)
}

func (b *Buffer) Snapshot() []*Request {
	cp := make([]*Request, len(b.requests))
	copy(cp, b.requests)
	return cp
}
