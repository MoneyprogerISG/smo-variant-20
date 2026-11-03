package smo

import "math/rand"

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

func (s *Simulator) Step() *Event {
	e := s.calendar.Next()
	if e == nil {
		return nil
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
