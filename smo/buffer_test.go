package smo

import "testing"

func TestBufferEmpty(t *testing.T) {
	b := NewBuffer(3)

	if !b.IsEmpty() {
		t.Errorf("IsEmpty() у нового буфера: получили false, ожидали true")
	}
	if b.IsFull() {
		t.Errorf("IsFull() у нового буфера: получили true, ожидали false")
	}
	if b.Size() != 0 {
		t.Errorf("Size() у нового буфера: получили %d, ожидали 0", b.Size())
	}
	if r := b.TakeFirst(); r != nil {
		t.Errorf("TakeFirst() на пустом буфере: получили %v, ожидали nil", r)
	}
	if r := b.RemoveLast(); r != nil {
		t.Errorf("RemoveLast() на пустом буфере: получили %v, ожидали nil", r)
	}
}

func TestBufferAddOrderAndPlace(t *testing.T) {
	b := NewBuffer(3)
	reqs := []*Request{
		NewRequest(1, 1, 0),
		NewRequest(2, 1, 0),
		NewRequest(1, 2, 0),
	}

	for i, r := range reqs {
		if place := b.Add(r, float64(i)); place != i+1 {
			t.Errorf("Add() заявки %d: место %d, ожидали %d", i, place, i+1)
		}
	}

	snap := b.Snapshot()
	if len(snap) != len(reqs) {
		t.Fatalf("длина Snapshot(): получили %d, ожидали %d", len(snap), len(reqs))
	}
	for i, r := range reqs {
		if snap[i] != r {
			t.Errorf("Snapshot()[%d]: получили %v, ожидали %v", i, snap[i], r)
		}
	}
}

func TestBufferAddSetsBufferInTime(t *testing.T) {
	b := NewBuffer(2)
	r := NewRequest(1, 1, 0)

	b.Add(r, 4.5)

	if r.bufferInTime != 4.5 {
		t.Errorf("bufferInTime после Add(): получили %v, ожидали 4.5", r.bufferInTime)
	}
}

func TestBufferFull(t *testing.T) {
	b := NewBuffer(2)

	b.Add(NewRequest(1, 1, 0), 0)
	if b.IsFull() {
		t.Errorf("IsFull() при 1 заявке из 2: получили true, ожидали false")
	}

	b.Add(NewRequest(1, 2, 0), 0)
	if !b.IsFull() {
		t.Errorf("IsFull() при 2 заявках из 2: получили false, ожидали true")
	}
	if b.IsEmpty() {
		t.Errorf("IsEmpty() у полного буфера: получили true, ожидали false")
	}
}

func TestBufferAddToFullPanics(t *testing.T) {
	b := NewBuffer(1)
	b.Add(NewRequest(1, 1, 0), 0)

	defer func() {
		if recover() == nil {
			t.Errorf("Add() в полный буфер: panic не произошёл")
		}
	}()

	b.Add(NewRequest(1, 2, 0), 0)
}

func TestNewBufferNegativeCapacityPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("NewBuffer(-1): panic не произошёл")
		}
	}()

	NewBuffer(-1)
}

func TestBufferTakeFirst(t *testing.T) {
	b := NewBuffer(3)
	reqs := []*Request{
		NewRequest(1, 1, 0),
		NewRequest(1, 2, 0),
		NewRequest(1, 3, 0),
	}
	for _, r := range reqs {
		b.Add(r, 0)
	}

	for i, want := range reqs {
		got := b.TakeFirst()
		if got != want {
			t.Fatalf("шаг %d: TakeFirst() вернул %v, ожидали %v", i, got, want)
		}
		if b.Size() != len(reqs)-i-1 {
			t.Errorf("шаг %d: Size() = %d, ожидали %d", i, b.Size(), len(reqs)-i-1)
		}
	}
	if !b.IsEmpty() {
		t.Errorf("после извлечения всех заявок буфер не пуст")
	}
}

func TestBufferRemoveLast(t *testing.T) {
	b := NewBuffer(3)
	reqs := []*Request{
		NewRequest(1, 1, 0),
		NewRequest(1, 2, 0),
		NewRequest(1, 3, 0),
	}
	for _, r := range reqs {
		b.Add(r, 0)
	}

	if got := b.RemoveLast(); got != reqs[2] {
		t.Errorf("RemoveLast() вернул %v, ожидали %v", got, reqs[2])
	}
	if b.Size() != 2 {
		t.Errorf("Size() после RemoveLast(): получили %d, ожидали 2", b.Size())
	}
	if got := b.TakeFirst(); got != reqs[0] {
		t.Errorf("TakeFirst() после RemoveLast() вернул %v, ожидали %v", got, reqs[0])
	}
}

func TestBufferKnockOut(t *testing.T) {
	b := NewBuffer(2)
	first := NewRequest(1, 1, 0)
	second := NewRequest(1, 2, 0)
	incoming := NewRequest(2, 1, 0)
	b.Add(first, 0)
	b.Add(second, 0)

	if !b.IsFull() {
		t.Fatalf("буфер должен быть полон перед выбиванием")
	}

	if out := b.RemoveLast(); out != second {
		t.Errorf("выбита заявка %v, ожидали хвост %v", out, second)
	}
	if place := b.Add(incoming, 5); place != 2 {
		t.Errorf("Add() после выбивания: место %d, ожидали 2", place)
	}

	snap := b.Snapshot()
	want := []*Request{first, incoming}
	for i, w := range want {
		if snap[i] != w {
			t.Errorf("Snapshot()[%d]: получили %v, ожидали %v", i, snap[i], w)
		}
	}
}

func TestBufferCapacityAfterShift(t *testing.T) {
	b := NewBuffer(2)
	b.Add(NewRequest(1, 1, 0), 0)
	b.Add(NewRequest(1, 2, 0), 0)

	for i := 0; i < 5; i++ {
		b.TakeFirst()
		if b.IsFull() {
			t.Fatalf("итерация %d: IsFull() после TakeFirst(): получили true, ожидали false", i)
		}
		b.Add(NewRequest(1, 3+i, 0), 0)
		if !b.IsFull() {
			t.Fatalf("итерация %d: IsFull() после Add(): получили false, ожидали true", i)
		}
	}
}

func TestBufferSnapshotIsCopy(t *testing.T) {
	b := NewBuffer(2)
	first := NewRequest(1, 1, 0)
	second := NewRequest(1, 2, 0)
	b.Add(first, 0)
	b.Add(second, 0)

	snap := b.Snapshot()
	snap[0] = nil
	snap = append(snap[:0], snap[1:]...)

	if b.Size() != 2 {
		t.Errorf("Size() после изменения Snapshot(): получили %d, ожидали 2", b.Size())
	}
	if got := b.TakeFirst(); got != first {
		t.Errorf("TakeFirst() после изменения Snapshot(): получили %v, ожидали %v", got, first)
	}
}
