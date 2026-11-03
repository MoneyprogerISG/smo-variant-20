package smo

type PlacementDispatcher struct {
	buffer     *Buffer
	devicePool *DevicePool
	calendar   *EventCalendar
	statistics *Statistics
}

func NewPlacementDispatcher(b *Buffer, p *DevicePool, c *EventCalendar, s *Statistics) *PlacementDispatcher {
	if b == nil {
		panic("placement: buffer must not be nil")
	}
	if p == nil {
		panic("placement: device pool must not be nil")
	}
	if c == nil {
		panic("placement: calendar must not be nil")
	}
	if s == nil {
		panic("placement: statistics must not be nil")
	}
	return &PlacementDispatcher{
		buffer:     b,
		devicePool: p,
		calendar:   c,
		statistics: s,
	}
}

func (d *PlacementDispatcher) Place(r *Request, now float64) {
	if dev := d.devicePool.FindFreeDevice(); dev != nil {
		d.sendToDevice(dev, r, now)
		return
	}
	if !d.buffer.IsFull() {
		d.putToBuffer(r, now)
		return
	}
	d.knockOutAndPut(r, now)
}

func (d *PlacementDispatcher) sendToDevice(dev *Device, r *Request, now float64) {
	releaseTime := dev.StartService(r, now)
	d.calendar.Schedule(NewReleaseEvent(releaseTime, dev))
	d.statistics.RecordWait(r, 0)
}

func (d *PlacementDispatcher) putToBuffer(r *Request, now float64) {
	d.buffer.Add(r, now)
}

func (d *PlacementDispatcher) knockOutAndPut(r *Request, now float64) {
	victim := d.buffer.RemoveLast()
	d.statistics.RecordRefused(victim, now)
	d.buffer.Add(r, now)
}
