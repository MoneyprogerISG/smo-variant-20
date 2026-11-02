package smo

import "testing"

func TestCalendarEmpty(t *testing.T) {
	var c EventCalendar

	if !c.IsEmpty() {
		t.Errorf("IsEmpty() у нового календаря: получили false, ожидали true")
	}
	if e := c.Next(); e != nil {
		t.Fatalf("Next() на пустом календаре: получили %v, ожидали nil", e)
	}

	c.Schedule(NewGenerationEvent(1, nil))
	if c.IsEmpty() {
		t.Errorf("IsEmpty() после Schedule: получили true, ожидали false")
	}
}

func TestCalendarOrder(t *testing.T) {
	var c EventCalendar
	for _, tm := range []float64{5, 1, 3} {
		c.Schedule(NewGenerationEvent(tm, nil))
	}

	want := []float64{1, 3, 5}
	for i, w := range want {
		e := c.Next()
		if e == nil {
			t.Fatalf("шаг %d: Next() вернул nil, ожидали событие с временем %v", i, w)
		}
		if e.Time() != w {
			t.Errorf("шаг %d: время события %v, ожидали %v", i, e.Time(), w)
		}
	}
	if !c.IsEmpty() {
		t.Errorf("после извлечения всех событий календарь не пуст")
	}
}

func TestCalendarReleaseBeforeGeneration(t *testing.T) {
	tests := []struct {
		name  string
		order []*Event
	}{
		{
			name: "GENERATION добавлен первым",
			order: []*Event{
				NewGenerationEvent(2, nil),
				NewReleaseEvent(2, nil),
			},
		},
		{
			name: "RELEASE добавлен первым",
			order: []*Event{
				NewReleaseEvent(2, nil),
				NewGenerationEvent(2, nil),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c EventCalendar
			for _, e := range tc.order {
				c.Schedule(e)
			}

			e := c.Next()
			if e == nil {
				t.Fatalf("Next() вернул nil, ожидали событие")
			}
			if e.Type() != RELEASE {
				t.Errorf("при равном времени %v первым вышло %v, ожидали RELEASE", e.Time(), e.Type())
			}
		})
	}
}

func TestCalendarSnapshot(t *testing.T) {
	var c EventCalendar
	events := []*Event{
		NewGenerationEvent(4, nil),
		NewGenerationEvent(2, nil),
		NewGenerationEvent(3, nil),
		NewReleaseEvent(1, nil),
		NewGenerationEvent(3, nil),
		NewReleaseEvent(3, nil),
	}
	for _, e := range events {
		c.Schedule(e)
	}

	snap := c.Snapshot()

	if len(snap) != len(events) {
		t.Fatalf("длина Snapshot(): получили %d, ожидали %d", len(snap), len(events))
	}
	for i := 1; i < len(snap); i++ {
		prev, cur := snap[i-1], snap[i]
		if cur.Time() < prev.Time() {
			t.Errorf("Snapshot()[%d].Time() = %v меньше предыдущего %v: нарушена сортировка",
				i, cur.Time(), prev.Time())
		}
		if cur.Time() == prev.Time() && prev.Type() == GENERATION && cur.Type() == RELEASE {
			t.Errorf("Snapshot() при времени %v: на позиции %d стоит %v перед %v, RELEASE должен идти первым",
				cur.Time(), i-1, prev.Type(), cur.Type())
		}
	}

	if len(c.events) != len(events) {
		t.Errorf("после Snapshot() в календаре %d событий, ожидали %d", len(c.events), len(events))
	}
	if e := c.Next(); e == nil || e.Time() != 1 {
		t.Errorf("после Snapshot() Next() вернул %v, ожидали событие с минимальным временем 1", e)
	}
}
