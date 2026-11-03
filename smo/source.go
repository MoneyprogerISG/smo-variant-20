package smo

import "math/rand"

type Source struct {
	id             int
	a              float64
	b              float64
	requestCounter int
	random         *rand.Rand
}

func NewSource(id int, a, b float64, rng *rand.Rand) *Source {
	if a < 0 || a > b {
		panic("source: interval must satisfy 0 <= a <= b")
	}
	if rng == nil {
		panic("source: rng must not be nil")
	}
	return &Source{
		id:     id,
		a:      a,
		b:      b,
		random: rng,
	}
}

func (s *Source) Generate(now float64) *Request {
	s.requestCounter++
	return NewRequest(s.id, s.requestCounter, now)
}

func (s *Source) NextArrivalTime(now float64) float64 {
	tau := s.a + (s.b-s.a)*s.random.Float64()
	return now + tau
}

func (s *Source) ID() int {
	return s.id
}
