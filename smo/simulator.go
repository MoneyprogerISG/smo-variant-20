package smo

import (
	"math"
	"math/rand"
)

const (
	tAlpha        = 1.643
	delta         = 0.1
	initialN      = 100
	epsilon       = 0.1
	maxN          = 1000000
	maxIterations = 20
)

type Iteration struct {
	N int
	P float64
}

type Config struct {
	Sources        int
	A, B           float64
	Devices        int
	Lambda         float64
	BufferCapacity int
	Seed           int64
}

type SourceState struct {
	ID                 int
	Generated          int
	Refused            int
	Served             int
	RefusalProbability float64
}

type DeviceState struct {
	ID          int
	Busy        bool
	Current     *Request
	ReleaseTime float64
}

type SystemState struct {
	Time           float64
	Sources        []SourceState
	Calendar       []*Event
	Buffer         []*Request
	BufferCapacity int
	Devices        []DeviceState
	Pointer        int
}

type Simulator struct {
	currentTime         float64
	calendar            *EventCalendar
	sources             []*Source
	buffer              *Buffer
	devicePool          *DevicePool
	placementDispatcher *PlacementDispatcher
	selectionDispatcher *SelectionDispatcher
	statistics          *Statistics
	maxRequests         int
}

func NewSimulator(cfg Config) *Simulator {
	if cfg.Sources <= 0 {
		panic("simulator: sources count must be positive")
	}
	if cfg.Devices <= 0 {
		panic("simulator: devices count must be positive")
	}
	if cfg.BufferCapacity <= 0 {
		panic("simulator: buffer capacity must be positive")
	}

	rng := rand.New(rand.NewSource(cfg.Seed))

	sources := make([]*Source, cfg.Sources)
	for i := range sources {
		sources[i] = NewSource(i, cfg.A, cfg.B, rng)
	}

	devices := make([]*Device, cfg.Devices)
	for j := range devices {
		devices[j] = NewDevice(j, cfg.Lambda, rng)
	}

	calendar := &EventCalendar{}
	buffer := NewBuffer(cfg.BufferCapacity)
	pool := NewDevicePool(devices)
	stats := NewStatistics(cfg.Sources, cfg.Devices)

	return &Simulator{
		calendar:            calendar,
		sources:             sources,
		buffer:              buffer,
		devicePool:          pool,
		placementDispatcher: NewPlacementDispatcher(buffer, pool, calendar, stats),
		selectionDispatcher: NewSelectionDispatcher(buffer, pool, calendar, stats),
		statistics:          stats,
	}
}

func (s *Simulator) Init() {
	for _, src := range s.sources {
		s.calendar.Schedule(NewGenerationEvent(src.NextArrivalTime(0), src))
	}
}

func (s *Simulator) limitReached() bool {
	if s.maxRequests == 0 {
		return false
	}
	total := 0
	for _, src := range s.sources {
		total += s.statistics.Generated(src.ID())
	}
	return total >= s.maxRequests
}

func (s *Simulator) Step() *Event {
	for {
		e := s.calendar.Next()
		if e == nil {
			return nil
		}

		if e.Type() == GENERATION && s.limitReached() {
			continue
		}

		s.currentTime = e.Time()

		switch e.Type() {
		case GENERATION:
			s.handleGeneration(e)
		case RELEASE:
			s.handleRelease(e)
		}
		return e
	}
}

func (s *Simulator) RunAuto(requestCount int) SummaryTable {
	if requestCount <= 0 {
		panic("simulator: requets count must be positive")
	}
	s.maxRequests = requestCount
	s.Init()
	for s.Step() != nil {
	}
	return s.statistics.BuildSummaryTable(s.currentTime)
}

func requiredRequests(p float64) int {
	n := int(math.Ceil(tAlpha * tAlpha * (1 - p) / (p * delta * delta)))
	if n < initialN {
		n = initialN
	}
	if n > maxN {
		n = maxN
	}
	return n
}

func FindRequestCount(cfg Config) ([]Iteration, SummaryTable) {
	var history []Iteration
	var table SummaryTable
	var prev float64
	n := initialN

	for i := 0; i < maxIterations; i++ {
		if i > 0 && n == maxN && history[i-1].N == maxN {
			break
		}

		run := cfg
		run.Seed = cfg.Seed + int64(i)
		table = NewSimulator(run).RunAuto(n)
		p := table.TotalRefusalProbability()
		history = append(history, Iteration{N: n, P: p})

		if i > 0 && prev > 0 && math.Abs(p-prev)/prev < epsilon {
			break
		}
		prev = p

		if p == 0 {
			n *= 2
			if n > maxN {
				n = maxN
			}
		} else {
			n = requiredRequests(p)
		}
	}

	return history, table
}

func (s *Simulator) GetState() SystemState {
	sources := make([]SourceState, len(s.sources))
	for i, src := range s.sources {
		id := src.ID()
		sources[i] = SourceState{
			ID:                 id,
			Generated:          s.statistics.Generated(id),
			Refused:            s.statistics.Refused(id),
			Served:             s.statistics.Served(id),
			RefusalProbability: s.statistics.RefusalProbability(id),
		}
	}

	devs := s.devicePool.Devices()
	devices := make([]DeviceState, len(devs))
	for j, d := range devs {
		devices[j] = DeviceState{
			ID:          d.ID(),
			Busy:        !d.IsFree(),
			Current:     d.Current(),
			ReleaseTime: d.ReleaseTime(),
		}
	}

	return SystemState{
		Time:           s.currentTime,
		Sources:        sources,
		Calendar:       s.calendar.Snapshot(),
		Buffer:         s.buffer.Snapshot(),
		BufferCapacity: s.buffer.capacity,
		Devices:        devices,
		Pointer:        s.devicePool.Pointer(),
	}
}

func (s *Simulator) handleGeneration(e *Event) {
	src := e.Source()
	now := e.Time()

	r := src.Generate(now)
	s.calendar.Schedule(NewGenerationEvent(src.NextArrivalTime(now), src))
	s.statistics.RecordGenerated(r)

	s.placementDispatcher.Place(r, now)
}

func (s *Simulator) handleRelease(e *Event) {
	dev := e.Device()
	now := e.Time()

	r := dev.FinishService(now)
	s.statistics.RecordServed(r, now)
	s.statistics.RecordBusyTime(dev, r.serviceEndTime-r.serviceStartTime)

	s.selectionDispatcher.OnDeviceFreed(now)
}
