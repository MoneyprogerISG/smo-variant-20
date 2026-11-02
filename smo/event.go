package smo

type EventType int

type Event struct {
	time   float64
	typ    EventType
	source *Source
	device *Device
}

const (
	GENERATION EventType = iota
	RELEASE
)

func (t EventType) String() string {
	switch t {
	case GENERATION:
		return "GENERATION"
	case RELEASE:
		return "RELEASE"
	default:
		return "UNKNOWN"
	}
}

func (e *Event) Time() float64 {
	return e.time
}

func (e *Event) Type() EventType {
	return e.typ
}

func (e *Event) Source() *Source {
	return e.source
}

func (e *Event) Device() *Device {
	return e.device
}

func NewGenerationEvent(time float64, src *Source) *Event {
	return &Event{
		time:   time,
		typ:    GENERATION,
		source: src,
	}
}

func NewReleaseEvent(time float64, dev *Device) *Event {
	return &Event{
		time:   time,
		typ:    RELEASE,
		device: dev,
	}
}
